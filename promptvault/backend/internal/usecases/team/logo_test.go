package team

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"promptvault/internal/models"
)

// ===== Mock TeamLogoRepository =====

type mockTeamLogoRepo struct{ mock.Mock }

func (m *mockTeamLogoRepo) Get(ctx context.Context, teamID uint) (*models.TeamLogoFile, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamLogoFile), args.Error(1)
}
func (m *mockTeamLogoRepo) Upsert(ctx context.Context, file *models.TeamLogoFile) error {
	return m.Called(ctx, file).Error(0)
}
func (m *mockTeamLogoRepo) Delete(ctx context.Context, teamID uint) error {
	return m.Called(ctx, teamID).Error(0)
}

// ===== Helpers =====

// makePNG генерирует валидный PNG-байтовый payload N×N px (RGBA, чёрный фон).
// Используется как корректный happy-path payload и как «слишком большой по pixel'ам».
func makePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return buf.Bytes()
}

// setupLogoMocks — общий стенд: команда acme/uid42 owner, план задаётся.
// Возвращает svc с подключённым logos repo.
func setupLogoMocks(t *testing.T, ownerPlan string, requesterRole models.TeamRole) (*Service, *mockTeamLogoRepo) {
	t.Helper()
	svc, tr, ur := newTestService()
	lr := new(mockTeamLogoRepo)
	svc.SetLogoRepo(lr)

	team := &models.Team{ID: 10, Slug: "acme", Name: "ACME", CreatedBy: 42}
	tr.On("GetBySlug", mock.Anything, "acme").Return(team, nil)
	tr.On("GetMember", mock.Anything, uint(10), uint(42)).
		Return(&models.TeamMember{TeamID: 10, UserID: 42, Role: requesterRole}, nil)
	ur.On("GetByID", mock.Anything, uint(42)).
		Return(&models.User{ID: 42, PlanID: ownerPlan}, nil)
	tr.On("UpdateBrandLogoSource", mock.Anything, uint(10), mock.Anything).Return(nil).Maybe()
	return svc, lr
}

// ===== UploadLogo =====

func TestUploadLogo_HappyPathPNG(t *testing.T) {
	svc, lr := setupLogoMocks(t, "max", models.RoleOwner)
	lr.On("Upsert", mock.Anything, mock.MatchedBy(func(f *models.TeamLogoFile) bool {
		return f.TeamID == 10 && f.ContentType == "image/png" && len(f.Bytes) > 0 && f.SHA256 != ""
	})).Return(nil)

	body := bytes.NewReader(makePNG(t, 200, 60))
	got, err := svc.UploadLogo(context.Background(), "acme", 42, body)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "image/png", got.ContentType)
	assert.True(t, got.SizeBytes > 0)
	lr.AssertCalled(t, "Upsert", mock.Anything, mock.Anything)
}

func TestUploadLogo_TooLarge(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleOwner)
	// 1 МБ PNG-magic + просто наполнитель: важно превысить порог.
	body := bytes.NewReader(append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, MaxLogoFileSize)...))
	_, err := svc.UploadLogo(context.Background(), "acme", 42, body)
	assert.ErrorIs(t, err, ErrLogoFileTooLarge)
}

func TestUploadLogo_Empty(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleOwner)
	_, err := svc.UploadLogo(context.Background(), "acme", 42, bytes.NewReader(nil))
	assert.ErrorIs(t, err, ErrLogoFileMissing)
}

func TestUploadLogo_PlainText(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleOwner)
	_, err := svc.UploadLogo(context.Background(), "acme", 42, strings.NewReader("not an image, just text"))
	assert.ErrorIs(t, err, ErrLogoFileBadFormat)
}

// SVG distinguishes itself by XML preamble: http.DetectContentType вернёт
// "text/xml" → не в whitelist. Покрытие XSS-вектора со SVG-как-логотип.
func TestUploadLogo_SVGRejected(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleOwner)
	svgPayload := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="200" height="60"><script>alert(1)</script></svg>`)
	_, err := svc.UploadLogo(context.Background(), "acme", 42, bytes.NewReader(svgPayload))
	assert.ErrorIs(t, err, ErrLogoFileBadFormat)
}

// Polyglot: PNG signature + битое тело. http.DetectContentType увидит PNG,
// но image.DecodeConfig провалится на разборе IHDR/IDAT.
func TestUploadLogo_PolyglotPNGHeaderBadBody(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleOwner)
	bogus := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte("XXXX"), 32)...)
	_, err := svc.UploadLogo(context.Background(), "acme", 42, bytes.NewReader(bogus))
	assert.ErrorIs(t, err, ErrLogoFileBadFormat)
}

func TestUploadLogo_ImageTooLarge(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleOwner)
	// 2000×2000 RGBA PNG — пиксельный размер выше 1024×1024.
	// Декодеру не нужно декодировать пиксели для DecodeConfig — только заголовок,
	// поэтому byte-size payload останется в пределах 1MiB.
	body := bytes.NewReader(makePNG(t, 2000, 1500))
	_, err := svc.UploadLogo(context.Background(), "acme", 42, body)
	assert.ErrorIs(t, err, ErrLogoImageTooLarge)
}

func TestUploadLogo_OwnerPro_Blocked(t *testing.T) {
	svc, _ := setupLogoMocks(t, "pro", models.RoleOwner)
	_, err := svc.UploadLogo(context.Background(), "acme", 42, bytes.NewReader(makePNG(t, 50, 50)))
	assert.ErrorIs(t, err, ErrBrandingMaxOnly)
}

func TestUploadLogo_NotOwner(t *testing.T) {
	svc, _ := setupLogoMocks(t, "max", models.RoleEditor)
	_, err := svc.UploadLogo(context.Background(), "acme", 42, bytes.NewReader(makePNG(t, 50, 50)))
	assert.ErrorIs(t, err, ErrNotOwner)
}

// ===== DeleteLogo =====

func TestDeleteLogo_HappyPath(t *testing.T) {
	svc, lr := setupLogoMocks(t, "max", models.RoleOwner)
	lr.On("Delete", mock.Anything, uint(10)).Return(nil)

	err := svc.DeleteLogo(context.Background(), "acme", 42)
	assert.NoError(t, err)
	lr.AssertCalled(t, "Delete", mock.Anything, uint(10))
}

func TestDeleteLogo_OwnerPro_Blocked(t *testing.T) {
	svc, _ := setupLogoMocks(t, "pro", models.RoleOwner)
	err := svc.DeleteLogo(context.Background(), "acme", 42)
	assert.ErrorIs(t, err, ErrBrandingMaxOnly)
}

// ===== GetLogo =====

func TestGetLogo_FileSourceMax_ReturnsFile(t *testing.T) {
	svc, tr, ur := newTestService()
	lr := new(mockTeamLogoRepo)
	svc.SetLogoRepo(lr)

	team := &models.Team{ID: 10, Slug: "acme", CreatedBy: 42, BrandLogoSource: "file"}
	tr.On("GetBySlug", mock.Anything, "acme").Return(team, nil)
	ur.On("GetByID", mock.Anything, uint(42)).Return(&models.User{ID: 42, PlanID: "max"}, nil)

	expected := &models.TeamLogoFile{TeamID: 10, ContentType: "image/png", SizeBytes: 100, SHA256: "abc", Bytes: []byte("payload")}
	lr.On("Get", mock.Anything, uint(10)).Return(expected, nil)

	got, err := svc.GetLogo(context.Background(), "acme")
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

// source='url' → 404 даже для Max-owner; bytes хранилища нет, отдавать нечего.
func TestGetLogo_URLSource_NotFound(t *testing.T) {
	svc, tr, _ := newTestService()
	lr := new(mockTeamLogoRepo)
	svc.SetLogoRepo(lr)

	team := &models.Team{ID: 10, Slug: "acme", CreatedBy: 42, BrandLogoSource: "url"}
	tr.On("GetBySlug", mock.Anything, "acme").Return(team, nil)

	_, err := svc.GetLogo(context.Background(), "acme")
	assert.Error(t, err)
}

// owner downgraded с Max → public-отдача скрывает логотип, симметрично GetBrandingForShare.
func TestGetLogo_OwnerDowngraded_NotFound(t *testing.T) {
	svc, tr, ur := newTestService()
	lr := new(mockTeamLogoRepo)
	svc.SetLogoRepo(lr)

	team := &models.Team{ID: 10, Slug: "acme", CreatedBy: 42, BrandLogoSource: "file"}
	tr.On("GetBySlug", mock.Anything, "acme").Return(team, nil)
	ur.On("GetByID", mock.Anything, uint(42)).Return(&models.User{ID: 42, PlanID: "free"}, nil)

	_, err := svc.GetLogo(context.Background(), "acme")
	assert.Error(t, err)
	lr.AssertNotCalled(t, "Get")
}
