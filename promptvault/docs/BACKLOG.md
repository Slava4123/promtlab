# Backlog — Post Phase 14.2

**Дата:** 2026-04-23
**Последний прошедший CI:** `e0c5ea3` (UTF-8 truncate в share)
**Prod:** `promtlabs.ru` актуален, включая Analytics v2 + pricing fix.

Документ фиксирует **что осталось** после двух Phase 14 фаз. Упорядочено от
критичного к косметике. Каждый пункт — потенциальный отдельный PR.

---

## 🔒 Безопасность

### 1. `github.com/docker/docker` CVE (high + medium)
- **Статус:** ждём upstream.
- **Детали:** dependabot требует `>= v29.3.1`. В Go-modules опубликована только
  `v28.5.2+incompatible`. Прямо зависит `testcontainers-go`.
- **Действия когда:** upstream зарелизит v29 в Go-modules → `go get
  github.com/docker/docker@v29.x` + `go mod tidy` + тесты.
- **Мониторинг:** dependabot alert закроется автоматически при bump'е.

---

## 🧪 Покрытие тестами (короткое, полезное)

### 2. Integration-тесты analytics repo через testcontainers
Методы без CI-валидации — только локальный `go test`:
- `AnalyticsRepository.GetTrendingPrompts` — CTE last_7 + prev_7. Сложный SQL.
- `AnalyticsRepository.CleanupPromptUsageByRetention` — DELETE по plan_id.
- `AnalyticsRepository.PromptUsageTimeline` — WHERE prompt_id.
- `AnalyticsRepository.UsageByModel` — GROUP BY model_used.

**Паттерн:** как `audit_repo_test.go` — `testcontainers-go/modules/postgres`,
AutoMigrate, seeded data, прогон под `go test ./...` (без `-short`).
**Оценка:** 2-3 часа.

### 3. HTTP-handler тесты для analytics
Покрытие сейчас — service-уровень. Transport-слой не тестирован:
- `Handler.Personal`, `Handler.Team`, `Handler.Prompt` — mapping query → service.
- `Handler.Insights` — 402 на non-Max.
- `Handler.RefreshInsights` — 429 на 2-й вызов в течение часа.
- `Handler.Export` — xlsx vs csv switching.

**Паттерн:** `httptest.NewRecorder`, mock `analyticsuc.Service`.
**Оценка:** 1-1.5 часа.

### 4. Frontend page-level тесты
- `prompt-analytics.tsx` — рендер, использование `usePromptAnalytics`.
- `team-analytics.tsx` — upgrade-gate на Free, MetricCard с delta.

**Оценка:** 1 час.

---

## 📊 Observability (Q3 из Phase 14 self-review)

### 5. Prometheus counters
- `share_quota_increment_failed_total` — revenue-leak сигнал (сейчас slog.Error,
  SRE не видит).
- `analytics_insights_refresh_total{result=success|rate_limited|error}`.
- `analytics_cleanup_deleted_total{table=team_activity|share_views|prompt_usage}`.

### 6. Sentry breadcrumbs
- InsightsRefresh trigger → breadcrumb (вижу активность Max-юзеров).
- Deployed xlsx generation — breadcrumb с sheet-count.

### 7. Alerts
- `InsightsComputeLoop` не проработал > 25 часов (зависший cron).
- `CleanupLoop` не проработал > 25 часов.

**Оценка:** 4-6 часов (если нет Prometheus stack в infra — больше).

---

## 🚀 Аналитика — продуктовые расширения

### 8. M8 — Full Smart Insights (4 из 7 типов сейчас заглушка)
Феатур-флаг `ANALYTICS_EXPERIMENTAL_INSIGHTS=false` скрывает:
- `most_edited` — агрегация по `prompt_versions` (простой GROUP BY).
- `possible_duplicates` — Levenshtein через `pg_trgm` или `fuzzystrmatch`.
- `orphan_tags` — теги без промптов (LEFT JOIN + IS NULL).
- `empty_collections` — аналогично.

**Самый сложный** — possible_duplicates (extension + threshold tuning).
**Оценка:** 1 день.

### 9. Drill-down по тегам и коллекциям
UI: `/analytics?tag=:id` / `?collection=:id` — фильтр на все метрики и графики.
**Backend:** `AnalyticsRepository` методы принимают опциональные `tagID`,
`collectionID` в фильтре.
**Frontend:** URL-params → TanStack Query params → передаются в хуки.
**Оценка:** 4-6 часов.

### 10. Email-уведомления по инсайтам
Цель: юзер возвращается на сайт. Триггер: cron-loop после computeInsights
отправляет email, если:
- появились новые `unused_prompts` (первое уведомление в месяц).
- `trending` обновился (еженедельный дайджест).

**Нужно:** `InsightsChangedHook` в `insights_loop.go`, email-шаблон,
rate-limit 1 письмо/неделю на юзера.
**Оценка:** 3-4 часа.

### 11. API-key доступ к analytics (Max feature из pricing)
`/pricing` обещает «Аналитика: 365 дней истории + CSV export + API». Но
API-key middleware сейчас не пускает на `/api/analytics/*`.
**Задача:** добавить analytics endpoints в `apikey/constants.go:KnownTools`
whitelist + проверка `plan.max_mcp_uses_daily` как квота.
**Оценка:** 2 часа.

---

## 🌙 Низкоприоритетные фиксы

### 12. L2 — Frontend HTTPS scheme validation
`components/teams/branded-header.tsx` — не проверяет что `branding.website`
начинается с `https://`. Backend уже валидирует, но defence-in-depth
предотвратит `javascript:alert()` если когда-то прорвётся через API.
**Оценка:** 5 минут.

### 13. GDPR M5 — refine actor_email маскирование
Сейчас вариант **B**: viewer не видит email вообще. Альтернативы:
- **C**: `a***@acme.com` (маскирование вместо скрытия) — компромисс между
  прозрачностью и privacy.
- **D**: viewer видит только имя без email (текущее поведение + auditlog
  trail на случай расследования).

**Требует:** бизнес-решение. Код-изменение в `ActivityHandler` + `mcpserver/
tools.go` — ~30 мин.

### 14. Рефакторинг `app.go` (540+ строк, 25+ wire-ups)
Кандидат на `wire` или `uber/fx`. Не критично сейчас.

### 15. MCP `checkMCPQuota()` как middleware
Сейчас каждый tool handler вызывает руками — легко забыть на новом tool.
`mcpserver/tools.go` — обернуть в middleware на уровне registration.
**Оценка:** 1 час.

### 16. Soft-delete защита в queries
Все queries collection/tag должны помнить `WHERE deleted_at IS NULL` —
сейчас на каждый новый query нужно помнить. Кандидат на default-scope в
GORM или отдельный query-builder.

---

## 📉 Технический долг (знаем, не горит)

- `quota.isWithinLimit` все ещё принимает `limit == -1` как sentinel — даже
  после миграции 000046 конкретных лимитов. Оставил для backward-compat.
  Если гарантированно все `-1` вычищены → удалить ветку.
- `models.IsUnlimited(limit int)` — deprecated shim. Грепнуть все вызовы и
  удалить совсем.
- Phase 14 Analytics — Totals для Team Dashboard не считает `ShareViews`
  (его нет в team scope). Визуально поле отсутствует — не баг, но если
  кто-то будет добавлять team share-views retention — учесть.

---

## ✅ Закрыто в этой фазе (для справки)

- Pricing undefined share-ссылок/день (DTO fix + миграция 000046).
- «Безлимит» → конкретные числа по всем тарифам.
- Analytics v2: refresh insights / compare prev period / per-prompt UI /
  retention cron / window-functions trending / xlsx export / model
  segmentation.
- Локализация: `Smart Insights` → «Умные инсайты», `daily_shares` в
  модале, `7d` → «7 дней», `share-ссылок` → «публичных ссылок».
- UTF-8 truncate в `share.truncateString` (L1).
- 38+ regression тестов (не включая integration через testcontainers).

---

## Метрика качества кода

| Слой | Unit | Integration | HTTP |
|------|------|-------------|------|
| `usecases/analytics` | ✅ (retention, nonnil) | ❌ (GetTrendingPrompts) | ❌ |
| `usecases/activity` | ✅ (service_test) | ❌ (trigger append-only) | ❌ |
| `usecases/team/branding` | ✅ (branding_test) | — | ❌ |
| `usecases/subscription` | ✅ | ✅ (renewal) | ✅ (response) |
| `usecases/share` | ✅ (share + truncate) | — | ❌ |
| `delivery/http/analytics` | ✅ (export DTO) | — | ❌ |
| `delivery/http/team` | ❌ | — | ❌ |
| `delivery/http/subscription` | ✅ (response) | — | ❌ |
| `infrastructure/postgres/repository` | — | ✅ (audit, testhelper) | — |

**Первый приоритет доделки:** HTTP-тесты для analytics handlers + integration
для новых repo-методов.
