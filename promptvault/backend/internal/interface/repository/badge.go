package repository

import (
	"context"
	"errors"

	"promptvault/internal/models"
)

// ErrBadgeAlreadyUnlocked возвращается из BadgeRepository.Unlock, когда
// запись уже существует (race condition при concurrent evaluate).
// Должна обрабатываться тихо в badge.Service.Evaluate — не попадать
// в newly_unlocked результат и не пробрасываться пользователю.
var ErrBadgeAlreadyUnlocked = errors.New("badge already unlocked")

// BadgeRepository изолирует работу с user_badges и все aggregation-методы,
// используемые для проверки условий разблокировки. Существующие
// PromptRepository / CollectionRepository намеренно НЕ расширяются, чтобы
// не раздувать их ответственность — все badge-специфичные counters здесь.
type BadgeRepository interface {
	// Unlock вставляет запись в user_badges. Если запись уже есть
	// (UNIQUE constraint violation на (user_id, badge_id)) — возвращает
	// ErrBadgeAlreadyUnlocked. Любая другая ошибка — прозрачно.
	Unlock(ctx context.Context, userID uint, badgeID string) error

	// UnlockedIDs возвращает set разблокированных badge_id для юзера.
	// Используется badge.Service.Evaluate для short-circuit «уже разблокирован?».
	// Для юзера без бейджей — пустой (не nil) map.
	UnlockedIDs(ctx context.Context, userID uint) (map[string]struct{}, error)

	// ListByUser возвращает все UserBadge для юзера, отсортированные по
	// unlocked_at DESC. Для страницы /badges и admin user detail.
	ListByUser(ctx context.Context, userID uint) ([]models.UserBadge, error)

	// DeleteByUserAndBadge удаляет запись из user_badges (admin revoke).
	// Используется в usecases/admin.RevokeBadge. Если записи нет — returns nil
	// (идемпотентно); факт revoke фиксируется в audit_log на уровне Service.
	DeleteByUserAndBadge(ctx context.Context, userID uint, badgeID string) error

	// --- aggregation методы для проверки условий ---

	// CountSoloPrompts — COUNT(prompts WHERE user_id=? AND team_id IS NULL AND deleted_at IS NULL).
	// Для бейджей first_prompt, architect.
	CountSoloPrompts(ctx context.Context, userID uint) (int64, error)

	// CountTeamPrompts — COUNT(prompts WHERE user_id=? AND team_id IS NOT NULL AND deleted_at IS NULL).
	// Для бейджей team_player, team_lead.
	CountTeamPrompts(ctx context.Context, userID uint) (int64, error)

	// CountAllPrompts — COUNT(prompts WHERE user_id=? AND deleted_at IS NULL).
	// Для флагманского бейджа prompt_master (25 промптов всего).
	CountAllPrompts(ctx context.Context, userID uint) (int64, error)

	// CountSoloCollections — COUNT(collections WHERE user_id=? AND team_id IS NULL AND deleted_at IS NULL).
	// Для бейджа collector.
	CountSoloCollections(ctx context.Context, userID uint) (int64, error)

	// CountTeamCollections — COUNT(collections WHERE user_id=? AND team_id IS NOT NULL AND deleted_at IS NULL).
	// Для бейджа team_librarian.
	CountTeamCollections(ctx context.Context, userID uint) (int64, error)

	// SumUsage — SUM(prompts.usage_count WHERE user_id=? AND deleted_at IS NULL).
	// Для бейджа advanced (50 использований).
	SumUsage(ctx context.Context, userID uint) (int64, error)

	// CountVersionedPrompts возвращает количество промптов юзера, у которых
	// >= minVersions записей в prompt_versions. Для бейджа refactorer.
	// Вызывается только при событии prompt_updated со short-circuit в Service.
	CountVersionedPrompts(ctx context.Context, userID uint, minVersions int) (int64, error)
}
