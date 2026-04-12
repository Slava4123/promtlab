package adminauth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// BackupCodeCount — количество одноразовых recovery-кодов, генерируемых при
// enroll и regenerate. 10 — стандарт (GitHub, Google, 1Password).
const BackupCodeCount = 10

// Issuer — отображаемое имя приложения в Google Authenticator / 1Password /
// Authy. Видно юзеру в списке TOTP secrets как "PromptVault: admin@example.com".
const Issuer = "PromptVault"

// Service — TOTP enrollment/verification + backup codes.
// Зависит от TOTPRepository (storage) и UserRepository (для получения email,
// который используется как account name в QR-URL).
type Service struct {
	totps repo.TOTPRepository
	users repo.UserRepository
}

func NewService(totps repo.TOTPRepository, users repo.UserRepository) *Service {
	return &Service{totps: totps, users: users}
}

// Enroll генерирует новый TOTP secret + backup codes для юзера.
// Требует role='admin'. Для unconfirmed enrollment допускает overwrite.
// Для confirmed требует сначала явный Disable (иначе легко случайно сбросить TOTP
// у работающего admin — security hazard).
func (s *Service) Enroll(ctx context.Context, userID uint) (*EnrollResult, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u.Role != models.RoleAdmin {
		return nil, ErrNotAdmin
	}

	// Проверить — не confirmed ли уже enrollment (тогда overwrite запрещён).
	existing, err := s.totps.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return nil, err
	}
	if existing != nil && existing.IsConfirmed() {
		return nil, ErrTOTPAlreadyConfirmed
	}

	// Генерация TOTP secret через pquerna/otp.
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
		AccountName: u.Email,
	})
	if err != nil {
		slog.Error("adminauth.enroll.generate_failed", "user_id", userID, "error", err)
		return nil, fmt.Errorf("%w: %v", ErrGenerateFailed, err)
	}

	// Upsert (unconfirmed по умолчанию).
	if err := s.totps.UpsertEnrollment(ctx, userID, key.Secret()); err != nil {
		return nil, err
	}

	// Генерация + hash backup codes, сохранение через ReplaceBackupCodes.
	rawCodes, hashes, err := generateBackupCodes(BackupCodeCount)
	if err != nil {
		return nil, err
	}
	if err := s.totps.ReplaceBackupCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}

	return &EnrollResult{
		Secret:      key.Secret(),
		QRURL:       key.URL(),
		BackupCodes: rawCodes,
	}, nil
}

// ConfirmEnrollment завершает enrollment проверкой первого TOTP кода.
// После этого Authenticator генерирует валидные коды, и юзер начинает
// использовать TOTP при login. Возвращает nil если код валиден.
func (s *Service) ConfirmEnrollment(ctx context.Context, userID uint, code string) error {
	t, err := s.totps.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrTOTPNotEnrolled
		}
		return err
	}
	if t.IsConfirmed() {
		return ErrTOTPAlreadyConfirmed
	}
	if !totp.Validate(code, t.Secret) {
		return ErrInvalidCode
	}
	return s.totps.MarkConfirmed(ctx, userID)
}

// Verify проверяет введённый код на валидность. Код может быть TOTP (6 цифр
// из Authenticator) или backup code (формат xxxxx-xxxxx). Порядок проверки:
// сначала TOTP (дешёво), потом backup (итерация по bcrypt-хешам).
// Возвращает VerifyResult с флагом UsedBackupCode и count оставшихся backup кодов.
// При неверном коде — ErrInvalidCode.
func (s *Service) Verify(ctx context.Context, userID uint, code string) (*VerifyResult, error) {
	t, err := s.totps.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrTOTPNotEnrolled
		}
		return nil, err
	}
	if !t.IsConfirmed() {
		return nil, ErrTOTPNotEnrolled
	}

	// TOTP check (cheap, time-based).
	if totp.Validate(code, t.Secret) {
		active, err := s.totps.ListActiveBackupCodes(ctx, userID)
		if err != nil {
			slog.Warn("adminauth.verify.count_backup_failed", "user_id", userID, "error", err)
		}
		return &VerifyResult{UsedBackupCode: false, RemainingBackupCodes: len(active)}, nil
	}

	// Backup code check — linear scan bcrypt. Дорогое, но только при TOTP-miss.
	active, err := s.totps.ListActiveBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, c := range active {
		if bcrypt.CompareHashAndPassword([]byte(c.CodeHash), []byte(code)) == nil {
			if err := s.totps.MarkBackupCodeUsed(ctx, c.ID); err != nil {
				slog.Warn("adminauth.verify.mark_used_failed", "user_id", userID, "code_id", c.ID, "error", err)
			}
			return &VerifyResult{
				UsedBackupCode:       true,
				RemainingBackupCodes: len(active) - 1, // только что использовали один
			}, nil
		}
	}

	return nil, ErrInvalidCode
}

// Disable удаляет TOTP + все backup codes (полный reset). Опасная операция —
// требует admin confirmation на слое выше (UI dialog, admin audit).
// После Disable юзер теряет 2FA и сможет logиниться только по password.
func (s *Service) Disable(ctx context.Context, userID uint) error {
	return s.totps.Delete(ctx, userID)
}

// RegenerateBackupCodes заменяет все backup codes новыми.
// Возвращает 10 новых plaintext-кодов (показываются юзеру один раз).
// Требует чтобы TOTP уже был confirmed — иначе regen бессмысленен.
func (s *Service) RegenerateBackupCodes(ctx context.Context, userID uint) ([]string, error) {
	t, err := s.totps.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrTOTPNotEnrolled
		}
		return nil, err
	}
	if !t.IsConfirmed() {
		return nil, ErrTOTPNotEnrolled
	}

	rawCodes, hashes, err := generateBackupCodes(BackupCodeCount)
	if err != nil {
		return nil, err
	}
	if err := s.totps.ReplaceBackupCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}
	return rawCodes, nil
}

// IsConfirmed — helper для middleware/login flow: проверить закончил ли юзер
// enrollment. Если false, login TOTP step должен вести на /admin/totp для enroll.
func (s *Service) IsConfirmed(ctx context.Context, userID uint) (bool, error) {
	t, err := s.totps.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return t.IsConfirmed(), nil
}

// generateBackupCodes создаёт N случайных кодов в формате "xxxxx-xxxxx"
// (10 alphanumeric chars, 50 bits of entropy — достаточно для one-time use).
// Возвращает два слайса одинаковой длины: raw plaintext codes (для отображения)
// и bcrypt hashes (для сохранения в БД).
func generateBackupCodes(n int) ([]string, []string, error) {
	raws := make([]string, 0, n)
	hashes := make([]string, 0, n)

	// Алфавит Crockford base32 без I/L/O/U для человеческой читаемости
	// (в типичных true base32 встречаются 0/O и 1/I/L, которые путаются).
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTVWXYZ"

	for range n {
		var buf [10]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return nil, nil, err
		}
		code := make([]byte, 11) // 5 + 1 (dash) + 5
		for i := range 5 {
			code[i] = alphabet[int(buf[i])%len(alphabet)]
		}
		code[5] = '-'
		for i := range 5 {
			code[6+i] = alphabet[int(buf[5+i])%len(alphabet)]
		}

		raws = append(raws, string(code))
		hash, err := bcrypt.GenerateFromPassword(code, bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, string(hash))
	}
	return raws, hashes, nil
}

