// Package admin содержит usecases для административной панели:
// - чтение списка и деталей юзеров,
// - freeze/unfreeze,
// - reset password (триггер email со кодом, новый пароль юзер задаёт сам),
// - grant/revoke badges,
// - change tier (admin override plan_id без оплаты, с отменой активной подписки).
//
// Destructive actions обязательно пишут запись в audit_log через inject'енный
// audit.Service. AdminRequestInfo (admin_id, IP, User-Agent) читается из ctx
// через audit.FromContext — context должен быть пропущен через
// middleware/admin.AdminAuditContext на уровне HTTP.
//
// Валидации:
// - FreezeUser / ChangeRole не допускают операции над собой (self-lockout).
// - RevokeBadge требует fresh TOTP — проверка на уровне HTTP-handler,
//   здесь только бизнес-логика.
package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	auditsvc "promptvault/internal/usecases/audit"
	authuc "promptvault/internal/usecases/auth"
	badgeuc "promptvault/internal/usecases/badge"
)

// Service — usecase для административных операций.
// Композит из существующих usecase-сервисов (audit, auth, badge) + admin-repo.
//
// notifier и frontendURL — опциональны (для email-уведомлений в ChangeTier).
// Если notifier == nil или frontendURL == "" — email не отправляется.
// Устанавливаются через SetTierChangeNotifier после конструктора, чтобы не
// раздувать сигнатуру NewService и не ломать существующие тесты.
type Service struct {
	admins      repo.AdminRepository
	users       repo.UserRepository
	audit       *auditsvc.Service
	auth        *authuc.Service
	badges      *badgeuc.Service
	plans       repo.PlanRepository
	subs        repo.SubscriptionRepository
	notifier    TierChangeNotifier
	frontendURL string
}

func NewService(
	admins repo.AdminRepository,
	users repo.UserRepository,
	audit *auditsvc.Service,
	auth *authuc.Service,
	badges *badgeuc.Service,
	plans repo.PlanRepository,
	subs repo.SubscriptionRepository,
) *Service {
	return &Service{
		admins: admins,
		users:  users,
		audit:  audit,
		auth:   auth,
		badges: badges,
		plans:  plans,
		subs:   subs,
	}
}

// SetTierChangeNotifier подключает email-уведомление об изменении тарифа.
// Вызывается из app.go после NewService. Если не вызван — ChangeTier не шлёт email.
func (s *Service) SetTierChangeNotifier(n TierChangeNotifier, frontendURL string) {
	s.notifier = n
	s.frontendURL = frontendURL
}

// ==================== read-only ====================

// ListUsers возвращает страницу юзеров по фильтру. Не пишет в audit_log
// (read-only operations не логируются согласно OWASP Logging Cheat Sheet).
func (s *Service) ListUsers(ctx context.Context, filter UserListFilter) (*UserListResult, error) {
	items, total, err := s.admins.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	page := max(filter.Page, 1)
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return &UserListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetUserDetail возвращает полное представление юзера с агрегациями.
func (s *Service) GetUserDetail(ctx context.Context, userID uint) (*repo.UserDetail, error) {
	detail, err := s.admins.GetUserDetail(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return detail, nil
}

// ==================== destructive: user state ====================

// FreezeUser устанавливает status='frozen'. Юзер теряет возможность логина.
// Нельзя заморозить свой аккаунт (защита от self-lockout).
func (s *Service) FreezeUser(ctx context.Context, targetID uint) error {
	info, ok := auditsvc.FromContext(ctx)
	if !ok {
		return auditsvc.ErrMissingRequestInfo
	}
	if targetID == info.AdminID {
		return ErrCannotFreezeSelf
	}

	user, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if user.Status == models.StatusFrozen {
		// Идемпотентно: уже frozen — не пишем в audit, не дёргаем UPDATE.
		return nil
	}

	before := userStateSnapshot(user)
	if err := s.admins.UpdateStatus(ctx, targetID, models.StatusFrozen); err != nil {
		return err
	}
	after := map[string]any{"status": string(models.StatusFrozen)}

	return s.audit.Log(ctx, auditsvc.LogInput{
		Action:      auditsvc.ActionFreezeUser,
		TargetType:  auditsvc.TargetUser,
		TargetID:    &targetID,
		BeforeState: before,
		AfterState:  after,
	})
}

// UnfreezeUser устанавливает status='active'. Зеркало FreezeUser.
func (s *Service) UnfreezeUser(ctx context.Context, targetID uint) error {
	if _, ok := auditsvc.FromContext(ctx); !ok {
		return auditsvc.ErrMissingRequestInfo
	}

	user, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if user.Status == models.StatusActive {
		return nil
	}

	before := userStateSnapshot(user)
	if err := s.admins.UpdateStatus(ctx, targetID, models.StatusActive); err != nil {
		return err
	}
	after := map[string]any{"status": string(models.StatusActive)}

	return s.audit.Log(ctx, auditsvc.LogInput{
		Action:      auditsvc.ActionUnfreezeUser,
		TargetType:  auditsvc.TargetUser,
		TargetID:    &targetID,
		BeforeState: before,
		AfterState:  after,
	})
}

// ResetPassword инициирует password reset flow: генерирует код, отправляет
// email, инвалидирует refresh tokens. Сам пароль на этом этапе не меняется —
// юзер задаст его через /reset-password UI с полученным кодом.
func (s *Service) ResetPassword(ctx context.Context, targetID uint) error {
	if _, ok := auditsvc.FromContext(ctx); !ok {
		return auditsvc.ErrMissingRequestInfo
	}

	user, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.auth.AdminResetUserPassword(ctx, user); err != nil {
		return err
	}

	// В before/after state не кладём password_hash (PII). Только id+email
	// для идентификации в audit feed.
	state := map[string]any{
		"id":    user.ID,
		"email": user.Email,
	}
	return s.audit.Log(ctx, auditsvc.LogInput{
		Action:      auditsvc.ActionResetPassword,
		TargetType:  auditsvc.TargetUser,
		TargetID:    &targetID,
		BeforeState: state,
		AfterState:  state,
	})
}

// ChangeTier устанавливает plan_id для юзера (admin override, без оплаты).
// Если у юзера есть текущая подписка (active/past_due/paused) — переводит её
// в expired немедленно (без grace period) через MarkExpired. plan_id юзера
// устанавливается в указанный tier (не обязательно free).
//
// reason — опциональная причина, попадает в audit_log.after_state.source = "admin_override".
// При успехе шлёт юзеру email-уведомление (non-blocking, если notifier подключён).
//
// Идемпотентность: смена на тот же тариф = no-op (без записи в audit, без email).
func (s *Service) ChangeTier(ctx context.Context, targetID uint, tier, reason string) error {
	if _, ok := auditsvc.FromContext(ctx); !ok {
		return auditsvc.ErrMissingRequestInfo
	}

	// Validate plan exists
	if s.plans == nil {
		return ErrInvalidTier
	}
	if _, err := s.plans.GetByID(ctx, tier); err != nil {
		return ErrInvalidTier
	}

	user, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	oldPlan := user.PlanID
	if oldPlan == tier {
		// Idempotent: тариф уже такой же — ничего не делаем, не пишем в audit, не шлём email.
		return nil
	}

	// Завершить текущую подписку (active/past_due/paused), если есть. MarkExpired
	// меняет только статус подписки; users.plan_id мы перезапишем сами в нужный.
	if sub, subErr := s.subs.GetCurrentByUserID(ctx, targetID); subErr == nil {
		if expErr := s.subs.MarkExpired(ctx, sub.ID); expErr != nil {
			slog.Error("admin.change_tier.expire_sub_failed",
				"target_user_id", targetID, "sub_id", sub.ID, "error", expErr)
			return fmt.Errorf("expire existing subscription: %w", expErr)
		}
	} else if !errors.Is(subErr, repo.ErrNotFound) {
		// Любая другая ошибка — прерываемся: иначе plan_id обновится, а sub останется в подвисшем состоянии.
		return fmt.Errorf("lookup current subscription: %w", subErr)
	}

	// Update user plan
	user.PlanID = tier
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	before := map[string]any{"plan_id": oldPlan}
	after := map[string]any{
		"plan_id": tier,
		"reason":  reason,
		"source":  "admin_override",
	}
	if err := s.audit.Log(ctx, auditsvc.LogInput{
		Action:      auditsvc.ActionChangeTier,
		TargetType:  auditsvc.TargetUser,
		TargetID:    &targetID,
		BeforeState: before,
		AfterState:  after,
	}); err != nil {
		return err
	}

	// Non-blocking email — ошибка только логируется.
	if s.notifier != nil && s.frontendURL != "" {
		if err := s.notifier.SendAdminTierChanged(user.Email, user.Name, oldPlan, tier, reason, s.frontendURL); err != nil {
			slog.Warn("admin.change_tier.email_failed",
				"target_user_id", targetID, "error", err)
		}
	}

	slog.Info("admin.change_tier.success",
		"target_user_id", targetID, "old_plan", oldPlan, "new_plan", tier, "reason_set", reason != "")
	return nil
}

// ==================== destructive: badges ====================

// GrantBadge вручную разблокирует бейдж юзеру. Возвращает определение бейджа
// из каталога для включения в response. Если бейдж уже разблокирован —
// ErrBadgeAlreadyUnlocked.
func (s *Service) GrantBadge(ctx context.Context, targetID uint, badgeID string) (*badgeuc.Badge, error) {
	if _, ok := auditsvc.FromContext(ctx); !ok {
		return nil, auditsvc.ErrMissingRequestInfo
	}

	badge, ok := s.badges.GetByID(badgeID)
	if !ok {
		return nil, ErrBadgeNotFound
	}

	if _, err := s.users.GetByID(ctx, targetID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := s.badges.Unlock(ctx, targetID, badgeID); err != nil {
		if errors.Is(err, repo.ErrBadgeAlreadyUnlocked) {
			return nil, ErrBadgeAlreadyUnlocked
		}
		return nil, err
	}

	after := map[string]any{
		"badge_id": badgeID,
		"title":    badge.Title,
	}
	if err := s.audit.Log(ctx, auditsvc.LogInput{
		Action:      auditsvc.ActionGrantBadge,
		TargetType:  auditsvc.TargetUser,
		TargetID:    &targetID,
		BeforeState: nil,
		AfterState:  after,
	}); err != nil {
		return nil, err
	}
	return &badge, nil
}

// RevokeBadge удаляет бейдж у юзера. Идемпотентно на уровне repo — если
// записи нет, ничего не делает. Требует fresh TOTP (проверка в HTTP handler).
func (s *Service) RevokeBadge(ctx context.Context, targetID uint, badgeID string) error {
	if _, ok := auditsvc.FromContext(ctx); !ok {
		return auditsvc.ErrMissingRequestInfo
	}

	badge, ok := s.badges.GetByID(badgeID)
	if !ok {
		return ErrBadgeNotFound
	}

	if _, err := s.users.GetByID(ctx, targetID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	before := map[string]any{
		"badge_id": badgeID,
		"title":    badge.Title,
	}
	if err := s.badges.Revoke(ctx, targetID, badgeID); err != nil {
		return err
	}

	return s.audit.Log(ctx, auditsvc.LogInput{
		Action:      auditsvc.ActionRevokeBadge,
		TargetType:  auditsvc.TargetUser,
		TargetID:    &targetID,
		BeforeState: before,
		AfterState:  nil,
	})
}

// ==================== helpers ====================

// userStateSnapshot возвращает минимальное описание юзера для audit_log.
// Исключает password_hash, token_nonce и любую чувствительную информацию.
func userStateSnapshot(u *models.User) map[string]any {
	return map[string]any{
		"id":     u.ID,
		"email":  u.Email,
		"role":   string(u.Role),
		"status": string(u.Status),
	}
}
