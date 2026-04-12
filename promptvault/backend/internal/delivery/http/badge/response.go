package badge

import (
	"time"

	badgeuc "promptvault/internal/usecases/badge"
)

// BadgeResponse — transport-представление одной записи каталога + состояние
// пользователя. Отделено от domain-типа badgeuc.BadgeWithState, чтобы
// переименование/реорганизация полей на уровне usecase не ломала контракт API.
type BadgeResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Category    string     `json:"category"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
	Progress    int64      `json:"progress"`
	Target      int64      `json:"target"`
}

// BadgeListResponse — ответ GET /api/badges. Включает агрегаты total_count
// и total_unlocked, чтобы фронту не пришлось считать отдельно.
type BadgeListResponse struct {
	Items         []BadgeResponse `json:"items"`
	TotalCount    int             `json:"total_count"`
	TotalUnlocked int             `json:"total_unlocked"`
}

// BadgeSummary — минимальный DTO для поля newly_unlocked_badges в ответах
// mutating API (POST /api/prompts, /api/collections, /api/prompts/{id}/use, etc).
// Сюда входит ровно то, что фронту нужно для toast-уведомления: id для ключа,
// title + icon + description для содержимого. Exported чтобы response-структуры
// других пакетов (delivery/http/prompt, /collection) могли импортировать этот тип.
type BadgeSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// NewBadgeResponse конвертит domain-объект в transport-DTO.
func NewBadgeResponse(b badgeuc.BadgeWithState) BadgeResponse {
	return BadgeResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Icon:        b.Icon,
		Category:    string(b.Category),
		Unlocked:    b.Unlocked,
		UnlockedAt:  b.UnlockedAt,
		Progress:    b.Progress,
		Target:      b.Target,
	}
}

// NewBadgeListResponse строит полный ответ GET /api/badges из списка состояний.
// Подсчитывает total_unlocked проходом по items — избегаем второго запроса к БД.
func NewBadgeListResponse(items []badgeuc.BadgeWithState) BadgeListResponse {
	responses := make([]BadgeResponse, 0, len(items))
	unlockedCount := 0
	for _, b := range items {
		responses = append(responses, NewBadgeResponse(b))
		if b.Unlocked {
			unlockedCount++
		}
	}
	return BadgeListResponse{
		Items:         responses,
		TotalCount:    len(items),
		TotalUnlocked: unlockedCount,
	}
}

// NewBadgeSummaries конвертит []badgeuc.Badge (возврат Service.Evaluate) в
// []BadgeSummary для встраивания в response-структуры других HTTP-пакетов
// (prompt.PromptResponse.NewlyUnlockedBadges, collection.CollectionResponse.NewlyUnlockedBadges).
// При пустом входе возвращает nil — чтобы с `json:",omitempty"` поле
// не попадало в JSON, когда ничего не разблокировалось.
func NewBadgeSummaries(badges []badgeuc.Badge) []BadgeSummary {
	if len(badges) == 0 {
		return nil
	}
	out := make([]BadgeSummary, 0, len(badges))
	for _, b := range badges {
		out = append(out, BadgeSummary{
			ID:          b.ID,
			Title:       b.Title,
			Description: b.Description,
			Icon:        b.Icon,
		})
	}
	return out
}
