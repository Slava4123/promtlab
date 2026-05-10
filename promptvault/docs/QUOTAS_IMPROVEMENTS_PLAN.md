# Пакет улучшений тарифной/лимитной модели PromptVault

> **Статус:** план утверждён, реализация ожидает старта.
> **Создан:** 2026-05-02. Последняя сверка с кодом — на момент создания.
> **Owner:** Slava Kovalchuk.

## Что входит в пакет

Восемь связанных улучшений тарифной/лимитной модели:

1. **П11** — мягкое 80% предупреждение через toast (наибольший ROI)
2. **П12** — объяснение когда сбросятся дневные лимиты (`retry_after_seconds`)
3. **П13** — метрика `quota_check_total` + Grafana дашборд
4. **П14** — динамические числа в `quota-exceeded-dialog` через `usePlans()`
5. **П16 (объединён с П22)** — наследование плана owner'а команды + ребрендинг маркетинга «Pro/Max и для команд тоже»
6. **П17** — архивирование при даунгрейде (выполняем UX-обещание)
7. **П18** — friendly rate-limit JSON-сообщения
8. **П19** — поведенческие подсказки на стрик-событиях

**Исключено пользователем:** П15 (public-prompts loophole), П22 как отдельный 4-й тариф, П23 (usage-based pricing).

---

## Onboarding для нового исполнителя

### Контекст проекта в одном абзаце

PromptVault — self-hosted SaaS для управления AI-промптами (соло + команды), на VPS в РФ. Стек: **Go 1.25 + Chi + GORM + PostgreSQL 18** на бэке, **React 19.2 + Vite 8 + TypeScript** на фронте, **MCP-сервер** встроен (34 tools, опубликован в Official Registry как `ru.promtlabs/promptvault`). Конфиг — koanf через `.env.dev`/`.env.prod`. Биллинг — T-Bank. UI на русском (i18n нет). Принцип «без AI на нашей стороне» (мы храним промпты, AI вызывает MCP-клиент или юзер). Тарифы Free 0 / Pro 599 / Max 1299 руб./мес.

**Подробнее:** `promptvault/CLAUDE.md` (читать первым), `promptvault/docs/FEATURES.md`, `promptvault/docs/MCP.md`.

### Поднять dev-стек

```bash
cd promptvault
docker compose -f docker-compose.dev.yml up -d --build
# ждать ~30-60 сек
```

После этого:
- Frontend: `http://localhost:5173`
- API: `http://localhost:8080` (Prometheus метрики на `http://localhost:8080/metrics`)
- PostgreSQL: `localhost:5433` (внутри контейнера 5432, на хосте 5433 — конфликт с локальным PG16 в Windows). Креденшелы: `postgres / postgres`, БД `promptvault`.

**ВАЖНО:** после изменения миграций или Go-кода всегда `--build` — иначе `go:embed` миграции не подхватятся в образе.

### Применить тестовые данные (test-юзеры с низкими лимитами)

В проекте есть готовый seed-script. Он создаёт три тестовых тарифа `test_free` / `test_pro` / `test_max` с лимитами в 10-50 раз ниже prod (чтобы упереться в стену за 3-5 действий) и трёх юзеров под них.

```bash
cd promptvault
cat scripts/seed-test-data.sql | docker compose -f docker-compose.dev.yml exec -T postgres psql -U postgres -d promptvault
```

После этого доступны три юзера:

| Email | План | Пароль |
|---|---|---|
| `e2e-free@test.local` | test_free (2 промпта / 1 коллекция / 1 цепочка / 2 шага / 1 share / 2 daily MCP-вызовов и т.д.) | `TestPass2026!` |
| `e2e-pro@test.local` | test_pro (4 / 2 / 2 / 3 / 2 / 3 ...) | `TestPass2026!` |
| `e2e-max@test.local` | test_max (6 / 3 / 3 / 4 / 3 / 5 ...) + `IsMaxTier=true` | `TestPass2026!` |

`test_*`-планы скрыты из публичного `/api/plans/` через фильтр в `plan_repo.GetActive` (`backend/internal/infrastructure/postgres/repository/plan_repo.go:72-85`) — на проде их не видно. Логиниться можно UI-формой `/sign-in` или через API:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e-free@test.local","password":"TestPass2026!"}'
```

### Сбросить state тестового юзера (для чистого старта теста)

Между тестами state копится (созданные промпты, коллекции, цепочки, daily-счётчики). Чтобы сбросить — есть **dev-only endpoint** (защищён `cfg.Server.IsDev() == true`, на проде route не существует):

```bash
curl -X POST "http://localhost:8080/api/test/cleanup?email=e2e-free@test.local"
```

Удаляет: prompts, prompt_versions, prompt_pins, prompt_usage_log, prompt_chain_steps, prompt_chain_executions, prompt_chains, share_links, tags, collections, daily_feature_usage, team_invitations, team_members, teams (где created_by=user), user_smart_insights, insight_notifications. Самого юзера, его plan_id и subscription **не удаляет**.

Endpoint принимает только email с суффиксом `@test.local` и префиксом `e2e-` (защита от случайного использования на не-тестовых юзерах).

Источник: `backend/internal/delivery/http/testcleanup/handler.go`.

### Команды разработки

```bash
# Backend
cd backend && go run ./cmd/server      # dev-сервер
go test -short ./...                   # unit-тесты (testcontainers integration пропущены)
go test ./...                          # +integration (нужен Docker)
golangci-lint run                      # lint

# Frontend
cd frontend && npm run dev             # Vite dev server
npm run test                           # vitest run
npm run test:e2e                       # Playwright (после поднятия стека)
npm run lint                           # ESLint 9 (flat config)
npm run build                          # tsc -b && vite build

# Docker
cd promptvault
docker compose -f docker-compose.dev.yml up -d --build
docker compose -f docker-compose.dev.yml restart api  # без миграций — быстро
docker compose -f docker-compose.dev.yml logs -f api  # логи backend
```

### Связанные документы — что читать первым

1. `promptvault/CLAUDE.md` — архитектура проекта, конвенции, Phase 14/15/16 контекст, ADR ссылки.
2. `promptvault/docs/FEATURES.md` — каталог 104 идей с tier-маркерами и статусом ✅.
3. `promptvault/docs/MCP.md` — справочник 34 MCP tools.
4. `promptvault/docs/FEATURE_PROMPT_CHAINS.md` — Phase 16 (текущая активная фаза в проекте).
5. Этот файл (`QUOTAS_IMPROVEMENTS_PLAN.md`) — текущий план улучшений.
6. ADR'ы в `promptvault/docs/ADR/` — 0001-0005 уже есть, 0007 будет создан в S19.

### Принятые решения (зафиксированы пользователем)

| Решение | Выбор | Когда зафиксировано |
|---|---|---|
| Как реализовать «Team plan» (П22) | **Вариант A**: наследование плана owner'а команды + ребрендинг маркетинга на /pricing «Pro/Max и для команд тоже». **Никаких новых tier'ов** в БД. Это объединяет с П16. | 2026-05-02 |
| Дыра с публичными промптами (П15) | **Отложено.** Не критично пока юзеров мало; вернёмся когда метрики покажут злоупотребление. | 2026-05-02 |
| Поведенческие подсказки (П19) | **Включены в план.** Пользователь сознательно принял риск шума, использует существующую engagement-инфраструктуру. | 2026-05-02 |
| Prometheus + Grafana | **Развёрнуты на VPS.** Дашборд импортируем JSON-ом в `infra/grafana/dashboards/quotas.json`. | 2026-05-02 |
| Usage-based слой (П23) | **Отложено.** Не делаем сейчас. | 2026-05-02 |

### Что было сделано в подготовительных сессиях ДО утверждения плана

Все эти изменения **уже в коде** на момент старта работы по плану. Не нужно повторно реализовывать:

| # | Что | Файлы (path:line — где смотреть) |
|---|---|---|
| 1 | Расширили `/pricing`: Chains-лимиты, блок «Доступно во всех тарифах», total share-links | `frontend/src/pages/pricing.tsx:53-100` (planFeatures), `frontend/src/api/types.ts:Plan` |
| 2 | Backend `PlanResponse` отдаёт `max_chains`, `max_steps_per_chain`, `max_saved_executions` | `backend/internal/delivery/http/subscription/response.go:11-49` |
| 3 | Скрыли `test_*`-планы из публичного API | `backend/internal/infrastructure/postgres/repository/plan_repo.go:72-85` (фильтр `id NOT LIKE 'test_%'`) |
| 4 | Создан `seed-test-data.sql` + 3 test-юзера | `promptvault/scripts/seed-test-data.sql` |
| 5 | Создан dev-only `POST /api/test/cleanup` | `backend/internal/delivery/http/testcleanup/handler.go`, регистрация в `routes.go` под `cfg.Server.IsDev()` |
| 6 | Заменили `alert(...)` в цепочках на `toast.error` через `reportMutationError` helper | `frontend/src/pages/chains/editor.tsx:60-67` (helper), 4 места вызова |
| 7 | Заменили `confirm(...)` в цепочках на компонент `ConfirmDialog` | `frontend/src/pages/chains/editor.tsx`, `chains/index.tsx`, `frontend/src/components/ui/confirm-dialog.tsx` |
| 8 | Расширили `quotaLabels` и `quotaBenefits` для `chains` и `chain_steps` | `frontend/src/components/subscription/quota-exceeded-dialog.tsx:17-115` |
| 9 | Все хардкоженные числа в `quotaBenefits` приведены к реальным значениям миграции 000046 + цены 20₽ Pro / 43₽ Max | `frontend/src/components/subscription/quota-exceeded-dialog.tsx:34-115`, `subscription-section.tsx:159` |
| 10 | Поправлен FAQ-ответ про тарифы | `frontend/src/pages/help.tsx:21-25` |
| 11 | Sonner-toast: `pointer-events: none` на toast-контейнере, `auto` на close-button — синхронизация прогресс-бара с реальным dismiss-таймером | `frontend/src/index.css:268-310` |

**Эти 11 пунктов описаны для контекста — менять/повторять их не нужно.** План S1-S20 ниже строит **поверх** этого состояния.

### Текущее состояние тестовой Phase 1 (Playwright)

В прошлой сессии настроили Playwright skeleton для **отдельной QA-ветки** (не часть этого плана):
- `promptvault/frontend/playwright.config.ts` — projects setup/free/pro/max.
- `promptvault/frontend/playwright/auth.setup.ts` — login для каждого тира + сохранение storageState.
- `promptvault/frontend/playwright/specs/smoke.{free,pro,max}.spec.ts` — 3 smoke-теста зелёные.
- `npm run test:e2e` работает.

S20 в плане ниже опционально расширяет это для quota-warning. Если решено НЕ делать E2E в рамках текущего плана — пропустить S20.

---

## Карта существующего кода

**Слои архитектуры (Clean Architecture):**
- `usecases/quota/` — доменная логика лимитов (Service + Check*-методы)
- `interface/repository/` — интерфейсы (PlanRepository, QuotaRepository, UserRepository)
- `infrastructure/postgres/repository/` — GORM реализации
- `delivery/http/<feature>/` — HTTP handlers, передают error → `errors.RespondQuotaError`
- `infrastructure/metrics/` — `promauto` Counter/CounterVec с zero-init в `init()`

**Эталонные файлы:**
- `backend/internal/usecases/quota/quota.go:35-47` (`getPlan`) — паттерн lookup плана юзера через `users.GetByID` → `plans.GetByID` (cache 5 мин). Переиспользуем для `effectivePlanForTeam`.
- `backend/internal/usecases/quota/quota.go:178-187` (`IsMaxTierUser`) — паттерн tier-check, returns bool с safe-default false. Переиспользуем для `IsMaxTierForTeam`.
- `backend/internal/delivery/http/errors/errors.go:54-65` (`RespondQuotaError`) — JSON 402 с полями `{error, quota_type, used, limit, plan, upgrade_url}`. Расширяем добавлением `retry_after_seconds`.
- `backend/internal/infrastructure/metrics/metrics.go:49-67` — паттерн `chains_created_total{scope}` + zero-init в `init()`. Эталон для `quota_check_total`.
- `frontend/src/hooks/use-badge-toast.ts:19-42` — паттерн success-trigger (sonner toast при unlock), invalidateQueries. Эталон для пунктов 11 и 19.
- `frontend/src/components/subscription/usage-meters.tsx:32-33` — пороги `pct >= 75 amber, >= 90 red`. Source-of-truth threshold для warning.
- `frontend/src/api/client.ts:124-138` — перехват 402 + dispatch в `useQuotaStore.show()`. Расширяем чтобы прокидывать `retry_after_seconds`.

**Тесты-эталоны:**
- `backend/internal/usecases/quota/quota_test.go:13-135` — fake in-memory repos с `incrementLog` трейсом, ассерты через `errors.As(&qe)`.
- `frontend/src/components/subscription/quota-exceeded-dialog.test.tsx:40-86` — table-driven cases по quota_type.

**Конвенции:**
- Error-handling: доменные ошибки в `usecases/<feature>/errors.go`, HTTP-маппинг в `delivery/http/<feature>/errors.go` через `errors.As/Is`.
- Logging: `slog.Info/Error` со структурными атрибутами.
- Metrics: `promauto.NewCounterVec` + zero-init всех label combinations в `init()`.
- Migrations: `000NNN_description.{up,down}.sql`, idempotent через `IF NOT EXISTS`.
- Frontend: TanStack Query для всех API, sonner toast (4s default), Zustand для глобального state.

---

## 1. Резюме

Пакет восьми улучшений тарифной/лимитной модели в одном плане: мягкие 80%-предупреждения, объяснение reset-времени дневных лимитов, метрика `quota_check_total` + Grafana-дашборд, динамические числа в quota-dialog, наследование owner-плана командой («Team plan» вариант A), архивирование при даунгрейде, friendly rate-limit сообщения, поведенческие подсказки на стрик-событиях. Цель — увеличить переход «Free → платный» на 2-2.5x за счёт UX-практик 2026, закрыть несовпадение «обещание UX vs реальность» (архивирование) и убрать монетизационные дыры (наследование). Public-prompts loophole отложен пользователем; usage-based pricing отложен. Ключевые архитектурные решения — `effectivePlanForTeam` helper для команд, отдельная колонка `archived_at` для архива (не путать с GORM soft-delete), label `quota_type` для метрики (низкая cardinality).

**Аудитория плана:** исполнитель (последовательная реализация по фазам S1-S20) + PR-ревьюер (понять trade-offs) + on-call (runbook для quota-аномалий).

---

## 2. Архитектурные решения

### Решение 1: `effectivePlanForTeam(userID, teamID *uint)` — единый helper для team-наследования

**Решение:** Все persistent-Check-методы (prompts, collections, share-links, chains, chain-steps, saved-executions) принимают необязательный `teamID *uint`. Если `teamID != nil`, helper возвращает план владельца команды (роль `owner`); иначе — план юзера. Daily-window квоты (`ext_daily`, `mcp_daily`, `daily_shares`) **остаются per-user** для anti-abuse.

**Альтернативы:**
- (A) Middleware injection: подмешивать `effective_plan_id` в context на entry-point. Минус: middleware не знает teamID без парсинга route + body.
- (B) Дублировать логику в каждом usecase. Минус: copy-paste 9 мест.
- (C) Helper в quota-сервисе с явным параметром teamID **(выбран)**. Плюс: один источник истины.

**Trade-offs:**
- ✅ Один источник истины.
- ✅ Тестируется в изоляции (один fake-repo с teams).
- ❌ Breaking-change в сигнатурах Check*-методов.
- ❌ Дополнительный DB-roundtrip для team-resource. Митигация: 5-минутный кеш на уровне `getPlan`.

**Источник:** Phase 16 fork-gate использует тот же подход в `chain.go:isMaxTierForChain`. [ДОПУЩЕНИЕ: verify через grep при реализации S5.]

### Решение 2: `archived_at TIMESTAMP` — отдельная колонка vs переиспользование `deleted_at`

**Решение:** Добавить новую nullable колонку `archived_at` в `prompts`, `collections`, `prompt_chains` (миграция 000059). Семантика: `deleted_at IS NULL AND archived_at IS NOT NULL` = архивный (read-only, не считается в квоту, виден в отдельном UI-разделе). При re-upgrade автоматически `archived_at = NULL`.

**Альтернативы:**
- (A) Переиспользовать `deleted_at` + bool `is_archived`. Минус: путает семантику trash и архива.
- (B) Отдельная таблица `archived_prompts`. Минус: дорого по операциям, FK-проблемы для chains.
- (C) Колонка `archived_at` **(выбран)**. Плюс: единая таблица.

**Trade-offs:**
- ✅ Юзер не теряет данные при downgrade.
- ✅ Auto-restore при re-upgrade.
- ❌ Все Count*-методы должны явно фильтровать.
- ❌ UI требует нового раздела «Архив».

### Решение 3: Cardinality метрики `quota_check_total{quota_type, status, plan}`

**Решение:** Три label'а: `quota_type` (10 значений в текущем коде, 11-е `saved_executions` зарезервировано), `status` (`ok | exceeded`), `plan` (5 значений: free/pro/max/pro_yearly/max_yearly). Итого = 100 time-series — низкая cardinality, безопасно. Без user_id/team_id/IP — это убило бы Prometheus.

**Альтернативы:**
- (A) Без label `plan`. Минус: потеряем разбивку «Free vs Pro упирается в X».
- (B) С label `team_scope` (`personal | team`). Плюс: видим как team-юзеры используют квоты. Минус: x2 series.
- (C) `(quota_type, status, plan)` **(выбран)**. Идеальный баланс.

**Источник:** [context7 prometheus/client_golang](https://context7.com/prometheus/client_golang/llms.txt) — best practice для CounterVec.

---

## 3. Изменения в коде

### Backend

**Создаём:**
- `backend/internal/infrastructure/postgres/migrations/000059_archived_at.up.sql` — `archived_at TIMESTAMP` в 3 таблицы + индексы.
- `backend/internal/infrastructure/postgres/migrations/000059_archived_at.down.sql` — DROP COLUMN.
- `backend/internal/usecases/archive/archive.go` (новый usecase) — Service: `ArchivePrompt`, `RestorePrompt`, `ListArchived`, `ArchiveAtDowngrade`.
- `backend/internal/usecases/archive/types.go`, `errors.go`, `archive_test.go`.
- `backend/internal/delivery/http/archive/handler.go` — `GET /api/archive`, `POST /api/archive/{id}/restore`.
- `backend/internal/delivery/http/archive/{request,response,errors}.go`.

**Меняем:**
- `backend/internal/usecases/quota/quota.go` — добавить `effectivePlanForTeam`, обновить 6 Check*-методов. `CheckTeamQuota`, `CheckExtensionQuota`, `CheckMCPQuota`, `CheckDailyShareCreation`, `CheckTeamMemberQuota` — **остаются per-user**. Дополнить `DowngradePreview` полями `OverChains`, `OverSavedExecutions`.
- `backend/internal/usecases/quota/types.go` — расширить `QuotaExceededError` полем `RetryAfterSeconds int`.
- `backend/internal/delivery/http/errors/errors.go:RespondQuotaError` — добавить optional `retry_after_seconds`.
- `backend/internal/infrastructure/metrics/metrics.go` — зарегистрировать `quota_check_total` + zero-init.
- `backend/internal/middleware/ratelimit/ratelimit.go` — JSON `{"error": ..., "retry_after_seconds": 60}` (S14).
- `backend/internal/usecases/subscription/subscription.go:Downgrade` — вызывает `archive.ArchiveAtDowngrade`.
- `backend/internal/interface/repository/prompt.go`, `collection.go`, `chain.go` — методы `Archive`, `Restore`, `ListArchived`. `Count*` исключают `archived_at IS NOT NULL`.
- `backend/internal/infrastructure/postgres/repository/prompt_repo.go`, `collection_repo.go`, `chain_repo.go` — реализации.
- `backend/internal/usecases/chain/chain.go:isMaxTierForChain` — переиспользовать `quota.IsMaxTierForTeam`.
- `backend/internal/app/app.go` — wire-up `archive.Service`, `archive.Handler`.
- `backend/internal/app/routes.go` — register `/api/archive` routes.

### Frontend

**Создаём:**
- `frontend/src/pages/archive.tsx` — раздел «Архив» (список + restore-кнопка).
- `frontend/src/hooks/use-archive.ts` — TanStack Query хук.
- `frontend/src/hooks/use-quota-warning.ts` — observer 80%-warning toast.
- `frontend/src/hooks/use-streak-toast.ts` — поведенческие подсказки на milestones.

**Меняем:**
- `frontend/src/components/subscription/quota-exceeded-dialog.tsx` — убрать хардкод `quotaBenefits`, читать через `usePlans()`. Добавить блок «Сбросится через X».
- `frontend/src/api/types.ts` — расширить `QuotaPayload` полем `retryAfterSeconds?: number`.
- `frontend/src/api/client.ts:124-138` — пробросить `retry_after_seconds` в `useQuotaStore.show()`.
- `frontend/src/stores/quota-store.ts` — добавить `retryAfterSeconds` в state.
- `frontend/src/hooks/use-subscription.ts:useUsage` — добавить `refetchInterval: 30_000`.
- `frontend/src/components/layout/app-layout.tsx` — смонтировать `<QuotaWarningObserver />`.
- `frontend/src/components/layout/app-sidebar.tsx` — добавить пункт «Архив».
- `frontend/src/pages/pricing.tsx` — обновить `planFeatures` маркетинг-копи на «и для команд тоже».
- `frontend/src/App.tsx` — route `/archive` (lazy).

### Сущности / типы / интерфейсы

```go
// backend/internal/usecases/quota/types.go
type QuotaExceededError struct {
    QuotaType         string
    PlanID            string
    Used              int
    Limit             int
    Message           string
    RetryAfterSeconds int  // NEW: 0 для persistent, seconds_until_midnight_UTC для daily
}

// backend/internal/usecases/quota/quota.go (новый helper)
func (s *Service) effectivePlanForTeam(ctx context.Context, userID uint, teamID *uint) (planID string, plan *models.SubscriptionPlan, err error)

// IsMaxTierForTeam (рефакторинг IsMaxTierUser + chain.isMaxTierForChain)
func (s *Service) IsMaxTierForTeam(ctx context.Context, userID uint, teamID *uint) bool
```

```go
// backend/internal/usecases/archive/types.go (новый)
type ArchiveService interface {
    ArchivePrompt(ctx context.Context, promptID, userID uint) error
    RestorePrompt(ctx context.Context, promptID, userID uint) error
    ListArchived(ctx context.Context, userID uint, teamID *uint) ([]ArchivedItem, error)
    ArchiveAtDowngrade(ctx context.Context, userID uint, targetPlanID string) (ArchiveSummary, error)
}

type ArchiveSummary struct {
    PromptsArchived     int
    CollectionsArchived int
    ChainsArchived      int
}
```

```typescript
// frontend/src/hooks/use-quota-warning.ts (новый)
export function useQuotaWarning(): void {
  // Observer: при usage.X.used / usage.X.limit >= 0.8 показывает toast.warning,
  // но не чаще одного раза за тип лимита за сессию (sessionStorage).
}
```

### Контракты между слоями

- **Quota Service → Repository:** новый метод `TeamRepository.GetOwnerUserIDs(teamID) []uint`.
- **Archive Service → Repository:** новые методы `Archive(id)`, `Restore(id)`. `Count*` исключают `archived_at IS NOT NULL`.
- **HTTP Handler → Quota Service:** persistent endpoints передают `teamID *uint` в `quota.CheckXQuota`.

---

## 4. Модель данных

### Миграция 000059_archived_at

```sql
-- forward
ALTER TABLE prompts ADD COLUMN archived_at TIMESTAMP NULL;
CREATE INDEX idx_prompts_archived_user ON prompts(user_id) WHERE archived_at IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE collections ADD COLUMN archived_at TIMESTAMP NULL;
CREATE INDEX idx_collections_archived_user ON collections(user_id) WHERE archived_at IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE prompt_chains ADD COLUMN archived_at TIMESTAMP NULL;
CREATE INDEX idx_chains_archived_user ON prompt_chains(user_id) WHERE archived_at IS NOT NULL AND deleted_at IS NULL;

-- rollback (000059_archived_at.down.sql)
DROP INDEX IF EXISTS idx_chains_archived_user;
DROP INDEX IF EXISTS idx_collections_archived_user;
DROP INDEX IF EXISTS idx_prompts_archived_user;
ALTER TABLE prompts DROP COLUMN IF EXISTS archived_at;
ALTER TABLE collections DROP COLUMN IF EXISTS archived_at;
ALTER TABLE prompt_chains DROP COLUMN IF EXISTS archived_at;
```

**Влияние на данные:** нет бэкфилла, downtime — нет.

**Обратимость rollback:** обратима. Если в архиве есть данные на момент rollback — они «вернутся в активные».

### Соглашение по фильтрам

- `WHERE deleted_at IS NULL AND archived_at IS NULL` — только активные.
- `WHERE deleted_at IS NULL AND archived_at IS NOT NULL` — архив.
- `WHERE deleted_at IS NOT NULL` — корзина (как сейчас).

---

## 5. API контракт

### Изменения существующих

**HTTP 402 ответ** — расширен:
```json
{
  "error": "Лимит MCP-вызовов исчерпан",
  "quota_type": "mcp_daily",
  "used": 5,
  "limit": 5,
  "plan": "free",
  "upgrade_url": "/pricing",
  "retry_after_seconds": 14823
}
```

**HTTP 429 (rate-limit)** — формат меняется (S14):
```json
{
  "error": "Слишком много попыток, попробуй через 30 секунд",
  "retry_after_seconds": 30
}
```

### Новые endpoints

**`GET /api/archive?team_id=<optional>`**
```json
{
  "prompts": [{"id": 42, "title": "...", "archived_at": "2026-05-02T..."}],
  "collections": [...],
  "chains": [...],
  "total": 12
}
```

**`POST /api/archive/{prompts|collections|chains}/{id}/restore`** → 200 OK или 402 если попытка восстановить превышает лимит.

---

## 6. Зависимости

**Внешние сервисы:** нет новых.
**Новые библиотеки:** нет. Используем существующие — sonner, TanStack Query, Zustand, GORM, Prometheus client_golang.
**Кросс-командные блокеры:** нет.
**Grafana:** дашборд импортируем JSON-ом (Grafana уже развёрнут).

---

## 7. План тестирования

### Unit
- **`quota.effectivePlanForTeam`**: 4 кейса (no team, single owner, multiple owners pick max, no owners fallback).
- **`quota.newQuotaExceeded` с RetryAfter**: для daily-методов проверяем seconds-to-midnight-UTC.
- **`archive.ArchiveAtDowngrade`**: 5 промптов → лимит 2 → 3 архивируется (от старых).
- **`archive.RestorePrompt`** при превышении квоты → возвращает `ErrQuotaExceeded`.
- **`metrics.QuotaCheckTotal` zero-init**: все 100 комбинаций видны в `/metrics`.

### Integration
- testcontainers PostgreSQL: миграция 000059, архивирование, restore.
- Rate-limit auth — JSON ответ соответствует контракту.

### E2E (отдельная ветка)
- Phase 2 Playwright — добавится отдельной задачей после реализации.

---

## 8. Наблюдаемость

### Метрики

| Имя | Тип | Labels | Назначение |
|---|---|---|---|
| `quota_check_total` | CounterVec | `quota_type`, `status` (`ok\|exceeded`), `plan` | Сколько раз каждый лимит проверяется и упирается |
| `quota_archive_total` | CounterVec | `resource` (`prompts\|collections\|chains`), `trigger` (`downgrade\|manual`) | Сколько ресурсов архивируется и почему |
| `quota_restore_total` | CounterVec | `resource`, `result` (`ok\|over_quota`) | Сколько restore-попыток |

Cardinality: 100 + 6 + 9 = 115 series. Безопасно.

### Логи

```
slog.Info("quota.check", "user_id", uid, "quota_type", qt, "status", "exceeded", "plan", plan)
slog.Warn("archive.downgrade", "user_id", uid, "from_plan", from, "to_plan", to, "archived", count)
slog.Error("archive.restore.failed", "user_id", uid, "resource_id", id, "error", err)
```

### Алерты Grafana

- **`high_free_quota_exceeded`**: rate(quota_check_total{plan="free", status="exceeded"}[5m]) > 5/min sustained 15min → predicting churn.
- **`archive_downgrade_anomaly`**: increase(quota_archive_total{trigger="downgrade"}[1h]) > 50 → массовый отток.

### SLO/SLI
- SLI: % Free-юзеров с хотя бы одним `quota_check status=exceeded`. Цель — 80% за 90 дней.

### Дашборд Grafana

`infra/grafana/dashboards/quotas.json`. Панели:
1. quota_check_total rate by `(quota_type, plan)`.
2. quota_check_total `status="exceeded"` rate.
3. quota_archive_total / quota_restore_total.
4. Time-to-first-limit (Free).

---

## 9. План внедрения

| ID | Шаг | Owner | Критерий готовности | Зависит |
|---|---|---|---|---|
| **S1** | Метрика `quota_check_total` + zero-init + инкремент в 10 Check*-методах. | backend | `curl /metrics \| grep quota_check_total` показывает 100 строк | — |
| **S2** | `RetryAfterSeconds` в `QuotaExceededError`. Daily-методы заполняют через `secondsUntilMidnightUTC()`. | backend | `quota_test.go` зелёный, новые кейсы | — |
| **S3** | `RespondQuotaError` JSON `retry_after_seconds`. `client.ts` пробрасывает в store. | backend + frontend | `curl` 402 на daily возвращает поле; в DevTools видно | S2 |
| **S4** | Frontend: показ «сбросится через X» в `quota-exceeded-dialog`. | frontend | Скриншот dialog'а с правильным текстом | S3 |
| **S5** | `quota.effectivePlanForTeam` + `IsMaxTierForTeam`. Refactor `chain.isMaxTierForChain`. | backend | `quota_test.go` + `chain_test.go` зелёные | — |
| **S6** | Обновить 6 Check*-методов чтобы принимать `teamID *uint`. Обновить call-sites. | backend | `go test ./internal/usecases/...` зелёный | S5 |
| **S7** | Маркетинг-копи на `/pricing`: блок «и для команд тоже». | frontend | Visual review | — |
| **S8** | `OverChains`, `OverSavedExecutions` в `DowngradePreview`. UI обновить. | backend + frontend | M-10 показывает все 6 over-полей | — |
| **S9** | Миграция 000059 (archived_at). | backend | `migrate up && migrate down && migrate up` без ошибок | — |
| **S10** | Repo-методы `Archive`/`Restore`/`ListArchived`. Обновить `Count*`. Integration-тесты. | backend | `go test -count=1 ./internal/infrastructure/postgres/...` зелёный | S9 |
| **S11** | `usecases/archive/` Service. `ArchiveAtDowngrade` встраивается в `subscription.Downgrade`. | backend | Downgrade Pro→Free архивирует превышение | S10 |
| **S12** | HTTP handlers `/api/archive` + frontend `/archive` page + sidebar. | backend + frontend | downgrade → /archive → restore → активные | S11 |
| **S13** | Frontend: `useUsage` polling 30s. `useQuotaWarning` хук с sessionStorage дедупликацией. | frontend | Manual: 4 из 5 промптов → toast | — |
| **S14** | Friendly rate-limit JSON. `client.ts` перехват 429 → `toast.warning`. | backend + frontend | Manual: 21 раз login → красивый toast | — |
| **S15** | `useStreakToast` хук — поведенческие подсказки на 7-day streak, 10-prompt/неделю. | frontend | На staging видим toast при 7-day streak | — |
| **S16** | Динамика чисел в `quotaBenefits` через `usePlans()`. | frontend | `quota-exceeded-dialog.test.tsx` зелёный с моком | — |
| **S17** | Grafana дашборд JSON в `infra/grafana/dashboards/quotas.json`. | DevOps | Дашборд видим, 4 панели рендерятся | S1 |
| **S18** | Алерты Grafana. | DevOps | Alert rules применены, тестовый trigger через `promtool` | S1, S17 |
| **S19** | ADR-0007 + Runbook `archive-downgrade.md`. | docs | Файлы существуют | S6, S11 |
| **S20** | E2E Playwright spec для warning toast (опционально). | QA / frontend | `npx playwright test quota-warning` зелёный | S13 |

**Atomicity:** каждый шаг мержится самостоятельно.

---

## 10. Rollout и kill-switch

### Стратегия раскатки

1. **Wave 1 — Observability (S1, S17, S18)**: метрика + дашборд. Безопасно.
2. **Wave 2 — UX-улучшения (S2-S4, S13-S16)**: warning + retry-after + behavioral.
3. **Wave 3 — Team inheritance (S5-S8)**: меняет логику квот для team-ресурсов. Под feature flag.
4. **Wave 4 — Archive (S9-S12)**: миграция + новый usecase. Под feature flag.

### Feature flags

- **`TEAM_QUOTA_INHERITANCE` (env, default `false`)**: fallback к старому поведению при выкл.
- **`ARCHIVE_ENABLED` (env, default `false`)**: при выкл downgrade не архивирует.

Frontend: `VITE_TEAM_QUOTA_INHERITANCE_ENABLED`, `VITE_ARCHIVE_ENABLED`.

### Kill-switch RTO
- Метрика: ~30-60 сек (env переключение + restart).
- Архив: при выключении новые архивации не происходят, существующие записи в БД сохранены.
- Team inheritance: не выключать без коммуникации (юзеры могут потерять доступ к ресурсам).

### Communication plan
- Changelog `/changelog`: «Команды получили общие возможности», «Новый раздел Архив», «Предупреждение перед лимитом».
- Email на 80% — отдельная задача, не в этом плане.

---

## 11. Документация

- **README:** N/A.
- **ADR 0007:** «Effective plan for team-scoped resources + archived_at column».
- **CLAUDE.md:** краткое упоминание новых helper'ов.
- **Runbook `docs/runbooks/archive-downgrade.md`**: что делать при `archive_downgrade_anomaly`.
- **`docs/QUOTAS.md`** (новый): сводный документ — какие лимиты, как считаются, что архивируется, что унаследуется.

---

## 12. Риски и митигации

### Технические риски

- **Утечка retry_after_seconds для не-daily квот** → frontend показывает блок только если `retry_after_seconds > 0`.
- **Race в IncrementDailyUsage UPSERT** — уже сейчас существует (atomic ON CONFLICT).
- **80%-warning спам** — sessionStorage дедупликация.
- **Кеш плана 5 мин** при изменении планов owner'а — рассинхрон до 5 минут, приемлемо.

### Pre-mortem (через 6 месяцев это сломалось)

1. **Метрика стала high-cardinality** (кто-то добавил `user_id`). Митигация: ADR 0007 + code review.
2. **Архивирование сработало неправильно** при downgrade. Митигация: integration-тест + ORDER BY `created_at ASC`.
3. **Team inheritance + leave-from-team unexpected cycle**. Митигация: при leave запускается `archive.ArchiveAtDowngrade(userID, currentPlan)`.
4. **80%-warning слишком назойлив**. Митигация: sessionStorage = max 1 раз на тип, NPS-feedback.

### Известные ограничения

- Grafana дашборд через ручной импорт JSON (нет provisioning automation).
- Email-warning на 80% — отдельная задача.
- П15 (public-prompts loophole) и П23 (usage-based) — отложены пользователем.

---

## 13. Метрики успеха

### Бизнес
- Конверсия Free → платный: вырастет на 1.5-2.5x в течение 90 дней после Wave 2.
- Churn на Free: должен снизиться.

### Технические
- `quota_check_total` rate: > 100/min на проде.
- API latency p95 для `POST /api/prompts`: не вырастет > +5ms.
- Error rate `archive.ArchiveAtDowngrade`: < 0.1%.
- 80%-warning toast click-through-rate: target 5-10%.

### Срок измерения
- Wave 1: неделя.
- Wave 2: 30 дней.
- Wave 3: 60 дней.
- Wave 4: 90 дней.

---

## 14. Открытые вопросы

1. **Frontend-метрика `quota_warning_shown_total`** — отдельный backend endpoint `/api/metrics/event` или забыть про backend-side учёт показов?
2. **Алерт `archive_downgrade_anomaly` runbook routing** — куда уведомление? Email/Telegram/PagerDuty?
3. **Friendly rate-limit (П18)** — русские сообщения для всех 429 или только auth?
4. **MCP daily-квота при team-наследовании** — daily остаётся per-user, корректно? Документируем в ADR.

---

## Приложение A: что уже сделано в подготовительной работе

В предыдущих сессиях, до утверждения этого плана:

1. Расширили страницу `/pricing` — Chains-лимиты, блок «Доступно во всех тарифах».
2. Скрыли тестовые планы `test_*` из публичного `/api/plans/`.
3. Создали `scripts/seed-test-data.sql` + 3 test-юзера.
4. Создали dev-only `POST /api/test/cleanup`.
5. Заменили `alert(...)` в цепочках на `toast.error` + `reportMutationError` helper.
6. Заменили `confirm(...)` в цепочках на `ConfirmDialog`.
7. Расширили `quotaLabels` и `quotaBenefits` для chains/chain_steps.
8. Поправили все хардкоженные числа в `quota-exceeded-dialog` под реальные лимиты миграции 000046 + цены 20₽/43₽.
9. Поправили устаревший FAQ-ответ про тарифы в `help.tsx`.
10. Поправили `pointer-events: none` для sonner-toast (синхронизация прогресс-бара с реальным таймером).

Эти изменения — **подготовка** к текущему плану. Они уже в коде, не требуют повторной работы.

---

## Sources / референсы

- [Phoenix Strategy — Freemium Behavioral Insights](https://www.phoenixstrategy.group/blog/freemium-vs-subscription-behavioral-insights)
- [Maxio — SaaS Pricing Models 2026](https://www.maxio.com/blog/guide-to-saas-pricing-models-strategies-and-best-practices)
- [RevenueCat — Freemium Tier Design](https://www.revenuecat.com/blog/growth/freemium-tier-design/)
- [GoSquared — Penny Gap](https://www.gosquared.com/blog/freemium-conversion-issues)
- [Revenera — SaaS Pricing 2026](https://www.revenera.com/blog/software-monetization/saas-pricing-models-guide/)
- [InfluenceFlow — Pricing Page Best Practices 2026](https://influenceflow.io/resources/saas-pricing-page-best-practices-complete-guide-for-2026/)
- [context7 prometheus/client_golang](https://context7.com/prometheus/client_golang/llms.txt) — CounterVec best practices

Внутренние:
- `promptvault/CLAUDE.md`
- `promptvault/backend/internal/usecases/quota/quota.go`
- `promptvault/backend/internal/infrastructure/postgres/migrations/000046_concrete_plan_limits.up.sql`
- `promptvault/frontend/src/components/subscription/quota-exceeded-dialog.tsx`
- `promptvault/frontend/src/pages/pricing.tsx`

---

## Glossary

| Термин | Определение |
|---|---|
| **Persistent quota** | Лимит на «накопленное за всё время»: prompts (50/500/10 000), collections, teams, share_links, chains, chain_steps, saved_executions. Soft-deleted ресурсы НЕ считаются. |
| **Daily-window quota** | Лимит на «количество за календарный день UTC»: ext_uses_daily, mcp_uses_daily, daily_shares. Сбрасывается через date-lookup (нет cron). Реализация: composite PK `(user_id, usage_date, feature_type)` в `daily_feature_usage`. |
| **Feature gate** | Не лимит, а binary флаг «доступно/недоступно». Сейчас один — Conditional Chains (Max-only) через `quota.IsMaxTierUser()`. |
| **Effective plan** | План, по которому считается лимит для конкретного действия. Сейчас всегда `user.plan_id`. После S5 для team-resource → план owner'а команды. См. Решение 1 в §2. |
| **Team-scope ресурс** | Ресурс с `team_id != NULL`: prompt/collection/chain, созданный в team-пространстве. Отображается под «Team Workspace» в UI. После S6 квота считается по плану owner'а. |
| **Personal-scope ресурс** | Ресурс с `team_id IS NULL`. Квота всегда по `user.plan_id`. |
| **Persistent → daily трансфер** | НЕ существует. Это разные счётчики, дневной не «компенсирует» накопительный. |
| **Quota-exceeded JSON** | Ответ HTTP 402 от `RespondQuotaError`: `{error, quota_type, used, limit, plan, upgrade_url, retry_after_seconds?}`. |
| **Soft-delete vs Archive** | **Soft-delete** (`deleted_at IS NOT NULL`) — корзина, 30-day purge. **Archive** (`archived_at IS NOT NULL` после S9) — read-only бесконечный, не считается в квоту, виден в `/archive`. Семантика разная. |
| **`quota_type` (label метрики)** | Одно из 10 значений: `prompts | collections | teams | team_members | share_links | daily_shares | ext_daily | mcp_daily | chains | chain_steps`. (Saved_executions не имеет отдельного Check-метода — его лимит проверяется внутри chain.go при start_execution.) |
| **`isMaxTierForChain` → `IsMaxTierForTeam`** | Phase 16 локальная функция в `chain.go` — проверяет план owner'а команды для chain-fork-gate. После S5 рефакторится в общий `quota.IsMaxTierForTeam(userID, teamID *uint)`. |
| **`getPlan(userID)`** | Существующий внутренний helper в `quota.go:35-47`. Загружает `User.PlanID` → `Plan` через 5-минутный кеш. После S5 НЕ переименовывается, появляется параллельный `effectivePlanForTeam(userID, teamID)`. |
| **М-5c** | Внутреннее обозначение «Phase 14 quota-warning email». Был placeholder в коде (`SetEmailNotifier`, `quotaWarningThreshold`), удалён в `62036ba` как unused. **Email-warning в этом плане НЕ делается** — заменяется in-app toast (S13). |
| **М-10** | Внутреннее обозначение «downgrade preview», `2178166`. `DowngradePreview` структура с `Over*` полями. Расширяется в S8 полями `OverChains`, `OverSavedExecutions`. |
| **`promauto.NewCounterVec`** | Способ регистрации Prometheus counter в проекте (`metrics.go`). Альтернатива `prometheus.NewCounterVec + MustRegister`. Промавто проще, выбран по конвенции. |
| **Zero-init label combinations** | Паттерн в `metrics.go:init()` — `.WithLabelValues(...).Add(0)` для всех expected комбинаций. Защита от false-positive `absent_over_time` алертов когда метрика тихая. |
| **`scope` (chain metric label)** | `personal | team` — есть ли у chain `team_id`. См. `chains_created_total{scope}` в `metrics.go:49-54`. |
| **TestPass2026!** | Пароль для всех трёх test-юзеров (`e2e-{free,pro,max}@test.local`). Закоммичен в `frontend/.env.test` как `E2E_TEST_PASSWORD`. Это OK потому что тестовые юзеры не существуют на проде. |
| **5433 (PG port)** | Хост-порт PostgreSQL контейнера. Внутри контейнера 5432, на хосте 5433 — иначе конфликт с локальным Windows PG16 service. См. `docker-compose.dev.yml`. |
| **`go:embed`** | Используется в `migrate.go` для встраивания SQL-миграций в бинарник. Поэтому **миграции применяются только после `--build`** Docker'а. Перезапустить API без билда — миграции не подхватятся. |
| **CHAINS_ENABLED** | Backend feature flag (env, default false на проде). При true — включается chain-сервис, /api/chains routes, MCP chain-tools. На dev (`.env.dev`) уже `true`. См. `infrastructure/config/chains.go`. |
| **VITE_CHAINS_ENABLED** | Frontend-зеркало того же флага (env). Скрывает sidebar-пункт «Цепочки» и SPA-роуты `/chains/*` если false. См. `App.tsx:154`, `app-sidebar.tsx:48-53`. |
