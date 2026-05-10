// MN-1 — feedback.Service.Submit unit-тесты.
// Покрывает: тип-валидацию, длину сообщения, error wrap из repository,
// happy-path (Create + ID присваивается).
package feedback

import (
	"context"
	"errors"
	"strings"
	"testing"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// fakeFeedbackRepo — самодельный мок без testify (минимальный — Create-only).
type fakeFeedbackRepo struct {
	createErr  error
	created    *models.Feedback
	assignedID uint // если non-zero — присваиваем feedback.ID после Create
}

func (m *fakeFeedbackRepo) Create(_ context.Context, fb *models.Feedback) error {
	m.created = fb
	if m.assignedID != 0 {
		fb.ID = m.assignedID
	}
	return m.createErr
}

// Остальные методы интерфейса — заглушки (Submit использует только Create).
func (m *fakeFeedbackRepo) List(context.Context, repo.FeedbackListFilter) ([]repo.FeedbackListItem, int64, error) {
	panic("unused")
}
func (m *fakeFeedbackRepo) GetByID(context.Context, uint) (*repo.FeedbackDetail, error) {
	panic("unused")
}
func (m *fakeFeedbackRepo) UpdateStatus(context.Context, uint, models.FeedbackStatus) error {
	panic("unused")
}
func (m *fakeFeedbackRepo) Delete(context.Context, uint) error { panic("unused") }

// --- Tests ---

func TestSubmit_HappyPath_Bug(t *testing.T) {
	r := &fakeFeedbackRepo{assignedID: 42}
	svc := NewService(r)

	res, err := svc.Submit(context.Background(), SubmitInput{
		UserID:  1,
		Type:    string(models.FeedbackBug),
		Message: "Кнопка не работает",
		PageURL: "/dashboard",
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if res.ID != 42 {
		t.Errorf("expected ID=42 from repo, got %d", res.ID)
	}
	if r.created == nil {
		t.Fatal("expected Create called")
	}
	if r.created.UserID != 1 {
		t.Errorf("UserID mismatch: got %d", r.created.UserID)
	}
	if r.created.Type != models.FeedbackBug {
		t.Errorf("Type: got %q, want %q", r.created.Type, models.FeedbackBug)
	}
}

func TestSubmit_AllValidTypes_Pass(t *testing.T) {
	for _, ty := range []models.FeedbackType{models.FeedbackBug, models.FeedbackFeature, models.FeedbackOther} {
		t.Run(string(ty), func(t *testing.T) {
			r := &fakeFeedbackRepo{}
			svc := NewService(r)
			_, err := svc.Submit(context.Background(), SubmitInput{
				UserID: 1, Type: string(ty), Message: "test",
			})
			if err != nil {
				t.Errorf("expected nil for %q, got %v", ty, err)
			}
		})
	}
}

func TestSubmit_InvalidType_Refused(t *testing.T) {
	r := &fakeFeedbackRepo{}
	svc := NewService(r)
	_, err := svc.Submit(context.Background(), SubmitInput{
		UserID: 1, Type: "spam", Message: "x",
	})
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
	if r.created != nil {
		t.Error("expected Create NOT called on invalid input")
	}
}

func TestSubmit_TooLongMessage_Refused(t *testing.T) {
	r := &fakeFeedbackRepo{}
	svc := NewService(r)
	tooLong := strings.Repeat("я", MaxMessageLen+1) // rune-aware (cyrillic)
	_, err := svc.Submit(context.Background(), SubmitInput{
		UserID: 1, Type: string(models.FeedbackBug), Message: tooLong,
	})
	if !errors.Is(err, ErrMessageTooLong) {
		t.Fatalf("expected ErrMessageTooLong, got %v", err)
	}
}

func TestSubmit_ExactlyMaxLen_Allowed(t *testing.T) {
	r := &fakeFeedbackRepo{}
	svc := NewService(r)
	exactlyMax := strings.Repeat("a", MaxMessageLen)
	_, err := svc.Submit(context.Background(), SubmitInput{
		UserID: 1, Type: string(models.FeedbackBug), Message: exactlyMax,
	})
	if err != nil {
		t.Errorf("expected nil at exactly MaxMessageLen, got %v", err)
	}
}

func TestSubmit_RepoError_Wrapped(t *testing.T) {
	dbErr := errors.New("db connection lost")
	r := &fakeFeedbackRepo{createErr: dbErr}
	svc := NewService(r)
	_, err := svc.Submit(context.Background(), SubmitInput{
		UserID: 1, Type: string(models.FeedbackBug), Message: "x",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr propagated, got %v", err)
	}
}
