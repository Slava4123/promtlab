// Package adminauth — unit-tests для Service.Verify, Enroll, Disable.
//
// CR-9 в REVIEW_2026-05-07.md: usecases/adminauth/ имел 0 unit-тестов
// (тестировался только storage в totp_repo_test.go). Один регресс в
// Verify = bypass 2FA на админке (полный доступ к freeze/reset_password/
// grant_badge). Этот файл закрывает базовый набор, документированный в
// REVIEW: TOTP success, backup-code one-shot, invalid code, Enroll
// non-admin, Enroll over confirmed (refused), CR-14 rate limiter.
package adminauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- Mocks ---

type mockTOTPRepo struct {
	enrollment *models.UserTOTP
	backups    []models.UserTOTPBackupCode

	upsertCalls    int
	confirmedCalls int
	deleteCalls    int
	replaceCalls   int
	markUsedCalls  int

	upsertErr error
	getErr    error
}

func (m *mockTOTPRepo) UpsertEnrollment(_ context.Context, userID uint, secret string) error {
	m.upsertCalls++
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.enrollment = &models.UserTOTP{UserID: userID, Secret: secret}
	return nil
}

func (m *mockTOTPRepo) GetByUserID(_ context.Context, _ uint) (*models.UserTOTP, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.enrollment == nil {
		return nil, repo.ErrNotFound
	}
	return m.enrollment, nil
}

func (m *mockTOTPRepo) MarkConfirmed(_ context.Context, _ uint) error {
	m.confirmedCalls++
	if m.enrollment != nil {
		now := time.Now()
		m.enrollment.ConfirmedAt = &now
	}
	return nil
}

func (m *mockTOTPRepo) Delete(_ context.Context, _ uint) error {
	m.deleteCalls++
	m.enrollment = nil
	m.backups = nil
	return nil
}

func (m *mockTOTPRepo) ReplaceBackupCodes(_ context.Context, userID uint, hashes []string) error {
	m.replaceCalls++
	m.backups = make([]models.UserTOTPBackupCode, 0, len(hashes))
	for i, h := range hashes {
		m.backups = append(m.backups, models.UserTOTPBackupCode{
			ID:       uint(i + 1),
			UserID:   userID,
			CodeHash: h,
		})
	}
	return nil
}

func (m *mockTOTPRepo) ListActiveBackupCodes(_ context.Context, _ uint) ([]models.UserTOTPBackupCode, error) {
	out := make([]models.UserTOTPBackupCode, 0, len(m.backups))
	for _, c := range m.backups {
		if c.UsedAt == nil {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *mockTOTPRepo) MarkBackupCodeUsed(_ context.Context, codeID uint) error {
	m.markUsedCalls++
	for i := range m.backups {
		if m.backups[i].ID == codeID {
			now := time.Now()
			m.backups[i].UsedAt = &now
			return nil
		}
	}
	return repo.ErrNotFound
}

// minimal stub for UserRepository — нужен только для Enroll (берёт email).
type stubAdminUserRepo struct {
	user *models.User
}

func (s *stubAdminUserRepo) GetByID(_ context.Context, _ uint) (*models.User, error) {
	if s.user == nil {
		return nil, repo.ErrNotFound
	}
	return s.user, nil
}

// все остальные методы — panic, не используются в этих тестах.
func (s *stubAdminUserRepo) Create(context.Context, *models.User) error { panic("unused") }
func (s *stubAdminUserRepo) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) Update(context.Context, *models.User) error  { panic("unused") }
func (s *stubAdminUserRepo) SetPlan(context.Context, uint, string) error { panic("unused") }
func (s *stubAdminUserRepo) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	panic("unused")
}
func (s *stubAdminUserRepo) TouchLastLogin(context.Context, uint) error { panic("unused") }
func (s *stubAdminUserRepo) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) MarkReengagementSent(context.Context, uint) error { panic("unused") }
func (s *stubAdminUserRepo) CountReferredBy(context.Context, string) (int64, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) GetByReferralCode(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) MarkReferralRewarded(context.Context, uint) (bool, error) {
	panic("unused")
}
func (s *stubAdminUserRepo) ListMaxUsers(context.Context) ([]uint, error) { panic("unused") }
func (s *stubAdminUserRepo) SetInsightEmailsEnabled(context.Context, uint, bool) error {
	panic("unused")
}

// --- Helpers ---

func newServiceForTests(t *testing.T) (*Service, *mockTOTPRepo, *stubAdminUserRepo) {
	t.Helper()
	totps := &mockTOTPRepo{}
	users := &stubAdminUserRepo{
		user: &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin},
	}
	return NewService(totps, users), totps, users
}

// enrollAndConfirm выполняет полный setup: создаёт enrollment, генерирует
// backup codes, помечает confirmed. Возвращает service + raw backup-codes
// для тестов backup-code path.
func enrollAndConfirm(t *testing.T, svc *Service, totps *mockTOTPRepo, userID uint) []string {
	t.Helper()
	ctx := context.Background()

	res, err := svc.Enroll(ctx, userID)
	if err != nil {
		t.Fatalf("enroll failed: %v", err)
	}

	// Manually MarkConfirmed (имитируем что юзер прошёл verify своим первым кодом).
	if err := totps.MarkConfirmed(ctx, userID); err != nil {
		t.Fatalf("MarkConfirmed: %v", err)
	}
	return res.BackupCodes
}

// --- Tests: Enroll ---

func TestEnroll_NonAdmin_Refused(t *testing.T) {
	totps := &mockTOTPRepo{}
	users := &stubAdminUserRepo{
		user: &models.User{ID: 2, Email: "user@example.com", Role: models.RoleUser},
	}
	svc := NewService(totps, users)

	_, err := svc.Enroll(context.Background(), 2)
	if !errors.Is(err, ErrNotAdmin) {
		t.Fatalf("expected ErrNotAdmin, got %v", err)
	}
	if totps.upsertCalls != 0 {
		t.Fatalf("expected no upsert for non-admin, got %d", totps.upsertCalls)
	}
}

func TestEnroll_OverConfirmedEnrollment_Refused(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	enrollAndConfirm(t, svc, totps, 1)

	// Повторный Enroll должен отказать (security hazard — потеря backup codes).
	_, err := svc.Enroll(context.Background(), 1)
	if !errors.Is(err, ErrTOTPAlreadyConfirmed) {
		t.Fatalf("expected ErrTOTPAlreadyConfirmed, got %v", err)
	}
}

func TestEnroll_OverUnconfirmedEnrollment_OK(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	if _, err := svc.Enroll(context.Background(), 1); err != nil {
		t.Fatalf("first enroll: %v", err)
	}
	// totps.enrollment.ConfirmedAt == nil → можно перезаписать.
	if _, err := svc.Enroll(context.Background(), 1); err != nil {
		t.Fatalf("re-enroll over unconfirmed should succeed, got %v", err)
	}
	if totps.upsertCalls != 2 {
		t.Fatalf("expected 2 upsert calls, got %d", totps.upsertCalls)
	}
}

// --- Tests: Verify TOTP ---

func TestVerify_TOTPCode_Success(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	enrollAndConfirm(t, svc, totps, 1)

	code, err := totp.GenerateCode(totps.enrollment.Secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}

	res, err := svc.Verify(context.Background(), 1, code)
	if err != nil {
		t.Fatalf("expected OK, got %v", err)
	}
	if res.UsedBackupCode {
		t.Fatalf("expected UsedBackupCode=false")
	}
}

func TestVerify_NotEnrolled_ReturnsErrTOTPNotEnrolled(t *testing.T) {
	svc, _, _ := newServiceForTests(t)
	_, err := svc.Verify(context.Background(), 1, "000000")
	if !errors.Is(err, ErrTOTPNotEnrolled) {
		t.Fatalf("expected ErrTOTPNotEnrolled, got %v", err)
	}
}

func TestVerify_InvalidCode_NoBackupConsumed(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	codes := enrollAndConfirm(t, svc, totps, 1)
	_ = codes // не используем — пробуем заведомо неверный

	_, err := svc.Verify(context.Background(), 1, "wrong-code")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
	if totps.markUsedCalls != 0 {
		t.Fatalf("expected no MarkBackupCodeUsed calls on invalid input, got %d", totps.markUsedCalls)
	}
}

// --- Tests: Verify backup-code ---

func TestVerify_BackupCode_OneShot(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	codes := enrollAndConfirm(t, svc, totps, 1)
	if len(codes) == 0 {
		t.Fatalf("no backup codes generated")
	}
	bk := codes[0]

	// Первый раз — успех.
	res, err := svc.Verify(context.Background(), 1, bk)
	if err != nil {
		t.Fatalf("first backup verify: %v", err)
	}
	if !res.UsedBackupCode {
		t.Fatalf("expected UsedBackupCode=true")
	}
	if totps.markUsedCalls != 1 {
		t.Fatalf("expected 1 MarkUsed call, got %d", totps.markUsedCalls)
	}

	// Второй раз тот же код — должен fail (one-shot).
	_, err = svc.Verify(context.Background(), 1, bk)
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode on reuse, got %v", err)
	}
}

func TestVerify_BackupCodeBcryptHashed(t *testing.T) {
	// Проверка что коды хранятся как bcrypt-хеши (не plaintext).
	svc, totps, _ := newServiceForTests(t)
	codes := enrollAndConfirm(t, svc, totps, 1)
	for _, c := range totps.backups {
		// Plaintext не должен матчиться напрямую (string equal).
		for _, raw := range codes {
			if c.CodeHash == raw {
				t.Fatalf("backup code stored as plaintext (security regression)")
			}
		}
		// Зато bcrypt должен матчиться с одним из raw.
		matched := false
		for _, raw := range codes {
			if bcrypt.CompareHashAndPassword([]byte(c.CodeHash), []byte(raw)) == nil {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("hash %s does not match any raw code", c.CodeHash)
		}
	}
}

// --- Tests: CR-14 rate limiter ---

func TestVerify_RateLimited_AfterBurst(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	enrollAndConfirm(t, svc, totps, 1)

	// Первые 5 попыток (burst=totpVerifyBurst) проходят (хоть и InvalidCode).
	for i := 0; i < totpVerifyBurst; i++ {
		_, err := svc.Verify(context.Background(), 1, "000000")
		if errors.Is(err, ErrTOTPRateLimited) {
			t.Fatalf("attempt %d unexpectedly rate-limited", i+1)
		}
	}

	// 6-я попытка должна быть rate-limited.
	_, err := svc.Verify(context.Background(), 1, "000000")
	if !errors.Is(err, ErrTOTPRateLimited) {
		t.Fatalf("expected ErrTOTPRateLimited on 6th attempt, got %v", err)
	}
}

func TestVerify_RateLimiter_PerUser(t *testing.T) {
	// Лимиты должны быть per-userID — два разных юзера не делят bucket.
	svc, totps, _ := newServiceForTests(t)
	totps.enrollment = &models.UserTOTP{UserID: 1, Secret: "JBSWY3DPEHPK3PXP"}
	now := time.Now()
	totps.enrollment.ConfirmedAt = &now

	// User 1: исчерпываем лимит.
	for i := 0; i < totpVerifyBurst; i++ {
		_, _ = svc.Verify(context.Background(), 1, "000000")
	}
	if _, err := svc.Verify(context.Background(), 1, "000000"); !errors.Is(err, ErrTOTPRateLimited) {
		t.Fatalf("user 1 must be rate-limited at this point, got %v", err)
	}

	// User 2: первая попытка не должна быть rate-limited (хоть mock возвращает
	// ErrTOTPNotEnrolled — это НЕ rate-limit error, что и проверяем).
	_, err := svc.Verify(context.Background(), 2, "000000")
	if errors.Is(err, ErrTOTPRateLimited) {
		t.Fatalf("user 2 must not be rate-limited (independent bucket)")
	}
}

// --- Tests: Disable + RegenerateBackupCodes ---

func TestDisable_RemovesEnrollment(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	enrollAndConfirm(t, svc, totps, 1)
	if totps.enrollment == nil {
		t.Fatalf("setup: expected enrollment")
	}

	if err := svc.Disable(context.Background(), 1); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if totps.enrollment != nil {
		t.Fatalf("expected enrollment cleared after Disable")
	}
}

func TestRegenerateBackupCodes_RequiresConfirmedEnrollment(t *testing.T) {
	svc, _, _ := newServiceForTests(t)

	// Без enrollment — отказ.
	_, err := svc.RegenerateBackupCodes(context.Background(), 1)
	if !errors.Is(err, ErrTOTPNotEnrolled) {
		t.Fatalf("expected ErrTOTPNotEnrolled, got %v", err)
	}
}

func TestRegenerateBackupCodes_ProducesFreshCodes(t *testing.T) {
	svc, totps, _ := newServiceForTests(t)
	oldCodes := enrollAndConfirm(t, svc, totps, 1)

	newCodes, err := svc.RegenerateBackupCodes(context.Background(), 1)
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if len(newCodes) != BackupCodeCount {
		t.Fatalf("expected %d new codes, got %d", BackupCodeCount, len(newCodes))
	}

	// Старые коды должны больше НЕ работать (replaced).
	_, err = svc.Verify(context.Background(), 1, oldCodes[0])
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("old backup code должен fail после regenerate, got %v", err)
	}
}
