package audit

import (
	"context"
	"encoding/json"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — usecase для записи админ-действий в audit_log.
// Все поля admin_id/ip/user_agent автоматически прокидываются через
// AdminRequestInfo из context (см. middleware/admin.AdminAuditContext).
type Service struct {
	repo repo.AuditRepository
}

func NewService(auditRepo repo.AuditRepository) *Service {
	return &Service{repo: auditRepo}
}

// LogInput — минимальный набор параметров для audit.Log.
// BeforeState/AfterState — arbitrary struct/map, будет marshal в JSON.
// Не кладите сюда password_hash, TOTP secrets или JWT — это PII utrust.
type LogInput struct {
	Action      Action
	TargetType  TargetType
	TargetID    *uint
	BeforeState any
	AfterState  any
}

// Log записывает одно событие в audit_log. AdminID/IP/UserAgent берутся из ctx.
// Если ctx не содержит AdminRequestInfo — возвращает ErrMissingRequestInfo.
// Вызывающая сторона должна либо (а) добавить AdminAuditContext middleware
// к роуту, либо (б) явно вручную положить AdminRequestInfo через WithContext.
func (s *Service) Log(ctx context.Context, in LogInput) error {
	info, ok := FromContext(ctx)
	if !ok {
		return ErrMissingRequestInfo
	}

	before, err := marshalState(in.BeforeState)
	if err != nil {
		return err
	}
	after, err := marshalState(in.AfterState)
	if err != nil {
		return err
	}

	entry := &models.AuditLog{
		AdminID:     info.AdminID,
		Action:      string(in.Action),
		TargetType:  string(in.TargetType),
		TargetID:    in.TargetID,
		BeforeState: before,
		AfterState:  after,
		IP:          info.IP,
		UserAgent:   info.UserAgent,
	}
	return s.repo.Log(ctx, entry)
}

// List — pass-through к repo.List для admin audit log feed endpoint.
func (s *Service) List(ctx context.Context, filter repo.AuditLogFilter) ([]models.AuditLog, int64, error) {
	return s.repo.List(ctx, filter)
}

// marshalState конвертит arbitrary state в json.RawMessage для jsonb-колонки.
// nil → nil (не marshal'ить в "null"), иначе — json.Marshal.
func marshalState(state any) ([]byte, error) {
	if state == nil {
		return nil, nil
	}
	return json.Marshal(state)
}
