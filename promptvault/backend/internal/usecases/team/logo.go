package team

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg" // регистрирует JPEG-декодер для image.DecodeConfig
	_ "image/png"  // регистрирует PNG-декодер для image.DecodeConfig
	"io"
	"net/http"
	"strings"
	"time"

	_ "golang.org/x/image/webp" // регистрирует WebP-декодер

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/usecases/subscription"
)

// Phase 16-X. Загрузка логотипа файлом (bytea storage).

const (
	// MaxLogoFileSize — лимит размера byte-payload (1 МБ). Соответствует CHECK
	// на size_bytes в миграции 000060. Превышение → ErrLogoFileTooLarge до
	// SQL-уровня.
	MaxLogoFileSize = 1 << 20 // 1 048 576

	// MaxLogoImageWidth/Height — отказ от очень больших изображений по pixel-
	// metrics. Декомпрессия PNG может уйти в гигабайты при «zip-bomb»-style
	// файлах меньше 1MiB; ограничение pixel-dim защищает RAM при image.Decode.
	MaxLogoImageWidth  = 1024
	MaxLogoImageHeight = 1024
)

var (
	// ErrLogoStorageDisabled — usecase используется без подключённого репо.
	// В рантайме не должен возникать (app.go всегда подключает); защита от
	// рефакторинга.
	ErrLogoStorageDisabled = errors.New("team/logo: хранилище не подключено")

	// ErrLogoFileMissing — multipart без файла или нулевая длина.
	ErrLogoFileMissing = errors.New("team/logo: файл не передан")

	// ErrLogoFileTooLarge — байтов больше MaxLogoFileSize.
	ErrLogoFileTooLarge = errors.New("team/logo: размер файла превышает 1 МБ")

	// ErrLogoFileBadFormat — content-type вне whitelist ИЛИ image.Decode failed.
	// Покрывает: SVG, txt-как-png, polyglot файлы.
	ErrLogoFileBadFormat = errors.New("team/logo: формат не поддерживается, нужен PNG, JPEG или WebP")

	// ErrLogoImageTooLarge — pixel dimensions выше MaxLogoImageWidth/Height.
	ErrLogoImageTooLarge = errors.New("team/logo: размеры изображения превышают 1024×1024 px")
)

var allowedLogoTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/webp": {},
}

// UploadLogo — owner+Max gate, валидация (size + magic-byte + image.Decode +
// pixel-dim), upsert bytea, переключение source='file'. Возвращает модель
// с заполненным sha256 (используется в ETag отдачи).
func (s *Service) UploadLogo(ctx context.Context, slug string, userID uint, body io.Reader) (*models.TeamLogoFile, error) {
	if s.logos == nil {
		return nil, ErrLogoStorageDisabled
	}

	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return nil, err
	}

	owner, err := s.users.GetByID(ctx, team.CreatedBy)
	if err != nil {
		return nil, err
	}
	if !subscription.IsMax(owner.PlanID) {
		return nil, ErrBrandingMaxOnly
	}

	// LimitReader на MaxLogoFileSize+1 — если body вернул >MaxLogoFileSize
	// байтов, мы это увидим по len(raw) и отвергнем без полного буфера в RAM
	// (после +1 байта truncate'нем). HTTP-уровень параллельно ставит
	// MaxBytesReader для жёсткого ограничения соединения.
	raw, err := io.ReadAll(io.LimitReader(body, MaxLogoFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, ErrLogoFileMissing
	}
	if int64(len(raw)) > MaxLogoFileSize {
		return nil, ErrLogoFileTooLarge
	}

	// Magic-byte detection — std lib читает первые 512 байт и возвращает
	// "image/png", "image/jpeg" или "image/webp" по signature, игнорируя имя.
	contentType := normaliseContentType(http.DetectContentType(raw))
	if _, ok := allowedLogoTypes[contentType]; !ok {
		return nil, ErrLogoFileBadFormat
	}

	// Полный декод config'а (без декомпрессии всех пикселей) даёт width/height +
	// отлавливает polyglot (PNG header + не-PNG body) — DecodeConfig упадёт.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, ErrLogoFileBadFormat
	}
	if cfg.Width > MaxLogoImageWidth || cfg.Height > MaxLogoImageHeight {
		return nil, ErrLogoImageTooLarge
	}

	sum := sha256.Sum256(raw)
	file := &models.TeamLogoFile{
		TeamID:      team.ID,
		ContentType: contentType,
		SizeBytes:   int64(len(raw)),
		SHA256:      hex.EncodeToString(sum[:]),
		Bytes:       raw,
		UploadedAt:  time.Now().UTC(),
	}
	if err := s.logos.Upsert(ctx, file); err != nil {
		return nil, err
	}
	if err := s.teams.UpdateBrandLogoSource(ctx, team.ID, string(models.LogoSourceFile)); err != nil {
		return nil, err
	}
	return file, nil
}

// DeleteLogo — owner+Max gate; идемпотентно убирает строку team_logo_files
// и переключает source='none'. Если файла не было — всё равно ставим 'none'
// (юзер мог переключить с url'а в "ничего"). LogoURL не трогаем — пусть
// останется на случай, если юзер захочет вернуться обратно в URL-режим.
func (s *Service) DeleteLogo(ctx context.Context, slug string, userID uint) error {
	if s.logos == nil {
		return ErrLogoStorageDisabled
	}
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return err
	}
	owner, err := s.users.GetByID(ctx, team.CreatedBy)
	if err != nil {
		return err
	}
	if !subscription.IsMax(owner.PlanID) {
		return ErrBrandingMaxOnly
	}
	if err := s.logos.Delete(ctx, team.ID); err != nil {
		return err
	}
	return s.teams.UpdateBrandLogoSource(ctx, team.ID, string(models.LogoSourceNone))
}

// GetLogo — public-отдача bytes для GET endpoint.
//   - Не проверяет membership (как /api/share/...): анонимы могут видеть лого.
//   - Защита от enumeration: 404 одинаков для «нет команды» и «нет файла».
//   - Скрывает логотип, если owner downgraded с Max (зеркалит политику
//     GetBrandingForShare — публичные брендинговые элементы доступны только
//     пока owner на Max).
func (s *Service) GetLogo(ctx context.Context, slug string) (*models.TeamLogoFile, error) {
	if s.logos == nil {
		return nil, repo.ErrNotFound
	}
	team, err := s.teams.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if team.BrandLogoSource != models.LogoSourceFile {
		return nil, repo.ErrNotFound
	}
	owner, err := s.users.GetByID(ctx, team.CreatedBy)
	if err != nil {
		return nil, err
	}
	if !subscription.IsMax(owner.PlanID) {
		return nil, repo.ErrNotFound
	}
	return s.logos.Get(ctx, team.ID)
}

// normaliseContentType режет hint'ы вида "image/png; charset=utf-8" до базового MIME.
func normaliseContentType(ct string) string {
	if i := strings.Index(ct, ";"); i != -1 {
		ct = ct[:i]
	}
	return strings.TrimSpace(strings.ToLower(ct))
}
