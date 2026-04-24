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
- **MCP**: `modelcontextprotocol/go-sdk` v1.5 — встроенный MCP-сервер (`internal/mcpserver/`) **v1.2.0**, 30 tools (CRUD промптов/коллекций/тегов, поиск, версии, корзина, команды, шаринг, `whoami`/`use_prompt`). Опубликован в [Official MCP Registry](https://registry.modelcontextprotocol.io/v0/servers?search=promtlab) как `ru.promtlabs/promptvault`, автопубликация по git-тегу `v*` через `.github/workflows/mcp-publish.yml` (DNS-верификация, Ed25519 private key в secret `MCP_DNS_PRIVATE_KEY`). Scoped API-keys с `allowed_tools` whitelist (синхронизирован между `apikey/constants.go` и `frontend/src/lib/mcp-tools.ts`). Квотирование: 13 из 30 tools едят дневную MCP-квоту (все write/destructive кроме idempotent UX-toggle `prompt_favorite`/`prompt_pin`/`prompt_increment_usage`).
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
│   │   ├── migrate.go                      #   golang-migrate (embedded SQL, 5+ миграций)
│   │   └── repository/                     #   GORM реализации интерфейсов
│   │       └── user_repo.go
│   └── email/                              #   SMTP-клиент
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
│       # Актуальные фичи: admin, adminauth, apikey, auth, badge,
│       # changelog, collection, feedback, prompt, search, share, starter,
│       # streak, tag, team, trash, user
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
│   │   └── ratelimit.go                    #   Rate limiting (sliding window): IP для auth
│   ├── sentry/                             #   Sentry/GlitchTip error tracking
│   │   └── sentry.go                       #     Handler() + UserContext (sentry.User{ID} из JWT claims)
│   └── admin/                              #   Admin panel middleware
│       ├── require_admin.go                #     RequireAdmin(users) — re-check role+status из БД
│       ├── audit_context.go                #     AdminAuditContext — кладёт AdminRequestInfo в ctx
│       └── constants.go                    #     FreshTOTPTTL = 12h
│
└── mcpserver/                              # встроенный MCP v1.2.0 (go-sdk v1.5)
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

## Документация (`docs/`)
- `PLAN.md` — 12-фазный план разработки
- `BACKLOG.md` — матрица качества + статус задач по волнам Phase 14.x
- `FEATURES.md` — каталог 104 идей с tier'ами (1-4) и ✅ маркерами закрытых
- `DEPLOY.md` — Docker Compose + GitHub Actions + VPS + GlitchTip setup
- `OBSERVABILITY.md` — Prometheus metrics + Sentry breadcrumbs + alert rules
- `SEO.md` — архитектура server-rendered HTML для ботов + sitemap + OG-images
- `SENTRY_NEXT_STEPS.md` — опциональные расширения Sentry (performance tracing, alerts); reference активировать по триггерам
- `MCP.md` — полный справочник сервера (30 tools с квотными аннотациями, resources, use_prompt, cursor-пагинация)
- `MCP-PUBLISHING.md` — автопубликация CI, DNS setup, roadmap по всем каталогам (Registry/Anthropic Connectors/Smithery/Glama/PulseMCP/Cline)
- `ANTHROPIC_CONNECTORS.md` — blueprint для подачи в Anthropic MCP Directory
- `cline-submission-draft.md` — готовые значения для формы Cline Marketplace (репо public, подача не сделана)
- `archive/` — выполненные/исторические документы: `TODO.md`, `LAUNCH_PLAN.md`, `RELEASE_READINESS.md`, `ANTHROPIC_CONNECTORS_TODO.md`, `PHASE14_SELF_REVIEW.md`, `PROGRESS_2026-04-20.md`, `SUBSCRIPTION_PLAN.md` (Phase 13 план), `SECURITY_AUDIT.md` (чеклист перед открытием репо)

## Admin Panel (кратко)

- **Bootstrap первого админа:** `go run ./cmd/create-admin --email=you@example.com`
- **TOTP 2FA обязательна** для всех destructive actions, 12h TTL по дефолту.
- **Audit log append-only** через PostgreSQL BEFORE UPDATE/DELETE триггеры.
- **Endpoints** `/api/admin/users/*`, `/api/admin/audit`, `/api/admin/health`.
- **Frontend** `/admin/users`, `/admin/users/:id`, `/admin/audit`, `/admin/health`, `/admin/totp`.
