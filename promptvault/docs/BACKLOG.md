# Backlog — Post Phase 14.3+

**Дата:** 2026-04-23
**Предыдущий CI:** `2cad5ad` (Phase 14.3 wave 1)
**Prod:** `promtlabs.ru` актуален.

В Phase 14.3 wave 1 закрыты 16+ пунктов (см. раздел ниже).
В Phase 14.3 wave 2 закрыты #17 (SMTP notifier), частично #14 (adapters вынесены),
пилот #9 (AnalyticsFilter + helpers).

---

## ✅ Закрыто в Phase 14.3 (этот коммит)

- #2 **Integration-тесты AnalyticsRepository** — `analytics_repo_test.go` (4 метода + 2 team-scope варианта, паттерн audit_repo_test.go).
- #3 **HTTP-тесты analytics** — `request_test.go` (parseRange), `handler_test.go` (breadcrumb no-hub). Полный mock-service — deferred (требует извлечь `type analyticsService interface` в handler-пакет).
- #4 **Frontend page-level тесты** — `prompt-analytics.test.tsx`, `team-analytics.test.tsx`, `use-analytics-filter.test.ts`, `url.test.ts`.
- #5, #6 **Prometheus + Sentry breadcrumbs** — пакет `infrastructure/metrics` (3 counter), endpoint `/metrics` за `SERVER_METRICS_ENABLED`. Breadcrumbs в `RefreshInsights` и `Export`.
- #7 **Alerts** — `infra/prometheus/alerts.yaml` (InsightsComputeLoopStalled/Dead, CleanupLoopStalled, ShareQuotaIncrementLeak). Проверка: `promtool check rules`.
- #8 **M8 Full Smart Insights** — миграция `000048_analytics_m8.sql` (pg_trgm + GIN index). 4 метода `AnalyticsRepository`: MostEditedPrompts, PossibleDuplicates (similarity ≥ 0.8), OrphanTags, EmptyCollections. Раскрыт stub `insights.go` за флагом `experimentalInsights` (default false).
- #11 **API-key доступ к analytics** — `analytics_read` псевдо-tool в `apikey/constants.go KnownTools` + `frontend/src/lib/mcp-tools.ts`.
- #12 **L2 HTTPS validation** — `frontend/src/lib/url.ts isSafeHttpsUrl` + `branded-header.tsx` условный рендер `<a>`.
- #13 **actor_email маска (вариант C)** — `delivery/http/utils/mask.go MaskEmail` + `mcpserver/mask.go` (zero-dep копия); применён в `activity_handler.go` и `mcpserver/tools.go team_activity_feed`.
- #16, #17, #18 **Email-уведомления каркас** — миграция `000049_insight_notifications.sql` (таблица + `users.insight_emails_enabled` opt-in по ФЗ-152). Интерфейс `InsightsNotifier` + `NoopNotifier` (default). `Service.SetNotifier` hook. Реальная SMTP-реализация — TODO.
- #28 **MCP quota lint-test** — `quota_lint_test.go` AST-парсит `tools.go` и проверяет что все 13 write/destructive handler'ов вызывают `t.checkMCPQuota`. Полноценный decorator оказался хрупким в SDK v1.5 generics — lint даёт ту же защиту без runtime-overhead.
- #29, #30 **Soft-delete regression** — `prompt_repo_softdelete_test.go`. Подтверждено что `DeletedAt gorm.DeletedAt` работает корректно (предположение Explore #3 о уязвимости опровергнуто).
- #31 **IsUnlimited / sentinel -1 cleanup** — удалена `models.IsUnlimited`, упрощён `isWithinLimit` и `over()` в quota.go.
- **Drill-down (#9) каркас** — `AnalyticsFilter` struct в `interface/repository/analytics.go` + frontend `useAnalyticsFilter` хук (URL-params ?tag/?collection). Миграция всех 16 методов на Filter-struct — deferred.

---

## ✅ Закрыто в Phase 14.3 wave 2

- #17 **SMTP insights notifier** — `EmailInsightsNotifier` в `infrastructure/email/insights_notifier.go`, rate-limit 1/неделя через `insight_notifications` repo, `InsightNotificationRepository` + GORM impl, `email.Service.SendInsightsDigest` plain-text. Opt-in через `users.insight_emails_enabled` + `PATCH /api/auth/notifications/insights`. NoopNotifier заменён на реальный в app.go.
- #14 **app.go split** — 4 адаптера вынесены в `app/adapters.go`. MountRoutes (335 строк) — в `app/routes.go`. app.go снижен с 845 до 439 строк.
- #9 пилот **AnalyticsFilter** helpers — `HasTag()`/`HasCollection()`.

---

## ✅ Закрыто в Phase 14.3 wave 3

- #9b **Drill-down по тегу/коллекции** — 5 новых filter-aware методов в `AnalyticsRepository`
  (UsagePerDayFiltered, TopPromptsFiltered, PromptsCreatedPerDayFiltered,
  PromptsUpdatedPerDayFiltered, UsageByModelFiltered) с приватным `applyPromptFilter`
  helper'ом (JOIN prompt_tags / prompt_collections). Service.GetPersonalDashboardFiltered,
  handler Personal парсит `?tag_id=:id&collection_id=:id`. Frontend:
  `PersonalAnalyticsFilter` + `fetchPersonalAnalytics(range, filter)` +
  `usePersonalAnalytics(range, filter)`. QueryKey содержит filter'ы для
  корректной инвалидации кэша.
- #17b **UI settings страница** — `pages/settings/notifications.tsx` с toggle
  «Smart Insights digest», добавлена в `_nav-config.ts` и роутер `App.tsx`.
  Хук `useSetInsightEmails` в `use-settings.ts`, User.insight_emails_enabled
  в типах, инвалидация `["me"]` при переключении.

---

## 🔒 Осталось (блокировано внешним)

### 1. `github.com/docker/docker` CVE (high + medium)
- **Статус:** `go get latest` не поднимает — `v29.x` ещё не в Go-modules.
- **Когда станет доступно:** `go get github.com/docker/docker@v29.x && go mod tidy` + прогон тестов.

### 8b. Smart Insights feature-flag toggle в prod
- **Что сделать:** установить `ANALYTICS_EXPERIMENTAL_INSIGHTS=true` в `.env.prod` после проверки что managed Postgres поддерживает `pg_trgm` (миграция 000048 прошла).
- **Мониторинг:** 24 часа через `analytics_insights_refresh_total{result="success"}`.

### 17b. Insights digest UI settings
- **Что сделать:** UI-toggle в `/settings/notifications` — checkbox «Email-уведомления по Smart Insights» с описанием ФЗ-152 opt-in. Вызывает `PATCH /api/auth/notifications/insights`.
- **Оценка:** 1 час (endpoint готов, osталась React-страница).

---

## 🔧 Остаётся (не блокировано, но отложено)

### 9c. Drill-down для TeamDashboard
Сейчас filter-aware методы подключены только к PersonalDashboard. TeamDashboard
(`/api/analytics/teams/{id}`) и export остаются без drill-down. Требует
аналогичного `GetTeamDashboardFiltered` в service.go + параметров в handler
Team/Export. Оценка: 1 час, риск низкий (паттерн уже есть).

### 14c. Лёгкая lifecycle.go экстракция
`StartBackground`/`Shutdown` остаются в app.go. Можно вынести в `app/lifecycle.go`
рядом с adapters/routes когда app.go в следующий раз станет shy > 500 строк.

### 15. MCP `checkMCPQuota()` как middleware
AST-lint-test (квmerged-тест) решает забываемость — декоратор дешевле не становится. Оставлено ручным.

### 16. Soft-delete audit (дополнительный)
Проверено только для `Prompt`. Остальные soft-deletable модели (`Collection`) — ручная проверка при изменении их `*_repo_test.go`.

---

## 📉 Технический долг (знаем, не горит)

- Phase 14 Analytics `Totals` для Team Dashboard не считает `ShareViews` (share-ссылки принадлежат юзеру, не команде). Визуально поле отсутствует — не баг; если когда-то захотим team share-views retention — учесть в `buildAnalyticsSummary`.

---

## 🧭 Метрика качества кода (обновлено)

| Слой | Unit | Integration | HTTP |
|------|------|-------------|------|
| `usecases/analytics` | ✅ | ✅ **добавлено 14.3** | ◐ |
| `usecases/activity` | ✅ | ❌ | ❌ |
| `usecases/team/branding` | ✅ | — | ❌ |
| `usecases/subscription` | ✅ | ✅ | ✅ |
| `usecases/share` | ✅ | — | ❌ |
| `usecases/quota` | ✅ | — | — |
| `delivery/http/analytics` | ✅ (+ breadcrumb) | — | ◐ (parseRange + breadcrumb; full handler — deferred) |
| `delivery/http/team` | ❌ | — | ❌ |
| `delivery/http/subscription` | ✅ | — | — |
| `delivery/http/utils` | ✅ (MaskEmail) | — | — |
| `infrastructure/metrics` | ✅ **новый** | — | — |
| `infrastructure/postgres/repository` | — | ✅ (audit, analytics, soft-delete) | — |
| `mcpserver` | ✅ (+ quota lint-test) | — | — |
| `frontend/lib/url` | ✅ **новый** | — | — |
| `frontend/hooks/use-analytics-filter` | ✅ **новый** | — | — |
| `frontend/pages/prompt-analytics` | ✅ **новый** | — | — |
| `frontend/pages/team-analytics` | ✅ **новый** | — | — |
