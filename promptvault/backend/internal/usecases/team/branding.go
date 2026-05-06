package team

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"promptvault/internal/models"
	"promptvault/internal/usecases/subscription"
)

// BrandingInput — входные параметры для SetBranding.
// Все поля optional; пустая строка очищает предыдущее значение.
//
// Phase 16-X: LogoSource — пустая строка = «не трогать» (backward compat для
// клиентов, которые не знают про источник). Конкретное значение ('url'|'file'|
// 'none') пишется в teams.brand_logo_source отдельным UPDATE.
type BrandingInput struct {
	LogoURL      string
	LogoSource   string
	Tagline      string
	Website      string
	PrimaryColor string
}

var (
	// ErrBrandingMaxOnly — Max gate.
	ErrBrandingMaxOnly = errors.New("team/branding: фича доступна только на тарифе Max")
	// ErrBrandingInvalidURL — logo/website не https или слишком длинный.
	ErrBrandingInvalidURL = errors.New("team/branding: URL должен начинаться с https:// и быть не длиннее 500 символов")
	// ErrBrandingInvalidColor — primary_color не hex #RRGGBB.
	ErrBrandingInvalidColor = errors.New("team/branding: цвет должен быть в формате #RRGGBB")
	// ErrBrandingInvalidTagline — tagline длиннее 200 символов.
	ErrBrandingInvalidTagline = errors.New("team/branding: tagline не длиннее 200 символов")
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// SetBranding обновляет brand-поля команды. Только owner и только на Max.
// Валидация: URLs — HTTPS-only, primary_color — hex, tagline ≤200 символов.
func (s *Service) SetBranding(ctx context.Context, slug string, userID uint, input BrandingInput) error {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return err
	}

	// Max gate — проверяем план owner'а команды (created_by).
	owner, err := s.users.GetByID(ctx, team.CreatedBy)
	if err != nil {
		return err
	}
	if !subscription.IsMax(owner.PlanID) {
		return ErrBrandingMaxOnly
	}

	// Валидация.
	if err := validateBrandingURL(input.LogoURL); err != nil {
		return err
	}
	if err := validateBrandingURL(input.Website); err != nil {
		return err
	}
	if len(input.Tagline) > 200 {
		return ErrBrandingInvalidTagline
	}
	if input.PrimaryColor != "" && !hexColorRegex.MatchString(input.PrimaryColor) {
		return ErrBrandingInvalidColor
	}

	if err := s.teams.UpdateBranding(ctx, team.ID, input.LogoURL, input.Tagline, input.Website, input.PrimaryColor); err != nil {
		return err
	}
	if input.LogoSource != "" {
		return s.teams.UpdateBrandLogoSource(ctx, team.ID, input.LogoSource)
	}
	return nil
}

func validateBrandingURL(url string) error {
	if url == "" {
		return nil // пустой — очистка, OK
	}
	if len(url) > 500 {
		return ErrBrandingInvalidURL
	}
	if !strings.HasPrefix(url, "https://") {
		return ErrBrandingInvalidURL
	}
	return nil
}

// GetBranding возвращает BrandingInfo team по slug. Доступен всем членам
// команды (для settings page). Пустой BrandingInfo если не настроено.
func (s *Service) GetBranding(ctx context.Context, slug string, userID uint) (*models.BrandingInfo, error) {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleViewer)
	if err != nil {
		return nil, err
	}
	return buildBrandingInfo(team), nil
}

// GetBrandingForShare — для share usecase (public /s/:token).
// Возвращает BrandingInfo только если owner команды на тарифе Max;
// для других — nil (скрыто в public response).
// Не проверяет membership — это unauthenticated endpoint.
func (s *Service) GetBrandingForShare(ctx context.Context, teamID uint) (*models.BrandingInfo, error) {
	team, err := s.teams.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	owner, err := s.users.GetByID(ctx, team.CreatedBy)
	if err != nil {
		return nil, err
	}
	if !subscription.IsMax(owner.PlanID) {
		return nil, nil
	}
	info := buildBrandingInfo(team)
	if info.IsEmpty() {
		return nil, nil
	}
	return info, nil
}

// buildBrandingInfo — единая точка резолвинга effective_logo_url.
// Phase 16-X: фронт не должен выбирать между LogoURL и uploaded-file —
// бэк отдаёт уже готовую ссылку в EffectiveLogoURL. LogoSource дублируется
// для UI (показ режима «URL» / «файл» в форме).
//
// Контракт effective_logo_url:
//   - source='url'  → BrandLogoURL (может быть пустой строкой = нет логотипа)
//   - source='file' → relative path к нашему GET endpoint; кэш-bust через ETag
//   - source='none' → пустая строка
//
// LogoSource по умолчанию = "url" (DB default из миграции 000060), даже если
// сам BrandLogoURL пуст — это эквивалентно "нет логотипа" в URL-режиме.
func buildBrandingInfo(team *models.Team) *models.BrandingInfo {
	info := &models.BrandingInfo{
		LogoURL:      team.BrandLogoURL,
		LogoSource:   team.BrandLogoSource,
		Tagline:      team.BrandTagline,
		Website:      team.BrandWebsite,
		PrimaryColor: team.BrandPrimaryColor,
	}
	if info.LogoSource == "" {
		info.LogoSource = "url"
	}
	switch info.LogoSource {
	case "file":
		info.EffectiveLogoURL = "/api/teams/" + team.Slug + "/branding/logo"
	case "url":
		info.EffectiveLogoURL = team.BrandLogoURL
	}
	return info
}
