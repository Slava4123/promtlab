# ПромтЛаб (PromptLab)

Приложение для управления AI-промптами. Соло + команды. Self-hosted на VPS в России.

## Стек (март 2026)

### Backend
- **Go 1.26** + Chi v5 (роутер) + GORM v2 (ORM) + PostgreSQL 18
- **Config**: koanf/v2 (env + .env → struct с вложенными секциями)
- **Auth**: JWT (golang-jwt/jwt v5) access 15m + refresh 7d, bcrypt, OAuth2 + PKCE (GitHub/Google/Yandex)
- **Rate Limiting**: in-memory sliding window (middleware/ratelimit — auth 20rpm/IP, AI per-user)
- **AI**: OpenRouter API, серверный ключ, SSE-стриминг
- **Логи**: slog (text в dev, JSON в prod)
- **Профилирование**: net/http/pprof (только в dev)
- **Валидация**: go-playground/validator/v10

### Frontend
- **React 19.2** + Vite 8 (Rolldown) + TypeScript
- **UI**: shadcn/ui (CLI v4, Radix) + Tailwind CSS v4.2 + Geist font, dark mode
- **State**: TanStack Query v5 (серверный) + Zustand v5 с devtools (клиентский)
- **Forms**: React Hook Form + Zod
- **Routing**: React Router 7.13

### Deploy
- Docker Compose: отдельные `docker-compose.dev.yml` и `docker-compose.prod.yml`
- Backend: `Dockerfile.dev` (single-stage) и `Dockerfile.prod` (multi-stage + healthcheck)
- Frontend: `Dockerfile.prod` (Vite build → nginx)

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
│   │   ├── config.go                       #   Config { Server, Database, JWT, OAuth, AI }
│   │   ├── server.go                       #   ServerConfig + IsDev(), IsProd()
│   │   ├── database.go                     #   DatabaseConfig { Host, Port, User, ... } + DSN()
│   │   ├── jwt.go                          #   JWTConfig
│   │   ├── oauth.go                        #   OAuthConfig + OAuthProvider
│   │   ├── ai.go                           #   AIConfig + ModelConfig
│   │   └── loader.go                       #   Load() + defaults
│   └── postgres/
│       ├── postgres.go                     #   GORM connection (принимает DatabaseConfig)
│       ├── migrate.go                      #   golang-migrate (embedded SQL migrations)
│       └── repository/                     #   GORM реализации интерфейсов
│           └── user_repo.go
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
│   └── <feature>/                          #   HTTP transport
│       ├── handler.go                      #   handlers (без DB!)
│       ├── request.go                      #   request DTOs + валидация
│       ├── response.go                     #   response DTOs + конверторы
│       └── errors.go                       #   маппинг доменных ошибок → HTTP
│
└── middleware/
    ├── auth/                               #   JWT middleware
    │   ├── auth.go                         #     Middleware()
    │   ├── types.go                        #     TokenValidator interface
    │   └── constants.go                    #     UserIDKey, BearerScheme
    ├── cors/
    │   └── cors.go                         #   CORS middleware
    └── ratelimit/
        └── ratelimit.go                    #   Rate limiting по IP (sliding window)
```

## Правила разработки

### Общие
- Язык интерфейса: русский
- Никаких западных SaaS (без Clerk, без Vercel hosting) — self-hosted
- AI-ключ серверный: один `OPENROUTER_API_KEY` в `.env`, пользователи НЕ вводят свои ключи
- Все переменные окружения через `.env` → koanf, Docker-compose только `env_file: .env`

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
- Структура: api/, components/{ui,layout,prompts,ai,teams,collections,tags,auth}, pages/, hooks/, stores/, lib/

### Docker
- Отдельные файлы: `docker-compose.dev.yml`, `docker-compose.prod.yml` (без base)
- Backend: `Dockerfile.dev` (single-stage), `Dockerfile.prod` (multi-stage + healthcheck)
- НЕ дублировать env переменные в docker-compose — всё через `env_file: .env`

## Ключевые решения
- Rate limiting: по userID для AI, по IP для auth (middleware/ratelimit)
- Команды с ролями: owner / editor / viewer
- Версионирование промптов: каждое изменение = новая PromptVersion
- SSE streaming для AI-ответов
- OAuth Account Linking: HMAC-подписанная cookie (oauth_link) + PKCE (S256) на всех провайдерах
- Установка пароля: двухшаговая через email-код (OAuth-юзеры)
- Смена пароля: через старый пароль + email-уведомление
- Забыли пароль: публичный flow через email-код (не раскрывает существование аккаунта)
- Единственная AI-модель: anthropic/claude-sonnet-4 (default_model в User)

## План разработки
- Полный план: `docs/PLAN.md` — 12 фаз
- Прогресс: `docs/TODO.md` — чеклист с отметками выполнения
