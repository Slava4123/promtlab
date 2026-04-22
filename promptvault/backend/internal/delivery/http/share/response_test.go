package share

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"promptvault/internal/models"
	shareuc "promptvault/internal/usecases/share"
)

// Регрессия-тест на BUG #2 (QA-сессия 2026-04-22):
// toPublicPromptResponse не копировал info.Branding в HTTP DTO — поле
// проваливалось на границе delivery-слоя. На Max-team с настроенным
// branding /api/s/:token возвращал ответ без logo/tagline/color.

func TestToPublicPromptResponse_CopiesBranding(t *testing.T) {
	branding := &models.BrandingInfo{
		LogoURL:      "https://cdn.example/logo.png",
		Tagline:      "QA Brand",
		Website:      "https://example.com",
		PrimaryColor: "#ff0066",
	}
	info := &shareuc.PublicPromptInfo{
		Title:     "Example",
		Content:   "body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Branding:  branding,
	}

	resp := toPublicPromptResponse(info)

	if assert.NotNil(t, resp.Branding, "Branding must be copied from usecase info to HTTP DTO") {
		assert.Equal(t, branding.LogoURL, resp.Branding.LogoURL)
		assert.Equal(t, branding.Tagline, resp.Branding.Tagline)
		assert.Equal(t, branding.Website, resp.Branding.Website)
		assert.Equal(t, branding.PrimaryColor, resp.Branding.PrimaryColor)
	}
}

func TestToPublicPromptResponse_NilBrandingStaysNil(t *testing.T) {
	info := &shareuc.PublicPromptInfo{Title: "Example"}
	resp := toPublicPromptResponse(info)
	assert.Nil(t, resp.Branding, "nil branding must remain nil (omitempty in JSON)")
}

func TestToPublicPromptResponse_BasicFieldsCopied(t *testing.T) {
	info := &shareuc.PublicPromptInfo{
		Title:   "Hello",
		Content: "Body",
		Model:   "claude-sonnet",
	}
	resp := toPublicPromptResponse(info)
	assert.Equal(t, "Hello", resp.Title)
	assert.Equal(t, "Body", resp.Content)
	assert.Equal(t, "claude-sonnet", resp.Model)
	assert.NotNil(t, resp.Tags, "tags slice initialized (empty, not nil)")
	assert.Len(t, resp.Tags, 0)
}
