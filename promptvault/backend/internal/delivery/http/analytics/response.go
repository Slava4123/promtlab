package analytics

import (
	"encoding/json"

	"promptvault/internal/models"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// Response DTO для analytics-эндпоинтов. Напрямую возвращаем структуры
// analyticsuc, но обёрнутые в anon map для совместимости с фронтенд-ожиданиями
// (рендер не зависит от внутренних JSON-тегов типов из usecase-слоя).
//
// Используем type alias'ы и прямые ссылки — без доп. копирования данных.
// TanStack Query на фронте type-safe, но хранит минимум преобразований.

type InsightResponse struct {
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	ComputedAt string          `json:"computed_at"` // RFC3339
}

func toInsightResponses(insights []models.SmartInsight) []InsightResponse {
	result := make([]InsightResponse, len(insights))
	for i, ins := range insights {
		result[i] = InsightResponse{
			Type:       ins.InsightType,
			Payload:    ins.Payload,
			ComputedAt: ins.ComputedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
	}
	return result
}

// Используем готовые структуры analyticsuc.{PersonalDashboard,TeamDashboard,
// PromptAnalytics} — JSON-теги уже на них заданы. Ничего не переконвертируем.
var (
	_ = (*analyticsuc.PersonalDashboard)(nil)
	_ = (*analyticsuc.TeamDashboard)(nil)
	_ = (*analyticsuc.PromptAnalytics)(nil)
)
