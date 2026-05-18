# Analytics UX Fixes — Design Doc

**Дата:** 2026-05-17
**Owner:** Slava Kovalchuk
**Branch:** `feat/pricing-iteration-v3` (поверх analytics redesign)
**Статус:** утверждён, готов к implementation plan

---

## Контекст

**Фича.** Серия UX-фиксов и доработок аналитики после первого smoke test пользователя на свежем редизайне `/analytics`. Состоит из двух волн:

1. **Wave 1 — Deep linking из Smart Insights:** новые insight pages (`/prompts/insights/:type`, `/tags?filter=orphan`, `/collections?filter=empty`) с inline actions (Delete, Merge, View). Backend endpoints поверх существующего `analytics_repo.go` SQL.
2. **Wave 2 — UI polish:** русификация labels, GitHub-style activity heatmap padding, NarrativeBanner edge cases, Sparkline для constant data, UsageChart tick format.

**Для кого.** Все юзеры на `/analytics` (Free / Pro / Max). Smart Insights actionable cards сейчас ведут на пустые страницы (broken deep links) — главный pain point.

**Зачем.** Семь конкретных проблем, выявленных в smoke test:

1. **CTA-ссылки broken.** Я сочинил deep links (`/prompts?filter=unused`, `/tags`, etc.), которые не существуют в роутере. Юзер кликает → пустая страница.
2. **ORPHAN-ТЕГИ английский** в локализации (CSS uppercase превращает «Orphan-теги» в `ORPHAN-ТЕГИ`).
3. **Activity heatmap 2 квадратика вместо 28.** Renders только existing points; нет padding для пустых дней.
4. **NarrativeBanner кликабелен, но ведёт на `/analytics`** (self-link, бессмысленно).
5. **UsageChart x-axis ticks** — даты в формате `05-07`, неравномерное spacing.
6. **Sparkline для constant data** выглядит как горизонтальная черта (`_`).
7. **NarrativeBanner edge cases** — «топ-модель Без модели (100%)» / «streak 0 дней» рендерятся даже когда не информативны.

**Жёсткие ограничения.**

- Sохраняем three-state Pro Insights teaser (Free → UpgradeGate, Pro → 2 типа + 5 locked, Max → все 7). Wave 1 deep linking учитывает текущий plan.
- Backend SQL queries для insights уже есть в `analytics_repo.go` — переиспользуем, не дублируем.
- shadcn/ui + Lucide + Recharts — никаких новых deps.
- Сохраняем существующий `/prompts`, `/tags`, `/collections` контракт для обычного browsing — фильтры insights не ломают base behavior.

**Заинтересованные стороны.** Исполнитель + PR-ревьюер (Slava). Без внешних stakeholder'ов.

**Аудит ясности.**

- **CTA navigation pattern** утверждён как «новые insight pages» (Approach A из brainstorming): отдельные routes `/prompts/insights/:type` чисто разделяют insights views от обычного listing.
- **Inline actions** на insight pages: для unused — «Использовать» / «Удалить»; для duplicates — «Объединить промпты»; для trending/declining — только просмотр (read-only); для most_edited — «Перейти к промпту»; для orphan_tags — «Удалить тег»; для empty_collections — «Удалить коллекцию».
- **Merge duplicates** — отдельный flow: показывать два промпта side-by-side, юзер выбирает «оставить A» / «оставить B» / «cancel». Реализуется как modal внутри `/prompts/insights/duplicates`.

---

## Карта существующего кода

**Слои.** Backend Clean Architecture: `delivery/http/<feature>/handler.go` → `usecases/<feature>/` → `interface/repository/` → `infrastructure/postgres/repository/`. Frontend: `pages/`, `components/`, `hooks/`, `api/`.

**Эталонные файлы.**

- `backend/internal/infrastructure/postgres/repository/analytics_repo.go:65 UnusedPrompts(ctx, userID, teamID *uint, before, limit)` — возвращает `[]repo.PromptUsageRow` (поле `prompt_id`, `title`, `uses`). Переиспользуем в новых endpoints. Аналогично `GetTrendingPrompts:268`, `PossibleDuplicates`, `MostEditedPrompts`, `OrphanTags`, `EmptyCollections`.
- `backend/internal/usecases/analytics/service.go:111 GetInsightsGated(ctx, userID, teamID *uint)` — гейт по плану через `s.insightsForPlan(planID)`. Тот же гейт применяется на новые prompt-insights endpoints.
- `backend/internal/delivery/http/prompt/handler.go` — существующий `GET /api/prompts` handler. **Не меняем** — новые routes отдельные.
- `frontend/src/pages/prompts/index.tsx` (если есть) — base list page. **Не меняем** — новые insight pages отдельные.
- `frontend/src/components/analytics/insight-action-card.tsx` — `href` prop. **Меняем** — новые корректные routes в `insights-panel.tsx:INSIGHT_META`.
- `frontend/src/components/analytics/activity-heatmap.tsx` — текущий рендер только existing points. **Меняем** — добавляем padding до 28 ячеек.
- `frontend/src/components/analytics/sparkline.tsx` — рендерит polyline для любых points. **Меняем** — для constant data рендерим dot или skip.
- `frontend/src/components/analytics/narrative-banner.tsx` — `href` prop. **Меняем** — удаляем href entirely.
- `frontend/src/lib/analytics-narrative.ts:buildTopModel/buildSummary` — **Меняем** — skip edge cases (model="", streak=0).
- `frontend/src/components/analytics/usage-chart.tsx` — `tickFormatter`. **Меняем** — формат `DD MMM`.

**Тесты-эталоны.**

- `backend/internal/usecases/analytics/insights_test.go` — table-driven по insight types. Копируем для usecase нового prompt-insights service.
- `frontend/src/components/analytics/activity-heatmap.test.tsx` — verify cells count. **Расширяем** — verify padding до 28.

**Свежий git log по затрагиваемым директориям.** 35 коммитов на ветке за день: Pricing v3 (Tasks 1-19) + Analytics Redesign (Tasks A1-A16). Активный refactor frontend analytics components. Backend — стабилизация.

**Конвенции.** Backend slog + handler.go pattern; frontend Lucide + tabular-nums + Tailwind tokens; Vitest explicit imports + `afterEach(cleanup)`; TanStack Query для всех API; React Router 7.13 `Link to=...`.

---

## 1. Резюме

Серия фиксов из 2 wave: Wave 1 добавляет реальный backend deep linking для Smart Insights (5 prompt-insights endpoints + 2 filter endpoints для tags/collections, frontend 5 dedicated pages + filter overlays); Wave 2 — UI polish (русификация, heatmap padding, banner href cleanup, sparkline для constant data, narrative edge cases, x-axis tick format). Обе wave мержатся последовательно поверх existing analytics-redesign, но dependency очень слабая — Wave 2 можно делать параллельно с Wave 1 как только insights endpoints контракт зафиксирован.

**Ключевые технические решения.**

1. **Insights pages — отдельные routes** (`/prompts/insights/:type`) вместо filter param на existing `/prompts`. Чистое разделение insights views от browsing. Pages типизированы по `:type ∈ {unused, duplicates, trending, declining, most-edited}`.
2. **Backend endpoints поверх existing SQL** (`analytics_repo.go`) без дублирования. Новый usecase `usecases/prompt_insights/service.go` оборачивает analytics-repo функции и применяет plan-gating через `insightsForPlan`.
3. **Inline actions через mutations** (TanStack Query). Delete prompt — existing endpoint `DELETE /api/prompts/:id`. Merge duplicates — новый endpoint `POST /api/prompts/:id/merge-with/:other_id`.
4. **GitHub-style heatmap padding** на frontend через date math — backend не меняется. Генерим 28 дат от today-27 до today, мерджим с usage_per_day points.

**Аудитория плана.** Исполнитель + PR-ревьюер. Implementation-ready: для каждого endpoint указан handler/usecase/repo path, для каждого frontend компонента — props.

---

## 2. Архитектурные решения

### Решение 1: Routes — отдельные insight pages vs filter param

- **Решение:** Новые dedicated routes `/prompts/insights/:type` (где type ∈ `unused | duplicates | trending | declining | most-edited`) + filter overlays на existing `/tags?filter=orphan` и `/collections?filter=empty`. Tags/collections используют query param (potentially меньше data чем prompts, не нужна dedicated page).
- **Альтернативы:**
  - (A) **Filter param на existing `/prompts?filter=unused`** — расширяем existing list endpoint. Минус: перегружает list-handler логику (smart-insights mix с regular browse); требует расширения list response shape.
  - (B) **Drawer/Modal на `/analytics`** без navigation. Минус: для actions (Delete, Merge) drawer недостаточно — нужен full-page experience.
  - (C) **Dedicated routes (выбран)** — чистое разделение, маленькие focused pages, легко добавить новые типы insights.
- **Trade-offs.**
  - ✅ Чистая separation of concerns: insights ≠ browsing.
  - ✅ Existing `/prompts` контракт не трогается.
  - ✅ Каждая insight page имеет свой layout, актуальный для типа (duplicates side-by-side, trending — list + sparkline, etc.).
  - ❌ 5 новых routes + 2 filter overlays = 7 новых page entries. Acceptable — это focused functionality.
- **Источник.** Linear «Backlog» views (отдельные routes), Stripe Sigma (saved reports как routes), GitHub Insights — все используют dedicated views.

### Решение 2: Backend — новый usecase vs прямые handlers поверх analytics_repo

- **Решение:** Новый usecase `usecases/prompt_insights/service.go` оборачивает existing `analytics_repo.go` функции. Handler в `delivery/http/prompt/insights_handler.go` (отдельный файл в существующем package) делает thin parsing + auth + dispatch в usecase. Plan-gating через `insightsForPlan(planID)` — переиспользуем helper из Pricing v3.
- **Альтернативы:**
  - (A) **Прямые handlers поверх analytics_repo** без usecase прослойки. Минус: нарушение Clean Architecture (handler владеет SQL); сложнее тестировать.
  - (B) **Расширить existing `usecases/analytics/service.go`** методами `GetInsightsList(type, range, limit)`. Минус: analytics service становится мега-сервисом (там уже compute loop, gated read, gated refresh).
  - (C) **Новый usecase (выбран)** — single responsibility, testable, легко расширяется новыми типами insights.
- **Trade-offs.**
  - ✅ Clean Architecture сохранена.
  - ✅ Plan-gating переиспользует Pricing v3 helper.
  - ✅ Тесты — отдельный файл, не загромождает analytics tests.
  - ❌ Ещё один usecase package (7-й в проекте) — но это разовая стоимость.

### Решение 3: Merge duplicates — flow и backend контракт

- **Решение:** Merge как UI-level операция через 2 действия + новый backend endpoint `POST /api/prompts/:id/merge-with/:other_id`. Frontend modal: показывает 2 промпта side-by-side, юзер выбирает keep-id, другой soft-delete'ится (попадает в корзину). Backend: 1) verify оба промпта принадлежат юзеру, 2) `prompts.SoftDelete(other_id)` через existing trash logic, 3) копирует usage_count с deleted в kept (опционально через flag `?merge_usage=true`).
- **Альтернативы:**
  - (A) **Только Delete без merge** — юзер видит дубликат, кликает Delete. Минус: теряется usage history дубликата.
  - (B) **Полный merge с reassignment** — копирует все связи (tags, collections, usage_log) с deleted на kept. Минус: complex SQL, FK конфликты.
  - (C) **Merge как UI выбор + soft-delete (выбран)** — минимально invasive. Usage transfer опциональный.
- **Trade-offs.**
  - ✅ Reuses existing trash flow.
  - ✅ User control over which prompt to keep.
  - ❌ Не копируем связи (tags/collections) — юзер может потерять metadata. Mitigation: в UI явно warn'аем «Промпт B будет удалён в корзину, теги/коллекции не переносятся».

---

## 3. Изменения в коде

### Создаём (Wave 1)

**Backend:**

- `backend/internal/usecases/prompt_insights/service.go` — `Service` с методами:
  - `ListUnused(ctx, userID, teamID, limit) ([]PromptInsightRow, error)`
  - `ListDuplicates(ctx, userID, teamID, limit) ([]DuplicatePair, error)`
  - `ListTrending(ctx, userID, teamID, limit) ([]PromptInsightRow, error)`
  - `ListDeclining(ctx, userID, teamID, limit) ([]PromptInsightRow, error)`
  - `ListMostEdited(ctx, userID, teamID, limit) ([]PromptInsightRow, error)`
  - `MergePrompts(ctx, userID, keepID, mergeIDs uint) error`
- `backend/internal/usecases/prompt_insights/types.go` — `PromptInsightRow`, `DuplicatePair`.
- `backend/internal/usecases/prompt_insights/errors.go` — `ErrUnknownInsightType`, `ErrPromptsNotOwned`.
- `backend/internal/usecases/prompt_insights/service_test.go` — table-driven по типам.
- `backend/internal/delivery/http/prompt/insights_handler.go` — handlers для 5 prompt-insights endpoints + merge.
- `backend/internal/delivery/http/prompt/insights_handler_test.go`.
- `backend/internal/delivery/http/tag/orphan_handler.go` — `GET /api/tags?filter=orphan` (или новый endpoint `GET /api/tags/orphan`).
- `backend/internal/delivery/http/collection/empty_handler.go` — аналогично для collections.

**Frontend:**

- `frontend/src/pages/prompts/insights/unused.tsx`
- `frontend/src/pages/prompts/insights/duplicates.tsx` — side-by-side pairs с merge modal
- `frontend/src/pages/prompts/insights/trending.tsx`
- `frontend/src/pages/prompts/insights/declining.tsx`
- `frontend/src/pages/prompts/insights/most-edited.tsx`
- `frontend/src/components/prompts/insights/merge-modal.tsx` — duplicate merge UI
- `frontend/src/components/prompts/insights/insight-prompt-row.tsx` — row компонент для list views (title + uses + actions)
- `frontend/src/hooks/use-prompt-insights.ts` — TanStack Query hooks (`useUnusedPrompts`, etc.)
- `frontend/src/api/prompt-insights.ts` — fetcher functions
- Tests для всех новых hooks и compoennts.

### Меняем (Wave 1)

- `backend/internal/app/app.go` — wire-up нового usecase `promptInsights.NewService(promptsRepo, analyticsRepo, ...)` + регистрация роутов.
- `backend/internal/app/routes.go` — добавить `/api/prompts/insights/*`, `/api/tags/orphan`, `/api/collections/empty`, `POST /api/prompts/:id/merge-with/:other_id`.
- `backend/internal/interface/repository/prompt.go` — добавить метод `MergeWith(ctx, keepID, mergeID uint) error` (если идём по soft-delete approach).
- `backend/internal/infrastructure/postgres/repository/prompt_repo.go` — реализация `MergeWith` (soft-delete merge target).
- `frontend/src/App.tsx` — 5 новых routes для prompts/insights.
- `frontend/src/components/analytics/insights-panel.tsx:INSIGHT_META` — обновить `href` на новые routes:
  - `unused_prompts.href` → `/prompts/insights/unused`
  - `possible_duplicates.href` → `/prompts/insights/duplicates`
  - `trending.href` → `/prompts/insights/trending`
  - `declining.href` → `/prompts/insights/declining`
  - `most_edited.href` → `/prompts/insights/most-edited`
  - `orphan_tags.href` → `/tags?filter=orphan`
  - `empty_collections.href` → `/collections?filter=empty`
- `frontend/src/components/analytics/insights-panel.tsx:INSIGHT_META.orphan_tags.title` — `"Orphan-теги"` → `"Теги без промптов"`.
- `frontend/src/pages/tags-page.tsx` — поддержать `?filter=orphan` query param: фильтрует только теги без активных промптов.
- `frontend/src/pages/collections-page.tsx` — поддержать `?filter=empty` query param.

### Меняем (Wave 2 — UI polish)

- `frontend/src/components/analytics/activity-heatmap.tsx` — добавить padding до 28 ячеек через date math:
  - Генерим dates от `today - 27` до `today`.
  - Для каждой даты ищем match в `points`, default `count=0` если нет.
  - Tooltip формат: «12 мая: 5 использований» (русский месяц).
- `frontend/src/components/analytics/narrative-banner.tsx` — убрать `<a href>` wrapper + ArrowRight icon. Banner становится статичным.
- `frontend/src/components/analytics/sparkline.tsx` — если все points равны (`Math.max(points) === Math.min(points)`), рендерим **одну точку** в правом конце (`<circle cx=120 cy=11 r=2.5 fill={color} />`) вместо линии. Это явный «no trend» сигнал.
- `frontend/src/components/analytics/usage-chart.tsx` — `tickFormatter` меняем на `formatDayShort(v)`: «7 мая», «16 мая». Helper в `frontend/src/lib/date-format.ts`.
- `frontend/src/lib/analytics-narrative.ts`:
  - `buildTopModel`: skip если `top.model === ""` ИЛИ `pct === 100 && usage_by_model.length === 1` (единственная неинформативная модель).
  - В page level (analytics.tsx): skip `streakSegment` если `current_streak === 0`.

### Сущности / типы

```go
// backend/internal/usecases/prompt_insights/types.go
type PromptInsightRow struct {
    PromptID  uint      `json:"prompt_id"`
    Title     string    `json:"title"`
    Uses      int       `json:"uses"`
    UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type DuplicatePair struct {
    PromptA    PromptInsightRow `json:"prompt_a"`
    PromptB    PromptInsightRow `json:"prompt_b"`
    Similarity float64          `json:"similarity"`
}
```

```ts
// frontend/src/api/prompt-insights.ts
export interface PromptInsightRow { prompt_id: number; title: string; uses: number; updated_at?: string }
export interface DuplicatePair { prompt_a: PromptInsightRow; prompt_b: PromptInsightRow; similarity: number }
```

### Контракты слоёв

- **HTTP → Usecase:** handlers тонкие, парсят query/path params, dispatch в usecase методы.
- **Usecase → Repo:** usecase оборачивает existing `analytics_repo.go` функции, применяет plan-gating, ничего нового в repo не добавляет (кроме `prompt.MergeWith`).
- **Frontend → Backend:** TanStack Query hooks с `keepPreviousData=true` для плавных переходов.

---

## 4. Модель данных

**Без новых таблиц.** `prompt.MergeWith(keepID, mergeID)` использует existing `deleted_at` field через soft-delete (уже реализован в trash flow). Никаких миграций.

---

## 5. API контракт

### Новые endpoints (Wave 1)

```
GET /api/prompts/insights/unused?limit=50
Response: 200 OK
{ "items": [{ "prompt_id": 42, "title": "Refactor X", "uses": 12, "updated_at": "..." }, ...] }
Errors: 401, 402 (ErrProRequired для Free), 500

GET /api/prompts/insights/duplicates?limit=20
Response: 200 OK
{ "items": [{ "prompt_a": {...}, "prompt_b": {...}, "similarity": 0.92 }, ...] }

GET /api/prompts/insights/trending?limit=10
Response: 200 OK { "items": [PromptInsightRow] }

GET /api/prompts/insights/declining?limit=10
Response: 200 OK { "items": [PromptInsightRow] }

GET /api/prompts/insights/most-edited?limit=10
Response: 200 OK { "items": [PromptInsightRow] }

POST /api/prompts/:id/merge-with/:other_id
Body: {}  (опционально { "merge_usage": true } чтобы скопировать usage_count)
Response: 200 OK { "kept_id": 42, "merged_id": 43 }
Errors: 400 (same prompt), 404 (not owned), 409 (uses transfer conflict)

GET /api/tags/orphan
Response: 200 OK { "items": [{ "id": 1, "name": "...", "uses": 0 }, ...] }

GET /api/collections/empty
Response: 200 OK { "items": [{ "id": 1, "name": "..." }, ...] }
```

### Plan-gating

Все insights endpoints применяют `insightsForPlan(planID)`. Если type не в allowed для plan'а — `HTTP 402 ErrProRequired` (как существующий `/api/analytics/insights`).

### Существующие endpoints — без изменений

`GET /api/prompts`, `GET /api/tags`, `GET /api/collections` — без изменений в контракте.

---

## 6. Зависимости

- **Внешние сервисы:** нет новых.
- **Новые библиотеки:** нет (Lucide, Recharts, shadcn/ui уже всё покрывают).
- **Кросс-командные блокеры:** нет.

---

## 7. План тестирования

### Unit (backend)

- `usecases/prompt_insights/service_test.go` — table-driven по типам: каждый метод (`ListUnused`, etc.) вызывает соответствующий `analytics_repo` метод, plan-gating работает.
- `MergePrompts` — happy path + ownership check + same-id error.

### Integration (backend)

- testcontainers PG: `prompt.MergeWith` корректно soft-delete'ит target, kept остаётся active.
- HTTP handlers integration через httptest: 200/402 для каждого endpoint, JSON shape.

### Unit (frontend)

- `useUnusedPrompts` и др. hooks — query keys уникальные, refetch при изменении range/limit.
- `MergeModal` — рендер с обоими промптами, click на «Оставить A» вызывает mutation с правильными params.
- `ActivityHeatmap` — 28 ячеек всегда (test обновлён).
- `Sparkline` — constant data → dot вместо linии.
- `buildNarrative` — skip empty/uninformative segments.

### E2E

- Manual smoke через Chrome DevTools (как в analytics redesign Task A15).

### Не тестируем

- Recharts internal tick rendering — upstream.
- shadcn/ui Modal accessibility (upstream).

---

## 8. Наблюдаемость

### Метрики

| Имя | Тип | Labels | Назначение |
|---|---|---|---|
| `prompt_insights_request_total` | CounterVec | `type` (unused/duplicates/etc.), `plan` (free/pro/max), `status` (ok/blocked) | Какие insights реально читают |
| `prompt_insights_merge_total` | Counter | — | Сколько merge операций |

### Логи

```
slog.Info("prompt_insights.requested", "type", t, "user_id", uid, "plan", plan, "items_count", len(items))
slog.Info("prompt_insights.merge", "user_id", uid, "kept_id", keep, "merged_id", merge)
slog.Warn("prompt_insights.merge_failed", "user_id", uid, "kept_id", keep, "merged_id", merge, "err", err)
```

### Sentry

- `prompt_insights.merge.failed.{ownership|db}` fingerprint.

### Алерты

- `prompt_insights_request_total{status="blocked"}` rate spike — кто-то пытается dosить insights endpoints без подписки.

---

## 9. План внедрения

### Wave 1 — Backend insights + Frontend pages

| ID | Шаг | Owner | Критерий готовности | Зависит |
|---|---|---|---|---|
| **B1** | `usecases/prompt_insights/types.go` + `errors.go` | backend | files exist, build clean | — |
| **B2** | `usecases/prompt_insights/service.go` — все 5 List* методов + unit tests | backend | `service_test.go` все 5 типов PASS | B1 |
| **B3** | `prompt.MergeWith` repo method + integration test | backend | testcontainers PG: merge soft-deletes target | — |
| **B4** | `usecases/prompt_insights/service.go:MergePrompts` + unit test | backend | unit + ownership/error cases | B3 |
| **B5** | `delivery/http/prompt/insights_handler.go` — 5 GET handlers | backend | handler_test.go integration через httptest | B2 |
| **B6** | `delivery/http/prompt/insights_handler.go:Merge` POST handler | backend | integration test | B4, B5 |
| **B7** | `delivery/http/tag/orphan_handler.go` + register route | backend | curl `/api/tags/orphan` returns items | — |
| **B8** | `delivery/http/collection/empty_handler.go` + register | backend | curl returns items | — |
| **B9** | `app.go` wire-up + `routes.go` registration | backend | server starts с новыми routes | B5, B6, B7, B8 |
| **F1** | `frontend/src/api/prompt-insights.ts` fetcher functions | frontend | TS types match backend JSON | — |
| **F2** | `frontend/src/hooks/use-prompt-insights.ts` — 5 hooks + useMergePrompts mutation | frontend | hooks render data | F1 |
| **F3** | `frontend/src/components/prompts/insights/insight-prompt-row.tsx` — reusable row | frontend | unit test | — |
| **F4** | 5 `frontend/src/pages/prompts/insights/<type>.tsx` pages | frontend | route navigation работает | F2, F3 |
| **F5** | `frontend/src/components/prompts/insights/merge-modal.tsx` + integration в duplicates page | frontend | merge through UI работает | F2, F4 |
| **F6** | `App.tsx` route registration | frontend | 5 routes доступны | F4 |
| **F7** | `tags-page.tsx` — `?filter=orphan` support | frontend | filter работает | — |
| **F8** | `collections-page.tsx` — `?filter=empty` support | frontend | filter работает | — |
| **F9** | `insights-panel.tsx:INSIGHT_META` — обновить hrefs + rename orphan_tags title | frontend | клик на cards ведёт на новые pages | F4, F7, F8 |

### Wave 2 — UI polish

| ID | Шаг | Owner | Критерий готовности | Зависит |
|---|---|---|---|---|
| **U1** | `activity-heatmap.tsx` — padding до 28 ячеек + русский tooltip | frontend | test обновлён, 28 cells всегда | — |
| **U2** | `narrative-banner.tsx` — убрать href / ArrowRight | frontend | banner статичный | — |
| **U3** | `sparkline.tsx` — constant data → dot | frontend | unit test расширен | — |
| **U4** | `usage-chart.tsx` + `lib/date-format.ts` — `DD MMM` ticks | frontend | x-axis показывает «7 мая» | — |
| **U5** | `analytics-narrative.ts` — skip empty topModel / streak=0 | frontend | unit tests расширены | — |
| **U6** | Audit русских строк insights (grep по англицизмам) | frontend | grep шлюзов чисто | — |

**Atomicity.** Каждый Wave 1 шаг ≤ 200 строк диффа (исключение F4 — 5 страниц могут быть бо́льшим). Wave 2 — мелкие (<50 строк каждый).

---

## 10. Rollout и kill-switch

### Стратегия

- **Wave 1 direct prod** после Wave 1 complete. Никаких feature flags — это новые endpoints, не меняем existing.
- **Wave 2 direct prod** — мелкие UI fixes, no risk.

### Feature flags

**N/A** — никаких runtime-флагов. Это additive functionality (new endpoints / fixed UI bugs).

### Kill-switch RTO

- Wave 1: revert PR или временный proxy 404. ~10 минут.
- Wave 2: revert single commit. ~5 минут.

### Communication

- Changelog `/changelog`: «Smart Insights теперь ведут на реальные страницы с действиями».

---

## 11. Документация

- **README** — N/A.
- **ADR** — N/A (нет долгосрочных архитектурных решений; Решения 1-3 в §2 локальны для этой итерации).
- **Runbook** — N/A.
- **OpenAPI** — добавить новые endpoints если есть автогенерация (грепнуть на наличие swagger gen в проекте).
- **CLAUDE.md** — добавить 1-2 строки про `/prompts/insights/:type` routes.

---

## 12. Риски и митигации

### Технические риски

- **Merge prompt — race condition.** Юзер на двух tabs кликает Merge одного и того же дубликата → race на `prompt.SoftDelete`. **Митигация:** existing soft-delete idempotent (повторный DELETE на уже deleted — no-op). UI показывает свежие data через `invalidateQueries`.
- **`/tags?filter=orphan` — performance.** Если у юзера 10k тегов, фильтр по «без промптов» — N+1 risk. **Митигация:** `analytics_repo.OrphanTags` уже использует JOIN + WHERE NOT EXISTS — single query, OK.
- **Frontend deep-link перезапуск чарта.** Юзер кликает «Растущие» → новая page рендерит свой график. Если использует ту же `useInsights` — кэш совпадёт. **Митигация:** разные query keys (`["prompt-insights", "trending"]`).
- **Plan downgrade race.** Юзер на Max странице `/prompts/insights/orphan-tags`, между загрузкой и кликом downgrade'ится в Pro → 402. **Митигация:** error handling — redirect на `/pricing` с toast.

### Pre-mortem: «через 6 месяцев это сломалось»

1. **Юзеры не используют insight pages.** Метрика `prompt_insights_request_total` низкая. **Митигация:** A/B copy CTA («Удалить 5 забытых» vs «Посмотреть») через rotating labels.
2. **Merge удалил не тот промпт.** Юзер кликнул «Оставить A», ушёл, понял что нужно было B. **Митигация:** merged промпт идёт в корзину (existing trash flow) с 30-day restore window.
3. **Backend insights endpoints DoS.** Бесплатные запросы (Free → 402, но 100 RPS rate-limit). **Митигация:** existing IP rate-limit + plan check ранее.

### Известные ограничения

- Merge **не переносит** теги и коллекции — юзер должен сам пересохранить metadata перед удалением. Это известно, документируется в UI warning.
- Trending/declining требуют usage_log данных — если юзер новый, эти страницы пустые.

---

## 13. Метрики успеха

### Бизнес

- **Click-through rate** Smart Insights cards → insight pages: цель >15% за месяц (vs текущие 0% потому что broken).
- **Merge / cleanup actions per user/month**: cancelable метрика — engagement с housekeeping flows.

### Технические

- `/api/prompts/insights/*` p95 latency: < 200ms (analytics_repo SQL уже оптимизирован).
- `prompt_insights_request_total{status="ok"}` > 100/день после launch.
- Bundle size delta: <15 KB gzipped (5 новых pages + hooks).

### Срок измерения

- Wave 1: 30 дней после deploy.
- Wave 2: 1 неделя observability.

---

## 14. Открытые вопросы

1. **Merge usage transfer** — переносить `usage_count` с deleted на kept (`?merge_usage=true`)? **Default:** да, потому что юзер ожидает что после merge статистика обоих промптов будет на kept. **Митигация:** опциональный flag, default true.
2. **Trending/declining empty state copy** — что писать когда новый юзер? Кому: пользователь. Блокирует: нет.
3. **`tags?filter=orphan` — visually highlight orphan-теги в list** или отдельный page? **Default:** filter overlay на existing tags-page. Блокирует: нет.
4. **Sparkline constant data — dot vs skip** — рендерим dot в правом конце или sparkline скрываем entirely если нет тренда? **Default:** dot — даёт минимальную визуальную информацию о value. Не блокер.

---

## Self-check

- [x] **Инвентаризация инструментов.** Evidence: AskUserQuestion (1, scope choice), Bash/Grep для routes inspection, TaskCreate/Update (9 tasks), Read для existing analytics.tsx (свежее из A13).
- [x] **Прочитан релевантный код.** Evidence: `analytics_repo.go:65 UnusedPrompts`, `analytics_repo.go:268 GetTrendingPrompts`, `insights-panel.tsx:INSIGHT_META`, `App.tsx:51 PromptAnalytics + lazy routes`.
- [x] **Внешняя документация.** N/A для этой итерации — используем existing patterns проекта.
- [x] **Архитектурные решения.** Evidence: §2 Решение 1 (3 routes альтернативы), Решение 2 (3 usecase альтернативы), Решение 3 (3 merge альтернативы).
- [x] **Нет over-engineering.** Evidence: переиспользуем existing analytics_repo SQL, не добавляем новых deps, soft-delete merge через existing trash flow.
- [x] **Edge cases / errors.** Evidence: §12 (merge race, plan downgrade, performance), §14 (empty states).
- [x] **Консистентность с проектом.** Evidence: Clean Architecture (usecase + handler), TanStack Query hooks, Lucide icons, shadcn/ui patterns.
- [x] **Критерии готовности.** Evidence: §9 — все шаги с конкретным acceptance.
- [x] **Допущения помечены.** Evidence: 4 Open Questions в §14 с defaults.
- [x] **Scope discipline.** Evidence: явно отделили Wave 1 (deep linking) от Wave 2 (polish). Не делаем merge with full reassignment, не делаем OpenAPI generation.
