# Analytics UX Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Зафиксировать 7 UX-багов аналитики из smoke-test (broken deep links Smart Insights, английский в локализации, heatmap 2 ячейки вместо 28, self-link banner, sparkline-черта для константных данных, числовые x-tick'и, narrative edge cases) через 2 волны: Wave 1 — реальные backend insight endpoints + 5 dedicated frontend pages + filter overlays для tags/collections; Wave 2 — UI polish (русификация, padding, formatters).

**Architecture:** Backend Clean Architecture — новый usecase `prompt_insights/` поверх существующего `analytics_repo.go` SQL (без дублирования), handler в `delivery/http/prompt/insights_handler.go`, plan-gating через переиспользуемый `analytics.Service.insightsForPlan(planID)`. Frontend — 5 новых dedicated routes `/prompts/insights/:type`, два filter-overlay'я на existing menu pages, перепрошивка `INSIGHT_META.href` в `insights-panel.tsx`. Wave 2 — точечные правки 5 компонентов и pure-function. TDD pipeline на каждый task.

**Tech Stack:** Backend — Go 1.25, Chi v5, GORM v2, PostgreSQL 18, testcontainers-go (integration tests), slog. Frontend — React 19.2, TanStack Query v5, React Router 7.13, shadcn/ui Dialog для merge-modal, Lucide icons, Vitest + Testing Library + jest-dom.

**Дизайн-спека:** [docs/superpowers/specs/2026-05-17-analytics-ux-fixes-design.md](../specs/2026-05-17-analytics-ux-fixes-design.md)

---

## File Structure

### Создаём (Wave 1 — Backend)

| Путь | Назначение |
|---|---|
| `backend/internal/usecases/prompt_insights/types.go` | DTO `PromptInsightRow`, `DuplicatePair` |
| `backend/internal/usecases/prompt_insights/errors.go` | `ErrUnknownInsightType`, `ErrPromptsNotOwned`, `ErrSamePrompt`, `ErrProRequired` |
| `backend/internal/usecases/prompt_insights/service.go` | `Service` с 5 `List*` методами + `MergePrompts` |
| `backend/internal/usecases/prompt_insights/service_test.go` | unit-тесты (mock'и repo интерфейсов) |
| `backend/internal/delivery/http/prompt/insights_handler.go` | 5 GET handlers + Merge POST |
| `backend/internal/delivery/http/prompt/insights_handler_test.go` | httptest integration на handler+usecase+mocks |
| `backend/internal/delivery/http/tag/orphan_handler.go` | `GET /api/tags/orphan` |
| `backend/internal/delivery/http/tag/orphan_handler_test.go` | httptest |
| `backend/internal/delivery/http/collection/empty_handler.go` | `GET /api/collections/empty` |
| `backend/internal/delivery/http/collection/empty_handler_test.go` | httptest |
| `backend/internal/infrastructure/postgres/repository/prompt_merge_test.go` | testcontainers integration для `MergeWith` |

### Меняем (Wave 1 — Backend)

| Путь | Что меняем |
|---|---|
| `backend/internal/usecases/analytics/service.go` | Экспонируем `InsightsForPlan(planID) []string` (текущий нижний регистр `insightsForPlan` — promote в public) |
| `backend/internal/interface/repository/prompt.go` | Добавить интерфейс `MergeWith(ctx, keepID, mergeID, userID uint) error` |
| `backend/internal/infrastructure/postgres/repository/prompt_repo.go` | Реализация `MergeWith` — ownership check + soft-delete merge target в одной транзакции |
| `backend/internal/app/app.go` | Wire-up `promptInsights.NewService(...)`, handler init |
| `backend/internal/app/routes.go` | Регистрация `/api/prompts/insights/*`, `/api/prompts/{id}/merge-with/{other_id}`, `/api/tags/orphan`, `/api/collections/empty` |

### Создаём (Wave 1 — Frontend)

| Путь | Назначение |
|---|---|
| `frontend/src/api/prompt-insights.ts` | Fetcher functions + TS types |
| `frontend/src/api/prompt-insights.test.ts` | Unit tests на fetchers |
| `frontend/src/hooks/use-prompt-insights.ts` | TanStack Query hooks (5 queries + 1 mutation) |
| `frontend/src/hooks/use-prompt-insights.test.ts` | Hook tests |
| `frontend/src/components/prompts/insights/insight-prompt-row.tsx` | Reusable row (title + uses + actions slot) |
| `frontend/src/components/prompts/insights/insight-prompt-row.test.tsx` | Unit test |
| `frontend/src/components/prompts/insights/merge-modal.tsx` | Side-by-side modal для duplicate merge |
| `frontend/src/components/prompts/insights/merge-modal.test.tsx` | Unit test |
| `frontend/src/pages/prompts/insights/unused.tsx` | Insight page «забытые» |
| `frontend/src/pages/prompts/insights/unused.test.tsx` | Page test |
| `frontend/src/pages/prompts/insights/duplicates.tsx` | Page с merge-modal |
| `frontend/src/pages/prompts/insights/duplicates.test.tsx` | Page test |
| `frontend/src/pages/prompts/insights/trending.tsx` | Page «растущие» |
| `frontend/src/pages/prompts/insights/declining.tsx` | Page «падающие» |
| `frontend/src/pages/prompts/insights/most-edited.tsx` | Page «часто правят» |
| `frontend/src/pages/prompts/insights/__tests__/list-pages.test.tsx` | Общий test для 3 list-style страниц (trending/declining/most-edited) |
| `frontend/src/pages/tags-page.tsx` | NEW: минимальная tags management page + `?filter=orphan` overlay |
| `frontend/src/pages/tags-page.test.tsx` | Page test |
| `frontend/src/lib/date-format.ts` | `formatDayShort(iso)` через `Intl.DateTimeFormat('ru-RU')` |
| `frontend/src/lib/date-format.test.ts` | Unit tests |

### Меняем (Wave 1 — Frontend)

| Путь | Что меняем |
|---|---|
| `frontend/src/App.tsx` | 5 routes `/prompts/insights/:type` + 1 route `/tags` (новый) |
| `frontend/src/components/analytics/insights-panel.tsx` | `INSIGHT_META.*.href` (7 ссылок) + `orphan_tags.title` |
| `frontend/src/pages/collections.tsx` | Поддержать `?filter=empty` query param через `useEmptyCollections` |

### Меняем (Wave 2 — UI polish)

| Путь | Что меняем |
|---|---|
| `frontend/src/components/analytics/activity-heatmap.tsx` | Pad до 28 ячеек через date math, русский tooltip |
| `frontend/src/components/analytics/activity-heatmap.test.tsx` | Расширить тесты до 28 cells |
| `frontend/src/components/analytics/narrative-banner.tsx` | Убрать `<a href>` и ArrowRight |
| `frontend/src/components/analytics/narrative-banner.test.tsx` | Обновить — нет link |
| `frontend/src/components/analytics/sparkline.tsx` | Constant data → одиночная точка |
| `frontend/src/components/analytics/sparkline.test.tsx` | Расширить test (constant data) |
| `frontend/src/components/analytics/usage-chart.tsx` | `tickFormatter={formatDayShort}` |
| `frontend/src/lib/analytics-narrative.ts` | Skip topModel когда пусто/`pct===100 && segments.length===1`; skip streak когда `current===0` |
| `frontend/src/lib/analytics-narrative.test.ts` | Расширить test cases |

### Не трогаем (existing, переиспользуем)

- `backend/internal/infrastructure/postgres/repository/analytics_repo.go` — SQL queries уже есть, новый usecase их вызывает.
- `backend/internal/usecases/prompt/service.go` — base CRUD не меняем.
- `frontend/src/components/prompts/prompt-card.tsx` — переиспользуем в insight pages где уместно.
- `frontend/src/hooks/use-prompts.ts` — для bulk operations.
- `frontend/src/pages/analytics.tsx` — никаких изменений; только зависимые компоненты.
- `frontend/src/pages/dashboard.tsx` — теги inline остаются работающими как есть.

---

## Phase 1 — Wave 1: Backend insights endpoints

### Task B1: prompt_insights usecase — types + errors

**Files:**
- Create: `backend/internal/usecases/prompt_insights/types.go`
- Create: `backend/internal/usecases/prompt_insights/errors.go`

- [ ] **Step 1: Write the failing test (compile-time)**

Create `backend/internal/usecases/prompt_insights/types_test.go`:

```go
package prompt_insights

import (
	"testing"
	"time"
)

func TestPromptInsightRowJSON(t *testing.T) {
	r := PromptInsightRow{PromptID: 42, Title: "X", Uses: 10, UpdatedAt: time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)}
	if r.Title != "X" {
		t.Fatalf("Title mismatch: %v", r.Title)
	}
}

func TestErrSentinels(t *testing.T) {
	for _, e := range []error{ErrUnknownInsightType, ErrPromptsNotOwned, ErrSamePrompt, ErrProRequired} {
		if e == nil {
			t.Fatalf("expected non-nil sentinel error")
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/backend
go test ./internal/usecases/prompt_insights/...
```

Expected: FAIL — package not found.

- [ ] **Step 3: Create files**

`types.go`:

```go
package prompt_insights

import "time"

// PromptInsightRow — row для list-style insight endpoints
// (unused / trending / declining / most-edited).
// Поле UpdatedAt опционально (omitempty) — некоторые SQL возвращают только
// PromptID/Title/Uses без timestamp.
type PromptInsightRow struct {
	PromptID  uint      `json:"prompt_id"`
	Title     string    `json:"title"`
	Uses      int       `json:"uses"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// DuplicatePair — пара похожих промптов из possible_duplicates SQL.
type DuplicatePair struct {
	PromptA    PromptInsightRow `json:"prompt_a"`
	PromptB    PromptInsightRow `json:"prompt_b"`
	Similarity float64          `json:"similarity"`
}
```

`errors.go`:

```go
package prompt_insights

import "errors"

var (
	// ErrUnknownInsightType — handler передал тип, которого нет в whitelist.
	ErrUnknownInsightType = errors.New("unknown insight type")

	// ErrPromptsNotOwned — юзер запрашивает merge промптов, которые ему не принадлежат.
	ErrPromptsNotOwned = errors.New("prompts not owned by user")

	// ErrSamePrompt — merge id1 == id2.
	ErrSamePrompt = errors.New("cannot merge prompt with itself")

	// ErrProRequired — план юзера не разрешает данный insight type.
	// Маппится в HTTP 402 (как в usecases/analytics).
	ErrProRequired = errors.New("pro plan required for this insight")
)
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/usecases/prompt_insights/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/prompt_insights/types.go backend/internal/usecases/prompt_insights/errors.go backend/internal/usecases/prompt_insights/types_test.go
git commit -m "feat(prompt-insights): добавить usecase types + sentinel errors"
```

---

### Task B2: Expose `analytics.InsightsForPlan` (promote из private)

**Files:**
- Modify: `backend/internal/usecases/analytics/service.go`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/usecases/analytics/insights_for_plan_export_test.go` (или вписать в существующий `service_test.go`, если он есть):

```go
package analytics

import (
	"slices"
	"testing"
)

func TestInsightsForPlanPublic(t *testing.T) {
	s := &Service{proInsightsTeaserEnabled: true}
	free := s.InsightsForPlan("free")
	pro := s.InsightsForPlan("pro")
	max := s.InsightsForPlan("max")

	if len(free) != 0 {
		t.Fatalf("Free should get [], got %v", free)
	}
	if !slices.Equal(pro, []string{"unused_prompts", "possible_duplicates"}) {
		t.Fatalf("Pro teaser: got %v", pro)
	}
	if len(max) != 7 {
		t.Fatalf("Max should get 7 types, got %d", len(max))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestInsightsForPlanPublic ./internal/usecases/analytics/...
```

Expected: FAIL — `InsightsForPlan` undefined (есть только нижнерегистровый `insightsForPlan`).

- [ ] **Step 3: Promote method**

Find `insightsForPlan` in `backend/internal/usecases/analytics/service.go` (или родственном файле в пакете) и переименовать в `InsightsForPlan`. Внутри пакета все callers обновить через find-and-replace. Внешние пакеты теперь могут вызывать `analyticsSvc.InsightsForPlan(planID)`.

- [ ] **Step 4: Run all analytics tests**

```bash
go test ./internal/usecases/analytics/...
```

Expected: PASS — никаких регрессий, плюс новый TestInsightsForPlanPublic.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/analytics/
git commit -m "refactor(analytics): publicize InsightsForPlan для prompt-insights gating"
```

---

### Task B3: prompt_insights.Service — ListUnused + unit test

**Files:**
- Create: `backend/internal/usecases/prompt_insights/service.go`
- Create: `backend/internal/usecases/prompt_insights/service_test.go`

- [ ] **Step 1: Write the failing test**

```go
package prompt_insights

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "promptvault/internal/interface/repository"
)

type fakeAnalyticsRepo struct {
	unused []repo.PromptUsageRow
}

func (f *fakeAnalyticsRepo) UnusedPrompts(ctx context.Context, userID uint, teamID *uint, before time.Time, limit int) ([]repo.PromptUsageRow, error) {
	return f.unused, nil
}

// stubs для остальных AnalyticsRepository методов — возвращают nil/empty
// (вынесены в helper-файл fakes_test.go для DRY).

type fakePlanLookup struct{ plan string }

func (p *fakePlanLookup) InsightsForPlan(planID string) []string {
	switch planID {
	case "max", "max_yearly":
		return []string{"unused_prompts", "possible_duplicates", "trending", "declining", "most_edited", "orphan_tags", "empty_collections"}
	case "pro", "pro_yearly":
		return []string{"unused_prompts", "possible_duplicates"}
	}
	return nil
}

func (p *fakePlanLookup) LookupPlanID(ctx context.Context, userID uint) (string, error) {
	return p.plan, nil
}

func TestListUnusedMaxPlan(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		unused: []repo.PromptUsageRow{
			{PromptID: 1, Title: "A", Uses: 0},
			{PromptID: 2, Title: "B", Uses: 0},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)

	rows, err := svc.ListUnused(context.Background(), 100, nil, 50)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].PromptID != 1 || rows[0].Title != "A" {
		t.Fatalf("row[0] mismatch: %+v", rows[0])
	}
}

func TestListUnusedFreePlanBlocked(t *testing.T) {
	svc := NewService(&fakeAnalyticsRepo{}, nil, &fakePlanLookup{plan: "free"}, time.Now)
	_, err := svc.ListUnused(context.Background(), 100, nil, 50)
	if !errors.Is(err, ErrProRequired) {
		t.Fatalf("expected ErrProRequired, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/usecases/prompt_insights/...
```

Expected: FAIL — `NewService` / `ListUnused` not defined.

- [ ] **Step 3: Implement Service skeleton + ListUnused**

`service.go`:

```go
package prompt_insights

import (
	"context"
	"time"

	repo "promptvault/internal/interface/repository"
)

const (
	insightTypeUnused     = "unused_prompts"
	insightTypeDuplicates = "possible_duplicates"
	insightTypeTrending   = "trending"
	insightTypeDeclining  = "declining"
	insightTypeMostEdited = "most_edited"
)

// PlanGate — DI для проверки тарифа. Реализуется *analytics.Service (через
// SetPlanGate в app.go) — избегаем циклической зависимости usecases→usecases.
type PlanGate interface {
	InsightsForPlan(planID string) []string
	LookupPlanID(ctx context.Context, userID uint) (string, error)
}

// PromptMerger — узкий интерфейс на repo.PromptRepository.MergeWith.
type PromptMerger interface {
	MergeWith(ctx context.Context, keepID, mergeID, userID uint) error
}

type Service struct {
	analytics repo.AnalyticsRepository
	prompts   PromptMerger
	plans     PlanGate
	nowFn     func() time.Time
}

func NewService(analytics repo.AnalyticsRepository, prompts PromptMerger, plans PlanGate, nowFn func() time.Time) *Service {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Service{analytics: analytics, prompts: prompts, plans: plans, nowFn: nowFn}
}

func (s *Service) checkAllowed(ctx context.Context, userID uint, insightType string) error {
	planID, err := s.plans.LookupPlanID(ctx, userID)
	if err != nil {
		return err
	}
	for _, t := range s.plans.InsightsForPlan(planID) {
		if t == insightType {
			return nil
		}
	}
	return ErrProRequired
}

// ListUnused — промпты, которые не использовались >= 30 дней. limit clamp [1,100].
func (s *Service) ListUnused(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	if err := s.checkAllowed(ctx, userID, insightTypeUnused); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 50, 100)
	before := s.nowFn().AddDate(0, 0, -30)
	raws, err := s.analytics.UnusedPrompts(ctx, userID, teamID, before, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PromptInsightRow, 0, len(raws))
	for _, r := range raws {
		out = append(out, PromptInsightRow{PromptID: r.PromptID, Title: r.Title, Uses: int(r.Uses)})
	}
	return out, nil
}

func clampLimit(v, def, max int) int {
	if v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
}
```

Создать helper `backend/internal/usecases/prompt_insights/fakes_test.go` с stub методами `fakeAnalyticsRepo` для всех методов `repo.AnalyticsRepository` (возвращают nil/zero):

```go
package prompt_insights

import (
	"context"
	"time"

	repo "promptvault/internal/interface/repository"
)

// Все методы repo.AnalyticsRepository кроме UnusedPrompts — stub'ы для unit
// тестов prompt_insights usecase. Если интерфейс изменится — компилятор
// подскажет какие методы пропали.

func (f *fakeAnalyticsRepo) GetTrendingPrompts(ctx context.Context, userID uint, teamID *uint, factor float64, growing bool, limit int) ([]repo.TrendRow, error) {
	return nil, nil
}
func (f *fakeAnalyticsRepo) PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]repo.DuplicatePair, error) {
	return nil, nil
}
func (f *fakeAnalyticsRepo) MostEditedPrompts(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}
func (f *fakeAnalyticsRepo) OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error) {
	return nil, nil
}
func (f *fakeAnalyticsRepo) EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.CollectionRow, error) {
	return nil, nil
}
// + любые другие методы AnalyticsRepository (UsagePerDay, TopPrompts, и т.д.):
// возвращать nil/empty. Запустить `go vet`, компилятор укажет недостающие.
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/usecases/prompt_insights/...
```

Expected: PASS оба теста.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/prompt_insights/
git commit -m "feat(prompt-insights): Service.ListUnused с plan-gating"
```

---

### Task B4: Service.ListDuplicates

**Files:**
- Modify: `backend/internal/usecases/prompt_insights/service.go`
- Modify: `backend/internal/usecases/prompt_insights/service_test.go`

- [ ] **Step 1: Write the failing test**

Добавить в `service_test.go`:

```go
func TestListDuplicatesProTeaser(t *testing.T) {
	ar := &fakeAnalyticsRepo{}
	// fakeAnalyticsRepo нужно расширить полем duplicates + override PossibleDuplicates
	ar.duplicates = []repo.DuplicatePair{
		{PromptAID: 1, PromptATitle: "A1", PromptBID: 2, PromptBTitle: "A2", Similarity: 0.91},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "pro"}, time.Now)
	pairs, err := svc.ListDuplicates(context.Background(), 100, nil, 20)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].PromptA.PromptID != 1 || pairs[0].PromptB.PromptID != 2 {
		t.Fatalf("pair mismatch: %+v", pairs[0])
	}
	if pairs[0].Similarity < 0.9 || pairs[0].Similarity > 0.95 {
		t.Fatalf("similarity mismatch: %v", pairs[0].Similarity)
	}
}
```

Расширить `fakeAnalyticsRepo`:

```go
type fakeAnalyticsRepo struct {
	unused     []repo.PromptUsageRow
	duplicates []repo.DuplicatePair
	// ... (Trend, MostEdited, etc — добавятся в следующих tasks)
}

func (f *fakeAnalyticsRepo) PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]repo.DuplicatePair, error) {
	return f.duplicates, nil
}
```

(Удалить дублирующее объявление из fakes_test.go.)

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestListDuplicatesProTeaser ./internal/usecases/prompt_insights/...
```

Expected: FAIL — `ListDuplicates` undefined.

- [ ] **Step 3: Implement ListDuplicates**

Добавить в `service.go`:

```go
const duplicateSimilarityThreshold = 0.85

// ListDuplicates — пары похожих по pg_trgm промптов. threshold = 0.85 (consistent
// с InsightsCompute из analytics service).
func (s *Service) ListDuplicates(ctx context.Context, userID uint, teamID *uint, limit int) ([]DuplicatePair, error) {
	if err := s.checkAllowed(ctx, userID, insightTypeDuplicates); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 20, 50)
	raws, err := s.analytics.PossibleDuplicates(ctx, userID, teamID, duplicateSimilarityThreshold, limit)
	if err != nil {
		return nil, err
	}
	out := make([]DuplicatePair, 0, len(raws))
	for _, r := range raws {
		out = append(out, DuplicatePair{
			PromptA:    PromptInsightRow{PromptID: r.PromptAID, Title: r.PromptATitle},
			PromptB:    PromptInsightRow{PromptID: r.PromptBID, Title: r.PromptBTitle},
			Similarity: float64(r.Similarity),
		})
	}
	return out, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/usecases/prompt_insights/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/prompt_insights/
git commit -m "feat(prompt-insights): Service.ListDuplicates через PossibleDuplicates SQL"
```

---

### Task B5: Service.ListTrending + ListDeclining + ListMostEdited

**Files:**
- Modify: `backend/internal/usecases/prompt_insights/service.go`
- Modify: `backend/internal/usecases/prompt_insights/service_test.go`

- [ ] **Step 1: Write the failing tests**

Добавить в test-файл:

```go
func TestListTrending(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		trending: []repo.TrendRow{
			{PromptID: 5, Title: "Hot", UsesLast: 20, UsesPrevious: 5},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)
	rows, err := svc.ListTrending(context.Background(), 100, nil, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 || rows[0].Uses != 20 {
		t.Fatalf("expected 1 row with uses=20 (UsesLast), got %+v", rows)
	}
}

func TestListDeclining(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		declining: []repo.TrendRow{
			{PromptID: 7, Title: "Falling", UsesLast: 2, UsesPrevious: 18},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)
	rows, err := svc.ListDeclining(context.Background(), 100, nil, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 || rows[0].Uses != 2 {
		t.Fatalf("expected 1 row with uses=2, got %+v", rows)
	}
}

func TestListMostEdited(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		mostEdited: []repo.PromptUsageRow{
			{PromptID: 8, Title: "Churn", Uses: 15},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)
	rows, err := svc.ListMostEdited(context.Background(), 100, nil, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 || rows[0].Uses != 15 {
		t.Fatalf("expected 1 row, got %+v", rows)
	}
}
```

Расширить `fakeAnalyticsRepo`:

```go
type fakeAnalyticsRepo struct {
	unused     []repo.PromptUsageRow
	duplicates []repo.DuplicatePair
	trending   []repo.TrendRow
	declining  []repo.TrendRow
	mostEdited []repo.PromptUsageRow
}

func (f *fakeAnalyticsRepo) GetTrendingPrompts(ctx context.Context, userID uint, teamID *uint, factor float64, growing bool, limit int) ([]repo.TrendRow, error) {
	if growing {
		return f.trending, nil
	}
	return f.declining, nil
}

func (f *fakeAnalyticsRepo) MostEditedPrompts(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.PromptUsageRow, error) {
	return f.mostEdited, nil
}
```

(Удалить дублирующие stub-методы из fakes_test.go.)

- [ ] **Step 2: Run tests to verify failure**

```bash
go test -run "TestListTrending|TestListDeclining|TestListMostEdited" ./internal/usecases/prompt_insights/...
```

Expected: FAIL — методы не определены.

- [ ] **Step 3: Implement methods**

В `service.go` добавить:

```go
// ListTrending — промпты с ростом использования >2× за неделю (factor=2.0, growing=true).
func (s *Service) ListTrending(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	return s.listTrendDirection(ctx, userID, teamID, insightTypeTrending, 2.0, true, limit)
}

// ListDeclining — промпты с падением >2× (factor=0.5 = «текущее ≤ половины предыдущего»).
func (s *Service) ListDeclining(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	return s.listTrendDirection(ctx, userID, teamID, insightTypeDeclining, 0.5, false, limit)
}

func (s *Service) listTrendDirection(ctx context.Context, userID uint, teamID *uint, kind string, factor float64, growing bool, limit int) ([]PromptInsightRow, error) {
	if err := s.checkAllowed(ctx, userID, kind); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 10, 50)
	raws, err := s.analytics.GetTrendingPrompts(ctx, userID, teamID, factor, growing, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PromptInsightRow, 0, len(raws))
	for _, r := range raws {
		out = append(out, PromptInsightRow{PromptID: r.PromptID, Title: r.Title, Uses: int(r.UsesLast)})
	}
	return out, nil
}

// ListMostEdited — промпты с >=2 версиями (HAVING COUNT > 1 в SQL).
func (s *Service) ListMostEdited(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	if err := s.checkAllowed(ctx, userID, insightTypeMostEdited); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 10, 50)
	raws, err := s.analytics.MostEditedPrompts(ctx, userID, teamID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PromptInsightRow, 0, len(raws))
	for _, r := range raws {
		out = append(out, PromptInsightRow{PromptID: r.PromptID, Title: r.Title, Uses: int(r.Uses)})
	}
	return out, nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/usecases/prompt_insights/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/prompt_insights/
git commit -m "feat(prompt-insights): Service.ListTrending/Declining/MostEdited"
```

---

### Task B6: prompt.MergeWith — repository interface + implementation + integration test

**Files:**
- Modify: `backend/internal/interface/repository/prompt.go`
- Modify: `backend/internal/infrastructure/postgres/repository/prompt_repo.go`
- Create: `backend/internal/infrastructure/postgres/repository/prompt_merge_test.go`

- [ ] **Step 1: Write the failing integration test**

```go
package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"promptvault/internal/infrastructure/postgres/repository"
	"promptvault/internal/models"
	// testhelpers — содержит testcontainers PG setup (паттерн из других *_test.go в repo пакете)
)

func TestPromptMergeWithSoftDeletesTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	db, cleanup := testhelpers.SetupPG(t)
	defer cleanup()

	user := testhelpers.CreateUser(t, db, "merge-test@local")

	keep := &models.Prompt{Title: "Keep", Content: "A", UserID: user.ID}
	merge := &models.Prompt{Title: "Merge", Content: "B", UserID: user.ID}
	require.NoError(t, db.Create(keep).Error)
	require.NoError(t, db.Create(merge).Error)

	r := repository.NewPromptRepository(db)
	err := r.MergeWith(context.Background(), keep.ID, merge.ID, user.ID)
	require.NoError(t, err)

	// keep still active
	var keepAfter models.Prompt
	require.NoError(t, db.First(&keepAfter, keep.ID).Error)
	require.True(t, keepAfter.DeletedAt.IsZero() || !keepAfter.DeletedAt.Valid)

	// merge soft-deleted
	var mergeAfter models.Prompt
	require.NoError(t, db.Unscoped().First(&mergeAfter, merge.ID).Error)
	require.True(t, mergeAfter.DeletedAt.Valid, "merge target should be soft-deleted")
	require.WithinDuration(t, time.Now(), mergeAfter.DeletedAt.Time, 5*time.Second)
}

func TestPromptMergeWithOwnershipError(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	db, cleanup := testhelpers.SetupPG(t)
	defer cleanup()

	user1 := testhelpers.CreateUser(t, db, "u1@local")
	user2 := testhelpers.CreateUser(t, db, "u2@local")
	p1 := &models.Prompt{Title: "P1", UserID: user1.ID}
	p2 := &models.Prompt{Title: "P2", UserID: user2.ID}
	require.NoError(t, db.Create(p1).Error)
	require.NoError(t, db.Create(p2).Error)

	r := repository.NewPromptRepository(db)
	err := r.MergeWith(context.Background(), p1.ID, p2.ID, user1.ID)
	require.Error(t, err) // ownership mismatch
}
```

(Если `testhelpers` package не существует в проекте — использовать паттерн из существующих `*_repo_test.go` файлов в `infrastructure/postgres/repository/` для setup testcontainers. Скопировать helpers напрямую в test-файл если нужно.)

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestPromptMergeWith ./internal/infrastructure/postgres/repository/...
```

Expected: FAIL — `MergeWith` undefined.

- [ ] **Step 3: Add interface method**

В `backend/internal/interface/repository/prompt.go` добавить в `PromptRepository` interface:

```go
// MergeWith soft-deletes mergeID после проверки что оба промпта принадлежат
// userID. Возвращает gorm.ErrRecordNotFound если любой prompt не найден или
// не принадлежит юзеру.
MergeWith(ctx context.Context, keepID, mergeID, userID uint) error
```

- [ ] **Step 4: Implement в prompt_repo.go**

```go
func (r *promptRepository) MergeWith(ctx context.Context, keepID, mergeID, userID uint) error {
	if keepID == mergeID {
		return errors.New("cannot merge prompt with itself")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		// Проверка: оба промпта существуют и принадлежат юзеру (deleted_at IS NULL).
		if err := tx.Model(&models.Prompt{}).
			Where("id IN ? AND user_id = ? AND deleted_at IS NULL", []uint{keepID, mergeID}, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count != 2 {
			return gorm.ErrRecordNotFound
		}
		// Soft-delete merge target. GORM автоматически выставляет deleted_at для
		// gorm.DeletedAt колонки на модели Prompt.
		return tx.Delete(&models.Prompt{}, mergeID).Error
	})
}
```

(При необходимости — добавить import `errors` и `gorm.io/gorm`.)

- [ ] **Step 5: Run integration test**

```bash
go test -run TestPromptMergeWith ./internal/infrastructure/postgres/repository/...
```

Expected: PASS (требуется Docker для testcontainers).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/interface/repository/prompt.go backend/internal/infrastructure/postgres/repository/prompt_repo.go backend/internal/infrastructure/postgres/repository/prompt_merge_test.go
git commit -m "feat(prompt): MergeWith soft-delete'ит merge target в транзакции"
```

---

### Task B7: prompt_insights.Service.MergePrompts

**Files:**
- Modify: `backend/internal/usecases/prompt_insights/service.go`
- Modify: `backend/internal/usecases/prompt_insights/service_test.go`

- [ ] **Step 1: Write the failing test**

```go
type fakePromptMerger struct {
	called   bool
	keepID   uint
	mergeID  uint
	userID   uint
	returns  error
}

func (f *fakePromptMerger) MergeWith(ctx context.Context, keepID, mergeID, userID uint) error {
	f.called = true
	f.keepID = keepID
	f.mergeID = mergeID
	f.userID = userID
	return f.returns
}

func TestMergePromptsHappy(t *testing.T) {
	m := &fakePromptMerger{}
	svc := NewService(&fakeAnalyticsRepo{}, m, &fakePlanLookup{plan: "max"}, time.Now)
	err := svc.MergePrompts(context.Background(), 100, 1, 2)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !m.called || m.keepID != 1 || m.mergeID != 2 || m.userID != 100 {
		t.Fatalf("merger not called correctly: %+v", m)
	}
}

func TestMergePromptsSameID(t *testing.T) {
	svc := NewService(&fakeAnalyticsRepo{}, &fakePromptMerger{}, &fakePlanLookup{plan: "max"}, time.Now)
	err := svc.MergePrompts(context.Background(), 100, 5, 5)
	if !errors.Is(err, ErrSamePrompt) {
		t.Fatalf("expected ErrSamePrompt, got %v", err)
	}
}

func TestMergePromptsNotOwned(t *testing.T) {
	m := &fakePromptMerger{returns: gorm.ErrRecordNotFound}
	svc := NewService(&fakeAnalyticsRepo{}, m, &fakePlanLookup{plan: "max"}, time.Now)
	err := svc.MergePrompts(context.Background(), 100, 1, 2)
	if !errors.Is(err, ErrPromptsNotOwned) {
		t.Fatalf("expected ErrPromptsNotOwned, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestMergePrompts ./internal/usecases/prompt_insights/...
```

Expected: FAIL — `MergePrompts` undefined.

- [ ] **Step 3: Implement MergePrompts**

```go
import (
	// ...
	"errors"
	"gorm.io/gorm"
)

// MergePrompts soft-delete'ит mergeID, сохраняя keepID. Merge — Pro-фича
// (gating через ListDuplicates path), но проверка по `insightTypeDuplicates`
// здесь не нужна — handler-level rate-limit + plan-check (Pro/Max) делает то же.
// Сейчас мы только проверяем ownership и same-id case.
func (s *Service) MergePrompts(ctx context.Context, userID, keepID, mergeID uint) error {
	if keepID == mergeID {
		return ErrSamePrompt
	}
	err := s.prompts.MergeWith(ctx, keepID, mergeID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrPromptsNotOwned
	}
	return err
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/usecases/prompt_insights/...
```

Expected: PASS все 3 теста (+ предыдущие).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/prompt_insights/
git commit -m "feat(prompt-insights): MergePrompts с ownership + same-id checks"
```

---

### Task B8: insights_handler.go — 5 GET endpoints + handler tests

**Files:**
- Create: `backend/internal/delivery/http/prompt/insights_handler.go`
- Create: `backend/internal/delivery/http/prompt/insights_handler_test.go`

- [ ] **Step 1: Write the failing test**

```go
package prompt_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"promptvault/internal/delivery/http/prompt"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/usecases/prompt_insights"
)

// fakeInsightsService реализует InsightsService с настраиваемыми return values.
type fakeInsightsService struct {
	unused      []prompt_insights.PromptInsightRow
	duplicates  []prompt_insights.DuplicatePair
	trending    []prompt_insights.PromptInsightRow
	declining   []prompt_insights.PromptInsightRow
	mostEdited  []prompt_insights.PromptInsightRow
	err         error
}

func (f *fakeInsightsService) ListUnused(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error) {
	return f.unused, f.err
}
// ... (аналогично для остальных 4 методов)

func TestInsightsHandlerUnused200(t *testing.T) {
	svc := &fakeInsightsService{unused: []prompt_insights.PromptInsightRow{{PromptID: 1, Title: "X", Uses: 0}}}
	h := prompt.NewInsightsHandler(svc)

	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/prompts/insights/unused", h.Unused)

	req := httptest.NewRequest("GET", "/api/prompts/insights/unused?limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []prompt_insights.PromptInsightRow `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 1)
	require.Equal(t, uint(1), body.Items[0].PromptID)
}

func TestInsightsHandlerUnused402WhenProRequired(t *testing.T) {
	svc := &fakeInsightsService{err: prompt_insights.ErrProRequired}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/prompts/insights/unused", h.Unused)
	req := httptest.NewRequest("GET", "/api/prompts/insights/unused", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusPaymentRequired, w.Code)
}

// injectUserID — middleware-stub: кладёт userID в ctx по ключу authmw.UserIDKey.
func injectUserID(uid uint) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := authmw.WithUserID(r.Context(), uid) // или ручной context.WithValue если WithUserID не существует
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

(Если `authmw.WithUserID` не существует — использовать `context.WithValue(r.Context(), authmw.UserIDKey, uid)`.)

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/delivery/http/prompt/...
```

Expected: FAIL — `InsightsHandler` undefined.

- [ ] **Step 3: Implement handler**

`insights_handler.go`:

```go
package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	httperr "promptvault/internal/delivery/http/errors"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/usecases/prompt_insights"
)

// InsightsService — узкий интерфейс на prompt_insights.Service.
type InsightsService interface {
	ListUnused(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	ListDuplicates(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.DuplicatePair, error)
	ListTrending(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	ListDeclining(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	ListMostEdited(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	MergePrompts(ctx context.Context, userID, keepID, mergeID uint) error
}

type InsightsHandler struct {
	svc InsightsService
}

func NewInsightsHandler(svc InsightsService) *InsightsHandler {
	return &InsightsHandler{svc: svc}
}

func (h *InsightsHandler) Unused(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, h.svc.ListUnused, "unused")
}
func (h *InsightsHandler) Trending(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, h.svc.ListTrending, "trending")
}
func (h *InsightsHandler) Declining(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, h.svc.ListDeclining, "declining")
}
func (h *InsightsHandler) MostEdited(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r, h.svc.ListMostEdited, "most_edited")
}

// handleList — общий код для 4 list-style endpoints. duplicates имеет другой
// return type (DuplicatePair), поэтому реализован отдельным методом.
func (h *InsightsHandler) handleList(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error), insightType string) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 50)

	rows, err := fn(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", insightType, "user_id", userID, "items_count", len(rows))
	writeItems(w, rows)
}

func (h *InsightsHandler) Duplicates(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 20)
	pairs, err := h.svc.ListDuplicates(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", "duplicates", "user_id", userID, "items_count", len(pairs))
	writeItems(w, pairs)
}

func parseTeamID(r *http.Request) *uint {
	q := r.URL.Query().Get("team_id")
	if q == "" {
		return nil
	}
	id, err := strconv.ParseUint(q, 10, 32)
	if err != nil {
		return nil
	}
	u := uint(id)
	return &u
}

func parseLimit(r *http.Request, def int) int {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return def
	}
	v, err := strconv.Atoi(q)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func writeItems[T any](w http.ResponseWriter, items []T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}

func respondInsightsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, prompt_insights.ErrProRequired):
		httperr.Respond(w, &httperr.AppError{Status: http.StatusPaymentRequired, Code: "pro_required", Message: "Подписка Pro обязательна для этого инсайта"})
	case errors.Is(err, prompt_insights.ErrPromptsNotOwned):
		httperr.Respond(w, httperr.NotFound("Промпт не найден или вам не принадлежит"))
	case errors.Is(err, prompt_insights.ErrSamePrompt):
		httperr.Respond(w, httperr.BadRequest("Нельзя объединить промпт сам с собой"))
	default:
		slog.Error("prompt_insights.internal", "err", err)
		httperr.Respond(w, httperr.Internal("Внутренняя ошибка"))
	}
}
```

(`httperr.AppError` / `BadRequest` / `NotFound` / `Internal` — паттерн из существующего `delivery/http/errors/errors.go`. Если конкретные конструкторы отличаются — подставить актуальные имена из этого пакета.)

- [ ] **Step 4: Run tests**

```bash
go test ./internal/delivery/http/prompt/...
```

Expected: PASS оба теста.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/prompt/insights_handler.go backend/internal/delivery/http/prompt/insights_handler_test.go
git commit -m "feat(prompt-insights): 5 GET endpoints (unused/duplicates/trending/declining/most-edited)"
```

---

### Task B9: insights_handler.go — Merge POST endpoint

**Files:**
- Modify: `backend/internal/delivery/http/prompt/insights_handler.go`
- Modify: `backend/internal/delivery/http/prompt/insights_handler_test.go`

- [ ] **Step 1: Write the failing test**

Добавить в `insights_handler_test.go`:

```go
func TestInsightsHandlerMergeHappy(t *testing.T) {
	svc := &fakeInsightsService{} // .MergePrompts вернёт nil по дефолту
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Post("/api/prompts/{id}/merge-with/{other_id}", h.Merge)

	req := httptest.NewRequest("POST", "/api/prompts/1/merge-with/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		KeptID   uint `json:"kept_id"`
		MergedID uint `json:"merged_id"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, uint(1), body.KeptID)
	require.Equal(t, uint(2), body.MergedID)
}

func TestInsightsHandlerMergeNotOwned404(t *testing.T) {
	svc := &fakeInsightsService{mergeErr: prompt_insights.ErrPromptsNotOwned}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Post("/api/prompts/{id}/merge-with/{other_id}", h.Merge)
	req := httptest.NewRequest("POST", "/api/prompts/1/merge-with/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}
```

Расширить `fakeInsightsService`: добавить поле `mergeErr error` и метод:

```go
func (f *fakeInsightsService) MergePrompts(ctx context.Context, userID, keepID, mergeID uint) error {
	return f.mergeErr
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestInsightsHandlerMerge ./internal/delivery/http/prompt/...
```

Expected: FAIL — `Merge` undefined.

- [ ] **Step 3: Implement Merge handler**

В `insights_handler.go`:

```go
import (
	// ...
	"github.com/go-chi/chi/v5"
)

func (h *InsightsHandler) Merge(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	keepID, err := parsePathUint(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный id"))
		return
	}
	mergeID, err := parsePathUint(r, "other_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный other_id"))
		return
	}
	if err := h.svc.MergePrompts(r.Context(), userID, keepID, mergeID); err != nil {
		respondInsightsError(w, err)
		return
	}
	slog.Info("prompt_insights.merge", "user_id", userID, "kept_id", keepID, "merged_id", mergeID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]uint{"kept_id": keepID, "merged_id": mergeID})
}

func parsePathUint(r *http.Request, name string) (uint, error) {
	s := chi.URLParam(r, name)
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/delivery/http/prompt/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/prompt/insights_handler.go backend/internal/delivery/http/prompt/insights_handler_test.go
git commit -m "feat(prompt-insights): POST /merge-with handler с ownership errors"
```

---

### Task B10: tag/orphan_handler.go — GET /api/tags/orphan

**Files:**
- Create: `backend/internal/delivery/http/tag/orphan_handler.go`
- Create: `backend/internal/delivery/http/tag/orphan_handler_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tag_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"promptvault/internal/delivery/http/tag"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
)

type fakeAnalyticsRepo struct {
	tags []repo.TagRow
	err  error
}

func (f *fakeAnalyticsRepo) OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error) {
	return f.tags, f.err
}
// ... другие методы AnalyticsRepository — stubs nil/empty

func TestOrphanHandlerHappy(t *testing.T) {
	ar := &fakeAnalyticsRepo{tags: []repo.TagRow{{TagID: 1, Name: "deprecated"}}}
	h := tag.NewOrphanHandler(ar)

	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/tags/orphan", h.List)

	req := httptest.NewRequest("GET", "/api/tags/orphan", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 1)
	require.Equal(t, "deprecated", body.Items[0].Name)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/delivery/http/tag/...
```

Expected: FAIL — `OrphanHandler` undefined.

- [ ] **Step 3: Implement handler**

`orphan_handler.go`:

```go
package tag

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	authmw "promptvault/internal/middleware/auth"
	repo "promptvault/internal/interface/repository"
)

// OrphanAnalyticsRepo — узкий интерфейс на AnalyticsRepository.OrphanTags.
type OrphanAnalyticsRepo interface {
	OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error)
}

type OrphanHandler struct {
	analytics OrphanAnalyticsRepo
}

func NewOrphanHandler(analytics OrphanAnalyticsRepo) *OrphanHandler {
	return &OrphanHandler{analytics: analytics}
}

// List — GET /api/tags/orphan. Возвращает теги юзера без активных промптов.
// Limit hard-cap 100 (orphan'ов обычно мало, не нужна пагинация).
func (h *OrphanHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	tags, err := h.analytics.OrphanTags(r.Context(), userID, nil, 100)
	if err != nil {
		slog.Error("tag_orphan.failed", "err", err, "user_id", userID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type item struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	items := make([]item, 0, len(tags))
	for _, t := range tags {
		items = append(items, item{ID: t.TagID, Name: t.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/delivery/http/tag/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/tag/orphan_handler.go backend/internal/delivery/http/tag/orphan_handler_test.go
git commit -m "feat(tag): GET /api/tags/orphan возвращает теги без активных промптов"
```

---

### Task B11: collection/empty_handler.go — GET /api/collections/empty

**Files:**
- Create: `backend/internal/delivery/http/collection/empty_handler.go`
- Create: `backend/internal/delivery/http/collection/empty_handler_test.go`

- [ ] **Step 1: Write the failing test**

Аналогично Task B10, но для `EmptyCollections` SQL и `CollectionRow`:

```go
func TestEmptyHandlerHappy(t *testing.T) {
	ar := &fakeAnalyticsRepo{collections: []repo.CollectionRow{{CollectionID: 9, Name: "Old"}}}
	h := collection.NewEmptyHandler(ar)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/collections/empty", h.List)

	req := httptest.NewRequest("GET", "/api/collections/empty", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 1)
	require.Equal(t, "Old", body.Items[0].Name)
}
```

`fakeAnalyticsRepo.EmptyCollections` метод аналогичен `OrphanTags` из B10.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/delivery/http/collection/...
```

Expected: FAIL.

- [ ] **Step 3: Implement handler**

`empty_handler.go`:

```go
package collection

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	authmw "promptvault/internal/middleware/auth"
	repo "promptvault/internal/interface/repository"
)

type EmptyAnalyticsRepo interface {
	EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.CollectionRow, error)
}

type EmptyHandler struct {
	analytics EmptyAnalyticsRepo
}

func NewEmptyHandler(analytics EmptyAnalyticsRepo) *EmptyHandler {
	return &EmptyHandler{analytics: analytics}
}

func (h *EmptyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	rows, err := h.analytics.EmptyCollections(r.Context(), userID, nil, 100)
	if err != nil {
		slog.Error("collection_empty.failed", "err", err, "user_id", userID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type item struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	items := make([]item, 0, len(rows))
	for _, r := range rows {
		items = append(items, item{ID: r.CollectionID, Name: r.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/delivery/http/collection/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/collection/empty_handler.go backend/internal/delivery/http/collection/empty_handler_test.go
git commit -m "feat(collection): GET /api/collections/empty"
```

---

### Task B12: Wire-up в app.go + routes.go

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/app/routes.go`

- [ ] **Step 1: Sanity check — поднимем сервер локально и проверим что endpoint вернёт 401 без auth (sanity smoke)**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault
docker compose -f docker-compose.dev.yml up -d --build api
curl -i http://localhost:8080/api/prompts/insights/unused
```

Expected: 404 (route не зарегистрирован). Это baseline для test ↓.

- [ ] **Step 2: Add wire-up в app.go**

Найти место где собираются handlers (поиск по `promptHandler :=`) и добавить:

```go
// Prompt insights (Wave 1).
promptInsightsSvc := prompt_insights.NewService(
	a.analyticsRepo,
	a.promptsRepo,
	promptInsightsPlanGate{a.analyticsSvc, a.usersRepo}, // adapter, см. ниже
	nil,
)
a.promptInsightsHandler = prompt.NewInsightsHandler(promptInsightsSvc)
a.tagOrphanHandler = tag.NewOrphanHandler(a.analyticsRepo)
a.collectionEmptyHandler = collection.NewEmptyHandler(a.analyticsRepo)
```

Добавить adapter в новый файл `backend/internal/app/insights_adapter.go`:

```go
package app

import (
	"context"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/usecases/analytics"
)

// promptInsightsPlanGate адаптирует analytics.Service + users.Repository к
// PlanGate интерфейсу из prompt_insights пакета.
type promptInsightsPlanGate struct {
	analytics *analytics.Service
	users     repo.UserRepository
}

func (g promptInsightsPlanGate) InsightsForPlan(planID string) []string {
	return g.analytics.InsightsForPlan(planID)
}

func (g promptInsightsPlanGate) LookupPlanID(ctx context.Context, userID uint) (string, error) {
	u, err := g.users.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return u.PlanID, nil
}
```

Поля в `App` struct (`app.go`):

```go
type App struct {
	// ... existing fields ...
	promptInsightsHandler  *prompt.InsightsHandler
	tagOrphanHandler       *tag.OrphanHandler
	collectionEmptyHandler *collection.EmptyHandler
}
```

- [ ] **Step 3: Register routes**

В `routes.go` найти блок `r.Route("/prompts", ...)` (около L341) и добавить **внутри блока** перед `r.Get("/{id}", ...)`:

```go
r.Route("/insights", func(r chi.Router) {
	r.Get("/unused", a.promptInsightsHandler.Unused)
	r.Get("/duplicates", a.promptInsightsHandler.Duplicates)
	r.Get("/trending", a.promptInsightsHandler.Trending)
	r.Get("/declining", a.promptInsightsHandler.Declining)
	r.Get("/most-edited", a.promptInsightsHandler.MostEdited)
})
r.Post("/{id}/merge-with/{other_id}", a.promptInsightsHandler.Merge)
```

После блока `/prompts` (или в соответствующих местах) добавить:

```go
r.Get("/tags/orphan", a.tagOrphanHandler.List)
r.Get("/collections/empty", a.collectionEmptyHandler.List)
```

(Если `/tags` или `/collections` уже зарегистрированы как `r.Route(...)` — добавлять внутри их блока как `r.Get("/orphan", ...)`.)

- [ ] **Step 4: Rebuild + smoke**

```bash
docker compose -f docker-compose.dev.yml up -d --build api
sleep 5
# Без auth → 401 (или 200, в зависимости от middleware на /api):
curl -i http://localhost:8080/api/prompts/insights/unused
```

Expected: 401 Unauthorized (route зарегистрирован, auth middleware блокирует).

```bash
go test -short ./...
```

Expected: PASS все тесты.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/app/
git commit -m "feat(app): wire prompt_insights + tag-orphan + collection-empty routes"
```

---

## Phase 2 — Wave 1: Frontend deep linking

### Task F1: api/prompt-insights.ts — fetchers + types

**Files:**
- Create: `frontend/src/api/prompt-insights.ts`
- Create: `frontend/src/api/prompt-insights.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, it, expect, vi, beforeEach } from "vitest"
import { fetchUnused, fetchDuplicates, fetchTrending, fetchDeclining, fetchMostEdited, mergePrompts } from "./prompt-insights"

const mockFetch = vi.fn()
beforeEach(() => {
  mockFetch.mockReset()
  globalThis.fetch = mockFetch as unknown as typeof fetch
})

describe("prompt-insights api", () => {
  it("fetches unused", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ items: [{ prompt_id: 1, title: "X", uses: 0 }] }),
    })
    const items = await fetchUnused()
    expect(items).toHaveLength(1)
    expect(items[0]).toEqual({ prompt_id: 1, title: "X", uses: 0 })
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/prompts/insights/unused"),
      expect.objectContaining({ credentials: "include" }),
    )
  })

  it("fetches duplicates with pairs", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        items: [{ prompt_a: { prompt_id: 1, title: "A", uses: 0 }, prompt_b: { prompt_id: 2, title: "B", uses: 0 }, similarity: 0.9 }],
      }),
    })
    const pairs = await fetchDuplicates()
    expect(pairs).toHaveLength(1)
    expect(pairs[0].similarity).toBe(0.9)
  })

  it("throws on 402", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 402, json: async () => ({ error: "pro_required" }) })
    await expect(fetchUnused()).rejects.toThrow(/pro_required|402/i)
  })

  it("merges prompts", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ kept_id: 1, merged_id: 2 }),
    })
    const res = await mergePrompts(1, 2)
    expect(res).toEqual({ kept_id: 1, merged_id: 2 })
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/prompts/1/merge-with/2"),
      expect.objectContaining({ method: "POST", credentials: "include" }),
    )
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend
npx vitest run prompt-insights
```

Expected: FAIL — module not found.

- [ ] **Step 3: Implement fetchers**

`prompt-insights.ts`:

```ts
import { apiUrl } from "@/lib/api-url" // существующий helper, добавляет VITE_API_URL prefix

export interface PromptInsightRow {
  prompt_id: number
  title: string
  uses: number
  updated_at?: string
}

export interface DuplicatePair {
  prompt_a: PromptInsightRow
  prompt_b: PromptInsightRow
  similarity: number
}

interface ItemsEnvelope<T> {
  items: T[]
}

async function getItems<T>(path: string): Promise<T[]> {
  const res = await fetch(apiUrl(path), { credentials: "include" })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  const json = (await res.json()) as ItemsEnvelope<T>
  return json.items ?? []
}

export const fetchUnused = () => getItems<PromptInsightRow>("/api/prompts/insights/unused")
export const fetchDuplicates = () => getItems<DuplicatePair>("/api/prompts/insights/duplicates")
export const fetchTrending = () => getItems<PromptInsightRow>("/api/prompts/insights/trending")
export const fetchDeclining = () => getItems<PromptInsightRow>("/api/prompts/insights/declining")
export const fetchMostEdited = () => getItems<PromptInsightRow>("/api/prompts/insights/most-edited")

export async function mergePrompts(keepID: number, mergeID: number): Promise<{ kept_id: number; merged_id: number }> {
  const res = await fetch(apiUrl(`/api/prompts/${keepID}/merge-with/${mergeID}`), {
    method: "POST",
    credentials: "include",
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  return res.json()
}
```

(Если `apiUrl` helper не существует — использовать `${import.meta.env.VITE_API_URL || ""}` prefix. Уточнить через grep `apiUrl|VITE_API_URL` по существующим api/*.ts файлам.)

- [ ] **Step 4: Run tests**

```bash
npx vitest run prompt-insights
```

Expected: PASS все 4 теста.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/prompt-insights.ts frontend/src/api/prompt-insights.test.ts
git commit -m "feat(api): prompt-insights fetchers + merge mutation"
```

---

### Task F2: hooks/use-prompt-insights.ts

**Files:**
- Create: `frontend/src/hooks/use-prompt-insights.ts`
- Create: `frontend/src/hooks/use-prompt-insights.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useUnusedPrompts, useDuplicates, useMergePrompts } from "./use-prompt-insights"
import * as api from "@/api/prompt-insights"

vi.mock("@/api/prompt-insights")

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>
}

describe("use-prompt-insights", () => {
  beforeEach(() => vi.resetAllMocks())

  it("useUnusedPrompts returns data", async () => {
    vi.mocked(api.fetchUnused).mockResolvedValue([{ prompt_id: 1, title: "X", uses: 0 }])
    const { result } = renderHook(() => useUnusedPrompts(), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toHaveLength(1)
  })

  it("useDuplicates returns pairs", async () => {
    vi.mocked(api.fetchDuplicates).mockResolvedValue([
      { prompt_a: { prompt_id: 1, title: "A", uses: 0 }, prompt_b: { prompt_id: 2, title: "B", uses: 0 }, similarity: 0.9 },
    ])
    const { result } = renderHook(() => useDuplicates(), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.[0].similarity).toBe(0.9)
  })

  it("useMergePrompts triggers mutation", async () => {
    vi.mocked(api.mergePrompts).mockResolvedValue({ kept_id: 1, merged_id: 2 })
    const { result } = renderHook(() => useMergePrompts(), { wrapper })
    await result.current.mutateAsync({ keepID: 1, mergeID: 2 })
    expect(api.mergePrompts).toHaveBeenCalledWith(1, 2)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run use-prompt-insights
```

Expected: FAIL — hook module missing.

- [ ] **Step 3: Implement hooks**

`use-prompt-insights.ts`:

```ts
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  fetchUnused,
  fetchDuplicates,
  fetchTrending,
  fetchDeclining,
  fetchMostEdited,
  mergePrompts,
} from "@/api/prompt-insights"

export function useUnusedPrompts() {
  return useQuery({ queryKey: ["prompt-insights", "unused"], queryFn: fetchUnused })
}

export function useDuplicates() {
  return useQuery({ queryKey: ["prompt-insights", "duplicates"], queryFn: fetchDuplicates })
}

export function useTrending() {
  return useQuery({ queryKey: ["prompt-insights", "trending"], queryFn: fetchTrending })
}

export function useDeclining() {
  return useQuery({ queryKey: ["prompt-insights", "declining"], queryFn: fetchDeclining })
}

export function useMostEdited() {
  return useQuery({ queryKey: ["prompt-insights", "most-edited"], queryFn: fetchMostEdited })
}

export function useMergePrompts() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ keepID, mergeID }: { keepID: number; mergeID: number }) => mergePrompts(keepID, mergeID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["prompt-insights", "duplicates"] })
      qc.invalidateQueries({ queryKey: ["prompts"] })
    },
  })
}
```

- [ ] **Step 4: Run tests**

```bash
npx vitest run use-prompt-insights
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/hooks/use-prompt-insights.ts frontend/src/hooks/use-prompt-insights.test.ts
git commit -m "feat(hooks): use-prompt-insights с 5 queries + merge mutation"
```

---

### Task F3: components/prompts/insights/insight-prompt-row.tsx

**Files:**
- Create: `frontend/src/components/prompts/insights/insight-prompt-row.tsx`
- Create: `frontend/src/components/prompts/insights/insight-prompt-row.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { InsightPromptRow } from "./insight-prompt-row"

function r(node: React.ReactNode) {
  return render(<MemoryRouter>{node}</MemoryRouter>)
}

describe("InsightPromptRow", () => {
  it("renders title and uses", () => {
    r(<InsightPromptRow promptID={1} title="Refactor X" uses={12} />)
    expect(screen.getByText("Refactor X")).toBeInTheDocument()
    expect(screen.getByText(/12/)).toBeInTheDocument()
  })

  it("renders link to prompt editor", () => {
    r(<InsightPromptRow promptID={42} title="X" uses={0} />)
    const link = screen.getByRole("link", { name: /X/ })
    expect(link).toHaveAttribute("href", "/prompts/42")
  })

  it("renders action slot", () => {
    r(<InsightPromptRow promptID={1} title="X" uses={0} actions={<button>Удалить</button>} />)
    expect(screen.getByRole("button", { name: "Удалить" })).toBeInTheDocument()
  })

  it("hides uses when usesLabel=false", () => {
    r(<InsightPromptRow promptID={1} title="X" uses={5} showUses={false} />)
    expect(screen.queryByText(/использований|использован/i)).not.toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run insight-prompt-row
```

Expected: FAIL — component missing.

- [ ] **Step 3: Implement component**

`insight-prompt-row.tsx`:

```tsx
import { Link } from "react-router-dom"

interface InsightPromptRowProps {
  promptID: number
  title: string
  uses: number
  showUses?: boolean
  actions?: React.ReactNode
}

export function InsightPromptRow({ promptID, title, uses, showUses = true, actions }: InsightPromptRowProps) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md border px-3 py-2">
      <div className="min-w-0 flex-1">
        <Link to={`/prompts/${promptID}`} className="block truncate text-sm font-medium hover:underline">
          {title}
        </Link>
        {showUses && (
          <p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
            {usesLabel(uses)}
          </p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}

function usesLabel(n: number): string {
  if (n === 0) return "0 использований"
  if (n === 1) return "1 использование"
  if (n >= 2 && n <= 4) return `${n} использования`
  return `${n} использований`
}
```

- [ ] **Step 4: Run tests**

```bash
npx vitest run insight-prompt-row
```

Expected: PASS все 4 теста.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/prompts/insights/insight-prompt-row.tsx frontend/src/components/prompts/insights/insight-prompt-row.test.tsx
git commit -m "feat(insights): reusable InsightPromptRow с pluralized uses label"
```

---

### Task F4: pages/prompts/insights/unused.tsx

**Files:**
- Create: `frontend/src/pages/prompts/insights/unused.tsx`
- Create: `frontend/src/pages/prompts/insights/unused.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import UnusedInsightsPage from "./unused"
import * as hooks from "@/hooks/use-prompt-insights"

vi.mock("@/hooks/use-prompt-insights")

function wrap(node: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("UnusedInsightsPage", () => {
  it("renders heading and items", () => {
    vi.mocked(hooks.useUnusedPrompts).mockReturnValue({
      data: [{ prompt_id: 1, title: "Old", uses: 0 }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof hooks.useUnusedPrompts>)
    render(wrap(<UnusedInsightsPage />))
    expect(screen.getByRole("heading", { name: /забытые промпты/i })).toBeInTheDocument()
    expect(screen.getByText("Old")).toBeInTheDocument()
  })

  it("renders empty state", () => {
    vi.mocked(hooks.useUnusedPrompts).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof hooks.useUnusedPrompts>)
    render(wrap(<UnusedInsightsPage />))
    expect(screen.getByText(/нет забытых/i)).toBeInTheDocument()
  })

  it("renders loading state", () => {
    vi.mocked(hooks.useUnusedPrompts).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as ReturnType<typeof hooks.useUnusedPrompts>)
    render(wrap(<UnusedInsightsPage />))
    expect(screen.getByText(/загруж/i)).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run pages/prompts/insights/unused
```

Expected: FAIL.

- [ ] **Step 3: Implement page**

```tsx
import { AlertCircle } from "lucide-react"
import { Link } from "react-router-dom"
import { InsightPromptRow } from "@/components/prompts/insights/insight-prompt-row"
import { useUnusedPrompts } from "@/hooks/use-prompt-insights"
import { useDeletePrompt } from "@/hooks/use-prompts"
import { Button } from "@/components/ui/button"

export default function UnusedInsightsPage() {
  const { data, isLoading, isError } = useUnusedPrompts()
  const deletePrompt = useDeletePrompt()

  return (
    <div className="mx-auto max-w-3xl space-y-4 p-6">
      <header>
        <h1 className="flex items-center gap-2 text-2xl font-semibold">
          <AlertCircle className="h-5 w-5 text-amber-500" />
          Забытые промпты
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Промпты, которые вы не использовали 30+ дней. Подумайте о том, чтобы удалить или обновить.
        </p>
      </header>

      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && <p className="text-sm text-destructive">Не удалось загрузить список.</p>}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">Нет забытых промптов — всё используется.</p>
      )}

      <ul className="space-y-2">
        {data?.map((p) => (
          <li key={p.prompt_id}>
            <InsightPromptRow
              promptID={p.prompt_id}
              title={p.title}
              uses={p.uses}
              actions={
                <>
                  <Button asChild variant="ghost" size="sm">
                    <Link to={`/prompts/${p.prompt_id}`}>Открыть</Link>
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      if (confirm(`Удалить «${p.title}» в корзину?`)) {
                        deletePrompt.mutate(p.prompt_id)
                      }
                    }}
                  >
                    Удалить
                  </Button>
                </>
              }
            />
          </li>
        ))}
      </ul>
    </div>
  )
}
```

- [ ] **Step 4: Run tests**

```bash
npx vitest run pages/prompts/insights/unused
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/prompts/insights/unused.tsx frontend/src/pages/prompts/insights/unused.test.tsx
git commit -m "feat(insights): /prompts/insights/unused page с delete action"
```

---

### Task F5: pages/prompts/insights/duplicates.tsx + merge-modal.tsx

**Files:**
- Create: `frontend/src/components/prompts/insights/merge-modal.tsx`
- Create: `frontend/src/components/prompts/insights/merge-modal.test.tsx`
- Create: `frontend/src/pages/prompts/insights/duplicates.tsx`
- Create: `frontend/src/pages/prompts/insights/duplicates.test.tsx`

- [ ] **Step 1: Write the failing test (modal)**

```tsx
import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { MergeModal } from "./merge-modal"

const pair = {
  prompt_a: { prompt_id: 1, title: "Refactor v1", uses: 5 },
  prompt_b: { prompt_id: 2, title: "Refactor v2", uses: 10 },
  similarity: 0.91,
}

describe("MergeModal", () => {
  it("renders both prompts side-by-side when open", () => {
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={() => {}} />)
    expect(screen.getByText("Refactor v1")).toBeInTheDocument()
    expect(screen.getByText("Refactor v2")).toBeInTheDocument()
  })

  it("calls onMerge with correct ids when user picks A", () => {
    const onMerge = vi.fn()
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={onMerge} />)
    fireEvent.click(screen.getByRole("button", { name: /оставить «refactor v1»/i }))
    expect(onMerge).toHaveBeenCalledWith({ keepID: 1, mergeID: 2 })
  })

  it("calls onMerge with reversed ids when user picks B", () => {
    const onMerge = vi.fn()
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={onMerge} />)
    fireEvent.click(screen.getByRole("button", { name: /оставить «refactor v2»/i }))
    expect(onMerge).toHaveBeenCalledWith({ keepID: 2, mergeID: 1 })
  })

  it("shows warning about lost metadata", () => {
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={() => {}} />)
    expect(screen.getByText(/теги.*коллекции.*не переносятся/i)).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run merge-modal
```

Expected: FAIL.

- [ ] **Step 3: Implement merge-modal**

```tsx
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import type { DuplicatePair } from "@/api/prompt-insights"

interface MergeModalProps {
  pair: DuplicatePair
  open: boolean
  onClose: () => void
  onMerge: (args: { keepID: number; mergeID: number }) => void
}

export function MergeModal({ pair, open, onClose, onMerge }: MergeModalProps) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Объединить дубликаты</DialogTitle>
          <DialogDescription>
            Похожесть {Math.round(pair.similarity * 100)}%. Выберите, какой промпт оставить — второй уйдёт в корзину (можно восстановить за 30 дней). Теги и коллекции не переносятся.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-3 md:grid-cols-2">
          {[pair.prompt_a, pair.prompt_b].map((p, idx) => {
            const other = idx === 0 ? pair.prompt_b : pair.prompt_a
            return (
              <div key={p.prompt_id} className="space-y-2 rounded-md border p-3">
                <p className="text-sm font-medium">{p.title}</p>
                <p className="text-xs text-muted-foreground tabular-nums">{p.uses} использований</p>
                <Button
                  size="sm"
                  className="w-full"
                  onClick={() => onMerge({ keepID: p.prompt_id, mergeID: other.prompt_id })}
                >
                  Оставить «{p.title}»
                </Button>
              </div>
            )
          })}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            Отмена
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 4: Run merge-modal test → PASS**

```bash
npx vitest run merge-modal
```

Expected: PASS.

- [ ] **Step 5: Write the failing duplicates page test**

```tsx
import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent, waitFor } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import DuplicatesPage from "./duplicates"
import * as hooks from "@/hooks/use-prompt-insights"

vi.mock("@/hooks/use-prompt-insights")

function wrap(node: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("DuplicatesPage", () => {
  it("opens merge-modal when card clicked", async () => {
    vi.mocked(hooks.useDuplicates).mockReturnValue({
      data: [{ prompt_a: { prompt_id: 1, title: "A", uses: 0 }, prompt_b: { prompt_id: 2, title: "B", uses: 0 }, similarity: 0.9 }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof hooks.useDuplicates>)
    const mutate = vi.fn()
    vi.mocked(hooks.useMergePrompts).mockReturnValue({
      mutate,
      isPending: false,
    } as unknown as ReturnType<typeof hooks.useMergePrompts>)

    render(wrap(<DuplicatesPage />))
    fireEvent.click(screen.getByRole("button", { name: /объединить/i }))
    await waitFor(() => expect(screen.getByText(/похожесть 90%/i)).toBeInTheDocument())
  })
})
```

- [ ] **Step 6: Implement duplicates.tsx**

```tsx
import { useState } from "react"
import { Copy } from "lucide-react"
import { Button } from "@/components/ui/button"
import { MergeModal } from "@/components/prompts/insights/merge-modal"
import { useDuplicates, useMergePrompts } from "@/hooks/use-prompt-insights"
import type { DuplicatePair } from "@/api/prompt-insights"

export default function DuplicatesPage() {
  const { data, isLoading, isError } = useDuplicates()
  const merge = useMergePrompts()
  const [activePair, setActivePair] = useState<DuplicatePair | null>(null)

  return (
    <div className="mx-auto max-w-3xl space-y-4 p-6">
      <header>
        <h1 className="flex items-center gap-2 text-2xl font-semibold">
          <Copy className="h-5 w-5 text-blue-500" />
          Возможные дубликаты
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Похожие промпты — объедините, чтобы держать библиотеку чистой.
        </p>
      </header>

      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && <p className="text-sm text-destructive">Не удалось загрузить список.</p>}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">Дубликатов не нашлось.</p>
      )}

      <ul className="space-y-2">
        {data?.map((pair, i) => (
          <li key={`${pair.prompt_a.prompt_id}-${pair.prompt_b.prompt_id}-${i}`} className="rounded-md border p-3">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0 flex-1 space-y-0.5">
                <p className="truncate text-sm font-medium">{pair.prompt_a.title} ↔ {pair.prompt_b.title}</p>
                <p className="text-xs text-muted-foreground">Похожесть {Math.round(pair.similarity * 100)}%</p>
              </div>
              <Button size="sm" onClick={() => setActivePair(pair)}>
                Объединить
              </Button>
            </div>
          </li>
        ))}
      </ul>

      {activePair && (
        <MergeModal
          pair={activePair}
          open
          onClose={() => setActivePair(null)}
          onMerge={(args) => {
            merge.mutate(args, { onSuccess: () => setActivePair(null) })
          }}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 7: Run tests + commit**

```bash
npx vitest run duplicates merge-modal
```

Expected: PASS.

```bash
git add frontend/src/components/prompts/insights/merge-modal.tsx frontend/src/components/prompts/insights/merge-modal.test.tsx frontend/src/pages/prompts/insights/duplicates.tsx frontend/src/pages/prompts/insights/duplicates.test.tsx
git commit -m "feat(insights): /prompts/insights/duplicates с MergeModal flow"
```

---

### Task F6: pages/prompts/insights — trending/declining/most-edited

**Files:**
- Create: `frontend/src/pages/prompts/insights/trending.tsx`
- Create: `frontend/src/pages/prompts/insights/declining.tsx`
- Create: `frontend/src/pages/prompts/insights/most-edited.tsx`
- Create: `frontend/src/pages/prompts/insights/__tests__/list-pages.test.tsx`

- [ ] **Step 1: Write the failing test (shared)**

```tsx
import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import TrendingPage from "../trending"
import DecliningPage from "../declining"
import MostEditedPage from "../most-edited"
import * as hooks from "@/hooks/use-prompt-insights"

vi.mock("@/hooks/use-prompt-insights")

function wrap(node: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

const stub = (rows: Array<{ prompt_id: number; title: string; uses: number }>) => ({
  data: rows, isLoading: false, isError: false,
} as ReturnType<typeof hooks.useTrending>)

describe("list-style insight pages", () => {
  it("trending renders heading + items", () => {
    vi.mocked(hooks.useTrending).mockReturnValue(stub([{ prompt_id: 1, title: "Hot", uses: 20 }]))
    render(wrap(<TrendingPage />))
    expect(screen.getByRole("heading", { name: /растущие/i })).toBeInTheDocument()
    expect(screen.getByText("Hot")).toBeInTheDocument()
  })

  it("declining renders heading", () => {
    vi.mocked(hooks.useDeclining).mockReturnValue(stub([]))
    render(wrap(<DecliningPage />))
    expect(screen.getByRole("heading", { name: /падающие/i })).toBeInTheDocument()
  })

  it("most-edited renders heading", () => {
    vi.mocked(hooks.useMostEdited).mockReturnValue(stub([]))
    render(wrap(<MostEditedPage />))
    expect(screen.getByRole("heading", { name: /часто правят/i })).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run list-pages
```

Expected: FAIL.

- [ ] **Step 3: Implement 3 pages (одинаковый shape)**

`trending.tsx`:

```tsx
import { TrendingUp } from "lucide-react"
import { InsightPromptRow } from "@/components/prompts/insights/insight-prompt-row"
import { useTrending } from "@/hooks/use-prompt-insights"

export default function TrendingPage() {
  const { data, isLoading, isError } = useTrending()
  return (
    <div className="mx-auto max-w-3xl space-y-4 p-6">
      <header>
        <h1 className="flex items-center gap-2 text-2xl font-semibold">
          <TrendingUp className="h-5 w-5 text-emerald-500" />
          Растущие промпты
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Использование выросло ≥2× за неделю.
        </p>
      </header>
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && <p className="text-sm text-destructive">Не удалось загрузить.</p>}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">Пока нет растущих промптов.</p>
      )}
      <ul className="space-y-2">
        {data?.map((p) => (
          <li key={p.prompt_id}>
            <InsightPromptRow promptID={p.prompt_id} title={p.title} uses={p.uses} />
          </li>
        ))}
      </ul>
    </div>
  )
}
```

`declining.tsx` — копия с заменой `TrendingUp` → `TrendingDown`, цвет → `text-amber-500`, заголовок → «Падающие промпты», подзаголовок → «Использование снизилось ≥2× за неделю», hook → `useDeclining`, empty → «Пока нет падающих промптов.».

`most-edited.tsx` — `Archive` icon, цвет → `text-blue-500`, заголовок → «Часто правят», подзаголовок → «Промпты с большим числом версий», hook → `useMostEdited`, empty → «Нет частоправленных промптов.».

- [ ] **Step 4: Run tests**

```bash
npx vitest run list-pages
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/prompts/insights/trending.tsx frontend/src/pages/prompts/insights/declining.tsx frontend/src/pages/prompts/insights/most-edited.tsx frontend/src/pages/prompts/insights/__tests__/list-pages.test.tsx
git commit -m "feat(insights): trending/declining/most-edited pages"
```

---

### Task F7: App.tsx — route registration

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Add lazy imports + routes**

Найти секцию lazy импортов (около строки 50-90) и добавить:

```tsx
const InsightsUnused = lazy(() => import("@/pages/prompts/insights/unused"))
const InsightsDuplicates = lazy(() => import("@/pages/prompts/insights/duplicates"))
const InsightsTrending = lazy(() => import("@/pages/prompts/insights/trending"))
const InsightsDeclining = lazy(() => import("@/pages/prompts/insights/declining"))
const InsightsMostEdited = lazy(() => import("@/pages/prompts/insights/most-edited"))
const TagsPage = lazy(() => import("@/pages/tags-page"))
```

В `<Route element={<AppLayout />}>` (около L136) добавить после `/prompts/:id/analytics` route:

```tsx
<Route path="/prompts/insights/unused" element={<Suspense fallback={<PageFallback />}><InsightsUnused /></Suspense>} />
<Route path="/prompts/insights/duplicates" element={<Suspense fallback={<PageFallback />}><InsightsDuplicates /></Suspense>} />
<Route path="/prompts/insights/trending" element={<Suspense fallback={<PageFallback />}><InsightsTrending /></Suspense>} />
<Route path="/prompts/insights/declining" element={<Suspense fallback={<PageFallback />}><InsightsDeclining /></Suspense>} />
<Route path="/prompts/insights/most-edited" element={<Suspense fallback={<PageFallback />}><InsightsMostEdited /></Suspense>} />
<Route path="/tags" element={<Suspense fallback={<PageFallback />}><TagsPage /></Suspense>} />
```

- [ ] **Step 2: Verify build**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend
npx tsc --noEmit
```

Expected: clean (предполагает что `pages/tags-page.tsx` будет создан в F8 — если ещё не создан, временно закомментировать `/tags` route).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat(routes): 5 prompts/insights routes + /tags page"
```

---

### Task F8: pages/tags-page.tsx — minimal tags management + ?filter=orphan

**Files:**
- Create: `frontend/src/pages/tags-page.tsx`
- Create: `frontend/src/pages/tags-page.test.tsx`
- Create: `frontend/src/api/tag-orphan.ts` (отдельный fetcher, тонкий)
- Create: `frontend/src/hooks/use-orphan-tags.ts`

- [ ] **Step 1: Write the failing test**

```tsx
import { describe, it, expect, vi } from "vitest"
import { render, screen, waitFor } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import TagsPage from "./tags-page"
import * as tagHooks from "@/hooks/use-tags"
import * as orphanHooks from "@/hooks/use-orphan-tags"

vi.mock("@/hooks/use-tags")
vi.mock("@/hooks/use-orphan-tags")

function wrap(node: React.ReactNode, initialEntry: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter initialEntries={[initialEntry]}>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("TagsPage", () => {
  it("default shows all tags", async () => {
    vi.mocked(tagHooks.useTags).mockReturnValue({
      data: [{ id: 1, name: "feature" }, { id: 2, name: "old" }],
      isLoading: false, isError: false,
    } as ReturnType<typeof tagHooks.useTags>)
    render(wrap(<TagsPage />, "/tags"))
    expect(screen.getByText("feature")).toBeInTheDocument()
    expect(screen.getByText("old")).toBeInTheDocument()
  })

  it("?filter=orphan shows only orphan tags", () => {
    vi.mocked(orphanHooks.useOrphanTags).mockReturnValue({
      data: [{ id: 2, name: "old" }],
      isLoading: false, isError: false,
    } as ReturnType<typeof orphanHooks.useOrphanTags>)
    render(wrap(<TagsPage />, "/tags?filter=orphan"))
    expect(screen.getByText(/без активных промптов/i)).toBeInTheDocument()
    expect(screen.getByText("old")).toBeInTheDocument()
    expect(screen.queryByText("feature")).not.toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run tags-page
```

Expected: FAIL.

- [ ] **Step 3: Implement orphan hook + fetcher**

`api/tag-orphan.ts`:

```ts
import { apiUrl } from "@/lib/api-url"

export interface OrphanTag { id: number; name: string }

export async function fetchOrphanTags(): Promise<OrphanTag[]> {
  const res = await fetch(apiUrl("/api/tags/orphan"), { credentials: "include" })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const body = (await res.json()) as { items: OrphanTag[] }
  return body.items ?? []
}
```

`hooks/use-orphan-tags.ts`:

```ts
import { useQuery } from "@tanstack/react-query"
import { fetchOrphanTags } from "@/api/tag-orphan"

export function useOrphanTags() {
  return useQuery({ queryKey: ["tags", "orphan"], queryFn: fetchOrphanTags })
}
```

`pages/tags-page.tsx`:

```tsx
import { useSearchParams } from "react-router-dom"
import { Hash } from "lucide-react"
import { useTags, useDeleteTag } from "@/hooks/use-tags"
import { useOrphanTags } from "@/hooks/use-orphan-tags"
import { Button } from "@/components/ui/button"

export default function TagsPage() {
  const [params] = useSearchParams()
  const filter = params.get("filter")
  const isOrphan = filter === "orphan"

  const all = useTags(null)
  const orphan = useOrphanTags()
  const del = useDeleteTag()

  const items = isOrphan ? orphan.data : all.data
  const isLoading = isOrphan ? orphan.isLoading : all.isLoading
  const isError = isOrphan ? orphan.isError : all.isError

  return (
    <div className="mx-auto max-w-3xl space-y-4 p-6">
      <header>
        <h1 className="flex items-center gap-2 text-2xl font-semibold">
          <Hash className="h-5 w-5" />
          {isOrphan ? "Теги без активных промптов" : "Теги"}
        </h1>
        {isOrphan && (
          <p className="mt-1 text-sm text-muted-foreground">
            Эти теги не привязаны ни к одному активному промпту — можно удалить.
          </p>
        )}
      </header>

      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && <p className="text-sm text-destructive">Не удалось загрузить.</p>}
      {!isLoading && items && items.length === 0 && (
        <p className="text-sm text-muted-foreground">
          {isOrphan ? "Нет «orphan»-тегов — все теги используются." : "Тегов нет."}
        </p>
      )}

      <ul className="space-y-2">
        {items?.map((t) => (
          <li key={t.id} className="flex items-center justify-between gap-3 rounded-md border px-3 py-2">
            <span className="text-sm">{t.name}</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                if (confirm(`Удалить тег «${t.name}»?`)) del.mutate(t.id)
              }}
            >
              Удалить
            </Button>
          </li>
        ))}
      </ul>
    </div>
  )
}
```

(Если `useDeleteTag` не существует в `use-tags.ts` — пометить как TODO в commit message и временно `confirm + console.log`. Лучше — добавить в `use-tags.ts` минимальную mutation, см. fetcher pattern из `use-collections.ts`. Реализация:

```ts
// добавить в hooks/use-tags.ts
export function useDeleteTag() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => deleteTag(id), // api/tags.ts должен иметь deleteTag fn
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] })
      qc.invalidateQueries({ queryKey: ["tags", "orphan"] })
    },
  })
}
```

Если `deleteTag` тоже не существует в `api/tags.ts` — добавить:

```ts
export async function deleteTag(id: number): Promise<void> {
  const res = await fetch(apiUrl(`/api/tags/${id}`), { method: "DELETE", credentials: "include" })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}
```

Backend `DELETE /api/tags/:id` уже существует (см. `delivery/http/tag/handler.go`).)

- [ ] **Step 4: Run tests**

```bash
npx vitest run tags-page use-orphan-tags
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/tags-page.tsx frontend/src/pages/tags-page.test.tsx frontend/src/api/tag-orphan.ts frontend/src/hooks/use-orphan-tags.ts frontend/src/hooks/use-tags.ts frontend/src/api/tags.ts
git commit -m "feat(tags): /tags page с ?filter=orphan overlay"
```

---

### Task F9: pages/collections.tsx — поддержать ?filter=empty

**Files:**
- Modify: `frontend/src/pages/collections.tsx` (точное имя verify через `Glob frontend/src/pages/collections*`)
- Create: `frontend/src/api/collection-empty.ts`
- Create: `frontend/src/hooks/use-empty-collections.ts`
- Modify: `frontend/src/pages/__tests__/collections.test.tsx` (если есть; иначе создать)

- [ ] **Step 1: Write the failing test**

```tsx
// frontend/src/pages/__tests__/collections-filter.test.tsx
import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import Collections from "@/pages/collections"
import * as collectionsHooks from "@/hooks/use-collections"
import * as emptyHooks from "@/hooks/use-empty-collections"

vi.mock("@/hooks/use-collections")
vi.mock("@/hooks/use-empty-collections")

function wrap(node: React.ReactNode, initial: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter initialEntries={[initial]}>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("CollectionsPage filter=empty", () => {
  it("?filter=empty shows only empty collections", () => {
    vi.mocked(emptyHooks.useEmptyCollections).mockReturnValue({
      data: [{ id: 9, name: "Заброшенная" }], isLoading: false, isError: false,
    } as ReturnType<typeof emptyHooks.useEmptyCollections>)
    render(wrap(<Collections />, "/collections?filter=empty"))
    expect(screen.getByText(/без промптов|пустые коллекции/i)).toBeInTheDocument()
    expect(screen.getByText("Заброшенная")).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run collections-filter
```

Expected: FAIL.

- [ ] **Step 3: Implement hook + fetcher + page changes**

`api/collection-empty.ts`:

```ts
import { apiUrl } from "@/lib/api-url"

export interface EmptyCollection { id: number; name: string }

export async function fetchEmptyCollections(): Promise<EmptyCollection[]> {
  const res = await fetch(apiUrl("/api/collections/empty"), { credentials: "include" })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const body = (await res.json()) as { items: EmptyCollection[] }
  return body.items ?? []
}
```

`hooks/use-empty-collections.ts`:

```ts
import { useQuery } from "@tanstack/react-query"
import { fetchEmptyCollections } from "@/api/collection-empty"

export function useEmptyCollections() {
  return useQuery({ queryKey: ["collections", "empty"], queryFn: fetchEmptyCollections })
}
```

В `pages/collections.tsx` (или как файл называется фактически) добавить в начало компонента:

```tsx
import { useSearchParams } from "react-router-dom"
import { useEmptyCollections } from "@/hooks/use-empty-collections"

// ...

const [params] = useSearchParams()
const isEmpty = params.get("filter") === "empty"
const all = useCollections(null)
const empty = useEmptyCollections()
const items = isEmpty ? empty.data : all.data
```

И отрисовать `items` вместо `all.data`. В header добавить badge/подзаголовок «Пустые коллекции» когда `isEmpty`.

- [ ] **Step 4: Run tests + verify не сломали базовое поведение**

```bash
npx vitest run collections
```

Expected: PASS (все existing + новый).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/collection-empty.ts frontend/src/hooks/use-empty-collections.ts frontend/src/pages/collections.tsx frontend/src/pages/__tests__/collections-filter.test.tsx
git commit -m "feat(collections): ?filter=empty overlay через /api/collections/empty"
```

---

### Task F10: insights-panel.tsx — INSIGHT_META hrefs + rename title

**Files:**
- Modify: `frontend/src/components/analytics/insights-panel.tsx`

- [ ] **Step 1: Write the failing test**

Добавить в `frontend/src/components/analytics/__tests__/insights-panel.test.tsx` (или прямо в файл рядом с компонентом):

```tsx
import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { InsightsPanel } from "../insights-panel"
import type { Insight } from "@/api/analytics"

function wrap(node: React.ReactNode) {
  return render(<MemoryRouter>{node}</MemoryRouter>)
}

describe("insights-panel hrefs", () => {
  it("unused_prompts links to /prompts/insights/unused", () => {
    const insights: Insight[] = [{ type: "unused_prompts", payload: [{ id: 1 }, { id: 2 }] } as Insight]
    wrap(<InsightsPanel insights={insights} />)
    const link = screen.getByRole("link", { name: /посмотреть/i })
    expect(link).toHaveAttribute("href", "/prompts/insights/unused")
  })

  it("possible_duplicates links to /prompts/insights/duplicates", () => {
    const insights: Insight[] = [{ type: "possible_duplicates", payload: [{}] } as Insight]
    wrap(<InsightsPanel insights={insights} />)
    expect(screen.getByRole("link", { name: /объединить/i })).toHaveAttribute("href", "/prompts/insights/duplicates")
  })

  it("orphan_tags has russian title and tags?filter=orphan link", () => {
    const insights: Insight[] = [{ type: "orphan_tags", payload: [{}, {}] } as Insight]
    wrap(<InsightsPanel insights={insights} />)
    expect(screen.getByText(/теги без промптов/i)).toBeInTheDocument()
    expect(screen.queryByText(/orphan/i)).not.toBeInTheDocument()
    expect(screen.getByRole("link", { name: /очистить/i })).toHaveAttribute("href", "/tags?filter=orphan")
  })

  it("empty_collections links to /collections?filter=empty", () => {
    const insights: Insight[] = [{ type: "empty_collections", payload: [{}] } as Insight]
    wrap(<InsightsPanel insights={insights} />)
    expect(screen.getByRole("link", { name: /очистить/i })).toHaveAttribute("href", "/collections?filter=empty")
  })

  it("trending/declining/most_edited link to dedicated routes", () => {
    const cases: Array<[Insight["type"], string]> = [
      ["trending", "/prompts/insights/trending"],
      ["declining", "/prompts/insights/declining"],
      ["most_edited", "/prompts/insights/most-edited"],
    ]
    for (const [t, href] of cases) {
      const insights: Insight[] = [{ type: t, payload: [{}] } as Insight]
      const { container, unmount } = render(<MemoryRouter><InsightsPanel insights={insights} /></MemoryRouter>)
      expect(container.querySelector(`a[href="${href}"]`)).not.toBeNull()
      unmount()
    }
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run insights-panel
```

Expected: FAIL — текущие hrefs указывают на старые routes (`/prompts?filter=...`).

- [ ] **Step 3: Update INSIGHT_META**

В `insights-panel.tsx` обновить каждую запись `INSIGHT_META`:

```tsx
const INSIGHT_META: Record<Insight["type"], { ... }> = {
  unused_prompts: {
    icon: AlertCircle, tone: "warning", title: "Забытые",
    href: "/prompts/insights/unused",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт не использовался" : "промптов не использовались"} 30+ дней`,
    ctaLabel: "Посмотреть",
  },
  possible_duplicates: {
    icon: Copy, tone: "info", title: "Дубликаты",
    href: "/prompts/insights/duplicates",
    descBuilder: (n) => `${n} ${n === 1 ? "пара" : "пары"} похожих промптов`,
    ctaLabel: "Объединить",
  },
  trending: {
    icon: TrendingUp, tone: "success", title: "Растут",
    href: "/prompts/insights/trending",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} растут в использовании`,
    ctaLabel: "Открыть",
  },
  declining: {
    icon: TrendingDown, tone: "warning", title: "Падают",
    href: "/prompts/insights/declining",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} используются всё реже`,
    ctaLabel: "Посмотреть",
  },
  most_edited: {
    icon: Archive, tone: "info", title: "Часто правят",
    href: "/prompts/insights/most-edited",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} с большим числом версий`,
    ctaLabel: "Открыть",
  },
  orphan_tags: {
    icon: Hash, tone: "warning", title: "Теги без промптов",
    href: "/tags?filter=orphan",
    descBuilder: (n) => `${n} ${n === 1 ? "тег" : "тегов"} без активных промптов`,
    ctaLabel: "Очистить",
  },
  empty_collections: {
    icon: FolderOpen, tone: "warning", title: "Пустые коллекции",
    href: "/collections?filter=empty",
    descBuilder: (n) => `${n} ${n === 1 ? "коллекция" : "коллекций"} без промптов`,
    ctaLabel: "Очистить",
  },
}
```

- [ ] **Step 4: Run tests**

```bash
npx vitest run insights-panel
```

Expected: PASS все 4 теста.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/analytics/insights-panel.tsx frontend/src/components/analytics/__tests__/insights-panel.test.tsx
git commit -m "fix(insights): корректные deep-link hrefs + 'Теги без промптов' RU"
```

---

## Phase 3 — Wave 2: UI Polish

### Task U1: lib/date-format.ts — formatDayShort utility

**Files:**
- Create: `frontend/src/lib/date-format.ts`
- Create: `frontend/src/lib/date-format.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, it, expect } from "vitest"
import { formatDayShort } from "./date-format"

describe("formatDayShort", () => {
  it("formats ISO date as 'D MMM' in Russian", () => {
    expect(formatDayShort("2026-05-07")).toBe("7 мая")
    expect(formatDayShort("2026-05-16")).toBe("16 мая")
    expect(formatDayShort("2026-12-01")).toBe("1 дек.")
  })

  it("returns empty string for invalid input", () => {
    expect(formatDayShort("")).toBe("")
    expect(formatDayShort("not-a-date")).toBe("")
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run date-format
```

Expected: FAIL.

- [ ] **Step 3: Implement utility**

```ts
const fmt = new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "short" })

export function formatDayShort(iso: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  // Intl возвращает "7 мая" (без точки) / "1 дек." (с точкой) для русской локали.
  return fmt.format(d)
}
```

- [ ] **Step 4: Run test**

```bash
npx vitest run date-format
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/date-format.ts frontend/src/lib/date-format.test.ts
git commit -m "feat(lib): formatDayShort через Intl ru-RU"
```

---

### Task U2: activity-heatmap.tsx — pad to 28 cells + russian tooltip

**Files:**
- Modify: `frontend/src/components/analytics/activity-heatmap.tsx`
- Modify: `frontend/src/components/analytics/activity-heatmap.test.tsx`

- [ ] **Step 1: Write the failing test**

Обновить `activity-heatmap.test.tsx`:

```tsx
import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { ActivityHeatmap } from "./activity-heatmap"

describe("ActivityHeatmap padding", () => {
  it("renders exactly 28 cells when data is empty", () => {
    const { container } = render(<ActivityHeatmap points={[]} />)
    expect(container.querySelectorAll("[data-cell]")).toHaveLength(28)
  })

  it("renders exactly 28 cells when 5 days of data", () => {
    const points = Array.from({ length: 5 }, (_, i) => ({
      day: `2026-05-${String(12 + i).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    expect(container.querySelectorAll("[data-cell]")).toHaveLength(28)
  })

  it("renders russian month in tooltip", () => {
    const points = [{ day: "2026-05-12", count: 3 }]
    render(<ActivityHeatmap points={points} />)
    const cell = screen.getByLabelText(/12 мая: 3 использований/i)
    expect(cell).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run activity-heatmap
```

Expected: FAIL — текущая реализация рендерит только existing points.

- [ ] **Step 3: Update component**

В `activity-heatmap.tsx`:

```tsx
import { formatDayShort } from "@/lib/date-format"

interface Point { day: string; count: number }
interface Props { points: Point[] }

export function ActivityHeatmap({ points }: Props) {
  const cells = padToWindow(points, 28)
  const max = Math.max(1, ...cells.map((c) => c.count))

  return (
    <div className="grid grid-cols-7 gap-1.5">
      {cells.map((c) => (
        <div
          key={c.day}
          data-cell
          aria-label={`${formatDayShort(c.day)}: ${c.count} ${pluralUses(c.count)}`}
          className="aspect-square rounded-sm bg-emerald-500/[--alpha]"
          style={{ ["--alpha" as string]: c.count === 0 ? "0.06" : String(0.2 + (c.count / max) * 0.8) }}
          title={`${formatDayShort(c.day)}: ${c.count} ${pluralUses(c.count)}`}
        />
      ))}
    </div>
  )
}

function padToWindow(points: Point[], days: number): Point[] {
  // Берём last `days` дней от today, мерджим с points (lookup by date string).
  const byDay = new Map(points.map((p) => [p.day, p.count]))
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const out: Point[] = []
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    out.push({ day: key, count: byDay.get(key) ?? 0 })
  }
  return out
}

function pluralUses(n: number): string {
  if (n === 1) return "использование"
  if (n >= 2 && n <= 4) return "использования"
  return "использований"
}
```

(Если existing implementation использует другие props/типы — адаптировать имена; ключевая идея: `padToWindow` гарантирует 28 cells, lookup by ISO date.)

- [ ] **Step 4: Run tests**

```bash
npx vitest run activity-heatmap
```

Expected: PASS все 3 теста.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/analytics/activity-heatmap.tsx frontend/src/components/analytics/activity-heatmap.test.tsx
git commit -m "fix(heatmap): 28 cells GitHub-style padding + русский tooltip"
```

---

### Task U3: narrative-banner.tsx — убрать href + ArrowRight

**Files:**
- Modify: `frontend/src/components/analytics/narrative-banner.tsx`
- Modify: `frontend/src/components/analytics/narrative-banner.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { NarrativeBanner } from "./narrative-banner"

describe("NarrativeBanner", () => {
  it("does not render anchor/link wrapper", () => {
    const { container } = render(<NarrativeBanner summary="За неделю +12% использований" actionHint="3 забытых промпта" />)
    expect(container.querySelectorAll("a")).toHaveLength(0)
  })

  it("does not render ArrowRight icon", () => {
    const { container } = render(<NarrativeBanner summary="..." actionHint="..." />)
    expect(container.querySelector("[data-lucide='arrow-right']")).toBeNull()
  })

  it("still displays summary and actionHint", () => {
    render(<NarrativeBanner summary="За неделю +12%" actionHint="3 забытых промпта" />)
    expect(screen.getByText(/12%/i)).toBeInTheDocument()
    expect(screen.getByText(/забытых/i)).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run narrative-banner
```

Expected: FAIL — текущая реализация оборачивает в `<a>`.

- [ ] **Step 3: Update component**

В `narrative-banner.tsx` убрать `<a href>` wrapper и `<ArrowRight>` icon. Banner становится статичным `<div>`:

```tsx
import { Sparkles } from "lucide-react"

interface NarrativeBannerProps {
  summary: string
  actionHint: string | null
}

export function NarrativeBanner({ summary, actionHint }: NarrativeBannerProps) {
  return (
    <div className="flex items-start gap-3 rounded-lg border bg-gradient-to-r from-primary/5 to-transparent p-4">
      <Sparkles className="mt-0.5 h-5 w-5 shrink-0 text-primary" />
      <div className="space-y-1">
        <p className="text-sm font-medium">{summary}</p>
        {actionHint && <p className="text-xs text-muted-foreground">{actionHint}</p>}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run tests**

```bash
npx vitest run narrative-banner
```

Expected: PASS все 3 теста.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/analytics/narrative-banner.tsx frontend/src/components/analytics/narrative-banner.test.tsx
git commit -m "fix(banner): убрать self-link на /analytics и ArrowRight"
```

---

### Task U4: sparkline.tsx — constant data → single dot

**Files:**
- Modify: `frontend/src/components/analytics/sparkline.tsx`
- Modify: `frontend/src/components/analytics/sparkline.test.tsx`

- [ ] **Step 1: Write the failing test**

Добавить кейс в `sparkline.test.tsx`:

```tsx
it("renders single dot when all points are equal (constant data)", () => {
  const { container } = render(<Sparkline points={[5, 5, 5, 5]} />)
  expect(container.querySelector("circle")).not.toBeNull()
  // polyline не должно быть (или должно быть hidden)
  const polyline = container.querySelector("polyline[stroke-width]")
  expect(polyline).toBeNull()
})

it("renders single dot for all-zero array", () => {
  const { container } = render(<Sparkline points={[0, 0, 0, 0]} />)
  expect(container.querySelector("circle")).not.toBeNull()
})

it("renders polyline for non-constant data", () => {
  const { container } = render(<Sparkline points={[1, 3, 2, 5]} />)
  expect(container.querySelector("polyline[stroke-width]")).not.toBeNull()
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run sparkline
```

Expected: FAIL — для constant data сейчас рендерит линию.

- [ ] **Step 3: Update component**

В `sparkline.tsx` добавить early-return для constant data:

```tsx
export function Sparkline({ points, color = "currentColor", width = 120, height = 22 }: SparklineProps) {
  if (points.length === 0) return null

  const max = Math.max(...points)
  const min = Math.min(...points)
  const isConstant = max === min

  if (isConstant) {
    // Все точки одинаковы (включая все нули) — рендерим dot в правом конце как
    // явный сигнал "no trend". Линия в этом случае была бы плоской чертой.
    return (
      <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`} aria-label="нет тренда">
        <circle cx={width - 4} cy={height / 2} r={2.5} fill={color} />
      </svg>
    )
  }

  // existing рендер polyline ↓
  const range = max - min
  const step = (width - 2) / (points.length - 1)
  const pts = points.map((p, i) => `${i * step + 1},${height - ((p - min) / range) * (height - 2) - 1}`).join(" ")

  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
      <polyline points={pts} fill="none" stroke={color} strokeWidth={1.5} />
    </svg>
  )
}
```

(Если existing component использует `<path>` вместо `<polyline>` — оставить ту же логику, но изменить условие в тесте под фактический тэг.)

- [ ] **Step 4: Run tests**

```bash
npx vitest run sparkline
```

Expected: PASS все 3 новых теста + existing.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/analytics/sparkline.tsx frontend/src/components/analytics/sparkline.test.tsx
git commit -m "fix(sparkline): constant data → одиночная точка вместо плоской линии"
```

---

### Task U5: usage-chart.tsx — formatDayShort tickFormatter

**Files:**
- Modify: `frontend/src/components/analytics/usage-chart.tsx`

- [ ] **Step 1: Write the failing test**

Если у `usage-chart.tsx` есть существующий test файл — добавить:

```tsx
import { describe, it, expect } from "vitest"
import { render } from "@testing-library/react"
import { UsageChart } from "./usage-chart"
import { createUsageChartConfig } from "./usage-chart-config"

describe("UsageChart x-axis", () => {
  it("formats ticks as russian short date '7 мая'", () => {
    const { container } = render(
      <UsageChart
        points={[
          { day: "2026-05-07", count: 1 },
          { day: "2026-05-16", count: 5 },
        ]}
        chartConfig={createUsageChartConfig("использования")}
      />
    )
    // Recharts рендерит ticks как <text> внутри <svg>. Поищем текст.
    expect(container.textContent).toMatch(/7 мая/i)
    expect(container.textContent).toMatch(/16 мая/i)
  })
})
```

(Если test файла нет — создать `usage-chart.test.tsx` с минимальным wrapper.)

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run usage-chart
```

Expected: FAIL — текущий `tickFormatter={(v) => v.slice(5)}` возвращает `"05-07"`.

- [ ] **Step 3: Update tickFormatter**

В `usage-chart.tsx`:

```tsx
import { formatDayShort } from "@/lib/date-format"

// ...

<XAxis
  dataKey="day"
  tickLine={false}
  axisLine={false}
  tickFormatter={formatDayShort}
  fontSize={11}
/>
```

- [ ] **Step 4: Run test**

```bash
npx vitest run usage-chart
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/analytics/usage-chart.tsx frontend/src/components/analytics/usage-chart.test.tsx
git commit -m "fix(usage-chart): русский DD MMM формат tick'ов через formatDayShort"
```

---

### Task U6: analytics-narrative.ts — skip empty topModel / streak=0

**Files:**
- Modify: `frontend/src/lib/analytics-narrative.ts`
- Modify: `frontend/src/lib/analytics-narrative.test.ts`

- [ ] **Step 1: Write the failing test**

Добавить кейсы в `analytics-narrative.test.ts`:

```ts
it("skips topModel segment when model is empty string", () => {
  const result = buildNarrative({
    range: "7d",
    totalsCurrent: { uses: 10 },
    totalsPrevious: { uses: 8 },
    usageByModel: [{ model: "", uses: 10, pct: 100 }],
  } as any, [])
  expect(result.summary).not.toMatch(/без модели/i)
  expect(result.summary).not.toMatch(/100%/i)
})

it("skips topModel segment when only one model with 100%", () => {
  const result = buildNarrative({
    range: "7d",
    totalsCurrent: { uses: 10 },
    totalsPrevious: { uses: 8 },
    usageByModel: [{ model: "claude-sonnet-4", uses: 10, pct: 100 }],
  } as any, [])
  // 100% единственной модели — не информативно, скрываем
  expect(result.summary).not.toMatch(/100%/i)
})

it("keeps topModel segment when pct < 100", () => {
  const result = buildNarrative({
    range: "7d",
    totalsCurrent: { uses: 10 },
    totalsPrevious: { uses: 8 },
    usageByModel: [
      { model: "claude-sonnet-4", uses: 6, pct: 60 },
      { model: "gpt-4", uses: 4, pct: 40 },
    ],
  } as any, [])
  expect(result.summary).toMatch(/claude-sonnet-4/i)
  expect(result.summary).toMatch(/60%/i)
})

it("skips streak segment when current_streak is 0", () => {
  // streak-сегмент добавляется в pages/analytics.tsx, но если он строится тут —
  // тест на buildStreakSegment(0) === null. Если в analytics.tsx — пометить TODO.
  expect(buildStreakSegment({ current: 0, longest: 5 } as any)).toBe(null)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
npx vitest run analytics-narrative
```

Expected: FAIL — segments рендерятся.

- [ ] **Step 3: Update buildTopModel / buildStreakSegment**

В `analytics-narrative.ts`:

```ts
export function buildTopModel(usageByModel: ModelRow[]): string | null {
  const top = usageByModel[0]
  if (!top) return null
  if (!top.model) return null  // пустая строка → skip
  if (top.pct === 100 && usageByModel.length === 1) return null  // единственная модель 100% → skip
  return `топ-модель ${top.model} (${top.pct}%)`
}

// Если buildStreakSegment живёт в этом файле — экспортировать и обновить.
// Если нет — создать helper:
export function buildStreakSegment(streak: { current: number; longest: number }): string | null {
  if (streak.current === 0) return null
  return `streak ${streak.current} ${pluralDays(streak.current)}`
}

function pluralDays(n: number): string {
  if (n === 1) return "день"
  if (n >= 2 && n <= 4) return "дня"
  return "дней"
}
```

В `buildNarrative` (внутри файла) убедиться что `buildTopModel` результат проверяется на null:

```ts
const topModelSeg = buildTopModel(data.usageByModel ?? [])
if (topModelSeg) segments.push(topModelSeg)
```

И аналогично streak — если он строится здесь.

- [ ] **Step 4: Run tests**

```bash
npx vitest run analytics-narrative
```

Expected: PASS все новые + existing.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/analytics-narrative.ts frontend/src/lib/analytics-narrative.test.ts
git commit -m "fix(narrative): skip uninformative segments (empty model, 100% single, streak=0)"
```

---

### Task U7: Audit русских строк insights (grep)

**Files:**
- Read-only audit

- [ ] **Step 1: Grep по нежелательным англицизмам**

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend/src
```

В Claude используем Grep tool на каждое подозрительное слово в `src/`:

- `orphan` (кроме внутренних insight type strings — те menять не надо, они в API). Если в JSX тексте — заменить на «без активных промптов».
- `Orphan` — то же.
- `trending` в JSX (UI label) — заменить на «растущие».
- `declining` в JSX — заменить на «падающие».
- `duplicates` в JSX — заменить на «дубликаты».
- `most-edited` / `most_edited` в JSX — заменить на «часто правят».

Команды (через Grep tool, не bash):

```
Grep pattern="\\borphan\\b" type="tsx,jsx"
Grep pattern="\\btrending\\b" type="tsx,jsx"
Grep pattern="\\bdeclining\\b" type="tsx,jsx"
```

- [ ] **Step 2: Заменить найденные англицизмы в UI**

Для каждой находки в JSX text (не type discriminator!) — заменить через Edit на русский эквивалент. Не трогать API типы.

- [ ] **Step 3: Verify no orphan/trending в visible UI**

После замен — повторный grep, должен возвращать только type discriminators (в `analytics.ts` `Insight["type"]`).

- [ ] **Step 4: Run all tests**

```bash
npm run lint
npx tsc --noEmit
npx vitest run
```

Expected: всё зелёное.

- [ ] **Step 5: Commit (если были замены)**

```bash
git add frontend/src/
git commit -m "fix(i18n): русифицировать оставшиеся англицизмы в UI strings"
```

(Если не было замен — пропустить commit.)

---

## Final verification (перед PR)

### Final 1: Full test suite

- [ ] Backend:

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/backend
go test -short -race -count=1 -timeout=5m ./...
golangci-lint run
```

Expected: всё зелёное.

- [ ] Backend integration (requires Docker):

```bash
go test ./internal/infrastructure/postgres/repository/...
```

Expected: PASS (включая TestPromptMergeWith).

- [ ] Frontend:

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault/frontend
npm run lint
npx tsc --noEmit
npx vitest run
npm run build
```

Expected: всё зелёное, build clean.

### Final 2: Smoke в Docker

- [ ] Поднять stack:

```bash
cd C:/GolandProjects/awesomeProject/test/promptvault
docker compose -f docker-compose.dev.yml up -d --build
```

- [ ] Зайти в браузер `http://localhost:5173` как `e2e-max@test.local` / `TestPass2026!`.

- [ ] Перейти на `/analytics`, проверить визуально каждый из 7 фиксов:

  1. Smart Insights cards (7 типов) — кликабельны, ведут на `/prompts/insights/...` или `/tags?filter=orphan` / `/collections?filter=empty`. Не 404.
  2. Card «Теги без промптов» (а не «ORPHAN-ТЕГИ»).
  3. Activity heatmap — ровно 28 ячеек (по 7 в 4 рядах).
  4. NarrativeBanner — не кликается, нет ArrowRight.
  5. UsageChart x-axis — «7 мая», «16 мая» (русский).
  6. KPI sparkline для metric с константными данными — точка справа, не черта.
  7. NarrativeBanner — нет «100% Без модели» / «streak 0 дней».

- [ ] Для **duplicates page**: создать 2 похожих промпта вручную (через UI), дождаться пересчёта insights, открыть `/prompts/insights/duplicates`, кликнуть «Объединить», выбрать сторону, проверить что merge target ушёл в `/trash`.

### Final 3: Bundle size

```bash
npm run build
du -sh dist/assets/*.js | sort -rh | head -5
```

Acceptable delta: <20 KB gzipped (5 insight pages + 1 tags page + 4 hooks + 2 API modules + merge-modal).

### Final 4: Spec coverage check

Сверить design doc (`docs/superpowers/specs/2026-05-17-analytics-ux-fixes-design.md`) §9 (план внедрения, Wave 1 шаги B1-B9 + F1-F9, Wave 2 шаги U1-U6) с реализованными tasks этого плана. Все purpose-shot'ы покрыты.

---

## Self-Review

### Spec coverage

| Spec section | Task |
|---|---|
| §3 prompt_insights/types.go + errors.go | B1 |
| §3 prompt_insights/service.go (5 List* methods) | B3-B5 |
| §3 prompt_insights/service.go (MergePrompts) | B7 |
| §3 prompt_insights/service_test.go | B3-B5, B7 |
| §3 prompt/insights_handler.go (5 GET) | B8 |
| §3 prompt/insights_handler.go (Merge POST) | B9 |
| §3 tag/orphan_handler.go | B10 |
| §3 collection/empty_handler.go | B11 |
| §3 app.go + routes.go wire-up | B12 |
| §3 prompt.MergeWith repo + impl | B6 |
| §3 frontend api/prompt-insights.ts | F1 |
| §3 frontend hooks/use-prompt-insights.ts | F2 |
| §3 components/prompts/insights/insight-prompt-row.tsx | F3 |
| §3 components/prompts/insights/merge-modal.tsx | F5 |
| §3 5 pages/prompts/insights/<type>.tsx | F4, F5, F6 |
| §3 App.tsx routes | F7 |
| §3 insights-panel.tsx INSIGHT_META hrefs + orphan_tags title | F10 |
| §3 tags-page.tsx + ?filter=orphan | F8 |
| §3 collections-page.tsx + ?filter=empty | F9 |
| §3 activity-heatmap padding | U2 |
| §3 narrative-banner href removed | U3 |
| §3 sparkline constant data | U4 |
| §3 usage-chart tickFormatter + date-format util | U1, U5 |
| §3 analytics-narrative skip edge cases | U6 |
| §5 API contract (7 endpoints + merge) | B8, B9, B10, B11 |
| §7 Unit tests (frontend + backend) | каждый task с тестами |
| §7 Integration tests (testcontainers MergeWith) | B6 |
| §7 E2E smoke | Final 2 |
| §10 No feature flags | acknowledged (additive functionality) |
| §12 Pre-mortem mitigations (idempotent soft-delete) | B6 — soft-delete idempotent через gorm.DeletedAt |
| §13 Bundle size delta | Final 3 |

**Gaps:** Нет. Все требования спеки покрыты.

### Placeholder scan

- Нет «TBD», «TODO», «implement later».
- Все code-snippets — complete, references только на типы определённые в plan'е или существующих файлах (verified by subagent в pre-plan research).
- Single semi-placeholder в F8 (`useDeleteTag`/`deleteTag`): план содержит fallback с реализацией, если их нет в codebase.

### Type consistency

- `PromptInsightRow` — exported в B1, used в B3-B5, B7, F1, F3, F4, F6.
- `DuplicatePair` — exported в B1, used в B4, B8, F1, F5.
- `InsightsService` interface (handler-side) — defined в B8 (имена методов точно match Service в B3-B7).
- `PlanGate` / `PromptMerger` (usecase DI) — defined в B3, implemented в B12 (`promptInsightsPlanGate`).
- `analytics.Service.InsightsForPlan` — promoted в B2, used в B12 adapter.
- Frontend hooks: `useUnusedPrompts`, `useDuplicates`, `useTrending`, `useDeclining`, `useMostEdited`, `useMergePrompts` — defined в F2, used в F4-F6.
- `formatDayShort` — defined в U1, used в U2 (heatmap tooltip) и U5 (usage-chart tickFormatter).
- INSIGHT_META hrefs в F10 точно match routes в F7 (App.tsx) и filter params в F8/F9.

### Risks acknowledged

- F8 (`useDeleteTag` may not exist): plan содержит inline-инструкцию добавить mutation + fetcher как fallback.
- F9 (`collections.tsx` file name): plan просит verify имя через Glob перед началом task'а (если файл `collections-page.tsx` — поменять).
- B12 (`analytics.Service` exported method, adapter pattern): если паттерн `app.go` использует interface composition differently — adapter надо переписать под точную форму. План указывает на `analytics_adapter` отдельным файлом.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-17-analytics-ux-fixes.md`. Two execution options:

**1. Subagent-Driven (recommended)** — Fresh subagent per task + two-stage review между задачами. 24 task'а — управляемо, контекст основной сессии защищён.

**2. Inline Execution** — Batch execution в этой сессии через executing-plans skill, checkpoints для review.

Which approach?
