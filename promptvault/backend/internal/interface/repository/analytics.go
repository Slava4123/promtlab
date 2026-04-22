package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// DateRange — полуоткрытый диапазон [From, To) для агрегаций.
type DateRange struct {
	From time.Time
	To   time.Time
}

// AnalyticsFilter — расширенный фильтр для drill-down по тегам/коллекциям
// (Phase 14 продолжение, задача #9 бэклога). Методы AnalyticsRepository
// постепенно мигрируют на приём этой структуры вместо отдельных args.
// TagID/CollectionID — опциональные, DateRange — обязательный.
type AnalyticsFilter struct {
	UserID       uint
	TeamID       *uint
	TagID        *uint
	CollectionID *uint
	Range        DateRange
}

// UsagePoint — точка таймсерии (день + count). Дата — day-precision UTC.
type UsagePoint struct {
	Day   time.Time `json:"day"`
	Count int64     `json:"count"`
}

// PromptUsageRow — элемент списка топ-промптов.
type PromptUsageRow struct {
	PromptID uint   `json:"prompt_id"`
	Title    string `json:"title"`
	Uses     int64  `json:"uses"`
}

// ContributorRow — строка contributors leaderboard для team-dashboard.
type ContributorRow struct {
	UserID         uint   `json:"user_id"`
	Email          string `json:"email"`
	Name           string `json:"name,omitempty"`
	PromptsCreated int64  `json:"prompts_created"`
	PromptsEdited  int64  `json:"prompts_edited"`
	Uses           int64  `json:"uses"`
}

// TrendRow — строка trending/declining инсайта.
// UsesLast7d/UsesPrev7d считаются SQL CTE'ями.
type TrendRow struct {
	PromptID     uint   `json:"prompt_id"`
	Title        string `json:"title"`
	UsesLast     int64  `json:"uses_last_7d"`
	UsesPrevious int64  `json:"uses_prev_7d"`
}

// ModelUsageRow — одна строка сегмента "использование по модели AI".
// Model = "" означает записи без заполненной модели промпта.
type ModelUsageRow struct {
	Model string `json:"model"`
	Uses  int64  `json:"uses"`
}

// AnalyticsRepository — aggregation-запросы для /api/analytics/*.
//
// Отделён от QuotaRepository: quota-repo отвечает за счётчики текущих
// лимитов (быстрые COUNT), analytics-repo — за исторические агрегации
// (таймсерии, топы, join'ы).
//
// Для всех методов: teamID == nil означает «личный скоуп юзера»
// (WHERE user_id = ? AND team_id IS NULL), teamID != nil означает
// «командный скоуп» (WHERE team_id = ?).
type AnalyticsRepository interface {
	// --- USAGE metrics (prompt_usage_log) ---

	// UsagePerDay — таймсерия count use'ов по дням.
	UsagePerDay(ctx context.Context, userID uint, teamID *uint, r DateRange) ([]UsagePoint, error)

	// TopPrompts — топ-N промптов по usage за период.
	TopPrompts(ctx context.Context, userID uint, teamID *uint, r DateRange, limit int) ([]PromptUsageRow, error)

	// UnusedPrompts — промпты пользователя без use с before (для Smart Insights).
	UnusedPrompts(ctx context.Context, userID uint, teamID *uint, before time.Time, limit int) ([]PromptUsageRow, error)

	// GetTrendingPrompts — растущие (growing=true) или падающие (growing=false)
	// промпты. Считается один SQL-запрос с двумя CTE (last-7d, prev-7d).
	// factor — коэффициент сравнения (2.0 для TRENDING, 0.5 для DECLINING).
	GetTrendingPrompts(ctx context.Context, userID uint, teamID *uint, factor float64, growing bool, limit int) ([]TrendRow, error)

	// PromptUsageTimeline — использование одного конкретного промпта по дням.
	// Отличается от UsagePerDay тем, что имеет WHERE prompt_id = ?
	// (UsagePerDay считает все промпты юзера разом).
	PromptUsageTimeline(ctx context.Context, promptID uint, r DateRange) ([]UsagePoint, error)

	// PromptShareViewsTimeline — просмотры share-ссылки конкретного промпта
	// по дням. Фильтр по sl.prompt_id для более точной картины чем общий
	// ShareViewsPerDay.
	PromptShareViewsTimeline(ctx context.Context, promptID uint, r DateRange) ([]UsagePoint, error)

	// UsageByModel — сегментация использования по AI-модели. Поле model_used
	// в prompt_usage_log заполняется при каждом IncrementUsage.
	UsageByModel(ctx context.Context, userID uint, teamID *uint, r DateRange) ([]ModelUsageRow, error)

	// --- CREATION activity (prompts.created_at + prompt_versions.created_at) ---

	PromptsCreatedPerDay(ctx context.Context, userID uint, teamID *uint, r DateRange) ([]UsagePoint, error)
	PromptsUpdatedPerDay(ctx context.Context, userID uint, teamID *uint, r DateRange) ([]UsagePoint, error)

	// Contributors — только для team (teamID required). Топ по суммарной активности.
	Contributors(ctx context.Context, teamID uint, r DateRange, limit int) ([]ContributorRow, error)

	// --- SHARE perf (share_view_log + share_links) ---

	// ShareViewsPerDay — просмотры всех активных шар-ссылок юзера.
	ShareViewsPerDay(ctx context.Context, userID uint, r DateRange) ([]UsagePoint, error)

	// TopSharedPrompts — топ промптов по просмотрам за период.
	TopSharedPrompts(ctx context.Context, userID uint, r DateRange, limit int) ([]PromptUsageRow, error)

	// LogShareView — вставка записи при просмотре (вызывается только для Pro+).
	LogShareView(ctx context.Context, view *models.ShareView) error

	// --- SMART INSIGHTS (Max only) ---

	// UpsertInsight — INSERT ... ON CONFLICT (user_id, COALESCE(team_id,0), insight_type)
	// DO UPDATE SET payload=?, computed_at=NOW().
	UpsertInsight(ctx context.Context, insight *models.SmartInsight) error

	// GetInsights — все активные инсайты для (userID, teamID).
	GetInsights(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error)

	// --- CLEANUP (cron) ---

	// DeleteShareViewsOlderThan — retention cleanup. Возвращает количество удалённых.
	DeleteShareViewsOlderThan(ctx context.Context, before time.Time) (int64, error)

	// CleanupShareViewsByRetention — массовый cleanup по plan_id владельца
	// share-ссылки. Pro=90д, Max=365д. Free не пишет в share_view_log, но
	// cleanup покрывает corner-case если план юзера был даунгрейднут.
	CleanupShareViewsByRetention(ctx context.Context) (int64, error)

	// CleanupPromptUsageByRetention — массовый cleanup prompt_usage_log
	// по plan_id юзера. Free=30д, Pro=90д, Max=365д. Без cleanup таблица
	// растёт линейно с usage — 100 записей/день активного Max-юзера ×
	// 10k юзеров × 5 лет = миллиард строк.
	CleanupPromptUsageByRetention(ctx context.Context) (int64, error)

	// --- SMART INSIGHTS M8 (feature-flag experimentalInsights) ---

	// MostEditedPrompts — промпты с наибольшим числом версий за период.
	// Использует COUNT(pv.id) GROUP BY p.id.
	MostEditedPrompts(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptUsageRow, error)

	// PossibleDuplicates — пары похожих промптов через pg_trgm.similarity().
	// threshold 0.8 по умолчанию — эмпирически выбрано, tuned during QA.
	// Требует CREATE EXTENSION pg_trgm (миграция 000048).
	PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]DuplicatePair, error)

	// OrphanTags — теги без единого промпта (тег был создан, но никогда
	// не ассоциирован или все ассоциации удалены).
	OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]TagRow, error)

	// EmptyCollections — коллекции без промптов.
	EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]CollectionRow, error)
}

// DuplicatePair — пара потенциально дублирующих промптов с similarity-скором.
type DuplicatePair struct {
	PromptAID    uint    `json:"prompt_a_id"`
	PromptATitle string  `json:"prompt_a_title"`
	PromptBID    uint    `json:"prompt_b_id"`
	PromptBTitle string  `json:"prompt_b_title"`
	Similarity   float32 `json:"similarity"`
}

// TagRow — минимальное представление тега для insights.
type TagRow struct {
	TagID uint   `json:"tag_id"`
	Name  string `json:"name"`
}

// CollectionRow — минимальное представление коллекции для insights.
type CollectionRow struct {
	CollectionID uint   `json:"collection_id"`
	Name         string `json:"name"`
}
