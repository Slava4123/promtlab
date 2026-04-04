# ПромтЛаб — План разработки (Go + React)

## Контекст

Приложение для сохранения и управления AI-промптами с AI-редактированием через OpenRouter (серверный ключ). Соло-пользователи + команды с ролями. Self-hosted деплой на VPS в России через Docker.

Пользователь знает Go, предпочитает раздельные backend/frontend.

---

## Стек технологий (актуальные версии, март 2026)

| Слой | Технология | Версия |
|------|-----------|--------|
| **Backend** | Go | 1.26 |
| **Router** | go-chi/chi | v5 |
| **ORM** | GORM | v2 (gorm.io/gorm) |
| **Database** | PostgreSQL | 18 |
| **JWT** | golang-jwt/jwt | v5 |
| **OAuth2** | golang.org/x/oauth2 | latest |
| **Пароли** | golang.org/x/crypto/bcrypt | latest |
| **Frontend** | React | 19.2 |
| **Bundler** | Vite (Rolldown) | 8.0 |
| **Routing** | React Router | 7.13 |
| **UI** | shadcn/ui (CLI v4, Radix) | latest |
| **CSS** | Tailwind CSS | v4.2 |
| **Server State** | TanStack Query | v5 |
| **Client State** | Zustand | v5 |
| **Forms** | React Hook Form + Zod | latest |
| **Icons** | lucide-react | latest |
| **Font** | Geist (Sans + Mono) | latest |
| **Deploy** | Docker + Docker Compose | latest |

### Почему эти выборы

- **Chi v5** — чистый, stdlib-совместимый, composable middleware. Не привязывает к фреймворку.
- **GORM v2** — самый популярный Go ORM, быстрый старт, автомиграции. Для MVP идеален.
- **Vite 8 (Rolldown)** — Rust-based бандлер, 10-30x быстрее Vite 7, один движок для dev и build.
- **TanStack Query v5** — кеширование, фоновый рефетч, оптимистичные обновления, devtools. Стандарт индустрии.
- **Zustand v5** — 1.2KB, простой API, не нужен Provider. Для глобального клиентского состояния (auth, theme).

---

## Структура проекта

```
promptvault/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Точка входа, роутер, запуск
│   ├── internal/
│   │   ├── config/
│   │   │   ├── config.go            # Конфигурация из env
│   │   │   └── models.go            # Список AI-моделей
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── team.go
│   │   │   ├── collection.go
│   │   │   ├── prompt.go
│   │   │   ├── tag.go
│   │   │   └── version.go
│   │   ├── handlers/
│   │   │   ├── auth.go
│   │   │   ├── prompts.go
│   │   │   ├── collections.go
│   │   │   ├── teams.go
│   │   │   ├── tags.go
│   │   │   ├── ai.go                # SSE streaming endpoints
│   │   │   ├── settings.go
│   │   │   └── search.go
│   │   ├── middleware/
│   │   │   ├── auth.go              # JWT validation
│   │   │   ├── cors.go
│   │   │   └── ratelimit.go
│   │   ├── services/
│   │   │   ├── ai.go                # OpenRouter HTTP client + SSE
│   │   │   └── auth.go              # JWT gen/validate, OAuth flows
│   │   └── database/
│   │       ├── database.go          # GORM connection
│   │       └── migrate.go           # AutoMigrate
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   │   └── client.ts            # Fetch wrapper с JWT + auto-refresh
│   │   ├── components/
│   │   │   ├── ui/                   # shadcn/ui
│   │   │   ├── layout/              # sidebar, header, app-layout
│   │   │   ├── prompts/             # card, editor, form, filters
│   │   │   ├── ai/                  # panel, model-selector, result
│   │   │   ├── teams/
│   │   │   ├── collections/
│   │   │   ├── tags/
│   │   │   └── auth/
│   │   ├── pages/                    # route components
│   │   ├── hooks/
│   │   │   ├── use-auth.ts
│   │   │   └── use-sse.ts           # SSE streaming hook
│   │   ├── stores/
│   │   │   └── auth-store.ts        # Zustand store
│   │   ├── lib/
│   │   │   ├── utils.ts             # cn() и прочее
│   │   │   └── queries.ts           # TanStack Query keys + queryFn
│   │   ├── App.tsx                   # React Router layout
│   │   └── main.tsx
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
├── nginx/
│   └── nginx.conf                    # SPA fallback + proxy /api → backend
├── docker-compose.yml
├── .env.example
└── docs/
    └── PLAN.md
```

---

## API Endpoints

### Auth
| Метод | Путь | Описание |
|-------|------|---------|
| POST | `/api/auth/register` | Регистрация (email + password) |
| POST | `/api/auth/login` | Вход → JWT access + refresh |
| POST | `/api/auth/refresh` | Обновить access token |
| GET | `/api/auth/me` | Текущий пользователь |
| GET | `/api/auth/github` | OAuth GitHub redirect |
| GET | `/api/auth/github/callback` | OAuth GitHub callback |
| GET | `/api/auth/google` | OAuth Google redirect |
| GET | `/api/auth/google/callback` | OAuth Google callback |

### Prompts
| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/prompts` | Список (query: collection, tags, search, favorite, page) |
| POST | `/api/prompts` | Создать |
| GET | `/api/prompts/:id` | Получить (с тегами) |
| PUT | `/api/prompts/:id` | Обновить (автоверсия) |
| DELETE | `/api/prompts/:id` | Удалить |
| POST | `/api/prompts/:id/favorite` | Toggle избранное |
| POST | `/api/prompts/:id/use` | +1 usageCount |
| GET | `/api/prompts/:id/versions` | История версий |
| POST | `/api/prompts/:id/revert/:version` | Откат к версии |

### Collections
| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/collections` | Список (личные + командные) |
| POST | `/api/collections` | Создать |
| PUT | `/api/collections/:id` | Обновить |
| DELETE | `/api/collections/:id` | Удалить |

### Teams
| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/teams` | Мои команды |
| POST | `/api/teams` | Создать (автор = owner) |
| GET | `/api/teams/:slug` | Детали + участники |
| PUT | `/api/teams/:slug` | Обновить |
| DELETE | `/api/teams/:slug` | Удалить |
| POST | `/api/teams/:slug/members` | Пригласить |
| PUT | `/api/teams/:slug/members/:userId` | Изменить роль |
| DELETE | `/api/teams/:slug/members/:userId` | Удалить участника |

### Tags
| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/tags` | Все теги |
| POST | `/api/tags` | Создать |

### AI (все SSE streaming)
| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/ai/models` | Доступные модели |
| POST | `/api/ai/enhance` | Улучшить промпт |
| POST | `/api/ai/rewrite` | Переписать (стиль в body) |
| POST | `/api/ai/analyze` | Анализ качества |
| POST | `/api/ai/variations` | 3 варианта |

### Settings
| Метод | Путь | Описание |
|-------|------|---------|
| PUT | `/api/settings/profile` | Обновить имя/аватар |
| PUT | `/api/settings/model` | Выбрать модель |

### Search
| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/search?q=...` | ILIKE по промптам, коллекциям, командам |

---

## Модель данных

*(Без изменений — та же схема что в docs/PLAN.md: User, Team, TeamMember, Collection, Prompt, PromptVersion, Tag, PromptTag)*

---

## Фазы разработки

### Фаза 1: Фундамент

**1.1 — Go backend init**
- `go mod init promptvault`, установить chi v5, gorm, jwt v5, bcrypt, godotenv, cors
- `cmd/server/main.go` — Chi router + graceful shutdown
- `internal/config/config.go` — чтение env
- `internal/database/database.go` — GORM + PostgreSQL
- `internal/middleware/cors.go` — CORS для dev

**1.2 — GORM модели + автомиграция**
- Все модели в `internal/models/`
- `internal/database/migrate.go` — AutoMigrate при старте

**1.3 — React frontend init**
- `npm create vite@latest frontend -- --template react-ts` (Vite 8)
- `npx shadcn@latest init -d --base radix`
- Установить: react-router-dom, @tanstack/react-query, zustand, react-hook-form, @hookform/resolvers, zod, lucide-react
- Настроить Geist font (literal names), dark mode, shadcn компоненты
- Настроить Vite proxy (`/api` → `localhost:8080`)

**1.4 — Docker Compose**
- `backend/Dockerfile` — Go multi-stage (build → scratch/alpine)
- `frontend/Dockerfile` — Vite build → nginx
- `nginx/nginx.conf` — SPA fallback + `/api` proxy
- `docker-compose.yml` — api + frontend + postgres
- `.env.example`

**1.5 — Git init**

### Фаза 2: Авторизация

**2.1 — Go: JWT + Register/Login**
- `internal/services/auth.go` — JWT access (15m) + refresh (7d), bcrypt
- `internal/middleware/auth.go` — Chi middleware: validate JWT, inject userID в context
- `internal/handlers/auth.go` — POST register, POST login, POST refresh, GET me

**2.2 — Go: OAuth (GitHub + Google)**
- `golang.org/x/oauth2` — Authorization Code flow
- State parameter + PKCE (OAuth 2.1)
- Callback: upsert User → выдать JWT

**2.3 — React: Auth**
- `stores/auth-store.ts` — Zustand: tokens, user, login/logout/refresh
- `api/client.ts` — fetch wrapper: Authorization header, auto-refresh на 401
- `pages/sign-in.tsx`, `pages/sign-up.tsx`
- React Router: ProtectedRoute компонент

### Фаза 3: Лейаут

**3.1 — App layout** — sidebar + header + content (React Router Outlet)
**3.2 — Sidebar** — навигация, коллекции, команды, responsive Sheet
**3.3 — Header** — поиск Cmd+K, "+ Новый", user menu

### Фаза 4: CRUD промптов

**4.1 — Go: Prompts handlers** — CRUD, фильтры, пагинация, проверка доступа
**4.2 — React: Dashboard** — TanStack Query для списка, карточки, фильтры
**4.3 — React: Editor** — форма с React Hook Form + Zod, двухколоночный лейаут
**4.4 — Удаление + Избранное + Использование**

### Фаза 5: Коллекции

**5.1 — Go: Collections handlers**
**5.2 — React: Collections UI** — список, создание в Dialog, фильтрация

### Фаза 6: Теги

**6.1 — Go: Tags handlers**
**6.2 — React: Tags UI** — autocomplete, цвета, фильтрация

### Фаза 7: История версий

**7.1 — Go: Auto-versioning** — при PUT prompt → новая PromptVersion
**7.2 — React: Version history** — список, diff, откат

### Фаза 8: AI-функции

**8.1 — Go: OpenRouter client**
- `internal/services/ai.go` — HTTP POST к OpenRouter, парсинг SSE-чанков, проксирование клиенту
- `internal/config/models.go` — список моделей
- `internal/middleware/ratelimit.go` — token bucket по userID

**8.2 — Go: AI SSE endpoints**
- `internal/handlers/ai.go` — enhance, rewrite, analyze, variations
- Content-Type: text/event-stream, flush каждый чанк

**8.3 — React: AI-панель**
- `hooks/use-sse.ts` — EventSource/fetch с ReadableStream
- `components/ai/ai-panel.tsx` — модель-селектор, 4 действия, результат
- "Применить" → PUT prompt с changeNote

### Фаза 9: Команды

**9.1 — Go: Teams handlers** — CRUD, участники, роли, проверка доступа
**9.2 — React: Teams UI** — страница команды, управление участниками

### Фаза 10: Настройки ✅

**10.1 — Settings API** — профиль, пароль (двухшаговый через email), linked accounts, тема
**10.2 — Settings UI** — секции профиля/безопасности/аккаунтов/темы, CSS variables для light/dark
**10.3 — OAuth Account Linking** — PKCE + HMAC cookie, привязка/отвязка GitHub/Google/Яндекс
**10.4 — Забыли пароль** — публичный flow email→код→новый пароль
**10.5 — Безопасность** — rate limiting auth, PKCE, email-уведомления, защита от пустого email

### Фаза 11: Поиск и UX

**11.1 — Go: Search** — ILIKE по title + content
**11.2 — React: UX** — Command palette, скелетоны, пустые состояния, responsive

### Фаза 12: Production

**12.1 — Безопасность** — rate limiting, CORS production, secure cookies, input sanitization
**12.2 — Docker optimize** — multi-stage, healthcheck, auto-migrate
**12.3 — Landing page**

---

## Порядок реализации

| # | Фаза | Зависит от | Критичность | Статус |
|---|------|-----------|-------------|--------|
| 1 | Фундамент | — | Блокирует всё | ✅ |
| 2 | Авторизация | 1 | Блокирует данные | ✅ |
| 3 | Лейаут | 2 | Блокирует UI | ✅ |
| 4 | CRUD промптов | 3 | Ядро | ✅ |
| 5 | Коллекции | 4 | Организация | ✅ |
| 6 | Теги | 4 | Организация | ✅ |
| 7 | Версии | 4 | Нужно для AI | ✅ |
| 8 | AI-функции | 4, 7 | Ключевая фича | ✅ |
| 9 | Команды | 4, 5 | Командная работа | ❌ |
| 10 | Настройки | 2, 8 | UX | ✅ |
| 11 | Поиск и UX | 4 | Полировка | ✅ |
| 12 | Production | Всё | Готовность | ✅ |
| 13 | Подписки | 12 | Монетизация | ❌ |

---

## Верификация

После каждой фазы:
1. `docker compose up --build` — всё стартует
2. `curl` к API — ответы корректны
3. Проверить UI в браузере

Финальная:
1. Регистрация → вход → коллекция → промпт → AI улучшение → применить → версия
2. Команда → участник → editor → доступ
3. Поиск → фильтры → теги → избранное
4. Чистая машина: `docker compose up --build` — всё работает
