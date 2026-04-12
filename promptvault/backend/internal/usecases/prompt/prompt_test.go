package prompt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

func newTestService() (*Service, *mockPromptRepo, *mockVersionRepo, *mockTagRepo, *mockCollectionRepo) {
	pr := new(mockPromptRepo)
	vr := new(mockVersionRepo)
	tr := new(mockTagRepo)
	cr := new(mockCollectionRepo)
	svc := NewService(pr, tr, cr, vr, nil, nil, nil, nil)
	return svc, pr, vr, tr, cr
}

func testPrompt() *models.Prompt {
	return &models.Prompt{
		ID:      1,
		UserID:  10,
		Title:   "Старое название",
		Content: "Старое содержимое",
		Model:   "gpt-4o",
	}
}

// ===== Update: автоматическое создание версии =====

func TestUpdate_CreatesVersionSnapshot(t *testing.T) {
	svc, pr, vr, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt()

	// GetByID для проверки доступа
	pr.On("GetByID", ctx, uint(1)).Return(p, nil)
	// CreateWithNextVersion — проверяем что снимок содержит СТАРЫЕ данные
	vr.On("CreateWithNextVersion", ctx, mock.MatchedBy(func(v *models.PromptVersion) bool {
		return v.PromptID == 1 &&
			v.Title == "Старое название" &&
			v.Content == "Старое содержимое" &&
			v.Model == "gpt-4o" &&
			v.ChangeNote == "Обновил контент"
	})).Return(nil)
	// Update промпта
	pr.On("Update", ctx, mock.AnythingOfType("*models.Prompt")).Return(nil)
	// Повторный GetByID после Update
	updatedPrompt := &models.Prompt{ID: 1, UserID: 10, Title: "Новое название", Content: "Новый контент", Model: "gpt-4o"}
	pr.On("GetByID", ctx, uint(1)).Return(updatedPrompt, nil)

	newTitle := "Новое название"
	newContent := "Новый контент"
	result, _, err := svc.Update(ctx, 1, 10, UpdateInput{
		Title:      &newTitle,
		Content:    &newContent,
		ChangeNote: "Обновил контент",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Новое название", result.Title)
	vr.AssertCalled(t, "CreateWithNextVersion", ctx, mock.Anything)
}

func TestUpdate_ForbiddenForOtherUser(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt() // UserID = 10

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)

	newTitle := "Хак"
	_, _, err := svc.Update(ctx, 1, 999, UpdateInput{Title: &newTitle}) // userID = 999 ≠ 10

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUpdate_PromptNotFound(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	newTitle := "Тест"
	_, _, err := svc.Update(ctx, 99, 10, UpdateInput{Title: &newTitle})

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestUpdate_VersionCreateFails_ReturnsError(t *testing.T) {
	svc, pr, vr, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt()

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)
	vr.On("CreateWithNextVersion", ctx, mock.Anything).Return(assert.AnError)

	newTitle := "Тест"
	_, _, err := svc.Update(ctx, 1, 10, UpdateInput{Title: &newTitle})

	assert.Error(t, err)
	pr.AssertNotCalled(t, "Update", mock.Anything, mock.Anything) // промпт НЕ обновлён
}

// ===== ListVersions =====

func TestListVersions_Success(t *testing.T) {
	svc, pr, vr, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt()

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)
	versions := []models.PromptVersion{
		{ID: 2, PromptID: 1, VersionNumber: 2, Title: "v2"},
		{ID: 1, PromptID: 1, VersionNumber: 1, Title: "v1"},
	}
	vr.On("ListByPromptID", ctx, uint(1), 1, 20).Return(versions, int64(2), nil)

	result, total, err := svc.ListVersions(ctx, 1, 10, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, uint(2), result[0].VersionNumber)
}

func TestListVersions_ForbiddenForOtherUser(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt() // UserID = 10

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)

	_, _, err := svc.ListVersions(ctx, 1, 999, 1, 20) // userID = 999

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestListVersions_PromptNotFound(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	_, _, err := svc.ListVersions(ctx, 99, 10, 1, 20)

	assert.ErrorIs(t, err, ErrNotFound)
}

// ===== RevertToVersion =====

func TestRevertToVersion_Success(t *testing.T) {
	svc, pr, vr, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt()

	// GetByIDForPrompt возвращает старую версию
	oldVersion := &models.PromptVersion{
		ID:            1,
		PromptID:      1,
		VersionNumber: 1,
		Title:         "Оригинал",
		Content:       "Оригинальный контент",
		Model:         "claude-sonnet",
	}
	vr.On("GetByIDForPrompt", ctx, uint(1), uint(1)).Return(oldVersion, nil)

	// Update flow: GetByID → CreateWithNextVersion → Update → GetByID
	pr.On("GetByID", ctx, uint(1)).Return(p, nil)
	vr.On("CreateWithNextVersion", ctx, mock.MatchedBy(func(v *models.PromptVersion) bool {
		return v.ChangeNote == "Откат к версии 1"
	})).Return(nil)
	pr.On("Update", ctx, mock.MatchedBy(func(p *models.Prompt) bool {
		return p.Title == "Оригинал" && p.Content == "Оригинальный контент"
	})).Return(nil)
	revertedPrompt := &models.Prompt{ID: 1, UserID: 10, Title: "Оригинал", Content: "Оригинальный контент", Model: "claude-sonnet"}
	pr.On("GetByID", ctx, uint(1)).Return(revertedPrompt, nil)

	result, _, err := svc.RevertToVersion(ctx, 1, 10, 1)

	assert.NoError(t, err)
	assert.Equal(t, "Оригинал", result.Title)
	assert.Equal(t, "Оригинальный контент", result.Content)
}

func TestRevertToVersion_VersionNotFound(t *testing.T) {
	svc, _, vr, _, _ := newTestService()
	ctx := context.Background()

	vr.On("GetByIDForPrompt", ctx, uint(99), uint(1)).Return(nil, repo.ErrNotFound)

	_, _, err := svc.RevertToVersion(ctx, 1, 10, 99)

	assert.ErrorIs(t, err, ErrVersionNotFound)
}

func TestRevertToVersion_ForbiddenUser(t *testing.T) {
	svc, pr, vr, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt() // UserID = 10

	oldVersion := &models.PromptVersion{ID: 1, PromptID: 1, VersionNumber: 1, Title: "T", Content: "C", Model: "M"}
	vr.On("GetByIDForPrompt", ctx, uint(1), uint(1)).Return(oldVersion, nil)
	pr.On("GetByID", ctx, uint(1)).Return(p, nil)

	_, _, err := svc.RevertToVersion(ctx, 1, 999, 1) // userID = 999

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestRevertToVersion_CreatesSnapshotBeforeRevert(t *testing.T) {
	svc, pr, vr, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt() // Title: "Старое название", Content: "Старое содержимое"

	oldVersion := &models.PromptVersion{ID: 1, PromptID: 1, VersionNumber: 1, Title: "Оригинал", Content: "Оригинальный контент", Model: "gpt-4o"}
	vr.On("GetByIDForPrompt", ctx, uint(1), uint(1)).Return(oldVersion, nil)

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)
	// Проверяем что снимок содержит ТЕКУЩЕЕ состояние (до отката)
	vr.On("CreateWithNextVersion", ctx, mock.MatchedBy(func(v *models.PromptVersion) bool {
		return v.Title == "Старое название" &&
			v.Content == "Старое содержимое" &&
			v.ChangeNote == "Откат к версии 1"
	})).Return(nil)
	pr.On("Update", ctx, mock.Anything).Return(nil)
	revertedPrompt := &models.Prompt{ID: 1, UserID: 10, Title: "Оригинал", Content: "Оригинальный контент"}
	pr.On("GetByID", ctx, uint(1)).Return(revertedPrompt, nil)

	_, _, err := svc.RevertToVersion(ctx, 1, 10, 1)

	assert.NoError(t, err)
	vr.AssertCalled(t, "CreateWithNextVersion", ctx, mock.MatchedBy(func(v *models.PromptVersion) bool {
		return v.Title == "Старое название" // снимок ДО отката
	}))
}

// ===== GetByID =====

func TestGetByID_NotFound(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	_, err := svc.GetByID(ctx, 99, 10)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetByID_Forbidden(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt() // UserID = 10

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)

	_, err := svc.GetByID(ctx, 1, 999)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestGetByID_Success(t *testing.T) {
	svc, pr, _, _, _ := newTestService()
	ctx := context.Background()
	p := testPrompt()

	pr.On("GetByID", ctx, uint(1)).Return(p, nil)

	result, err := svc.GetByID(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, p.Title, result.Title)
}

