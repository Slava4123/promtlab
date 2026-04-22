# Phase 14 — Self Review и план действий

**Дата:** 2026-04-22
**Scope:** audit feed + analytics + daily share limits + branded pages
**Статус:** реализация завершена, self-review проведён, **не готов к merge** (7 High). Нужен pass фиксов.

---

## 0. TL;DR — с чего начать завтра

1. Прочитай это целиком (~10 мин). Ориентир в коде через `file:line` ссылки.
2. **Ничего не закоммичено.** В рабочей копии перемешаны:
   - Твоя прежняя работа (AI-removal + file-import компоненты) — тоже uncommitted.
   - Мои изменения Phase 14 — поверх неё.
3. Запусти автопроверки (см. §5) — они зелёные.
4. Реши что фиксить (см. §7 «Что делать сейчас»).
5. Реши open questions (§6) — особенно GDPR-вопрос про `actor_email` для viewer.
6. После фиксов — коммит (один или несколько) и PR.

---

## 1. Контекст: что строим и зачем

Phase 14 — 4 продуктовых улучшения PromptVault по утверждённому плану в `C:\Users\Пользователь\.claude\plans\reactive-spinning-jellyfish.md`:

1. **Team activity feed** — кто/когда/что менял в команде (промпт/коллекция/шар/роль).
2. **Analytics dashboard** — personal + team метрики с tiered retention.
   - Free = 7 дней истории.
   - Pro = 90 дней + CSV export.
   - Max = 365 дней + Smart Insights + API.
3. **Daily public share limits** — 10/100/1000 шар-ссылок в день (Free/Pro/Max), fixed window UTC. Заменяет прежний cap на «total active».
4. **Branded share pages** (Max only) — кастомизация `/s/:token` под бренд team: логотип, tagline, website, primary color.

### Зачем
- Команды получают прозрачность (кто что менял) и метрики (topPrompts, contributors).
- Монетизация: retention как paywall, daily shares как soft-cap, Smart Insights / branded pages как эксклюзив Max.
- AI-ассистенты через MCP видят `team_activity_feed`, `analytics_summary` — новые tools.

---

## 2. Архитектура: слои и конвенции

Проект использует **Clean Architecture** (см. `promptvault/CLAUDE.md`):

```
cmd/server → app/app.go (DI)
  delivery/http/<feature>/ — handler/request/response/errors (БЕЗ *gorm.DB!)
    ↓ через usecase-интерфейсы
  usecases/<feature>/ — business logic, errors.go, types.go, constants.go
    ↓ через repo-интерфейсы
  interface/repository/ — только интерфейсы
    ↓
  infrastructure/postgres/repository/ — GORM реализации
  models/ — один пакет для всех сущностей
```

**Ключевые правила** (нарушены в нескольких местах Phase 14 — см. §6):
- Handler НИКОГДА не владеет `*gorm.DB` и не делает `users.GetByID` напрямую.
- Доменные ошибки в `usecases/<feature>/errors.go`, НЕ inline в service.go.
- Service setter'ы принимают **интерфейсы** (для DIP + тестируемости).

**Стек:** Go 1.25 + Chi v5 + GORM v2 + PostgreSQL 18 / React 19 + Vite 8 + TanStack Query v5 + Tailwind v4.2 + shadcn/ui + `@base-ui/react`.

---

## 3. Что реализовано (статус)

### 3.1. Миграции БД — 12 файлов (000039–000045, все с down)

| # | Что | Детали |
|---|---|---|
| 000039 | `prompt_versions.changed_by` | Nullable `BIGINT REFERENCES users(id) ON DELETE SET NULL` + backfill (автор = владелец промпта) + индекс. |
| 000040 | `team_activity_log` | Append-only с **row-level** UPDATE-триггером (разрешает менять только `actor_*` поля для GDPR anonymize + FK cascade). Индексы: `(team_id, created_at DESC)`, `(target_type, target_id)`, `(team_id, actor_id, created_at DESC)`. |
| 000041 | `prompt_usage_log.team_id` | Denormalized + backfill + замена `idx_usage_log_prompt` на compound `(prompt_id, used_at DESC)`. |
| 000042 | `share_view_log` | Timeline просмотров для Pro+ owner'ов. Индексы: `(share_link_id, viewed_at DESC)` + `(viewed_at)` для cleanup. |
| 000043 | `user_smart_insights` | UNIQUE на expression `(user_id, COALESCE(team_id, 0), insight_type)`. |
| 000044 | `subscription_plans.max_daily_shares` | `NOT NULL DEFAULT 10` + UPDATE Free=10 / Pro+Pro_yearly=100 / Max+Max_yearly=1000. |
| 000045 | `teams.brand_*` | 4 nullable поля: `brand_logo_url`, `brand_tagline`, `brand_website`, `brand_primary_color`. |

**Все миграции идемпотентны** (`IF EXISTS`/`IF NOT EXISTS`), backfill-запросы защищены `WHERE ... IS NULL`.

### 3.2. Backend — новые файлы

**Models:**
- `models/team_activity.go` — `TeamActivityLog` + константы `Activity*` event-типов + `Target*` target-типов + `AnonymizedActor*`.
- `models/share_view.go` — `ShareView`.
- `models/smart_insight.go` — `SmartInsight` + константы `Insight*`.
- `models/team.go` (**modified**): добавлены поля `BrandLogoURL/BrandTagline/BrandWebsite/BrandPrimaryColor` + тип `BrandingInfo` с методом `IsEmpty()`.
- `models/version.go` (**modified**): добавлено `ChangedBy *uint`.
- `models/subscription.go` (**modified**): добавлено `MaxDailyShares int`.

**Repository интерфейсы + реализации:**
- `interface/repository/team_activity.go` + `infrastructure/postgres/repository/team_activity_repo.go` (new).
- `interface/repository/analytics.go` + `infrastructure/postgres/repository/analytics_repo.go` (new).
- `interface/repository/user.go` (**modified**): + `ListMaxUsers`.
- `interface/repository/team.go` (**modified**): + `GetByID`, `UpdateBranding`.

**Usecases:**
- `usecases/activity/` (new): `service.go`, `types.go` — `Log/LogSafe/ListByTeam/GetPromptHistory/AnonymizeActor`.
- `usecases/analytics/` (new): `service.go` (dashboards), `retention.go` (ClampRange), `insights.go` (compute trend), `cleanup.go` (loop для activity + share_view retention), `insights_loop.go` (daily compute).
- `usecases/team/branding.go` (new): `SetBranding`, `GetBranding`, `GetBrandingForShare`.
- `usecases/prompt/prompt.go` (**modified**): `SetActivity` setter + hooks в Create/Update/Delete + `ChangedBy` в версиях.
- `usecases/collection/collection.go` (**modified**): `SetActivity` + hooks.
- `usecases/share/share.go` (**modified**): `SetActivity` + `SetViewLogger` + `SetBrandingLookup` + hooks + `logShareView` goroutine (Pro+ only) + `CheckDailyShareCreation`/`IncrementShareCreation` интеграция + `ViewMeta` параметр в `GetPublicPrompt`.
- `usecases/team/team.go` (**modified**): `SetActivity` + hooks в AcceptInvitation/UpdateMemberRole/RemoveMember.
- `usecases/quota/quota.go` + `types.go` (**modified**): `CheckDailyShareCreation` + `IncrementShareCreation` + `DailySharesToday` в `UsageSummary` + константа `FeatureShareCreate`.

**HTTP delivery:**
- `delivery/http/analytics/` (new): handler + request + response + errors. 5 endpoints: `personal/teams/prompts/insights/export`.
- `delivery/http/team/activity_handler.go` (new): `GET /api/teams/:slug/activity`.
- `delivery/http/team/branding_handler.go` (new): `GET/PUT /api/teams/:slug/branding`.
- `delivery/http/prompt/handler.go` (**modified**): `SetHistoryDeps` setter + `GetHistory` метод + `enrichVersionsWithActors` helper.

**MCP layer:**
- `mcpserver/interfaces.go` (**modified**): `ActivityService`, `AnalyticsService`.
- `mcpserver/tools.go` (**modified**): 3 новых tool (`team_activity_feed`, `analytics_summary`, `analytics_team_summary`) + расширение `getPromptVersions` (actor через user JOIN).
- `mcpserver/server.go` (**modified**): `NewMCPServer` signature расширен activitySvc+analyticsSvc + обновлён `serverInstructions`.
- `mcpserver/types.go` (**modified**): `VersionResponse` + `ActivityItemResponse` + `AnalyticsSummaryResponse` + `PromptUsageSummary`/`ContributorSummary`.
- `usecases/apikey/constants.go` (**modified**): `KnownTools` + 3 новых.

**DI (app.go, modified):** добавлены `activitySvc`, `analyticsSvc`, `activityCleanupLoop`, `insightsLoop`, `analyticsHandler`, `teamActivityHandler`, `teamBrandingHandler`. Все setter'ы подключены. Routes в MountRoutes.

### 3.3. Frontend — новые файлы

**API клиенты (api/):** `analytics.ts`, `activity.ts`, `branding.ts`. Обновлён `types.ts` (добавлены `MaxDailyShares`, `DailySharesToday`, `ChangedBy*`, `branding`).

**Hooks (hooks/):** `use-analytics.ts`, `use-team-activity.ts`, `use-prompt-history.ts`, `use-branding.ts`.

**Pages:** `analytics.tsx`, `team-analytics.tsx`, `team-activity.tsx`, `team-branding.tsx`.

**Components:**
- `components/analytics/` (9): `metric-card`, `usage-chart`, `top-prompts-table`, `contributors-leaderboard`, `quota-progress`, `range-picker`, `upgrade-gate`, `insights-panel`.
- `components/activity/` (3): `activity-timeline`, `activity-item`, `activity-filters`.
- `components/teams/` (2 new): `branding-form`, `branded-header`.

**Modified:** `App.tsx` (routes), `pages/team-view.tsx` (кнопки Analytics/Activity/Branding), `pages/versions.tsx` (actor info), `pages/pricing.tsx` (новая таблица лимитов), `pages/shared-prompt.tsx` (BrandedHeader), `components/prompts/share-dialog.tsx` (daily progress + upgrade CTA), `lib/mcp-tools.ts` (3 новых tool в whitelist).

**Установлены пакеты:** `recharts`, `date-fns` + shadcn компоненты `tabs`, `progress`, `badge`, `table`, `chart`.

**ВАЖНО:** Tremor (по плану) **не используется** — несовместим с Tailwind v4. Используется `shadcn chart` + Recharts напрямую.

### 3.4. Cron jobs (активны, стартуют в `app.go:StartBackground`)

- `analyticsCleanupLoop` (interval 24h) — retention per-plan SQL (Free 30д / Pro 90д / Max 365д) для `team_activity_log` + `share_view_log`.
- `insightsLoop` (interval 24h) — пересчёт Smart Insights для Max-юзеров через `UserRepository.ListMaxUsers` + `analytics.Service.ComputeInsights`.

### 3.5. Что работает / что не работает

| Фича | Состояние |
|---|---|
| Daily share limits | ✅ End-to-end (backend + UI progress bar + upgrade CTA) |
| Team activity feed — запись | ✅ hooks в 4 usecase |
| Team activity feed — чтение | ✅ HTTP + MCP + frontend timeline (cursor-style через useInfiniteQuery) |
| prompt history (склейка versions + activity) | ✅ |
| Analytics dashboard personal / team | ✅ |
| CSV export | ✅ stream |
| Smart Insights (Max) | ⚠️ **половина**: 3 из 7 типов реализованы (unused, trending, declining); 4 — stub (see `analytics/insights.go:41-46` + M8 в §6) |
| Branded share pages | ✅ (form + public render) |
| LogShareView | ✅ Pro+ only, async |
| Anonymize actor (GDPR) | ⚠️ метод написан, но **никто не вызывает** — нет flow `user.DeleteAccount` в проекте. Готов активироваться когда flow появится. |

---

## 4. Текущий git-статус

- **Ничего не закоммичено.** Рабочая копия содержит:
  - Старые твои правки: AI-removal (удаление OpenRouter), новая фича file-import + markdown editor (untracked `components/prompts/file-import-*`, `markdown-editor.tsx`, `prompt-content.*`, `prompt-split-editor.tsx`, `prompt-view.tsx`, `lib/file-import/`).
  - Мои правки Phase 14 (см. §3).
- Пересечения в файлах: `models/subscription.go`, `models/version.go`, `app.go`, `frontend/src/api/types.ts`, `frontend/src/pages/shared-prompt.tsx` — смешанные изменения.
- Твоё предпочтение было **«всё одним большим коммитом»**.

---

## 5. Как запустить проверки завтра

### Backend (из `promptvault/backend/`)
```bash
go build ./...                              # компилятор
go vet ./...                                # статанализ
go test -short -race -count=1 ./...         # unit + integration (race), без testcontainers
go test -count=1 ./...                      # + integration с testcontainers-go (нужен Docker)
golangci-lint run --timeout=3m              # линтер
```

### Frontend (из `promptvault/frontend/`)
```bash
npm run build                               # tsc + vite build
npm run lint                                # ESLint 9
npx vitest run                              # все тесты
```

### Dev-среда
```bash
# из promptvault/
docker compose -f docker-compose.dev.yml up
# → postgres + api + frontend; миграции применяются автоматически при старте API
```

### Известные локальные nits (не блокеры)

- Backend `golangci-lint`:
  - `branding_handler.go:82` — S1016 (можно заменить struct literal на type conversion).
  - `quota/quota.go:36` — `quotaWarningThreshold` unused (**не моё** — существовало до Phase 14).
- Frontend `npm run lint`:
  - `api/analytics.ts:119` — unused `// eslint-disable-next-line @typescript-eslint/no-explicit-any`.
  - 2 pre-existing warnings в `prompt-editor.tsx` (не мои).

---

## 6. Self-review findings (4 параллельных агента + ручной обзор)

Ни один не поднял Critical. Все findings структурированы по Severity.

### 🟠 High (7) — блокеры merge

#### H1 — `GetHistory` теряет activity errors
**Файл:** `backend/internal/delivery/http/prompt/handler.go:382-398`
**Цитата:**
```go
if h.activity != nil && prompt.TeamID != nil {
    events, err := h.activity.GetPromptHistory(r.Context(), id, 100)
    if err == nil {
        for _, e := range events { ... }
    }
}
```
**Проблема:** DB-ошибка / timeout / миграционный drift → `activity: []` в ответе без любого сигнала. Юзер решит «событий нет», реальная причина — инфраструктура.
**Фикс:** в `else`-ветку `slog.WarnContext(r.Context(), "history.activity.failed", "prompt_id", id, "error", err)`. Не блокируем response (versions остаются).
**Конфиденс:** высокая.

#### H2 — Integer overflow в offset-пагинации activity feed
**Файл:** `backend/internal/delivery/http/team/activity_handler.go:102`
**Цитата:**
```go
filter.Limit = pageSize * page
```
**Проблема:** `pageSize = 200`, `page = 10_000_000` → `Limit = 2 000 000 000`. Repo клампит до 200 внутри, но на handler-уровне защиты нет. Если repo-реализация изменится — DoS.
**Фикс:** `if page > 1000 { page = 1000 }` перед вычислением. Или (лучше) полный редизайн: cursor-based и для HTTP, как в MCP.
**Конфиденс:** средняя.

#### H3 — Prefix-matching для plan_id (bug-in-waiting)
**Файлы (5 мест):**
- `backend/internal/usecases/share/share.go:319` — `planID[:3] == "pro" || planID[:3] == "max"`.
- `backend/internal/usecases/analytics/retention.go:16-20` — `strings.HasPrefix(planID, "pro")`.
- `backend/internal/usecases/team/branding.go:48` — `strings.HasPrefix(owner.PlanID, "max")`.
- `backend/internal/delivery/http/analytics/handler.go:92` — `strings.HasPrefix(user.PlanID, "max")` (в handler! см. H5).
- SQL: `analytics_repo.go:244-247`, `team_activity_repo.go:116-120` — `LIKE 'pro%'` и `LIKE 'max%'`.

**Проблема:** `planID = "professional"` или `"proto"` → считается Pro. `"maximus"` → считается Max. Сейчас в БД только `free, pro, pro_yearly, max, max_yearly` — **не эксплойт-в-продакшене**, но bug в защите: любой новый план с префиксом `pro*`/`max*` получит платные фичи или неправильный retention.
**Фикс:**
- Go: whitelist — `planID == "pro" || planID == "pro_yearly"`.
- SQL: `IN ('pro', 'pro_yearly')` или `LIKE 'pro' OR LIKE 'pro\_%' ESCAPE '\'`.
**Альтернатива:** ввести helper `isPaid(planID string) bool` / `tier(planID string) "free"|"pro"|"max"` в одном месте (например `models/subscription.go` или `usecases/subscription/plan.go`), использовать везде — исправляется как H3 + устраняется дублирование логики.
**Конфиденс:** высокая.

#### H4 — Доменные ошибки inline (нарушение конвенции)
**Файлы:**
- `backend/internal/usecases/activity/service.go:14-18` — `ErrMissingTeam/ErrMissingEventType/ErrMissingActor` inline.
- `backend/internal/usecases/analytics/service.go:13-16` — `ErrForbidden/ErrNotFound` inline.

**Проблема:** `promptvault/CLAUDE.md` (строки 133-143, 194) требует вынесения доменных ошибок в `usecases/<feature>/errors.go`. Все существующие пакеты (share/team/prompt/collection/quota/trash/…) следуют паттерну.
**Фикс:** создать `usecases/activity/errors.go` и `usecases/analytics/errors.go`, переместить `var (...)` блоки.
**Конфиденс:** высокая.

#### H5 — Handler владеет plan-логикой (layer violation)
**Файл:** `backend/internal/delivery/http/analytics/handler.go:88-92, 122-126`
**Цитата:**
```go
func (h *Handler) Insights(...) {
    user, err := h.users.GetByID(r.Context(), userID)
    ...
    if !strings.HasPrefix(user.PlanID, "max") {
        respondTierRequired(w, "insights", user.PlanID, "Max")
        return
    }
    ...
}
```
**Проблема:** Handler владеет `users repo.UserRepository` и делает prefix-check плана — это доменная логика, должна быть в service. `CLAUDE.md:191` — "Handler НИКОГДА не владеет *gorm.DB" — технически тут через repo-интерфейс, но идея та же: handler не должен решать "Max или нет".
**Фикс:**
- В `usecases/analytics/` ввести `ErrMaxRequired` и `ErrProRequired`.
- `analytics.Service.GetInsights` сам проверяет план — возвращает `ErrMaxRequired`.
- `analytics.Service.ExportCSV` возвращает `ErrProRequired` для Free.
- Handler только маппит → `respondTierRequired` (уже есть в `delivery/http/analytics/errors.go`).
**Конфиденс:** высокая.

#### H6 — `BrandingLookup` как typed callback вместо интерфейса
**Файл:** `backend/internal/usecases/share/share.go:30-34, 87-89`
**Цитата:**
```go
type BrandingLookup func(ctx context.Context, teamID uint) (*models.BrandingInfo, error)
func (s *Service) SetBrandingLookup(fn BrandingLookup) { s.brandingLookup = fn }
```
**Проблема:** Все остальные setter'ы (`SetEmail`, `SetActivity`, `SetViewLogger`, `SetEmailNotifier`) принимают **интерфейсы** или `*Service`. Typed function alias уникален → тестам приходится закрывать замыкание вместо mock'а интерфейса.
**Фикс:**
```go
type BrandingProvider interface {
    GetBrandingForShare(ctx context.Context, teamID uint) (*models.BrandingInfo, error)
}
func (s *Service) SetBrandingLookup(p BrandingProvider) { s.brandingProvider = p }
```
`teamSvc` уже удовлетворяет интерфейсу — вызов в `app.go` **не меняется**.
**Конфиденс:** средняя.

#### H7 — `SetViewLogger` — 3 ортогональных зависимости, `plans` dead
**Файл:** `backend/internal/usecases/share/share.go:79-83`
**Цитата:**
```go
func (s *Service) SetViewLogger(analytics repo.AnalyticsRepository, users repo.UserRepository, plans repo.PlanRepository) {
    s.viewLogger = analytics
    s.users = users
    s.plans = plans
}
```
**Проблема:** `plans` **не используется** в `logShareView` (проверяется `users.PlanID`). Нарушение ISP, dead dependency.
**Фикс:**
- Убрать `plans` из сигнатуры.
- Или (лучше) вынести логику "pro+ only" в `analytics.Service.LogShareViewIfPaid(ctx, shareLinkID, ownerID, meta)` — share вызывает одним-методом, pro+check внутри analytics.
**Конфиденс:** высокая.

### 🟡 Medium (9)

#### M1 — `has_more` некорректно при точной границе
**Файл:** `backend/internal/delivery/http/team/activity_handler.go:127`
**Цитата:** `"has_more": len(events) == pageSize`
**Проблема:** если в БД ровно `pageSize` событий — `has_more = true` → клиент делает лишний запрос, возвращается пустой массив.
**Фикс:** sentinel — запрашивать `pageSize+1`, возвращать первые `pageSize`, `has_more := raw > pageSize`.

#### M2 — `IncrementShareCreation` fail → `Warn`, не `Error`
**Файл:** `backend/internal/usecases/share/share.go:147`
**Цитата:** `slog.Warn("share.quota.increment_failed", ...)`
**Проблема:** revenue-leak — юзер получил ссылку, счётчик не увеличился. SRE не увидит на уровне Warn.
**Фикс:** `slog.Error` + Prometheus counter `share_quota_increment_failed_total` (если есть метрика-система).

#### M3 — `ComputeInsights` молча пропускает repo-fail
**Файл:** `backend/internal/usecases/analytics/insights.go:26-39`
**Цитата:**
```go
unused, err := s.analytics.UnusedPrompts(...)
if err == nil && len(unused) > 0 { s.upsertSafe(...) }
```
**Проблема:** при repo-fail ни одна ветвь не логирует, функция возвращает `nil`. Loop в `insights_loop.go` пишет `ok++`, хотя ни один insight не посчитан.
**Фикс:** каждая из 3 ветвей → `if err != nil { slog.WarnContext(ctx, "insights.<type>.failed", "err", err, "user_id", userID); continue }`.

#### M4 — invalid query params silently ignored
**Файл:** `backend/internal/delivery/http/team/activity_handler.go:67-86`
**Цитата:** `if v, perr := strconv.ParseUint(s, 10, 32); perr == nil { ... }` — и `time.Parse`.
**Проблема:** `?from=invalid_date` или `?actor_id=abc` игнорируются silently → UI думает фильтр применён, backend вернул нефильтрованный feed.
**Фикс:** `return httperr.BadRequest("неверный формат <field>: ожидается ...")` при err.

#### M5 — GDPR: `actor_email` виден viewer'у (open question, см. §7)
**Файлы:** `backend/internal/mcpserver/tools.go:1085`, `backend/internal/delivery/http/team/activity_handler.go:38`
**Проблема:** viewer получает реальные email коллег в MCP и HTTP responses — нарушает минимизацию данных (GDPR/152-ФЗ).
**Фикс (если решим скрывать):** возвращать `actor_email` только owner/editor; viewer получает только `actor_name` + `actor_id` (или маскированный `a***@acme.com`).
**Требует бизнес-решения.**

#### M6 — per-user lookup fails silently в enrichVersions
**Файлы:**
- `backend/internal/delivery/http/prompt/handler.go:325-348` (`enrichVersionsWithActors`).
- `backend/internal/mcpserver/tools.go:989-993` (в `getPromptVersions`).

**Цитата:** `if u, err := h.users.GetByID(ctx, uid); err == nil { actorMap[uid] = u }` — ошибка тихо потеряна.
**Проблема:** DB-hiccup → UI показывает «неизвестный автор» для части строк, без признака ошибки.
**Фикс:** `slog.Warn` на первую ошибку с count + добавить в response флаг `actors_partial: true`.

#### M7 — Content-Disposition без sanitize (хрупко)
**Файл:** `backend/internal/delivery/http/analytics/handler.go:171`
**Цитата:** `w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")`
**Проблема:** `filename` формируется из `dash.Range` (clamp'нутый, safe **сейчас**) + `teamID64` (uint, safe). Реальной атаки нет, но хрупкий паттерн — если `parseRange` изменится, Response Splitting возможен.
**Фикс:** `mime.FormatMediaType("attachment", map[string]string{"filename": filename})`.

#### M8 — Smart Insights half-feature (MCP обещает, 4 из 7 stub)
**Файлы:**
- `backend/internal/usecases/analytics/insights.go:41-46` — комментарий "MOST_EDITED / POSSIBLE_DUPLICATES / ORPHAN_TAGS / EMPTY_COLLECTIONS оставлены заглушками".
- `backend/internal/models/smart_insight.go` — все 7 констант `Insight*` экспортированы.
- `backend/internal/mcpserver/server.go:49-54` — `serverInstructions` обещает клиенту insights.
- Frontend `components/analytics/insights-panel.tsx` — рендерит все 7 типов из labels.

**Проблема:** фича «обещана», но end-to-end для 4 типов не работает. UI покажет пустой раздел без объяснения причины.
**Фикс (на выбор):**
a) Реализовать оставшиеся 4 (Levenshtein для duplicates — через `pg_trgm`/`fuzzystrmatch`; остальные — простые SQL).
b) Спрятать недоделанные InsightType за feature-flag.
c) Убрать из `models/smart_insight.go` 4 неиспользуемых константы и из `LABELS` в `insights-panel.tsx`, чтобы фронт не знал о них.
**Требует бизнес-решения.**

#### M9 — hot-path `users.GetByID` на каждый публичный просмотр
**Файл:** `backend/internal/usecases/share/share.go:289-315` (`logShareView`).
**Цитата:** `owner, err := s.users.GetByID(bgCtx, ownerID)`.
**Проблема:** `/s/:token` — hot path. `users.GetByID` — PK lookup, но каждый view запускает лишний SELECT. GORM preload уже загрузил `p.User` (см. share.go:259 — `p.User.Name`/`p.User.AvatarURL`).
**Фикс:** расширить preload до `User.PlanID` и читать из уже загруженного объекта. Убрать `s.users` из `SetViewLogger` (связан с H7).

### 🟢 Low (4)

- **L1** `share/share.go:326-329` `truncateString` по байтам ломает UTF-8 на границе 500. PostgreSQL отклонит INSERT с `invalid byte sequence for encoding "UTF8"` → `slog.Warn` → потеря записи. Фикс: `utf8.ValidString` + движение назад до валидной границы.
- **L2** `frontend/src/components/teams/branded-header.tsx:24` — нет frontend scheme-validation для `website`. Backend защищает (HTTPS-only), но defence-in-depth: `branding.website?.startsWith("https://") ? ... : undefined`.
- **L3** `backend/internal/delivery/http/team/branding_handler.go:82` — staticcheck S1016: `BrandingResponse{...}` → `BrandingResponse(req)` (type conversion).
- **L4** `frontend/src/api/analytics.ts:119` — unused `// eslint-disable-next-line @typescript-eslint/no-explicit-any`. Удалить.

### Nit / Open Observations

- **N1** `share.go:319` `planID[:3] == "pro"` вместо `strings.HasPrefix` — несовместимо со стилем остального codebase (см. H3).
- **N2** `team/branding.go:32` — hex-regex дублируется с `validator:"hexcolor"` на request (defence-in-depth, не баг).
- **N3** `insights_loop.go:74` — единственный TODO в Phase 14 (обоснован).
- **N4** Backfill `max_daily_shares = 10` для Free при миграции: юзеры, создающие >10 ссылок сегодня, упрутся на 11-й попытке в день миграции (counter начинается с 0, не зависит от existing share_links — не блокер).

---

## 7. Open Questions (нужно твоё решение)

### Q1 — GDPR: `actor_email` в team activity feed для viewer
- **Вариант A (текущий):** все члены видят email — максимум прозрачности, но нарушает минимизацию GDPR.
- **Вариант B:** viewer получает только `actor_name` + `actor_id`; email только owner/editor.
- **Вариант C:** маскирование (`a***@acme.com`) для всех.
- **Рекомендация:** B — не ломает UX (имена видны), закрывает GDPR.

### Q2 — Smart Insights (M8)
- **A:** Реализовать оставшиеся 4 типа — ~1 день работы (SQL+Levenshtein).
- **B:** Скрыть за feature-flag до реализации.
- **C:** Удалить 4 неиспользуемые константы из модели и UI.
- **Рекомендация:** B (feature-flag) — меньше кода править, фичу потом доделать.

### Q3 — Observability
- Сейчас в Phase 14 только `slog`, нет Prometheus counter'ов и нет alert'ов. В проекте есть Sentry (Glitchtip), но метрик нет.
- **Вопрос:** добавить метрики в Phase 14 или отдельный тикет?
- **Рекомендация:** отдельный тикет на уровне проекта — Phase 14 не должна тянуть metrics-инфраструктуру.

### Q4 — Unit-тесты
- Для новых пакетов (`activity/`, `analytics/`, `team/branding`, `delivery/http/analytics/`) — **0 тестов**.
- **A:** Добавить критичные (ClampRange, SetBranding, activity.Log, append-only trigger через testcontainers) — ~4 часа перед merge.
- **B:** Follow-up тикет, merge с пометкой «known debt».
- **Рекомендация:** A для Critical gaps #1–#3 из test-coverage отчёта (ClampRange, SetBranding, activity.Log). Остальное — follow-up.

### Q5 — Feature flags (план упоминал)
- `PHASE14_ANALYTICS_ENABLED`, `PHASE14_ACTIVITY_FEED_ENABLED`, `PHASE14_DAILY_SHARE_LIMIT_ENABLED`, `PHASE14_BRANDING_ENABLED`.
- **A:** Реализовать для postепенного rollout.
- **B:** Пропустить (риск rollback ниже, чем писать флаги).
- **Рекомендация:** B — фича детерминистична, миграции reversible.

---

## 8. Что делать сейчас — рекомендованный план фиксов

### Группа A — Быстрые фиксы, точно блокеры (~1.5 часа)

| # | What | File | Effort |
|---|---|---|---|
| H3 | Prefix-matching → whitelist exact | 5 мест (см. §6) | 30 мин |
| H4 | Вынести `errors.go` в activity/ и analytics/ | 2 файла new | 10 мин |
| H6 | `BrandingLookup` → `BrandingProvider` interface | share.go | 10 мин |
| H2 | `if page > 1000 { page = 1000 }` | activity_handler.go | 2 мин |
| L3 | S1016 fix | branding_handler.go:82 | 1 мин |
| L4 | Unused eslint-disable | analytics.ts:119 | 1 мин |

### Группа B — Silent-failure + UX (~1 час)

| # | What | File |
|---|---|---|
| H1 | slog.Warn для activity err | prompt/handler.go:GetHistory |
| M1 | Sentinel-пагинация `pageSize+1` | activity_handler.go |
| M2 | `Error` + metric placeholder | share.go:147 |
| M3 | slog.Warn в каждой ветке | insights.go:26-39 |
| M4 | BadRequest на invalid params | activity_handler.go:67-86 |

### Группа C — Архитектура (~1.5 часа)

| # | What |
|---|---|
| H5 | Перенести plan-check из handler в service (`ErrMaxRequired`/`ErrProRequired`) |
| H7 | Убрать `plans` из `SetViewLogger` (или вынести в analytics-фасад) |
| M7 | `mime.FormatMediaType` для Content-Disposition |
| M9 | Preload `User.PlanID` в `share.GetPublicPrompt` |
| M6 | Флаг `actors_partial` в response |

### Группа D — Откладываемые (follow-up тикеты)

- **M5** GDPR-решение по actor_email (требует Q1 answer).
- **M8** Smart Insights (требует Q2 answer).
- **L1** UTF-8 truncate.
- **L2** Frontend scheme validation.
- **Observability** (Q3).
- **Unit-тесты** (Q4) — минимум ClampRange/SetBranding/activity.Log перед merge.
- **Feature flags** (Q5).

### Ориентиры по времени

- **Минимум для зелёного PR (A только):** ~1.5 часа. Закрывает все блокеры конвенций/безопасности.
- **Хороший PR (A+B):** ~2.5 часа. Плюс silent-failure fixes.
- **Полный PR (A+B+C+Unit-тесты Critical):** ~6 часов. Плюс архитектурные фиксы + regression-тесты Critical gaps.

---

## 9. Checklist для PR (завтра)

- [ ] Q1-Q5 решены (см. §7)
- [ ] Группа A фиксов применена, `go build + go vet + go test -race + lint` зелёные
- [ ] Группа B (опционально) применена
- [ ] Группа C (опционально) применена
- [ ] Unit-тесты Critical gaps (опционально; решение Q4)
- [ ] Перезапустил `npm run build + lint + vitest` — зелёные
- [ ] Проверил git diff: нет merge-маркеров / debug-артефактов / секретов
- [ ] Закоммитил — либо один большой коммит (как хотел), либо разбил по темам (Phase 14 + AI-removal + file-import отдельно)
- [ ] Открыл PR — черновик description ниже

---

## 10. Черновик PR description

**Title:** `[PROMPT-14] Phase 14: team activity feed + analytics + daily shares + branded pages`

**Summary:**
- Team activity feed: кто/когда/что менял в команде через `team_activity_log` (append-only с row-level триггером).
- Analytics dashboard: personal + team с tiered retention (Free 7д / Pro 90д / Max 365д) + Smart Insights (Max) + CSV export.
- Daily share limits: Free 10 / Pro 100 / Max 1000 в день, fixed window UTC. Заменяет total-active-cap.
- Branded share pages: Max-команды кастомизируют `/s/:token` (логотип, tagline, website, color).

**How to test:**
- `docker compose -f docker-compose.dev.yml up` — поднять dev stack (автоматически мигрирует 000039-000045).
- Backend: `go test -short -race ./...` — все зелёные.
- Frontend: `npm run test` — 178 passed.
- Manual smoke test через UI: `/analytics`, `/teams/:slug/analytics`, `/teams/:slug/activity`, `/teams/:slug/branding`, обновлённый `/pricing`, ShareDialog daily progress.
- MCP: `list_tools` должен показать `team_activity_feed`, `analytics_summary`, `analytics_team_summary`.

**Risks/rollout:**
- Миграции реверсивны (`down.sql` для каждой).
- Существующие endpoints backward-compat (optional поля).
- `daily_shares` заменяет semantics total-active → юзеры с >10 активными ссылками smarth сохраняют их, просто 11-е создание в день = 429.
- **Known issues (follow-up tickets):** M5 GDPR actor_email, M8 Smart Insights half-feature, Unit tests Critical gaps — см. `docs/PHASE14_SELF_REVIEW.md`.

**Related:**
- Плана: `C:\Users\Пользователь\.claude\plans\reactive-spinning-jellyfish.md` (phase 14 approved).
- Self-review: `docs/PHASE14_SELF_REVIEW.md` (этот файл).

---

## 11. Где что лежит (быстрая навигация)

### Plan
- `C:\Users\Пользователь\.claude\plans\reactive-spinning-jellyfish.md` — полный plan Phase 14 (A+B+C+Branded).

### Project docs
- `promptvault/CLAUDE.md` — правила архитектуры, стек, команды.
- `promptvault/docs/PHASE14_SELF_REVIEW.md` — **этот файл**.
- Соседние: `promptvault/docs/PLAN.md`, `FEATURES.md`, `MONETIZATION.md`, `MCP.md`, `SUBSCRIPTION_PLAN.md`.

### Ключевые пути
- Backend:
  - Миграции: `backend/internal/infrastructure/postgres/migrations/000039*.sql` … `000045*.sql`.
  - Usecases: `backend/internal/usecases/{activity,analytics}/`, `usecases/team/branding.go`.
  - HTTP: `backend/internal/delivery/http/analytics/`, `.../team/activity_handler.go`, `.../team/branding_handler.go`, `.../prompt/handler.go:GetHistory`.
  - MCP: `backend/internal/mcpserver/{tools.go,interfaces.go,types.go,server.go}`.
  - DI: `backend/internal/app/app.go`.

- Frontend:
  - Pages: `frontend/src/pages/{analytics,team-analytics,team-activity,team-branding}.tsx`.
  - Components: `frontend/src/components/{analytics,activity,teams/branding-form,teams/branded-header}/`.
  - Hooks: `frontend/src/hooks/use-{analytics,team-activity,prompt-history,branding}.ts`.
  - API: `frontend/src/api/{analytics,activity,branding}.ts`.
  - Обновлены: `App.tsx`, `pages/{team-view,versions,pricing,shared-prompt}.tsx`, `components/prompts/share-dialog.tsx`, `lib/mcp-tools.ts`, `api/types.ts`.

### Task list (персистентный)

Полная реализация → completed tasks (16-24):
- 16 Phase B.1 HTTP handlers ✅
- 17 Phase B.2 LogShareView ✅
- 18 Phase B.3 MCP tools ✅
- 19 Phase B.4 UserRepository.ListMaxUsers ✅
- 20 Phase C.1 shadcn chart + API client + hooks ✅
- 21 Phase C.2 /analytics personal ✅
- 22 Phase C.3 Team analytics + activity timeline ✅
- 23 Phase C.4 History tab + ShareDialog + pricing ✅
- 24 Phase D Branded share pages ✅

Phase A (backend foundation) — completed ранее (миграции + repo + сервисы + cron).

---

## 12. Что делать завтра пошагово

```
1. Прочитать этот файл (§0 TL;DR → §6 findings → §7 open questions → §8 план фиксов).
2. Решить Q1–Q5 (§7).
3. Запустить проверки (§5) — убедиться что всё зелёное.
4. Сделать Группу A фиксов (§8, ~1.5 часа):
   - Grep prefix patterns в коде и заменить на whitelist.
   - Вынести errors.go.
   - BrandingLookup → BrandingProvider.
   - page cap.
   - 2 lint-nits.
5. Перезапустить проверки.
6. (опц) Группа B, C, Unit-тесты.
7. git add + commit (один большой или несколько — на твой выбор).
8. git push + открыть PR с description из §10.
9. Создать follow-up тикеты на Группу D.
```

**Удачи. Вопросы — обращайся, full контекст в этом файле.**
