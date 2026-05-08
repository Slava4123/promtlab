# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# ПромтЛаб (PromptLab)

Приложение для управления AI-промптами. Соло + команды. Self-hosted на VPS в России.

## Стек (март 2026)

### Backend
- **Go 1.25** + Chi v5 (роутер) + GORM v2 (ORM) + PostgreSQL 18
- **Config**: koanf/v2 (env + .env → struct с вложенными секциями)
- **Auth**: JWT (golang-jwt/jwt v5) access 15m + refresh 7d, bcrypt, OAuth2 + PKCE (GitHub/Google/Yandex)
- **Rate Limiting**: in-memory sliding window (middleware/ratelimit — auth 20rpm/IP)
- **Логи**: slog (text в dev, JSON в prod)
- **Профилирование**: net/http/pprof (только в dev)
- **Валидация**: go-playground/validator/v10
- **Error tracking**: sentry-go v0.44+ с GlitchTip self-hosted (Sentry-compatible API), через feature flag `SENTRY_ENABLED`
- **MCP**: `modelcontextprotocol/go-sdk` v1.5 — встроенный MCP-сервер (`internal/mcpserver/`) **v1.5.0**, 30 tools (CRUD промптов/коллекций/тегов, поиск, версии, корзина, команды, шаринг, `whoami`/`use_prompt`) + 4 chains tools (Phase 16, dark launch). Опубликован в [Official MCP Registry](https://registry.modelcontextprotocol.io/v0/servers?search=promtlab) как `ru.promtlabs/promptvault`, автопубликация по git-тегу `v*` через `.github/workflows/mcp-publish.yml` (DNS-верификация, Ed25519 private key в secret `MCP_DNS_PRIVATE_KEY`). Scoped API-keys с `allowed_tools` whitelist (синхронизирован между `apikey/constants.go` и `frontend/src/lib/mcp-tools.ts`). Квотирование: 13 из 30 tools едят дневную MCP-квоту (все write/destructive кроме idempotent UX-toggle `prompt_favorite`/`prompt_pin`/`prompt_increment_usage`).
- **Email**: SMTP-клиент (`internal/infrastructure/email/`) для писем верификации/сброса пароля
- **Миграции**: `golang-migrate/migrate` v4 (embedded SQL, папка `migrations/`)

### Frontend
- **React 19.2** + Vite 8 (Rolldown) + TypeScript
- **UI**: shadcn/ui (CLI v4, Radix) + Tailwind CSS v4.2 + Geist font, dark mode
- **State**: TanStack Query v5 (серверный) + Zustand v5 с devtools (клиентский)
- **Forms**: React Hook Form + Zod
- **Routing**: React Router 7.13
- **Error tracking**: @sentry/react v10 (с GlitchTip backend), generic `browserTracingIntegration()` (v7 router-specific не используется), через `VITE_SENTRY_ENABLED` feature flag
- **ESLint 9** flat config (`eslint.config.js`, не `.eslintrc`)
- **TypeScript strict**, path alias `@/* → src/*`
- **Тесты**: Vitest + jsdom, конфиг в `vite.config.ts` (отдельного `vitest.config.ts` нет), setup — `src/test/setup.ts`

### Deploy
- Docker Compose: отдельные `docker-compose.dev.yml` и `docker-compose.prod.yml`
- Backend: `Dockerfile.dev` (single-stage) и `Dockerfile.prod` (multi-stage + healthcheck)
- Frontend: `Dockerfile.prod` (Vite build → nginx)

## Команды разработки

### Backend (из `promptvault/backend/`)
```bash
go run ./cmd/server                                        # запуск dev-сервера
go run ./cmd/create-admin --email=you@example.com          # bootstrap первого админа
go build -o server ./cmd/server                            # сборка бинарника
go test -short ./...                                       # unit-тесты (через testing.Short() пропускаем testcontainers)
go test ./...                                              # + integration (нужен Docker для testcontainers-go v0.41)
go test -v -run TestName ./internal/usecases/prompt/       # один тест
go test -short -race -count=1 -timeout=5m ./...            # как в CI
golangci-lint run                                          # lint (конфиг: .golangci.yml)
```

### Frontend (из `promptvault/frontend/`)
```bash
npm run dev                                                # Vite dev server (proxy /api → VITE_API_URL || localhost:8080)
npm run test                                               # vitest run (jsdom, setup: src/test/setup.ts)
npm run test:watch                                         # watch-режим
npx vitest run src/hooks/use-prompts.test.ts               # один тест-файл
npm run build                                              # production build (tsc -b && vite build)
npm run lint                                               # ESLint 9 (flat config)
npm run preview                                            # vite preview билда
```

### Docker (из `promptvault/`)
```bash
docker compose -f docker-compose.dev.yml up                # full dev stack (postgres + api + frontend)
docker compose -f docker-compose.prod.yml up -d            # production (+ GlitchTip, nginx)
```

### Миграции БД
- SQL-файлы: `backend/internal/infrastructure/postgres/migrations/000NNN_description.{up,down}.sql`
- Применяются автоматически при старте сервера через `postgres.RunMigrations()`
- Новая миграция: создать пару файлов с инкрементным номером

### CI (GitHub Actions: `.github/workflows/deploy.yml`)
Pipeline: lint → test-backend + test-frontend → build-push (GHCR) → deploy (VPS via SSH)

## Архитектура Backend (Clean Architecture)

### Слои и правила зависимостей

```
cmd/server/main.go                    → config, postgres, app
  ↓
internal/app/app.go                   → repos, usecases, handlers, middleware (единая сборка)
  ↓
internal/delivery/http/<feature>/     → usecases (НИКОГДА не импортирует gorm/DB напрямую)
  ↓
internal/usecases/<feature>/          → interface/repository (через интерфейсы, НЕ реализации)
  ↓
internal/interface/repository/        → models (только интерфейсы, без gorm)
  ↓
internal/infrastructure/postgres/     → models, gorm (реализации интерфейсов)
```

### Структура

```
backend/internal/
├── app/
│   └── app.go                              # единая сборка: repos → usecases → handlers → MountRoutes
│
├── interface/repository/                   # ИНТЕРФЕЙСЫ (чистые, без gorm)
│   ├── user.go, prompt.go, collection.go
│   ├── tag.go, team.go, version.go
│
├── infrastructure/
│   ├── config/                             # koanf конфигурация
│   │   ├── config.go                       #   Config { Server, Database, JWT, OAuth, SMTP, Sentry, MCP, Payment }
│   │   ├── server.go                       #   ServerConfig + IsDev(), IsProd()
│   │   ├── database.go                     #   DatabaseConfig { Host, Port, User, ... } + DSN()
│   │   ├── jwt.go                          #   JWTConfig
│   │   ├── oauth.go                        #   OAuthConfig + OAuthProvider
│   │   ├── smtp.go                         #   SMTPConfig (email для верификации/сброса)
│   │   ├── sentry.go                       #   SentryConfig { Enabled, Dsn, Environment, Release, TracesSampleRate, Debug }
│   │   ├── mcp.go                          #   MCPConfig (встроенный MCP-сервер)
│   │   └── loader.go                       #   Load() + defaults
│   ├── postgres/
│   │   ├── postgres.go                     #   GORM connection (принимает DatabaseConfig)
│   │   ├── migrate.go                      #   golang-migrate (embedded SQL, 61+ миграция)
│   │   └── repository/                     #   GORM реализации интерфейсов
│   │       └── user_repo.go
│   ├── email/                              #   SMTP-клиент
│   ├── metrics/                            #   Prometheus counters (loop panics, branding, chains)
│   ├── payment/                            #   T-Bank SDK + webhook signature/idempotency
│   └── telemetry/                          #   OpenTelemetry tracing (Tempo)
│
├── models/                                 # GORM entities (один пакет, НЕ разносить по папкам)
│   ├── user.go, team.go, collection.go
│   ├── prompt.go, tag.go, version.go
│
├── usecases/<feature>/                     # бизнес-логика
│   ├── <feature>.go                        #   Service с полными flows
│   ├── types.go                            #   входные/выходные типы
│   ├── errors.go                           #   доменные ошибки
│   └── constants.go                        #   константы
│
├── delivery/http/
│   ├── errors/errors.go                    #   AppError, BadRequest(), Respond() — общее
│   ├── utils/                              #   WriteJSON, DecodeJSON, Pagination — общее
│   │   ├── response.go, request.go, pagination.go
│   └── <feature>/                          #   HTTP transport (по одному пакету на фичу)
│       ├── handler.go                      #   handlers (без DB!)
│       ├── request.go                      #   request DTOs + валидация
│       ├── response.go                     #   response DTOs + конверторы
│       └── errors.go                       #   маппинг доменных ошибок → HTTP
│       #
│       # Актуальные фичи: activity, admin, adminauth, analytics, apikey,
│       # audit, auth, badge, chain (Phase 16), changelog, collection,
│       # engagement, feedback, metadata, oauth_server, prompt, quota,
│       # search, seo, share, starter, streak, subscription, tag, team,
│       # teamcheck, testcleanup, trash, user, webhook
│
├── middleware/
│   ├── auth/                               #   JWT + API-key middleware
│   │   ├── auth.go                         #     JWT Middleware()
│   │   ├── apikey.go                       #     API-key валидация
│   │   ├── combined.go                     #     JWT ИЛИ API-key (комбинированный)
│   │   ├── types.go                        #     TokenValidator interface
│   │   └── constants.go                    #     UserIDKey, BearerScheme
│   ├── cors/
│   │   └── cors.go                         #   CORS middleware (go-chi/cors)
│   ├── logger/
│   │   └── logger.go                       #   slog request logger
│   ├── ratelimit/
│   │   └── ratelimit.go                    #   Rate limiting (sliding window): IP для auth, ByUserID 401-fail-closed
│   ├── ipallowlist/
│   │   └── ipallowlist.go                  #   IP/CIDR whitelist для webhook + /metrics
│   ├── metrics/
│   │   └── metrics.go                      #   Prometheus middleware (RED метрики на HTTP)
│   ├── secheaders/                         #   X-Frame-Options/XCTO/Referrer/HSTS (MJ-12)
│   ├── sentry/                             #   Sentry/GlitchTip error tracking
│   │   └── sentry.go                       #     Handler() + UserContext (sentry.User{ID} из JWT claims)
│   └── admin/                              #   Admin panel middleware
│       ├── require_admin.go                #     RequireAdmin(users) — re-check role+status из БД
│       ├── audit_context.go                #     AdminAuditContext(trustProxy) — кладёт AdminRequestInfo в ctx
│       └── constants.go                    #     FreshTOTPTTL = 12h
│
├── pkg/safeloop/                           # CR-6: defer recover() helper для background loops
│   └── safeloop.go                         #     RunWithRecover(name, fn) + Prometheus counter
│
└── mcpserver/                              # встроенный MCP v1.5.0 (go-sdk v1.5)
    ├── server.go                             #   NewMCPServer — регистрация tools/resources/prompts, DNS/github auth
    ├── tools.go                              #   30 tool handlers; parseMCPQuota/incrementMCPUsage у write-операций
    ├── interfaces.go                         #   {Prompt,Collection,Tag,Search,Share,Team,Trash,User}Service — mock-able
    ├── scope.go                              #   enforceScope + enforceTeamID для scoped API-keys (allowed_tools)
    ├── errors.go                             #   mapDomainError — доменные → string для JSON-RPC
    ├── cursor.go                             #   keyset cursor-пагинация для list_prompts
    ├── cache.go                              #   30s in-memory TTL для list_collections/list_tags
    └── notifier.go                           #   resources/updated подписки (SDK трекает, notifier шлёт)
```

## Правила разработки

### Общие
- Язык интерфейса: русский
- Никаких западных SaaS (без Clerk, без Vercel hosting) — self-hosted
- Все переменные окружения через `.env` → koanf, Docker-compose только `env_file: .env`
- Error tracking: GlitchTip self-hosted (НЕ Sentry.io — заблокирован для РФ с сентября 2024), SDK совместим с Sentry (drop-in replacement через DSN swap при необходимости)

### Backend — Go
- **Каждая фича = отдельный пакет** в usecases/, delivery/http/, middleware/
- **Внутри пакета** — разделять на файлы: handler.go, types.go, errors.go, constants.go
- **Handler НИКОГДА не владеет *gorm.DB** — только вызов usecase Service
- **Usecase владеет полным flow** — Register(), Login(), не отдельные HashPassword()
- **Repository — интерфейсы** в `interface/repository/`, реализации в `infrastructure/postgres/repository/`
- **Ошибки**: доменные в `usecases/<feature>/errors.go`, HTTP-маппинг в `delivery/http/<feature>/errors.go`
- **Общие утилиты**: `delivery/http/utils/` (WriteJSON, DecodeJSON), `delivery/http/errors/` (AppError, Respond)
- **Config**: вложенные секции `cfg.Server.Port`, `cfg.Database.DSN()`, `cfg.JWT.Secret`
- **Логи**: slog, НЕ log.Printf
- **models** — один пакет, не разносить по папкам (избежать circular imports с GORM relations)

### Frontend — React
- TanStack Query для всех API-запросов
- Zustand с devtools middleware для auth/theme
- shadcn/ui компоненты, НЕ кастомные с нуля
- fetch wrapper с JWT auto-refresh, НЕ axios
- Структура: api/, components/{ui,layout,auth,prompts,teams,collections,tags,settings,feedback}, pages/, hooks/, stores/, lib/, test/

### Docker
- Отдельные файлы: `docker-compose.dev.yml`, `docker-compose.prod.yml` (без base)
- Backend: `Dockerfile.dev` (single-stage), `Dockerfile.prod` (multi-stage + healthcheck)
- НЕ дублировать env переменные в docker-compose — всё через `env_file: .env`

## Ключевые решения
- Rate limiting: по IP для auth (middleware/ratelimit)
- Команды с ролями: owner / editor / viewer
- Версионирование промптов: каждое изменение = новая PromptVersion
- Error tracking: GlitchTip (Sentry-compatible) self-hosted в `docker-compose.prod.yml` (web+worker+valkey), external Postgres (отдельная БД `glitchtip` на том же managed инстансе), source maps upload через `sentry-cli` в GitHub Actions (НЕ `@sentry/vite-plugin` — Vite 8 slowdown), feature flag `SENTRY_ENABLED` для gradual rollout, `RespondWithRequest(w,r,err)` в `delivery/http/errors` для захвата 5xx с user.id из Sentry Hub
- OAuth Account Linking: HMAC-подписанная cookie (oauth_link) + PKCE (S256) на всех провайдерах
- Установка пароля: двухшаговая через email-код (OAuth-юзеры)
- Смена пароля: через старый пароль + email-уведомление
- Забыли пароль: публичный flow через email-код (не раскрывает существование аккаунта)
- MCP autopublish: bump `promptvault/server.json` → commit → `git tag v1.3.0 && git push --tags` → workflow сам логинится в Registry через DNS (secret `MCP_DNS_PRIVATE_KEY`), публикует и создаёт GitHub Release. Локально `mcp-publisher` ставить не нужно. Не-bump'нутые тэги падают на шаге Verify server.json version matches tag.
- SPA chunk-load-error после деплоя: `src/components/error-boundary.tsx` детектит "Failed to fetch dynamically imported module" и один раз `location.reload()` (флаг в sessionStorage, очищается при mount в `main.tsx`).
- **Phase 15 (доделка базовых мест):**
  - **Smart Insights:** kill-switch `ANALYTICS_EXPERIMENTAL_INSIGHTS` (default `true`). Все 7 типов; `possible_duplicates` skip если pg_trgm недоступен (probe `postgres.DetectExtensions` на старте). Team scope добавлен через `TeamRepository.ListOwnedTeams` + второй проход в `InsightsComputeLoop`.
  - **Admin ChangeTier:** доделан в `usecases/admin/admin.go`. Заменил `subs.CancelAtPeriodEnd` → `MarkExpired` + ловит paused-подписки через `subs.GetCurrentByUserID`. Audit-payload содержит `reason`+`source:"admin_override"`. Email через `TierChangeNotifier` (non-blocking).
  - **Slug:** транслитерация cyrillic через `mozillazg/go-unidecode` (раньше "Мой промпт" → `p-<id>`, теперь `moi-promt-<id>`). Backward-compat через `prompts.slug_aliases jsonb` — старая опубликованная ссылка продолжает резолвиться.
  - **M9 (analytics hot-path):** `Claims.PlanID` в JWT access-токене. `analytics.Service.lookupPlanID` читает из ctx через injected callback (`SetPlanFromCtx` в `app.go`), fallback на `users.GetByID` для legacy-JWT. Минус 1 DB-hit на каждый /api/analytics/* запрос.
  - **Search FTS:** `prompts.search_tsv` GENERATED STORED tsvector с `russian_unaccent` + english stemming. GIN-индекс. `websearch_to_tsquery` устойчив к user-input. Гибрид с pg_trgm `similarity(title::text, ?::text) > 0.3` для опечаток (НЕ оператор `%` — см. ADR 0004 про gotcha с varchar/text cast). Score в ORDER BY = `ts_rank_cd × 0.7 + similarity × 0.3`.
  - **Extension Sentry:** реальная отправка envelope NDJSON POST в GlitchTip (раньше только console). `WXT_SENTRY_DSN` в .env, host_permissions добавлены, rate-limit 10/min.
- **Phase 16-X (Branding UX):** загрузка логотипа файлом (bytea в `team_logo_files`, миграция 000060) + visual color picker (12 brand-чипов из `lib/branding/colors.ts` + native `<input type="color">`, без новых JS-deps). API: `POST/DELETE /api/teams/:slug/branding/logo` (owner Max-only, ratelimit 10/час/userID), public `GET /api/teams/:slug/branding/logo` с ETag (sha256) + `Cache-Control: public, max-age=86400`. Magic-byte валидация через std `image.Decode` (PNG/JPEG) + `golang.org/x/image/webp`. Дискриминатор источника `teams.brand_logo_source ∈ {url, file, none}` — backward compat для существующих `brand_logo_url`. ADR 0006 фиксирует выбор bytea vs FS+nginx vs MinIO. Метрики `team_branding_logo_uploads_total{result}`, `team_branding_logo_size_bytes`, `team_branding_logo_serve_total{cache_hit}`. Runbook `docs/runbooks/TeamBrandingUploadErrors.md`.
- **Phase 16 (Prompt Chains, dark launch):**
  - Цепочка = упорядоченная последовательность шагов; output одного → input следующего. Run-mode wizard на фронте; LLM-вызов делает MCP-клиент или сам юзер (copy-paste). Принцип «без AI на нашей стороне» сохраняется (маржа 92%).
  - **Schema (миграция 000053):** `prompt_chains`, `prompt_chain_steps`, `prompt_chain_executions`. `prompt_chain_steps.uq_prompt_chain_steps_position` — DEFERRABLE INITIALLY DEFERRED (для атомарного reorder в одной транзакции). `prompt_chain_executions.chain_snapshot JSONB` — заморозка структуры + контента промптов на момент Start (Edit во время Run не ломает execution).
  - **Tier-лимиты:** Free 1 цепочка × 3 шага × 0 saved exec; Pro 5×10×10; Max 100×50×1000. Конкретные числа (паттерн 000046), не sentinel -1.
  - **API:** `/api/chains` (CRUD + steps + reorder + executions), `/api/executions/:id` (read + advance). Status: `in_progress` → `completed`/`abandoned`. AdvanceStep пишет output под ключом `step_<id>` в `step_outputs JSONB`.
  - **MCP (4 tools):** `list_chains`, `get_chain` (read), `start_chain_execution`, `advance_chain_step` (write, едят MCP-квоту). Loop: клиент сам вызывает LLM между Start и Advance. **Эксперимент** с Cursor/Claude Code на multi-step tool-loop — обязателен перед launch.
  - **Feature flag:** `CHAINS_ENABLED` (env, default `false`) → backend nil-wires chainSvc/handler/MCP-tools, не регистрирует routes. `VITE_CHAINS_ENABLED` (frontend) → скрывает sidebar item «Цепочки» и не регистрирует роуты `/chains/*`. Включается одновременно после QA.
  - **Frontend:** `pages/chains/{index,editor,run,runs,canvas}.tsx` (5 страниц). Tree-canvas через `@xyflow/react` + `elkjs` для DAG-layout (заменили `@dnd-kit` после Phase 16 v2). Multi-step wizard через `useState<RunStep>`. Пункт меню `Link2` icon в `app-sidebar.tsx`.
  - **Phase B (Conditional Chains, Max-only):** реализовано dark launch. Миграция 000054 (`step_type`, `conditions JSONB` + CHECK constraint). 7 matchers (`contains, not_contains, regex, equals, starts_with, ends_with, length_gt, length_lt`) + AND/OR/NOT с MaxConditionDepth=10. Cycle detection через DFS на graph branches (запрещает self-reference). ReDoS защита: Go RE2 + MaxRegexPatternLen=500. Tier-check: `quotas.IsMaxTierUser` (plan_id ∈ {max, max_yearly}); Pro юзер → 403. Frontend: JSON-textarea condition builder (visual editor — Phase 1 polish), badge «условный» в editor + run-wizard.
  - **MCP server bump v1.4.0 → v1.5.0** (4 новых chains tools).
  - **Pre-activation fixes:** `trashRepo.PurgeExpired`/`EmptyTrash` skip prompts с `id IN (SELECT prompt_id FROM prompt_chain_steps)` — иначе FK 23503. Prometheus counters: `chains_created_total{scope}`, `chain_executions_started_total`, `chain_executions_completed_total{status}`, `chain_conditional_evaluated_total{result}` (последний — legacy v1 DSL, не инкрементится; удалить отдельно).
  - **Team RBAC (Wave team-aware):** `viewer = читатель + runner` — может Read/StartExecution/AdvanceStep (свой initiated), но **не** Create/Update/Delete/AddStep/RemoveStep/UpdateStep/MoveStep. `editor`/`owner` — полный write. Backend через `checkReadAccess` (read-only RBAC) и `checkEditAccess` (RequireEditor) в `chain.go:1010-1028`; `GetExecution` дополнительно делает initiator-only check (`exec.UserID != userID`) + актуальный `checkReadAccess` к chain (security fix: юзер выгнан из команды между Start и Advance → 403). Frontend hook `useCurrentTeamRole` в `pages/chains/index.tsx` и `editor.tsx` — viewer не видит «Создать»/«Удалить»/Save/Add/Remove/Move кнопки, лейбл «Редактор» на карточке заменяется на «Просмотр».
  - **Fork tier-gate (справедливая модель):** для personal-цепочки проверяется `userID.plan_id ∈ {max, max_yearly}`; для team-цепочки — план **любого owner** команды (`isMaxTierForChain` в `chain.go`). Если хотя бы один owner на Max → все editor'ы команды могут добавлять fork (даже Pro/Free). Это поощряет owner'а апгрейдиться (его tier «дарит» команде feature). Frontend в team-mode оптимистично показывает кнопку «+ Развилка» enabled (бэк проверит и вернёт `ErrForkRequiresMax` если нужно).
  - **Activity feed events** (Phase 14 wire-up): `chain.created`, `chain.updated`, `chain.deleted`, `chain.execution_started`, `chain.execution_completed` (`models.ActivityChain*` в `team_activity.go`). Записываются только для team-цепочек через `s.activity.LogSafe`. Step-level events намеренно НЕ трекаются (noisy для длинных цепочек).
  - **Phase C (отложено):** 3 starter chains (PRD/Code Review/Контент) — отдельный мини-PR, требует расширения starter catalog. Task #17 (unit-тесты chain.Service) и task #18 (Playwright E2E) — тех.долг. MCP multi-step loop эксперимент с Cursor/Claude Code — обязателен ДО включения CHAINS_ENABLED=true в prod.

## Phase 15 dev-workflow gotchas (от QA findings)

- **`docker compose up -d` БЕЗ `--build` использует устаревший образ** — миграции/embed файлы из go:embed не подхватятся. Всегда: `docker compose -f docker-compose.dev.yml up -d --build` после изменения миграций или Go-кода.
- **Frontend anonymous volume `/app/node_modules`** не пересоздаётся при `--build`. После `npm install` нового пакета: `docker compose down` (без `-v` — pgdata save) + `docker compose up -d --build`.
- **pg_trgm `%`/`%%` оператор + GORM `Raw()`** — источник regressions. Использовать функцию `similarity(title::text, ?::text) > threshold` вместо оператора. См. ADR 0004.
- **In-memory rate-limit** для destructive endpoints (e.g. `/insights/refresh` 1/час) хранится per-process. Сбрасывается через `docker compose restart api`, НЕ через DELETE из БД.
- **Юнит-тесты с моками не ловят SQL-regressions** — для repository методов с raw SQL нужен integration test через testcontainers с реальной PG + extensions.
- **`UsageChart` (analytics): chartConfig — per-instance prop, не module-scope.** Если новый график использует тот же `UsageChart`, передавайте `chartConfig={createUsageChartConfig("…label…")}` из `components/analytics/usage-chart-config.ts`. Без этого tooltip покажет дефолтный лейбл «Использования» — раньше было багом графика «Создание промптов по дням».
- **Шаблонные `{{var}}` имеют идентичную грамматику на frontend (`lib/template/parse.ts`) и backend (`internal/template/template.go`)** — менять одну сторону без другой нельзя. Подсветка в редакторе использует тот же regex (`lib/codemirror/template-variable-highlight.ts`). Решение оставить custom-парсер вместо Mustache/Handlebars — см. ADR 0005.

## Документация (`docs/`)
- `PLAN.md` — 16+ фазный план разработки (Phase 13 платежи, 14 collab, 15 polish, 16 chains)
- `FEATURES.md` — каталог 104 идей с tier'ами (1-4) и ✅ маркерами закрытых
- `DEPLOY.md` — Docker Compose + GitHub Actions + VPS + GlitchTip setup
- `OBSERVABILITY.md` — Prometheus metrics + Sentry breadcrumbs + alert rules
- `SEO.md` — архитектура server-rendered HTML для ботов + sitemap + OG-images
- `SENTRY_NEXT_STEPS.md` — опциональные расширения Sentry (performance tracing, alerts); reference активировать по триггерам
- `MCP.md` — полный справочник сервера (30 tools с квотными аннотациями, resources, use_prompt, cursor-пагинация)
- `MCP-PUBLISHING.md` — автопубликация CI, DNS setup, roadmap по всем каталогам (Registry/Anthropic Connectors/Smithery/Glama/PulseMCP/Cline)
- `ANTHROPIC_CONNECTORS.md` — blueprint для подачи в Anthropic MCP Directory
- `MIGRATIONS.md` — best-practices для SQL-миграций (CONCURRENTLY, NOT VALID, 3-шаговый GENERATED rollout)
- `BUSINESS_RESEARCH.md` — анализ конкурентов + ценообразование Free/Pro/Max
- `FEATURE_PROMPT_CHAINS.md` — полная спецификация Phase 16 (tree-canvas, fork, RBAC)
- `QUOTAS_IMPROVEMENTS_PLAN.md` — TTL-модель квот (заменяет active-count + daily-create из Phase 13)
- `SLO.md` — service level objectives + alert thresholds
- `REVIEW_2026-05-07.md` — последний полный архитектурный аудит (17 Critical + 40 Major + 80 Minor)
- `adr/` — architecture decision records (FTS, JWT PlanID, pg_trgm gotcha, template syntax, bytea logo)
- `runbooks/` — инциденты SRE: `CleanupLoopStalled`, `InsightsComputeLoopStalled`, `TeamBrandingUploadErrors`
- `archive/` — выполненные/исторические документы: `TODO.md`, `LAUNCH_PLAN.md`, `RELEASE_READINESS.md`, `ANTHROPIC_CONNECTORS_TODO.md`, `PHASE14_SELF_REVIEW.md`, `PROGRESS_2026-04-20.md`, `SUBSCRIPTION_PLAN.md` (Phase 13 план), `SECURITY_AUDIT.md` (чеклист перед открытием репо), `BACKLOG_PHASE14.md` (матрица закрытия Phase 14.3 wave 1-4), `REVIEW_2026-05-07_v1.md` (предыдущая итерация ревью)

## Admin Panel (кратко)

- **Bootstrap первого админа:** `go run ./cmd/create-admin --email=you@example.com`
- **TOTP 2FA обязательна** для всех destructive actions, 12h TTL по дефолту.
- **Audit log append-only** через PostgreSQL BEFORE UPDATE/DELETE триггеры.
- **Endpoints** `/api/admin/users/*`, `/api/admin/audit`, `/api/admin/health`.
- **Frontend** `/admin/users`, `/admin/users/:id`, `/admin/audit`, `/admin/health`, `/admin/totp`.
