# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# ПромтЛаб (PromptLab)

Приложение для управления AI-промптами. Соло + команды. Self-hosted на VPS в России.

## Стек (март 2026)

### Backend
- **Go 1.25** + Chi v5 (роутер) + GORM v2 (ORM) + PostgreSQL 18
- **Config**: koanf/v2 (env + .env → struct с вложенными секциями)
- **Auth**: JWT (golang-jwt/jwt v5) access 15m + refresh 7d, bcrypt, OAuth2 + PKCE (GitHub/Google/Yandex)
- **Rate Limiting**: in-memory sliding window (middleware/ratelimit — auth 20rpm/IP, AI per-user)
- **AI**: OpenRouter API, серверный ключ, SSE-стриминг
- **Логи**: slog (text в dev, JSON в prod)
- **Профилирование**: net/http/pprof (только в dev)
- **Валидация**: go-playground/validator/v10
- **Error tracking**: sentry-go v0.44+ с GlitchTip self-hosted (Sentry-compatible API), через feature flag `SENTRY_ENABLED`
- **MCP**: `modelcontextprotocol/go-sdk` v1.5 — встроенный MCP-сервер (`internal/mcpserver/`) для Claude
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
│   │   ├── config.go                       #   Config { Server, Database, JWT, OAuth, SMTP, AI, Sentry, MCP }
│   │   ├── server.go                       #   ServerConfig + IsDev(), IsProd()
│   │   ├── database.go                     #   DatabaseConfig { Host, Port, User, ... } + DSN()
│   │   ├── jwt.go                          #   JWTConfig
│   │   ├── oauth.go                        #   OAuthConfig + OAuthProvider
│   │   ├── smtp.go                         #   SMTPConfig (email для верификации/сброса)
│   │   ├── ai.go                           #   AIConfig + ModelConfig
│   │   ├── sentry.go                       #   SentryConfig { Enabled, Dsn, Environment, Release, TracesSampleRate, Debug }
│   │   ├── mcp.go                          #   MCPConfig (встроенный MCP-сервер)
│   │   └── loader.go                       #   Load() + defaults
│   ├── postgres/
│   │   ├── postgres.go                     #   GORM connection (принимает DatabaseConfig)
│   │   ├── migrate.go                      #   golang-migrate (embedded SQL, 5+ миграций)
│   │   └── repository/                     #   GORM реализации интерфейсов
│   │       └── user_repo.go
│   ├── openrouter/                         #   клиент OpenRouter API (AI, SSE-стриминг)
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
│       # Актуальные фичи: admin, adminauth, ai, apikey, auth, badge,
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
│   │   └── ratelimit.go                    #   Rate limiting (sliding window): IP для auth, userID для AI
│   ├── sentry/                             #   Sentry/GlitchTip error tracking
│   │   └── sentry.go                       #     Handler() + UserContext (sentry.User{ID} из JWT claims)
│   └── admin/                              #   Admin panel middleware
│       ├── require_admin.go                #     RequireAdmin(users) — re-check role+status из БД
│       ├── audit_context.go                #     AdminAuditContext — кладёт AdminRequestInfo в ctx
│       └── constants.go                    #     FreshTOTPTTL = 12h
│
└── mcpserver/                              # встроенный MCP-сервер для Claude (go-sdk v1.5)
```

## Правила разработки

### Общие
- Язык интерфейса: русский
- Никаких западных SaaS (без Clerk, без Vercel hosting) — self-hosted
- AI-ключ серверный: один `OPENROUTER_API_KEY` в `.env`, пользователи НЕ вводят свои ключи
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
- Структура: api/, components/{ui,layout,auth,prompts,ai,teams,collections,tags,settings,feedback}, pages/, hooks/, stores/, lib/, test/

### Docker
- Отдельные файлы: `docker-compose.dev.yml`, `docker-compose.prod.yml` (без base)
- Backend: `Dockerfile.dev` (single-stage), `Dockerfile.prod` (multi-stage + healthcheck)
- НЕ дублировать env переменные в docker-compose — всё через `env_file: .env`

## Ключевые решения
- Rate limiting: по userID для AI, по IP для auth (middleware/ratelimit)
- Команды с ролями: owner / editor / viewer
- Версионирование промптов: каждое изменение = новая PromptVersion
- SSE streaming для AI-ответов
- Error tracking: GlitchTip (Sentry-compatible) self-hosted в `docker-compose.prod.yml` (web+worker+valkey), external Postgres (отдельная БД `glitchtip` на том же managed инстансе), source maps upload через `sentry-cli` в GitHub Actions (НЕ `@sentry/vite-plugin` — Vite 8 slowdown), feature flag `SENTRY_ENABLED` для gradual rollout, `RespondWithRequest(w,r,err)` в `delivery/http/errors` для захвата 5xx с user.id из Sentry Hub
- OAuth Account Linking: HMAC-подписанная cookie (oauth_link) + PKCE (S256) на всех провайдерах
- Установка пароля: двухшаговая через email-код (OAuth-юзеры)
- Смена пароля: через старый пароль + email-уведомление
- Забыли пароль: публичный flow через email-код (не раскрывает существование аккаунта)
- Единственная AI-модель: anthropic/claude-sonnet-4 (default_model в User)

## Документация (`docs/`)
- `PLAN.md` — 12-фазный план разработки
- `TODO.md` — чеклист с отметками выполнения
- `FEATURES.md` — детализация фич по фазам
- `DEPLOY.md` — Docker Compose + GitHub Actions + VPS
- `MONETIZATION.md` — тарифы Free/Pro/Max, маржа, квоты
- `ADMIN.md` — bootstrap админа, TOTP 2FA, audit log, admin actions
- `SENTRY_NEXT_STEPS.md` — GlitchTip self-hosted setup, source maps
- `MCP.md`, `MCP-PUBLISHING.md` — MCP-интеграция и публикация
- `EXTENSION_ROADMAP.md` — планы браузерного расширения

## Admin Panel (кратко)

- **Bootstrap первого админа:** `go run ./cmd/create-admin --email=you@example.com`
- **TOTP 2FA обязательна** для всех destructive actions, 12h TTL по дефолту.
- **Audit log append-only** через PostgreSQL BEFORE UPDATE/DELETE триггеры.
- **Endpoints** `/api/admin/users/*`, `/api/admin/audit`, `/api/admin/health`.
- **Frontend** `/admin/users`, `/admin/users/:id`, `/admin/audit`, `/admin/health`, `/admin/totp`.
- Полная документация — [`docs/ADMIN.md`](docs/ADMIN.md).
