package team

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"promptvault/internal/models"
)

// setupBrandingMocks подсовывает в моки стандартные ответы:
// team со slug="acme", owner=uid=42; requester=uid=42 (роль Owner);
// план владельца — ownerPlan. UpdateBranding разрешён с любыми параметрами
// (happy-path тесты на этом заканчиваются; error-path просто не доходят).
// Возвращает готовый svc для SetBranding/GetBrandingForShare.
func setupBrandingMocks(ownerPlan string) (*Service, *mockTeamRepo, *mockUserRepo) {
	svc, tr, ur := newTestService()
	team := &models.Team{ID: 10, Slug: "acme", Name: "ACME", CreatedBy: 42}
	tr.On("GetBySlug", context.Background(), "acme").Return(team, nil)
	tr.On("GetMember", context.Background(), uint(10), uint(42)).
		Return(&models.TeamMember{TeamID: 10, UserID: 42, Role: models.RoleOwner}, nil)
	tr.On("UpdateBranding",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(nil).Maybe()
	ur.On("GetByID", context.Background(), uint(42)).
		Return(&models.User{ID: 42, PlanID: ownerPlan}, nil)
	return svc, tr, ur
}

// TestSetBranding_MaxGateBlocksFree — SetBranding на free-тарифе → ErrBrandingMaxOnly.
func TestSetBranding_MaxGateBlocksFree(t *testing.T) {
	svc, _, _ := setupBrandingMocks("free")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		LogoURL:      "https://cdn.example/logo.png",
		PrimaryColor: "#ff0066",
	})
	assert.ErrorIs(t, err, ErrBrandingMaxOnly)
}

// TestSetBranding_MaxGateBlocksPro — Pro не Max, тоже blocked.
func TestSetBranding_MaxGateBlocksPro(t *testing.T) {
	svc, _, _ := setupBrandingMocks("pro")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		LogoURL: "https://cdn.example/logo.png",
	})
	assert.ErrorIs(t, err, ErrBrandingMaxOnly)
}

// TestSetBranding_MaxGateBlocksPrefixLookalike — "maximum" НЕ должен проходить
// gate (H3 защита от strings.HasPrefix).
func TestSetBranding_MaxGateBlocksPrefixLookalike(t *testing.T) {
	svc, _, _ := setupBrandingMocks("maximum")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		LogoURL: "https://cdn.example/logo.png",
	})
	assert.ErrorIs(t, err, ErrBrandingMaxOnly)
}

// TestSetBranding_InvalidHexColor — primary_color не hex → ErrBrandingInvalidColor.
func TestSetBranding_InvalidHexColor(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		PrimaryColor: "red",
	})
	assert.ErrorIs(t, err, ErrBrandingInvalidColor)
}

// TestSetBranding_RejectsHTTPLogo — http:// недопустим, только https://.
func TestSetBranding_RejectsHTTPLogo(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		LogoURL: "http://cdn.example/logo.png",
	})
	assert.ErrorIs(t, err, ErrBrandingInvalidURL)
}

// TestSetBranding_RejectsOverlongURL — длиннее 500 символов → ErrBrandingInvalidURL.
func TestSetBranding_RejectsOverlongURL(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max")
	long := "https://cdn.example/" + strings.Repeat("x", 600)
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		LogoURL: long,
	})
	assert.ErrorIs(t, err, ErrBrandingInvalidURL)
}

// TestSetBranding_RejectsLongTagline — > 200 символов → ErrBrandingInvalidTagline.
func TestSetBranding_RejectsLongTagline(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		Tagline: strings.Repeat("x", 201),
	})
	assert.ErrorIs(t, err, ErrBrandingInvalidTagline)
}

// TestSetBranding_HappyPathMax — валидный input, max-owner → nil error.
func TestSetBranding_HappyPathMax(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		LogoURL:      "https://cdn.example/logo.png",
		Tagline:      "ACME Prompts",
		Website:      "https://acme.example",
		PrimaryColor: "#ff0066",
	})
	assert.NoError(t, err)
}

// TestSetBranding_HappyPathMaxYearly — max_yearly тоже должен пройти gate (H3 whitelist).
func TestSetBranding_HappyPathMaxYearly(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max_yearly")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		PrimaryColor: "#00aaff",
	})
	assert.NoError(t, err)
}

// TestSetBranding_AllowsEmptyClear — пустые значения очищают branding, это разрешено.
func TestSetBranding_AllowsEmptyClear(t *testing.T) {
	svc, _, _ := setupBrandingMocks("max")
	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{})
	assert.NoError(t, err)
}

// TestSetBranding_ForbidsNonOwner — editor/viewer не может менять branding.
func TestSetBranding_ForbidsEditor(t *testing.T) {
	svc, tr, _ := newTestService()
	team := &models.Team{ID: 10, Slug: "acme", CreatedBy: 100}
	tr.On("GetBySlug", context.Background(), "acme").Return(team, nil)
	tr.On("GetMember", context.Background(), uint(10), uint(42)).
		Return(&models.TeamMember{TeamID: 10, UserID: 42, Role: models.RoleEditor}, nil)

	err := svc.SetBranding(context.Background(), "acme", 42, BrandingInput{
		PrimaryColor: "#ff0066",
	})
	assert.Error(t, err)
	// Проверяем что это именно "недостаточно прав" (ErrNotOwner из checkAccess),
	// а не Max-gate ошибка — до плановой проверки вообще не должны дойти.
	assert.NotErrorIs(t, err, ErrBrandingMaxOnly)
}

// TestGetBrandingForShare_HidesForFreeOwner — владелец free → nil branding (не Max).
func TestGetBrandingForShare_HidesForFreeOwner(t *testing.T) {
	svc, tr, ur := newTestService()
	tr.On("GetByID", context.Background(), uint(10)).
		Return(&models.Team{ID: 10, CreatedBy: 42, BrandLogoURL: "https://x.example/a.png"}, nil)
	ur.On("GetByID", context.Background(), uint(42)).
		Return(&models.User{ID: 42, PlanID: "free"}, nil)

	info, err := svc.GetBrandingForShare(context.Background(), 10)
	assert.NoError(t, err)
	assert.Nil(t, info)
}

// TestGetBrandingForShare_ReturnsForMax — владелец max + непустое branding → info.
func TestGetBrandingForShare_ReturnsForMax(t *testing.T) {
	svc, tr, ur := newTestService()
	tr.On("GetByID", context.Background(), uint(10)).
		Return(&models.Team{
			ID: 10, CreatedBy: 42,
			BrandLogoURL: "https://x.example/a.png",
			BrandTagline: "Prompts for all",
		}, nil)
	ur.On("GetByID", context.Background(), uint(42)).
		Return(&models.User{ID: 42, PlanID: "max"}, nil)

	info, err := svc.GetBrandingForShare(context.Background(), 10)
	assert.NoError(t, err)
	if assert.NotNil(t, info) {
		assert.Equal(t, "https://x.example/a.png", info.LogoURL)
		assert.Equal(t, "Prompts for all", info.Tagline)
	}
}

// TestGetBrandingForShare_TeamNotFound — repo-error → ошибка пробрасывается.
func TestGetBrandingForShare_TeamNotFound(t *testing.T) {
	svc, tr, _ := newTestService()
	tr.On("GetByID", context.Background(), uint(10)).Return(nil, errors.New("not found"))

	info, err := svc.GetBrandingForShare(context.Background(), 10)
	assert.Error(t, err)
	assert.Nil(t, info)
}
