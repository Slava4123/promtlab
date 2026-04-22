# Backlog — Post Phase 14.3+

**Дата:** 2026-04-23
**Предыдущий CI:** `550a75b` (Phase 14.3 wave 3)
**Prod:** `promtlabs.ru` актуален.

Все пункты исходного `BACKLOG.md` (16 + техдолг) закрыты в четыре волны Phase 14.3.
Единственный открытый пункт — `#1` docker/docker CVE, заблокирован upstream
(Go-modules содержат только v28; v29 ещё не опубликован).

---

## ✅ Закрыто в Phase 14.3 wave 1

- #2 **Integration-тесты AnalyticsRepository** — `analytics_repo_test.go` (4 метода + team-scope).
- #3 **HTTP-тесты analytics** — `request_test.go` (parseRange), `handler_test.go` (breadcrumb).
- #4 **Frontend page-level тесты** — `prompt-analytics.test.tsx`, `team-analytics.test.tsx`, `use-analytics-filter.test.ts`, `url.test.ts`.
- #5, #6 **Prometheus + Sentry breadcrumbs** — пакет `infrastructure/metrics` + endpoint `/metrics` за `SERVER_METRICS_ENABLED`.
- #7 **Alerts** — `infra/prometheus/alerts.yaml`.
- #8 **M8 Full Smart Insights** — миграция `000048_analytics_m8.sql` (pg_trgm + GIN), 4 метода `AnalyticsRepository`, раскрыт stub за `experimentalInsights`.
- #11 **API-key доступ к analytics** — `analytics_read` в `apikey/constants.go` + `mcp-tools.ts`.
- #12 **L2 HTTPS validation** — `frontend/src/lib/url.ts isSafeHttpsUrl`.
- #13 **actor_email маска (вариант C)** — `utils.MaskEmail` + zero-dep копия в `mcpserver`.
- #16, #17, #18 **Email-уведомления каркас** — миграция `000049` + `InsightsNotifier` interface.
- #28 **MCP quota AST-lint** — `quota_lint_test.go` (compile-time защита от забываемости).
- #29, #30 **Soft-delete regression (prompt)** — `prompt_repo_softdelete_test.go`.
- #31 **IsUnlimited / sentinel -1 cleanup** — удалены из `models` и `quota.go`.
- **Drill-down (#9) каркас** — `AnalyticsFilter` struct + frontend `useAnalyticsFilter`.

## ✅ Закрыто в Phase 14.3 wave 2

- #17 **SMTP insights notifier** — `EmailInsightsNotifier` с rate-limit 1/неделя + `PATCH /api/auth/notifications/insights` + `SetInsightEmailsEnabled` в user repo/auth service.
- #14 (частично) **app.go split** — 4 адаптера вынесены в `app/adapters.go`.
- #9 **AnalyticsFilter** helpers — `HasTag()`/`HasCollection()`.

## ✅ Закрыто в Phase 14.3 wave 3

- #9b **Drill-down по тегу/коллекции (Personal)** — 5 filter-aware методов +
  `applyPromptFilter` helper + `Service.GetPersonalDashboardFiltered` +
  handler `?tag_id`/`?collection_id` + frontend queryKey с фильтрами.
- #14b **MountRoutes split** — `app/routes.go` (357 строк), `app.go` 845 → 439.
- #17b **UI settings** — `pages/settings/notifications.tsx` + nav + `useSetInsightEmails`.

## ✅ Закрыто в Phase 14.3 wave 4 (этот коммит)

- #3 **Полноценные HTTP handler тесты analytics** — `delivery/http/analytics/full_handler_test.go`
  с testify/mock через новый локальный `analyticsService` интерфейс.
  Покрыты: Personal OK + drill-down filter pass-through + service error 500;
  Team OK + invalid ID 400; Prompt 404; Insights 402/OK; Export (invalid format 400,
  Free 402, team без team_id 400, CSV OK).
- #9c **Drill-down для TeamDashboard** — `GetTeamDashboardFiltered` в service,
  handler Team парсит `tag_id`/`collection_id`, frontend `fetchTeamAnalytics(id, range, filter)`
  + `useTeamAnalytics(id, range, filter)`.
- #14c **lifecycle.go split** — `StartBackground`/`Shutdown` вынесены в `app/lifecycle.go`.
  `app.go` дополнительно облегчён.
- #16 **Collection soft-delete regression** — `collection_repo_softdelete_test.go`
  (regression на `gorm.DeletedAt` scope + `.Unscoped()` для trash flow).
- **Team handler unit-тесты** — `activity_handler_test.go`: GDPR маска через
  матрицу ролей (owner/editor видят raw email, viewer — `a***@domain`, пустой
  email → пустая маска).
- **#8b env-документация** — `.env.example` обновлён: `SERVER_METRICS_ENABLED`
  с комментарием про IP-allowlist; `ANALYTICS_EXPERIMENTAL_INSIGHTS` с prod
  чек-листом (проверка `pg_extension pg_trgm` перед включением).
- **#15** оставлено как AST-lint — заменить на runtime decorator нельзя без
  синхронной правки 26 точек в `tools.go`; lint даёт ту же гарантию без риска
  quota drift при двойном инкременте.
- **#17b** — закрыт волной 3 (UI уведомлений + хук `useSetInsightEmails`).

---

## 🔒 Осталось (блокировано upstream)

### #1 `github.com/docker/docker` CVE (high + medium)
- **Проверено:** `go list -m -versions github.com/docker/docker` отдаёт максимум `v28.5.2+incompatible`. `v29.x` ещё не в Go-modules.
- **Следующий шаг:** `go get github.com/docker/docker@v29.x && go mod tidy` + прогон тестов сразу после upstream publish.

---

## 🧭 Метрика качества кода (после wave 4)

| Слой | Unit | Integration | HTTP |
|------|------|-------------|------|
| `usecases/analytics` | ✅ | ✅ | ✅ (через handler) |
| `usecases/activity` | ✅ | — | ✅ (handler masking) |
| `usecases/team/branding` | ✅ | — | — |
| `usecases/subscription` | ✅ | ✅ | ✅ |
| `usecases/share` | ✅ | — | — |
| `usecases/quota` | ✅ | — | — |
| `delivery/http/analytics` | ✅ | — | ✅ **wave 4** |
| `delivery/http/team` | ✅ **wave 4** | — | ✅ **wave 4** (activity_handler_test) |
| `delivery/http/subscription` | ✅ | — | — |
| `delivery/http/utils` | ✅ (MaskEmail) | — | — |
| `infrastructure/metrics` | ✅ | — | — |
| `infrastructure/postgres/repository` | — | ✅ (audit, analytics, prompt+collection soft-delete) | — |
| `mcpserver` | ✅ (+ quota AST-lint) | — | — |
| `frontend/lib/url` | ✅ | — | — |
| `frontend/hooks/use-analytics-filter` | ✅ | — | — |
| `frontend/pages/prompt-analytics` | ✅ | — | — |
| `frontend/pages/team-analytics` | ✅ | — | — |
