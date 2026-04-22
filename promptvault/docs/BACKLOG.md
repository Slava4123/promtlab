# Backlog — Post Phase 14.3

**Дата:** 2026-04-23
**Предыдущий CI:** `e0c5ea3` (UTF-8 truncate в share)
**Prod:** `promtlabs.ru` актуален, включая Analytics v2 + pricing fix.

Ранее документ содержал 16 пронумерованных пунктов. В Phase 14.3 закрыты
#2, #3 (частично), #4, #5, #6, #8, #12, #13, #15, #16 (каркас), #18,
#21, #22, #28 (AST-lint вместо декоратора), #29, #30, #31.

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

## 🔒 Осталось (блокировано внешним)

### 1. `github.com/docker/docker` CVE (high + medium)
- **Статус:** ждём upstream — `v29.x` в Go-modules.
- **Когда станет доступно:** `go get github.com/docker/docker@v29.x && go mod tidy` + прогон тестов.

### 8b. Smart Insights feature-flag toggle в prod
- **Что сделать:** установить `ANALYTICS_EXPERIMENTAL_INSIGHTS=true` в `.env.prod` после проверки что managed Postgres поддерживает `pg_trgm` (миграция 000048 прошла).
- **Мониторинг:** 24 часа через `analytics_insights_refresh_total{result="success"}`.

### 17. Email notifications — SMTP реализация
- **Что сделать:** `infrastructure/email/insights_notifier.go` с rate-limit 1/неделю через `insight_notifications` + HTML-шаблон `insights_digest.html` + settings UI `/settings/notifications`.
- **Ждёт:** решение продукта по UX digest и opt-in формулировкам.

---

## 🔧 Остаётся (не блокировано, но отложено)

### 9b. Drill-down migration — остальные методы AnalyticsRepository
Pilot `AnalyticsFilter` struct в интерфейсе есть, 16 методов ещё принимают отдельные `teamID *uint, rng DateRange`. Поэтапная миграция — отдельный PR без API-breakage.

### 14. Рефакторинг `app.go` (540+ строк)
Кандидат на `app/repos.go`/`usecases.go`/`handlers.go`/`loops.go`/`routes.go` split. Не критично сейчас.

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
