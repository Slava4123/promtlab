package feedback

import (
	"context"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — бизнес-логика обратной связи.
type Service struct {
	feedbacks repo.FeedbackRepository
}

func NewService(feedbacks repo.FeedbackRepository) *Service {
	return &Service{feedbacks: feedbacks}
}

// Submit валидирует и сохраняет обратную связь от пользователя.
func (s *Service) Submit(ctx context.Context, input SubmitInput) (*SubmitResult, error) {
	// Валидация типа
	switch models.FeedbackType(input.Type) {
	case models.FeedbackBug, models.FeedbackFeature, models.FeedbackOther:
		// ok
	default:
		return nil, ErrInvalidType
	}

	// Валидация длины сообщения
	if len([]rune(input.Message)) > MaxMessageLen {
		return nil, ErrMessageTooLong
	}

	fb := &models.Feedback{
		UserID:  input.UserID,
		Type:    models.FeedbackType(input.Type),
		Message: input.Message,
		PageURL: input.PageURL,
	}

	if err := s.feedbacks.Create(ctx, fb); err != nil {
		return nil, err
	}

	return &SubmitResult{ID: fb.ID}, nil
}
