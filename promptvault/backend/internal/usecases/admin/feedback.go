package admin

import (
	"context"
	"errors"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	auditsvc "promptvault/internal/usecases/audit"
)

// FeedbackListResult — DTO для admin GET /admin/feedbacks с пагинацией.
type FeedbackListResult struct {
	Items    []repo.FeedbackListItem
	Total    int64
	Page     int
	PageSize int
}

// SetFeedbackRepository — late-bind setter для FeedbackRepository.
// Используется тем же паттерном, что и SetTierChangeNotifier — чтобы не
// раздувать сигнатуру NewService и не ломать существующие call-site'ы.
//
// Вызывается из app.go после feedbackRepo создания. Если не вызван —
// list/get/update/delete вернут ErrFeedbackUnavailable.
func (s *Service) SetFeedbackRepository(f repo.FeedbackRepository) {
	s.feedbacks = f
}

// ErrFeedbackUnavailable — fail-fast если admin handler пытается работать
// с feedbacks без вызова SetFeedbackRepository в wire-up.
var ErrFeedbackUnavailable = errors.New("feedback repository not configured")

// ==================== read-only ====================

// ListFeedbacks возвращает страницу отзывов под фильтром.
// Без audit-логирования (read-only operation).
func (s *Service) ListFeedbacks(ctx context.Context, filter repo.FeedbackListFilter) (*FeedbackListResult, error) {
	if s.feedbacks == nil {
		return nil, ErrFeedbackUnavailable
	}
	items, total, err := s.feedbacks.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	page := max(filter.Page, 1)
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return &FeedbackListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetFeedbackDetail возвращает один отзыв с user-полями.
// Без audit-логирования.
func (s *Service) GetFeedbackDetail(ctx context.Context, id uint) (*repo.FeedbackDetail, error) {
	if s.feedbacks == nil {
		return nil, ErrFeedbackUnavailable
	}
	d, err := s.feedbacks.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrFeedbackNotFound
		}
		return nil, err
	}
	return d, nil
}

// ==================== mutations ====================

// UpdateFeedbackStatus меняет status отзыва. Требует AdminRequestInfo в ctx
// (через middleware/admin.AdminAuditContext).
//
// Идемпотентность: если новый status == текущему, no-op (без audit-записи).
// Это нормально — клик «пометить как прочитанный» дважды не создаёт спам в audit.
func (s *Service) UpdateFeedbackStatus(ctx context.Context, id uint, newStatus models.FeedbackStatus) error {
	if s.feedbacks == nil {
		return ErrFeedbackUnavailable
	}
	if !newStatus.IsValid() {
		return ErrInvalidFeedbackStatus
	}

	// Snapshot before (для audit + idempotency check).
	before, err := s.feedbacks.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrFeedbackNotFound
		}
		return err
	}
	if before.Status == newStatus {
		return nil // идемпотентно
	}

	if err := s.feedbacks.UpdateStatus(ctx, id, newStatus); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrFeedbackNotFound
		}
		return err
	}

	return s.audit.Log(ctx, auditsvc.LogInput{
		Action:     auditsvc.ActionUpdateFeedbackStatus,
		TargetType: auditsvc.TargetFeedback,
		TargetID:   &id,
		BeforeState: map[string]any{
			"status": before.Status,
		},
		AfterState: map[string]any{
			"status": newStatus,
		},
	})
}

// DeleteFeedback удаляет отзыв навсегда. Требует AdminRequestInfo в ctx
// и fresh TOTP (проверяется на уровне HTTP handler перед вызовом).
//
// BeforeState содержит снэпшот (без full message — обрезается до 200 символов,
// чтобы не раздувать audit_log при большом тексте).
func (s *Service) DeleteFeedback(ctx context.Context, id uint) error {
	if s.feedbacks == nil {
		return ErrFeedbackUnavailable
	}

	before, err := s.feedbacks.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrFeedbackNotFound
		}
		return err
	}

	if err := s.feedbacks.Delete(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrFeedbackNotFound
		}
		return err
	}

	return s.audit.Log(ctx, auditsvc.LogInput{
		Action:     auditsvc.ActionDeleteFeedback,
		TargetType: auditsvc.TargetFeedback,
		TargetID:   &id,
		BeforeState: map[string]any{
			"id":              before.ID,
			"user_id":         before.UserID,
			"user_email":      before.UserEmail,
			"type":            before.Type,
			"status":          before.Status,
			"message_excerpt": truncate(before.Message, 200),
			"page_url":        before.PageURL,
			"created_at":      before.CreatedAt,
		},
		AfterState: nil, // удалено
	})
}

// truncate обрезает строку до n символов (rune-safe для кириллицы).
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
