# ПромтЛаб — TODO Tracker

---

## Дополнения после аудита зависимостей

### Backend — добавить
| Пакет | Зачем | Фаза |
|-------|-------|------|
| `koanf/v2` + `koanf/providers/env` + `koanf/providers/dotenv` | Типизированная конфигурация (замена godotenv + getEnv) | 1.1 |
| `slog` (stdlib) | Структурированные логи (замена log.Printf) | 1.1 |
| `net/http/pprof` (stdlib) | CPU/memory/goroutine profiling, только в dev | 1.1 |
| `go-playground/validator/v10` | Валидация request body (email, required, min/max) | 2.1 |
| `golang.org/x/oauth2` | OAuth2 flow (GitHub, Google) | 2.2 |
| rate limiter (`sethvargo/go-limiter`) | Token bucket для AI эндпоинтов | 8.1 |

### Frontend — добавить
| Пакет | Зачем | Фаза |
|-------|-------|------|
| `@tanstack/react-query-devtools` (dev) | Инспекция кеша/запросов в dev | 1.3 |
| `zustand/middleware` devtools | Zustand → Redux DevTools в браузере | 2.3 |
| `sonner` | Toast-уведомления (shadcn рекомендует) | 3.1 |
| `date-fns` | Форматирование дат ("2 часа назад") | 4.2 |
| `cmdk` | Command palette (Cmd+K) | 11.2 |
| `react-diff-viewer-continued` | Diff версий промптов | 7.2 |

---

## Фаза 1: Фундамент

### 1.1 — Go backend init
- [x] `go mod init promptvault`, установить зависимости (chi v5, gorm, jwt v5, bcrypt, cors)
- [x] Заменить `godotenv` + `getEnv` на `koanf/v2` — типизированный конфиг с unmarshal в struct
- [x] `cmd/server/main.go` — Chi router + graceful shutdown
- [x] `internal/config/config.go` — переписать на koanf (env + .env файл + struct tags)
- [x] `internal/database/database.go` — GORM подключение к PostgreSQL
- [x] `internal/middleware/cors.go` — CORS для dev
- [x] Заменить `log` на `slog` — структурированные логи с уровнями (info/warn/error)
- [x] Подключить `net/http/pprof` на `/debug/pprof` (только при ENVIRONMENT=development)

### 1.2 — GORM модели + автомиграция
- [x] `internal/models/user.go` — User (id, email, password_hash, name, avatar_url, created_at)
- [x] `internal/models/team.go` — Team + TeamMember (роли: owner/editor/viewer)
- [x] `internal/models/collection.go` — Collection (user_id, team_id, name, description)
- [x] `internal/models/prompt.go` — Prompt (collection_id, title, content, model, favorite, usage_count, soft delete)
- [x] `internal/models/tag.go` — Tag + PromptTag (many-to-many)
- [x] `internal/models/version.go` — PromptVersion (prompt_id, content, change_note)
- [x] `internal/database/migrate.go` — AutoMigrate всех моделей при старте

### 1.3 — React frontend init
- [x] `npm create vite@latest frontend -- --template react-ts` (Vite 8)
- [x] `npx shadcn@latest init -d` + настройка dark mode, Geist font
- [x] Установить зависимости: react-router-dom, @tanstack/react-query, zustand, react-hook-form, @hookform/resolvers, zod, lucide-react
- [x] Установить dev: `@tanstack/react-query-devtools`
- [x] `vite.config.ts` — proxy `/api` → `localhost:8080`
- [x] Базовая структура папок: api/, components/, pages/, hooks/, stores/, lib/

### 1.4 — Docker Compose
- [x] `backend/Dockerfile` — Go multi-stage build (builder → alpine)
- [x] `frontend/Dockerfile` — Vite build → nginx (root context для доступа к nginx/)
- [x] `nginx/nginx.conf` — SPA fallback + `/api` proxy + SSE support
- [x] `docker-compose.yml` — сервисы: api, frontend, postgres + healthcheck
- [x] `docker-compose.dev.yml` — dev override: volume mounts, hot-reload, ENVIRONMENT=development
- [x] `docker-compose.prod.yml` — prod override: restart always, no exposed DB port
- [x] `.env.example` — шаблон переменных окружения

### 1.5 — Git init
- [x] `.gitignore` (Go + Node + .env* + IDE + OS)
- [ ] Первый коммит

---

## Фаза 2: Авторизация

### 2.1 — Go: JWT + Register/Login (Clean Architecture)
- [x] Установить `go-playground/validator/v10` — валидация request body
- [x] `internal/interface/repository/` — 6 интерфейсов (user, prompt, collection, tag, team, version)
- [x] `internal/infrastructure/postgres/repository/user_repo.go` — GORM реализация UserRepository
- [x] `internal/usecases/auth/` — auth.go (полные flows: Register, Login, Refresh, Me), types, errors, constants
- [x] `internal/delivery/http/auth/` — handler (без DB!), request.go, response.go
- [x] `internal/delivery/http/httperr/` — AppError, единый Respond()
- [x] `internal/delivery/http/httputil/` — WriteJSON, DecodeJSON, Pagination
- [x] `internal/middleware/auth.go` — TokenValidator interface
- [x] `internal/app/app.go` — единая сборка: repos → usecases → handlers → MountRoutes
- [x] Роуты подключены в main.go

### 2.2 — Go: OAuth (GitHub + Google + Yandex)
- [x] `usecases/auth/oauth.go` — OAuthService: GitHub, Google, Yandex (upsert User → JWT)
- [x] `delivery/http/auth/oauth_handler.go` — redirect + callback для каждого провайдера
- [x] State validation через cookie
- [x] Config: `oauth.go` + Yandex endpoint
- [x] Ошибки: ErrOAuthNotConfigured, ErrOAuthExchangeFailed, ErrOAuthStateMismatch
- [x] Роуты: /api/auth/oauth/{github,google,yandex} + /callback

### 2.3 — React: Auth UI
- [x] `api/types.ts` — User, TokenPair, AuthResponse, ApiError
- [x] `api/client.ts` — fetch wrapper: Authorization header, auto-refresh на 401, 204 handling
- [x] `stores/auth-store.ts` — Zustand с `devtools` middleware: tokens, user, login/logout/refresh/restoreSession
- [x] `pages/sign-in.tsx` — centered-card дизайн, OAuth с SVG иконками, toggle пароля, noValidate + Zod
- [x] `pages/sign-up.tsx` — paste блокировка на подтверждении пароля, очистка ошибок при onChange
- [x] `components/auth/auth-layout.tsx` — общий layout с лого и dot-pattern фоном
- [x] `components/auth/protected-route.tsx` — redirect на /sign-in если нет токена
- [x] `App.tsx` — React Router + QueryClientProvider + ReactQueryDevtools + restoreSession

### 2.4 — Email верификация
- [x] SMTP-сервис (`infrastructure/email/email.go`) — поддержка портов 465 (SMTPS) и 587 (STARTTLS)
- [x] Модель EmailVerification, 6-значный код, 15 минут
- [x] `pages/verify-email.tsx` — 6 полей для кода, countdown таймер 60с, "Отправить заново"
- [x] Register → отправка кода async → redirect на verify-email
- [x] Повторная регистрация с неподтверждённым email → обновление пароля + переотправка кода

### 2.5 — Account Linking (привязка аккаунтов)
- [x] Модель `LinkedAccount` — many providers per user
- [x] Миграция Provider/ProviderID из User в LinkedAccount
- [x] OAuth: автоматическая привязка по email к существующему аккаунту
- [x] Register: ошибка "Войдите через {provider}" если email занят OAuth
- [x] API: POST /api/auth/set-password, GET /api/auth/linked-accounts, DELETE /api/auth/unlink/{provider}
- [x] Все ошибки на русском языке

---

## Фаза 3: Лейаут

### 3.1 — App layout
- [x] Установить `sonner` — toast-уведомления
- [x] `components/layout/app-layout.tsx` — SidebarProvider + AppSidebar + AppHeader + Outlet + Toaster

### 3.2 — Sidebar
- [x] `components/layout/app-sidebar.tsx` — навигация (промпты, коллекции, команды, AI, настройки)
- [x] shadcn Sidebar компонент — responsive (Sheet на мобильных из коробки)
- [x] Список коллекций в sidebar (сворачиваемый, с поиском, max-height + scroll)

### 3.3 — Header
- [x] `components/layout/app-header.tsx` — SidebarTrigger + поиск (Cmd+K placeholder) + кнопка "+ Новый промпт"
- [x] `components/layout/user-menu.tsx` — аватар, имя, профиль, тема, выход
- [x] `stores/theme-store.ts` — Zustand + persist + devtools, dark/light с применением класса на html

---

## Фаза 4: CRUD промптов

### 4.1 — Go: Prompts handlers
- [x] `delivery/http/prompt/handler.go` — GET /api/prompts (фильтры: collection, tags, search, favorite, page)
- [x] POST /api/prompts — создание
- [x] GET /api/prompts/:id — получение с тегами и коллекциями
- [x] PUT /api/prompts/:id — обновление
- [x] DELETE /api/prompts/:id — soft delete
- [x] POST /api/prompts/:id/favorite — toggle избранное
- [x] POST /api/prompts/:id/use — +1 usageCount

### 4.2 — React: Dashboard
- [x] `pages/dashboard.tsx` — список промптов с TanStack Query, stats, skeleton
- [x] `components/prompts/prompt-card.tsx` — карточка с lift hover, glow, цветные точки моделей
- [x] Фильтры: по коллекции, избранному, поиск с debounce + ⌘K
- [x] Пагинация

### 4.3 — React: Editor
- [x] `pages/prompt-editor.tsx` — форма в карточке, счётчик символов, multi-select коллекций
- [x] React Hook Form + Zod валидация в prompt-editor.tsx

### 4.4 — Удаление + Избранное + Использование
- [x] UI для toggle избранного на карточке (звёздочка с hover)
- [x] Soft delete через API
- [x] POST /api/prompts/:id/use — +1 счётчик

---

## Фаза 5: Коллекции

### 5.1 — Go: Collections handlers
- [x] `delivery/http/collection/handler.go` — CRUD (GET list, GET by id, POST create, PUT update, DELETE)
- [x] Цвет и иконка для каждой коллекции (color, icon)
- [x] CountPrompts через many-to-many

### 5.2 — React: Collections UI
- [x] `pages/collections.tsx` — список коллекций с цветными карточками, Lucide иконки, skeleton
- [x] Диалог создания/редактирования с палитрой цветов (8) и иконок (15) с тултипами
- [x] Кастомный диалог удаления (не browser confirm)
- [x] `pages/collection-view.tsx` — страница коллекции с хлебными крошками
- [x] Кнопка "Новый промпт" → создание с автопривязкой к коллекции
- [x] Кнопка "Из списка" → модалка с чекбоксами для добавления существующих промптов

### 5.3 — Many-to-many коллекции
- [x] Таблица `prompt_collections` (many-to-many вместо collection_id)
- [x] Миграция данных из старой колонки
- [x] Multi-select коллекций в редакторе промпта (чипы)
- [x] Промпт может быть в нескольких коллекциях одновременно
- [x] "Из списка" добавляет к существующим коллекциям (не заменяет)

---

## Фаза 6: Теги

### 6.1 — Go: Tags handlers
- [x] `delivery/http/tag/handler.go` — GET /api/tags, POST /api/tags, DELETE /api/tags/{id}
- [x] `usecases/tag/` — tag service + errors
- [x] Привязка тегов к промптам (many-to-many) — уже работало, добавлен HTTP слой
- [x] Парсинг `tag_ids` в фильтре промптов (GET /api/prompts?tag_ids=1,2,3)

### 6.2 — React: Tags UI
- [x] `components/tags/tag-input.tsx` — combobox с autocomplete, создание на лету, цветные чипы
- [x] Фильтрация промптов по тегам на дашборде (сворачиваемые чипы, "Ещё N+")
- [x] Отображение тегов на карточках промптов (уже было)
- [x] Интеграция тегов в редактор промпта

---

## Дополнительные улучшения (вне фаз)

### UX
- [x] Бесконечный скролл на дашборде (useInfiniteQuery + IntersectionObserver) — заменил пагинацию
- [x] Поиск коллекций в sidebar (клиентская фильтрация, показывается при >5)
- [x] Поиск коллекций в редакторе промпта (сворачиваемый блок, "Ещё N+")
- [x] Поиск промптов в диалоге "Добавить в коллекцию"
- [x] Скроллируемый dropdown тегов (max-height: 240px)
- [x] Теги ярче (убран opacity-60, поднята яркость цвета)
- [x] Адаптивный grid статистики (grid-cols-2 sm:grid-cols-4)

### Безопасность и надёжность (аудит)
- [x] ErrorBoundary в App.tsx — fallback UI при ошибке рендера
- [x] Отдельная переменная `SERVER_FRONTEND_URL` для OAuth redirect
- [x] AutoMigrate только в dev или `FORCE_MIGRATE=true`
- [x] pprof закомментирован в nginx.conf
- [x] N+1 запрос коллекций → один SQL с JOIN + GROUP BY
- [x] Транзакции в prompt Update (Tags + Collections + Save)
- [x] Исправлены unchecked errors (collection_repo, auth, oauth)
- [x] Auth middleware — JSON ответы с Content-Type
- [x] Каскадное удаление — tag Delete чистит prompt_tags
- [x] Query invalidation — delete/create prompt инвалидирует collections и tags
- [x] SMTP retry — 3 попытки с exponential backoff
- [x] Nginx security headers (X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Referrer-Policy)
- [x] Nginx gzip compression
- [x] JWT secret validation в production
- [x] ConnMaxLifetime для DB pool (1 час)
- [x] DB индексы на prompts.user_id и collections.user_id
- [x] Двойной 401 исправлен — retry только после успешного refresh
- [x] collection-view.tsx адаптирован под useInfiniteQuery
- [x] SMTP переменные добавлены в .env.example
- [x] VerificationCodeExpiry вынесен в константу

---

## Фаза 7: История версий

### 7.1 — Go: Auto-versioning
- [x] Расширена модель PromptVersion: Title, Model, VersionNumber + uniqueIndex(prompt_id, version_number)
- [x] VersionRepository: `CreateWithNextVersion` (атомарный SELECT MAX FOR UPDATE + INSERT в транзакции)
- [x] `GetByIDForPrompt(versionID, promptID)` — безопасное получение с проверкой принадлежности
- [x] При PUT /api/prompts/:id → снимок старого состояния перед мутацией
- [x] GET /api/prompts/:id/versions — список версий с пагинацией (page, page_size, total, has_more)
- [x] POST /api/prompts/:id/revert/:versionId — откат (создаёт снимок + change_note "Откат к версии N")
- [x] change_note в UpdatePromptRequest
- [x] Cascade delete: `Versions []PromptVersion` с `OnDelete:CASCADE` в модели Prompt

### 7.2 — React: Version history
- [x] `react-diff-viewer-continued` — split/unified с word wrap (как GitHub/GitLab, `minWidth: unset`)
- [x] `hooks/use-versions.ts` — `useInfiniteQuery` + IntersectionObserver (бесконечный скролл)
- [x] `pages/versions.tsx` — responsive (вертикальный стек на мобильных, grid на десктопе)
- [x] `components/prompts/version-diff.tsx` — unified на мобильных, split на десктопе
- [x] Кнопка "Откатить" с double-click guard (`isPending`)
- [x] Кнопка "История версий" + поле "Заметка к изменению" в редакторе
- [x] Роут `/prompts/:id/versions`, навигация через `navigate(-1)`
- [x] Query invalidation с `exact: false` для промптов

### 7.3 — Тесты (24 теста, 100% PASS)
- [x] `usecases/prompt/prompt_test.go` — 14 unit тестов (Update, ListVersions, RevertToVersion, GetByID)
- [x] `delivery/http/prompt/handler_test.go` — 10 handler тестов (пагинация, 400/403/404, revert)
- [x] Mock репозитории с testify для PromptRepo, VersionRepo, TagRepo, CollectionRepo

---

## Фаза 8: AI-функции

### 8.1 — Go: OpenRouter client
- [x] `infrastructure/openrouter/client.go` — HTTP POST к OpenRouter API, парсинг SSE-чанков, проксирование клиенту
- [x] `infrastructure/config/ai.go` — список доступных моделей (5 топовых)
- [x] `usecases/ai/ratelimit.go` — in-memory rate limiter по userID (sliding window)
- [x] Rate limit проверяется до отправки SSE headers (двухфазные ошибки)

### 8.2 — Go: AI SSE endpoints
- [x] `delivery/http/ai/handler.go` — GET /api/ai/models (список моделей)
- [x] POST /api/ai/enhance — улучшить промпт (SSE stream)
- [x] POST /api/ai/rewrite — переписать в стиле (SSE stream)
- [x] POST /api/ai/analyze — анализ качества (SSE stream)
- [x] POST /api/ai/variations — 3 варианта (SSE stream)

### 8.3 — React: AI-панель
- [x] `hooks/use-sse.ts` — hook для SSE streaming (fetch + ReadableStream)
- [x] `components/ai/ai-panel.tsx` — выбор модели, 4 кнопки действий, область результата
- [x] `components/ai/model-selector.tsx` — селектор модели
- [x] Кнопка "Применить" → PUT prompt с changeNote от AI

### 8.4 — Код-ревью исправления (30 тестов)
- [x] SSE newline splitting — multiline chunks не ломают SSE framing
- [x] SSE event: error парсинг на фронте — ошибки показываются как toast, не как текст
- [x] Flusher comma-ok check — нет паники при обёрнутом ResponseWriter
- [x] finishSSE — user-friendly ошибки + slog.Error логирование
- [x] Rate limiter — memory cleanup (delete empty entries) + limit<=0 = unlimited
- [x] OpenRouter 404 → ErrModelNotFound (отдельный case)
- [x] Malformed SSE chunks логируются (slog.Warn)
- [x] Stream убран из ChatRequest (wire struct внутри клиента)
- [x] Models() возвращает копию слайса
- [x] disabled guard при пустой модели в AI-панели
- [x] clipboard.writeText awaited
- [x] ModelSelector показывает ошибку загрузки
- [x] streamEndpoint generic helper (handler.go ~190 → ~130 строк)
- [x] 30 новых тестов: ratelimit (5), client (10), service (9), handler (6)

### 8.5 — AI Research & Оптимизация (из AI-RESEARCH.md)
- [x] Исправлен model ID Claude: `anthropic/claude-sonnet-4-20250514` → `anthropic/claude-sonnet-4`
- [x] Per-model temperature: Claude 0.4 (confirmed Anthropic docs)
- [x] System prompts переведены на английский + anti-preamble (без "Sure, here's...")
- [x] Usage logging из финального SSE-чанка: prompt_tokens, completion_tokens, cost_usd, cached_tokens, duration_ms
- [x] Cache_control для Anthropic: top-level `{"cache_control": {"type": "ephemeral"}}` (работает при >1024 токенов)
- [x] Reasoning effort для GPT-5: `{"reasoning": {"effort": "low"}}` (формат OpenRouter, confirmed)
- [x] Оптимизация system prompts до качества 9.6/10 (8 режимов протестированы)
- [x] Лимит промпта снижен до 10,000 символов (было 50,000) на всех уровнях (backend + frontend + Zod)
- [x] Счётчик символов: `{len}/10 000` с цветовой индикацией (зелёный → жёлтый >7500 → красный >9000)

### 8.6 — Единственная модель: Claude Sonnet 4
- [x] Убраны 4 модели (GPT-5, Gemini, DeepSeek, GPT-4o Mini), оставлена только Claude Sonnet 4
- [x] Убран UI выбора модели (ModelSelector скрыт, auto-select)
- [x] ReasoningEffort и Temperature настроены per-model в конфиге

### 8.7 — UI/UX улучшения AI-панели
- [x] Dropdown модели/стилей: shadcn Select с `modal={false}` (тёмный, без блокировки скролла)
- [x] Loading animation: skeleton-полоски + bouncing dots до первого токена
- [x] Таймер генерации: "{N} сек" рядом со спиннером
- [x] Пульсирующий курсор при streaming
- [x] JWT auto-refresh в SSE hook (`use-sse.ts`) — retry при 401 с обновлением токена

---

## Фаза 9: Команды

### 9.1 — Go: Teams handlers
- [ ] `internal/handlers/teams.go` — CRUD команд (GET list, POST create, GET :slug, PUT update, DELETE)
- [ ] POST /api/teams/:slug/members — приглашение по email
- [ ] PUT /api/teams/:slug/members/:userId — смена роли
- [ ] DELETE /api/teams/:slug/members/:userId — удаление участника
- [ ] Проверка доступа: owner/editor/viewer

### 9.2 — React: Teams UI
- [ ] `pages/team.tsx` — страница команды с участниками
- [ ] `components/teams/member-list.tsx` — список участников с ролями
- [ ] `components/teams/invite-dialog.tsx` — приглашение нового участника
- [ ] Переключение контекста: личный / команда

---

## Фаза 10: Настройки

### 10.1 — Go: Settings API
- [x] `POST /api/auth/set-password/initiate` — отправить код на email (двухшаговая установка)
- [x] `POST /api/auth/set-password/confirm` — проверить код + установить пароль
- [x] `GET /api/auth/linked-accounts` — список привязанных провайдеров
- [x] `DELETE /api/auth/unlink/{provider}` — отвязать провайдер (с защитой от отвязки последнего)
- [x] `PUT /api/auth/profile` — обновить имя, аватар
- [x] `PUT /api/auth/password` — смена пароля + email-уведомление
- [x] ~~`PUT /api/settings/model`~~ — не нужно (единственная модель Claude Sonnet 4)

### 10.2 — React: Settings UI
- [x] `pages/settings.tsx` — страница настроек:
  - [x] Секция "Привязанные аккаунты" — GitHub/Google/Яндекс с кнопками привязать/отвязать
  - [x] Секция "Пароль" — установить через email-код / изменить через старый пароль
  - [x] Секция "Профиль" — имя, OAuth-аватар, email (read-only)
  - [x] ~~Выбор AI-модели~~ — не нужно (единственная модель)
  - [x] Переключение темы (light/dark) — CSS variables, работает в обеих темах
  - [x] Confirm dialog перед отвязкой провайдера

### 10.3 — OAuth Account Linking
- [x] `POST /api/auth/link/{provider}` — инициация привязки (protected, HMAC-подписанная cookie)
- [x] Link flow в callbacks — oauth_link cookie → linkProvider вместо upsert
- [x] `LinkGitHub/LinkGoogle/LinkYandex` — exchange + привязка к текущему юзеру
- [x] Проверка конфликтов: ErrProviderLinkedToOther, ErrProviderAlreadyLinked
- [x] Frontend: POST → получение redirect_url → OAuth → /settings?linked=github

### 10.4 — Забыли пароль
- [x] `POST /api/auth/forgot-password` — отправка кода сброса (публичный, не раскрывает аккаунт)
- [x] `POST /api/auth/reset-password` — проверка кода + новый пароль
- [x] `pages/forgot-password.tsx` — двухшаговая страница (email → код + пароль)
- [x] Ссылка "Забыли?" на sign-in → /forgot-password

### 10.5 — Безопасность настроек (аудит)
- [x] PKCE (S256) для всех OAuth провайдеров (verifier в cookie)
- [x] Rate limiting 20 req/min по IP на auth endpoints
- [x] Защита от пустого email в OAuth (reject вместо crash)
- [x] Email-уведомление при смене пароля
- [x] `default_model` обновлён на `anthropic/claude-sonnet-4`
- [x] Email service рефакторинг: `buildMessage()`, Base64 UTF-8, 4 шаблона
- [x] Koanf defaults fix: вложенная map (совместимость с env provider Unflatten)

---

## Фаза 11: Поиск и UX

### 11.1 — Go: Search
- [x] `delivery/http/search/handler.go` — GET /api/search?q= (ILIKE по промптам, коллекциям, тегам)
- [x] `usecases/search/` — Service с группированным ответом (prompts/collections/tags, лимиты 5/3/3)
- [x] `SearchByQuery` в 3 интерфейсах + 3 GORM-репозиториях
- [x] Wiring в app.go, 8 тестов (usecase + handler)

### 11.2 — React: UX polish
- [x] Установить `cmdk` — command palette (`npx shadcn add dialog command`)
- [x] Command palette (Cmd+K) — поиск + навигация (`components/command-palette.tsx`)
- [x] Кнопка "Поиск... ⌘K" в header (всегда видна)
- [x] Хук `useSearch()` — TanStack Query, debounce 300ms, enabled >= 2 символов
- [x] Группы результатов: Промпты, Коллекции, Теги + Навигация
- [x] Skeleton загрузки для списков
- [x] Пустые состояния (нет промптов, нет коллекций)
- [x] Responsive дизайн (мобильная адаптация) — частично (stats grid, sidebar)
- [x] Toast уведомления (sonner)

---

## Фаза 12: Production

### 12.1 — Безопасность
- [x] Rate limiting на auth эндпоинты (20 req/min по IP, middleware/ratelimit)
- [x] Rate limiting на protected эндпоинты (60 req/min по IP)
- [x] CORS — comma-separated origins, валидация HTTPS + запрет wildcard в production
- [x] Secure HttpOnly cookie для refresh token (SameSite=Strict, Path=/api/auth)
- [x] `POST /api/auth/logout` — очистка cookie
- [x] Frontend: `credentials: "include"`, убран localStorage для refresh token
- [x] Session restore через cookie (refresh → access token)
- [x] Input sanitization — `html.EscapeString` на всех user-input полях (prompt, collection, tag, profile)
- [x] Валидация avatar_url — только http/https (блок javascript:, data:)
- [x] `golangci-lint` v2 — `.golangci.yml` создан, линтер работает
- [x] Nginx: HSTS, Content-Security-Policy, Permissions-Policy

### 12.2 — Docker optimize
- [x] Multi-stage Dockerfile (минимальный образ)
- [x] Healthcheck для сервисов
- [x] Auto-migrate при старте контейнера (dev) / FORCE_MIGRATE (prod)
- [x] Docker Compose production profile

### 12.3 — Landing page
- [x] `pages/landing.tsx` — Hero, 4 фичи, кнопки "Войти"/"Регистрация", footer
- [x] `/` → лендинг (неавторизованные), `/dashboard` → дашборд (авторизованные)
- [x] Redirect авторизованных с `/` на `/dashboard`
- [x] Обновлены все навигационные ссылки (sidebar, command palette, sign-in, verify-email, oauth)

---

## Фаза 13: Подписки и оплата

> Обновлено 2026-04-14 — Фаза реализована (sandbox). Остаток: prod-терминал, 54-ФЗ, v2-фичи.

### 13.1 — Модель подписки ✅
- [x] Модель Subscription (миграции 000019-000023: plans, subscriptions, payments, daily_feature_usage, users.plan_id)
- [x] Тарифы: Free (5 запросов ВСЕГО) / Pro 599₽/мес (10/день) / Max 1299₽/мес (30/день)
- [ ] Годовые планы: Pro 4990₽/год / Max 10990₽/год (скидка 30%) — v2
- [x] Rate limiting по тарифу (usecases/quota с CheckAIQuota/CheckPromptQuota/...)

### 13.2 — Интеграция оплаты ✅
- [x] T-Bank Acquiring API V2 (провайдер выбран вместо ЮKassa)
- [x] Webhook для подтверждения оплаты (подпись SHA-256, idempotent, rate-limited)
- [ ] Автопродление подписки — v2 (SubStatusPastDue зарезервирован)
- [x] Страница "Тарифы" с карточками Free/Pro/Max

### 13.3 — UI подписок ✅
- [x] Страница `/pricing` — сравнение тарифов
- [x] `QuotaExceededDialog` на 402 (баннер «Осталось N запросов» заменён на dialog)
- [x] Upgrade flow: Free → Pro/Max через `useCheckout` → T-Bank redirect → polling 2 мин
- [x] Управление подпиской в `SubscriptionSection` (отмена/downgrade/usage-meters)

### 13.4 — Осталось для prod-релиза
- [ ] Production-терминал T-Bank (sandbox работает)
- [ ] IP allowlist T-Bank в middleware (ждём диапазоны)
- [ ] 54-ФЗ фискализация: Receipt в T-Bank Init (для физлиц РФ)
- [ ] Публичная оферта + условия возврата (юридический документ)
