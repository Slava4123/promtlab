# PromptVault Chrome Extension — Roadmap

> Дата актуализации: **2026-04-11 (v2)** — после завершения плана C (commercial-grade)
> Текущая версия extension: `0.1.0`
> Статус: **Готово к CWS submission** (ожидает операционных шагов — privacy, screenshots, $5 dev account)
> Репо: `promptvault/extension/` (WXT 0.20.20 + React 19 + Vite 8 + Tailwind v4 + shadcn/ui + Radix UI)

## TL;DR текущего состояния

**Код полностью готов к публикации** — собран, протестирован, CI pipeline добавлен. Backend и SPA доработаны, безопасно задеплоятся через обычный `git push main`. Остаются только **operational** задачи для первой публикации в Chrome Web Store: оплата dev-аккаунта, скриншоты, promo tile, upload zip и ручное заполнение CWS listing.

| Метрика | Значение |
|---|---|
| Extension bundle size (unpacked) | **448 kB** (с запасом под 5 MB) |
| Extension bundle size (CWS zip) | **~130 kB** (с огромным запасом под 15 MB CWS лимит) |
| Unit tests (vitest) | **22/22 passed** |
| Backend Go tests | **всё зелёное** |
| Frontend SPA tests | **72/72 passed** |
| TypeScript strict | **zero errors** |
| CI jobs (lint/test/build/deploy) | все добавлены в `.github/workflows/deploy.yml` |

---

## Текущее состояние

### ✅ Backend изменения (задеплоятся автоматически при merge)

1. `internal/middleware/auth/apikey.go` — вынесенный `APIKeyAuth` middleware с узким интерфейсом `APIKeyValidator`
2. `internal/middleware/auth/combined.go` — `CombinedAuth(jwt, apiKeys)` распознаёт тип токена по префиксу `pvlt_`
3. `internal/middleware/cors/cors.go` — `AllowOriginFunc` пропускает любой `chrome-extension://*` + static whitelist SPA origins
4. `internal/infrastructure/config/loader.go` — prod-валидатор CORS пропускает `chrome-extension://` минуя HTTPS check
5. `internal/app/app.go` — `apiKeyValidatorAdapter` + `CombinedAuth` применён к protected группе `/api/*`

**Backwards compatible**: existing JWT flow не сломан (проверено тестами + живым login в SPA).

### ✅ Frontend SPA изменения (задеплоятся автоматически при merge)

1. `frontend/src/components/settings/api-keys-section.tsx` — использует shadcn/ui `<Input>`, label сверху, короткий placeholder `Claude Code`, кнопки в отдельном flex-row `justify-end`, `flex-wrap` на метаданные ключей
2. `frontend/src/components/settings/extension-promo-section.tsx` — **новая секция** в Settings:
   - Smart browser detection (Chromium/Firefox/Safari)
   - Badge «Скоро» пока `CHROME_WEB_STORE_URL = ""`
   - Описание + список преимуществ
   - Кнопка `variant="outline" disabled` до публикации, переключается на `variant="default"` + `onClick=window.open()` когда URL задан
   - Ссылка на GitHub releases (fallback для early adopters)
   - Ссылка на Privacy Policy
3. `frontend/src/pages/legal/extension-privacy.tsx` — **новая public страница** `/legal/extension-privacy`:
   - 10 секций по CWS требованиям (tl;dr, data collected, recipients, target sites, permissions, storage, third parties, children, changes, contacts, open source)
   - Таблица permissions с обоснованием каждого (для CWS permission justifications)
   - 3 способа удалить данные
   - Контакт `privacy@promtlabs.ru` + GitHub issues
   - Ссылка на open source репозиторий
4. `frontend/src/pages/settings.tsx` — рендерит `<ExtensionPromoSection />` между Linked Accounts и API Keys sections
5. `frontend/src/App.tsx` — lazy-loaded роут `/legal/extension-privacy` (public, без `ProtectedRoute`, чтобы CWS reviewers открывали без логина)

### ✅ Extension v0.1.0 (готов к publishing)

**Core архитектура**:
- WXT workspace `promptvault/extension/` (Vite 8 + React 19 + TypeScript strict + Tailwind v4 + shadcn/ui + Radix UI)
- Side Panel UI: `ApiKeySetup` → `Home` → `VariableForm` → `SettingsView`
- Background service worker с brokering между Side Panel и content scripts
- 4 content scripts (chatgpt / claude / gemini / perplexity) с общим `insertPrompt` cascade
- Cascade insertion: `nativeSetter` (textarea) → `execCommand('insertText')` (contenteditable) → paste event (fallback)
- Template parser `{{переменные}}` с Unicode (поддержка кириллицы)

**UI / UX фичи**:
- **Workspace switcher** — dropdown в header, переключение Личное / Команда1 / Команда2, persist в `chrome.storage.local`
- **Collection filter** — под поиском, зависит от выбранного workspace
- **Tag filter** — подготовлен через `PromptFilter.tagIds` (UI selector в backlog)
- **Tabs навигация**: Все / Закреплённые / Недавние / ⭐ Избранное
- **Infinite scroll** для главного списка (`useInfiniteQuery`, page_size=50)
- **Skeleton loaders** вместо spinner
- **Cyrillic search** с debounce 300ms
- **Hover preview** — floating tooltip с полным content промпта (delay 400ms)
- **Active-tab badge** — зелёный «ChatGPT» на supported сайтах, жёлтый «Не поддерживается» иначе
- **Streak badge** — огненный 🔥 счётчик в header, подтягивается через `/api/streaks`
- **Active refresh** (⌘R) с smart retry (не retry 4xx, exponential backoff)
- **Sync с SPA** — `refetchOnWindowFocus: true` + кнопка Refresh в header
- **Auto-insert для промптов без переменных** — skip VariableForm, прямой insert
- **Cyrillic variables** в VariableForm, live preview с подсветкой заполненных/пустых
- **Char + token counter** в Preview
- **Saved variable values per-prompt** — второй запуск того же промпта с уже заполненными полями
- **Copy-to-clipboard** (⌘⇧C) — альтернатива Insert
- **Multi-target insert** (⚡) — во все открытые supported tabs сразу
- **Share prompt link** — создаёт public URL через `/api/prompts/:id/share` и кладёт в clipboard
- **Success flash + toast с Undo** (5 сек) — карточка мигает зелёным, в toast кнопка «Отменить»
- **Favorite/Pin** прямо из карточки на hover — mutations через `invalidateQueries`
- **Per-prompt Local recent** — backup 20 последних вставленных в `chrome.storage.local` (если backend недоступен)
- **Цветные теги** из backend (`t.color` → bg/border/text прозрачные оттенки)
- **Welcome empty state** с кнопкой «Открыть ПромтЛаб» (открывает apiBase в новой вкладке)

**Keyboard-first**:
- `Ctrl+Shift+K` — открыть Side Panel (system shortcut)
- `Ctrl+K` — фокус в поиск
- `↑↓` — навигация по карточкам
- `Enter` — открыть промпт
- `Esc` — очистить поиск / назад из VariableForm
- `Ctrl+Enter` — submit VariableForm
- `Ctrl+Shift+C` — copy to clipboard в VariableForm
- `Ctrl+R` — refresh всех queries

**Settings view** (клик на ⚙️ в header):
- API base URL editor (с валидацией http/https)
- API key replacement (с backend validation)
- Theme selector через **shadcn/ui Radix Select** (Системная / Светлая / Тёмная), иконки Monitor/Sun/Moon
- Health indicator (зелёная ✓ / красная ✗) с RefreshCw retry
- Logout
- Version info

**Qualify / hardening**:
- **Error Boundary** (class component) — ловит React errors, показывает friendly UI с кнопкой Reload
- **Onboarding overlay** — 4-step wizard при первом открытии (Welcome → Search → Keyboard → Favorites), progress bar, persist через `chrome.storage.local`
- **ApiKeySetup с friendly errors** — валидация `http(s)://` префикса base URL, detailed messages (unauthorized / network / unknown)
- **Light Sentry wrapper** — `lib/sentry.ts` с breadcrumbs + PII scrubbing (token/key/content/password → `[REDACTED]`), feature flag `VITE_SENTRY_ENABLED`, без внешних зависимостей (stub — реальная отправка в GlitchTip откладывается до первой prod-установки)
- **Offline cache foundation** — `getCachedPrompts()`/`setCachedPrompts()` в storage с TTL 5 мин (интеграция в api.ts — TODO)
- **Auto-retry** с exponential backoff (до 3s) и 4xx-bypass

**shadcn/ui компоненты**:
- `ui/select.tsx` — Radix UI Select с portal-рендерингом, keyboard nav, анимациями
- `ui/button.tsx`, `ui/input.tsx`, `ui/textarea.tsx`, `ui/label.tsx`, `ui/skeleton.tsx`, `ui/badge.tsx`
- `ui/toaster.tsx` — полностью custom toast без `sonner`, ~150 строк, 4 variants (success/error/info/default), action button (`Undo` с Undo2 icon)

**Иконки**:
- `public/icon/{16,32,48,128}.png` — сгенерированы через `scripts/gen-icons.mjs` (Node без внешних зависимостей, PNG через built-in `zlib.deflateSync`)
- Дизайн: фиолетовый round-rect + белая буква «П»

### ✅ Testing (Phase 6)

- `vitest.config.ts` + `tests/template.test.ts` + `tests/selectors.test.ts` — **22 unit теста** покрывающих template parser (extract/render, Unicode, dedup, order, regex metacharacters, spaces, digits) и selectors (coverage всех 4 сайтов, primary selectors confirmed)
- `package.json` scripts: `test`, `test:watch`, `compile`, `build`, `zip`

### ✅ CI/CD (Phase 6 + Phase 7)

`.github/workflows/deploy.yml` — новый job `test-extension`:
- `needs: lint`
- TypeScript strict check (`npx tsc --noEmit`)
- Unit tests (`npx vitest run`)
- Build (`npx wxt build`)
- Zip (`npx wxt zip`)
- Size check: < 5 MB unpacked, < 15 MB zipped (CWS limit)
- Upload artifact `promptvault-extension-chrome-<sha>` (30 дней retention)
- **Блокирует `build-push`** — если extension тесты падают, backend deploy не пойдёт
- Артефакт доступен через GitHub Actions UI → Actions → run → Artifacts

### 🟡 Что реализовано, но отложено (deferred до user request)

- **DnD reorder Pinned** — требует backend миграцию `prompt_pins.order` + dnd-kit ~80KB + 90-120 мин. Сейчас pinned sorted by `created_at`, что приемлемо для <100 pinned промптов.
- **Bulk actions** — selection mode + batch operations. Нужно только пользователям с 100+ промптами. Не в MVP.
- **Полноценный Sentry** — сейчас `lib/sentry.ts` — stub с структурированными логами. Доделаем когда будет real prod usage + GlitchTip project для extension.
- **Offline cache integration** — foundation в storage готов, интеграция в `api.ts` (когда fetch fail → fallback на cache) — TODO, ~30 мин.

---

## Changelog (2026-04-11)

### v0.1.0 → v0.1.0 (план C commercial-grade)

#### Added
- **Workspace switcher** (Phase 1)
- **Collection filter** (Phase 1)
- **Sync refresh button** + smart retry (Phase 2)
- **Copy-to-clipboard** в VariableForm (Phase 2)
- **Welcome state** с кнопкой «Открыть ПромтЛаб» (Phase 2)
- **Local recent history** backup (Phase 2)
- **Hover preview** на PromptCard (Phase 3)
- **Share prompt link** кнопка (Phase 3)
- **Multi-target insert** (⚡) (Phase 3)
- **Streak badge** в header (Phase 3)
- **Error boundary** class component (Phase 4)
- **Onboarding overlay** 4-step wizard (Phase 4)
- **Lightweight Sentry** wrapper (Phase 4)
- **Offline cache** foundation (Phase 4)
- **Request retry** с exponential backoff (Phase 4)
- **Vitest** + 22 unit тестов (Phase 6)
- **CI test-extension job** в `deploy.yml` (Phase 6)
- **shadcn/ui Radix Select** — WorkspaceSelector, CollectionSelector, Theme в Settings (Phase 7 polish)
- **Extension promo section** в SPA Settings
- **Privacy Policy** страница `/legal/extension-privacy`

#### Fixed
- Theme switcher не работал из-за `@media (prefers-color-scheme: dark)` вместо `.dark` class selector. Заменено на `.dark { --var: ... }` в `tailwind.css` — переключение работает мгновенно.
- TogglePin API response не полный `Prompt`, а `{pinned, team_wide}` — mutation не обновляла UI. Убрал optimistic update, использую `invalidateQueries` — надёжнее.
- `patchPromptInData` не обрабатывал `useInfiniteQuery` cache shape `{pages: [...]}` — favorite/pin не обновлялись в главном списке. Исправлено тем же `invalidateQueries` вместо custom patching.
- ApiKeySetup: placeholder был слишком длинным и обрезался в узких карточках. Перестроено на shadcn паттерн `<Label>` + короткий placeholder + кнопки вертикально на мобильном.
- `default_locale: "ru"` в manifest вызывал Chrome error «missing `_locales/ru/messages.json`». Убрано — все строки уже на русском хардкодом.
- `docker compose up -d` не пересобирал image с новым кодом — gotcha зафиксирован в документации. Сейчас CI всегда пересобирает при каждом build.

#### Deferred (не сделано осознанно)
- **DnD reorder Pinned** (Phase 5) — нужна backend миграция + dnd-kit. Польза для MVP низкая.
- **Bulk actions** (Phase 5) — нужен selection mode UI. Полезно только для 100+ промптов.

---

## Что нужно сделать дальше

Задачи разбиты на 2 категории: (A) **автодеплой** — просто `git push main`, всё едет через CI/CD; (B) **operational** — ручные шаги для публикации в CWS.

### A. Автодеплой (готово к `git push main`)

Эти задачи **не требуют ручной работы** — код готов, CI pipeline настроен. Просто push в main → через ~5-10 минут всё в production.

- [x] Backend `apikey.go`, `combined.go`, `cors.go`, `loader.go`, `app.go` — пересоберётся в `ghcr.io/slava4123/promtlab-api:latest`
- [x] Frontend SPA Settings с промо-кнопкой + Privacy Policy — пересоберётся в `ghcr.io/slava4123/promtlab-frontend:latest`
- [x] SSH deploy на VPS + `docker compose pull && up -d`
- [x] Health check через `/api/health`
- [x] Extension zip попадёт в artifacts (но **не публикуется автоматически** — это безопасно)

**Что проверить после первого deploy:**
1. SPA Settings страница → появился блок «Chrome-расширение» с badge «Скоро»
2. `https://promtlabs.ru/legal/extension-privacy` открывается без логина
3. Existing user flow (login → dashboard → create prompt) работает как раньше — CombinedAuth не должен ничего сломать
4. `docker logs promptvault-api-1` содержит `apikey.auth.success` при запросах с ключами

### B. Operational шаги для CWS submission

Эти задачи **требуют человеческого участия**, не автоматизируются.

#### B1. Подготовка аккаунта (одноразово)

- [ ] Оплатить **$5** за Chrome Web Store Developer Account на `https://chrome.google.com/webstore/devconsole/register`
- [ ] (Опционально, параллельно) Зарегистрировать Microsoft Partner Center для Edge Add-ons — бесплатно

#### B2. Контент для listing

- [ ] **Screenshots (3 штуки) 1280×800**. Могу сгенерировать через Playwright MCP:
   - Screenshot 1: Home со списком промптов, workspace selector открыт, hover preview
   - Screenshot 2: VariableForm с заполненными переменными, подсветка `{{var}}`, char counter
   - Screenshot 3: Success toast «Вставлено в ChatGPT» с кнопкой Undo
- [ ] **Promo tile 440×280**. Простой SVG→PNG с логотипом + tagline. Могу сгенерить Node-скриптом.
- [ ] **Описание на русском** (до 132 символов short + до 16000 full):
   - Short: `ПромтЛаб — вставка ваших AI-промптов в ChatGPT, Claude, Gemini, Perplexity одним кликом`
   - Full: описание с преимуществами, permission justifications, ссылка на Privacy Policy
- [ ] **(Опционально) Описание на английском** для международных юзеров

#### B3. Privacy Policy на prod

- [x] React route `/legal/extension-privacy` уже в коде
- [ ] После `git push main` → проверить что `https://promtlabs.ru/legal/extension-privacy` открывается
- [ ] **Заменить email `privacy@promtlabs.ru`** в тексте страницы на реальный, если такой почты нет
- [ ] **Обновить GitHub ссылки** если repo URL не `slava4123/promtlab`

#### B4. Upload в CWS

- [ ] Скачать `.output/promptvault-extension-0.1.0-chrome.zip` либо локально через `npx wxt zip`, либо из CI artifacts
- [ ] На CWS dashboard → Upload new item → загрузить zip
- [ ] Заполнить все поля:
   - **Название**: `ПромтЛаб — библиотека AI-промптов`
   - **Short description**: см. B2
   - **Detailed description**: см. B2
   - **Category**: Productivity или Developer Tools
   - **Language**: Russian (primary) + English (optional)
   - **Icon**: auto-detect из manifest (128×128)
   - **Screenshots**: 3 шт из B2
   - **Privacy policy URL**: `https://promtlabs.ru/legal/extension-privacy`
   - **Single purpose**: «Вставка пользовательских AI-промптов в поля ввода ChatGPT, Claude, Gemini, Perplexity»
   - **Permissions justifications**: скопировать текст из таблицы в Privacy Policy (секция 4)
   - **Distribution**: Countries → выбрать RU + СНГ (или Worldwide)
   - **Visibility**: Public (или Unlisted для тестирования)
- [ ] **Submit for review**
- [ ] Ждать **1-3 недели** на review (иногда быстрее)
- [ ] Следить за email — reviewers могут запросить дополнительную информацию

#### B5. После approve

- [ ] Получить **Extension ID** (длинная строка типа `bigfphphnadlglhdgppanhlckpbpafk`)
- [ ] Собрать CWS URL: `https://chromewebstore.google.com/detail/<EXTENSION_ID>`
- [ ] В `frontend/src/components/settings/extension-promo-section.tsx` заменить:
   ```typescript
   const CHROME_WEB_STORE_URL = "" // пусто = кнопка в "Скоро" состоянии
   ```
   на:
   ```typescript
   const CHROME_WEB_STORE_URL = "https://chromewebstore.google.com/detail/<EXTENSION_ID>"
   ```
- [ ] `git commit -am "feat(ext): enable CWS install button" && git push main`
- [ ] CI пересоберёт frontend, задеплоит — промо-кнопка автоматически станет активной: `variant="default"`, onClick открывает CWS, badge «Скоро» исчезает

#### B6. (Опционально) Edge Add-ons — параллельный путь

Edge Add-ons review обычно **быстрее** CWS (дни vs недели) и бесплатно.

- [ ] Зарегистрироваться на `https://partner.microsoft.com/dashboard/microsoftedge`
- [ ] Upload тот же zip
- [ ] Заполнить listing (те же описания, screenshots, privacy policy)
- [ ] Submit
- [ ] После approve — добавить Edge Store URL в promo-section, показывать Edge-юзерам

#### B7. (Опционально) Post-launch

- [ ] **Sentry интеграция** для extension — зарегистрировать отдельный GlitchTip project, получить DSN, обновить `lib/sentry.ts` с реальной отправкой событий, добавить `VITE_EXTENSION_SENTRY_DSN` в CI secrets
- [ ] **Automation publishing workflow** — отдельный `release-extension.yml` с `workflow_dispatch` триггером для автоматического upload в CWS через `chrome-webstore-upload-cli`
- [ ] **Firefox port** — `wxt build -b firefox`, правка manifest под `sidebar_action`, submit в Mozilla Add-ons (бесплатно)

---

## Roadmap после MVP launch (Этапы из предыдущей версии документа)

### Этап 1 — Функциональная полнота (закрыто)

| Пункт | Статус | Комментарий |
|---|---|---|
| 1.1 Workspace switcher + Collection filter | ✅ Сделано | Radix Select с иконками, portal, keyboard nav |
| 1.2 Sync с SPA | ✅ Сделано | `refetchOnWindowFocus: true` + кнопка Refresh |
| 1.3 Copy-to-clipboard | ✅ Сделано | кнопка + hotkey `Ctrl+Shift+C` + toast |
| 1.4 Drag-and-drop reorder Pinned | 🟡 Deferred | Требует backend миграцию + dnd-kit. Польза низкая. |
| 1.5 Bulk actions | 🟡 Deferred | Полезно только для 100+ промптов. |

### Этап 2 — Public launch (в процессе)

| Пункт | Статус | Что дальше |
|---|---|---|
| 2.1 Privacy Policy | ✅ Код готов | Задеплоить через push + проверить URL |
| 2.2 Screenshots × 3 | ⏳ В очереди | Снять через Playwright MCP или вручную |
| 2.3 Promo tile 440×280 | ⏳ В очереди | Сгенерить SVG→PNG |
| 2.4 CWS submission | ⏳ Блокер: $5 account | Ручной процесс, 1-3 недели review |
| 2.5 Edge Add-ons submission | ⏳ Опционально | Параллельно с CWS, бесплатно, быстрее |
| 2.6 Промо-кнопка в SPA | ✅ Код готов | Автоактивируется после подмены `CHROME_WEB_STORE_URL` |

### Этап 3 — Production integration

| Пункт | Статус | Комментарий |
|---|---|---|
| 3.1 Production API URL | ✅ Сделано | Дефолт `https://promtlabs.ru` в `lib/storage.ts` |
| 3.2 Sentry integration | 🟡 Stub | `lib/sentry.ts` — stub с breadcrumbs. Реальная отправка в GlitchTip — позже. |
| 3.3 Pro-only gate | 🟡 Отложен | До реализации Monetization Phase 1 в backend |

### Этап 4 — Cross-browser

| Пункт | Статус | Объём |
|---|---|---|
| 4.1 Firefox port | 🟡 Backlog | 1-2 часа + AMO submission (бесплатно) |
| 4.2 Safari port | 🟡 Very low priority | 4-6 часов + Apple Dev $99/год |

### Этап 5 — Bonus features (в основном сделано)

| Пункт | Статус |
|---|---|
| 5.1 Per-prompt keyboard shortcuts | 🟡 Backlog (через `chrome.commands`) |
| 5.2 Streaks UI | ✅ `StreakBadge` в header |
| 5.3 Share prompt link | ✅ Сделано |
| 5.4 Multi-target inserts | ✅ Сделано (⚡ кнопка) |
| 5.5 Hover preview | ✅ Сделано |
| 5.6 Tag filter | 🟡 Foundation готов (`PromptFilter.tagIds`), UI — backlog |
| 5.7 Recent history local | ✅ Сделано (`addLocalRecent`) |

---

## Критические файлы для навигации

**Extension**:
- `promptvault/extension/wxt.config.ts` — manifest metadata + permissions
- `promptvault/extension/lib/api.ts` — все API вызовы (teams, collections, prompts, streak, share)
- `promptvault/extension/lib/messages.ts` + `lib/messages-helpers.ts` — type-safe message passing
- `promptvault/extension/lib/storage.ts` — chrome.storage.local wrapper (apiKey, theme, workspace, savedVars, lastInsert, localRecent, onboarding, promptCache)
- `promptvault/extension/lib/insert.ts` — cascade insertion
- `promptvault/extension/lib/selectors.ts` — per-site DOM selectors
- `promptvault/extension/lib/content-handler.ts` — shared content-script logic + undo
- `promptvault/extension/lib/sentry.ts` — lightweight wrapper с breadcrumbs + PII scrub
- `promptvault/extension/entrypoints/background.ts` — SW brokering + insertIntoActiveTab + insertIntoAllSupportedTabs + undo
- `promptvault/extension/entrypoints/sidepanel/main.tsx` — React root + Sentry init
- `promptvault/extension/components/app.tsx` — view state machine + toast handling
- `promptvault/extension/components/home.tsx` — главный экран + workspace integration + keyboard shortcuts
- `promptvault/extension/components/variable-form.tsx` — форма переменных + preview + actions (insert/copy/multi/share)
- `promptvault/extension/components/settings-view.tsx` — Settings screen
- `promptvault/extension/components/error-boundary.tsx` — class ErrorBoundary
- `promptvault/extension/components/onboarding-overlay.tsx` — first-run wizard
- `promptvault/extension/components/workspace-selector.tsx` + `collection-selector.tsx` — Radix Select dropdowns
- `promptvault/extension/components/ui/select.tsx` — shadcn/Radix Select composition
- `promptvault/extension/components/ui/toaster.tsx` — custom toast system
- `promptvault/extension/scripts/gen-icons.mjs` — Node PNG generator без deps
- `promptvault/extension/tests/template.test.ts` + `selectors.test.ts` — unit тесты

**Backend**:
- `backend/internal/middleware/auth/apikey.go` — APIKeyAuth middleware
- `backend/internal/middleware/auth/combined.go` — CombinedAuth (JWT + API-key)
- `backend/internal/middleware/cors/cors.go` — CORS с `chrome-extension://` support
- `backend/internal/app/app.go` — wiring

**Frontend SPA**:
- `frontend/src/components/settings/extension-promo-section.tsx` — промо-секция
- `frontend/src/pages/legal/extension-privacy.tsx` — Privacy Policy страница
- `frontend/src/pages/settings.tsx` — включает `<ExtensionPromoSection />`
- `frontend/src/App.tsx` — роут `/legal/extension-privacy`

**CI/CD**:
- `.github/workflows/deploy.yml` — job `test-extension` добавлен, блокирует `build-push` если падает

---

## Verification checklist при первом prod deploy

### Backend (должно работать out of the box)

- [ ] `curl https://promtlabs.ru/api/health` → `{"status":"ok"}` или подобное
- [ ] `curl https://promtlabs.ru/api/auth/me -H "Authorization: Bearer <JWT>"` → 200 JSON user (existing SPA flow)
- [ ] `curl https://promtlabs.ru/api/auth/me -H "Authorization: Bearer pvlt_<key>"` → 200 JSON user (CombinedAuth)
- [ ] `curl -X OPTIONS https://promtlabs.ru/api/prompts -H "Origin: chrome-extension://abcd"` → `Access-Control-Allow-Origin: chrome-extension://abcd`
- [ ] `docker logs promptvault-api-1 --tail 50` → видны записи `apikey.auth.success` если кто-то уже использует ключи
- [ ] Existing SPA user login → dashboard → создать промпт → работает как раньше (no regression)

### SPA

- [ ] `https://promtlabs.ru/settings` → видна секция **«Chrome-расширение»** с badge «Скоро» и disabled кнопкой
- [ ] `https://promtlabs.ru/legal/extension-privacy` открывается **без логина**
- [ ] Privacy Policy страница корректно рендерится в light и dark темах
- [ ] Клик на «Политика конфиденциальности» из промо-секции → ведёт на `/legal/extension-privacy`

### Extension (локальный тест через Load unpacked)

- [ ] Extension загружается без ошибок в `chrome://extensions/`
- [ ] Side Panel открывается через клик на иконку
- [ ] Первый запуск показывает **Onboarding overlay** (4 шага, можно пропустить)
- [ ] После `chrome.storage.local.clear()` → onboarding показывается снова
- [ ] Theme switcher (Settings) переключает мгновенно — light/dark/system
- [ ] Workspace selector показывает «Личное» + список команд (если есть)
- [ ] Collection selector появляется только если коллекции есть
- [ ] Favorite/Pin toggles работают (invalidate → UI обновляется через ~150 мс)
- [ ] Multi-target insert (⚡) вставляет в все открытые ChatGPT+Claude+Gemini+Perplexity вкладки
- [ ] Undo button в toast очищает поле ввода после insert
- [ ] Copy to clipboard (Ctrl+Shift+C или кнопка 📋) работает
- [ ] Share link создаёт URL и кладёт в clipboard
- [ ] Hover preview на PromptCard через 400ms

---

## Что точно не делаем (final)

- **Client-side Pro enforcement** — security theater, enforcement только на backend
- **Inline payment через CWS** — Google убрал API в 2018, только external subscription flow
- **WebAuthn / 2FA в extension** — overkill для API-key auth модели
- **Ручная локализация (`_locales`)** — русский хардкод, English fallback не нужен для РФ рынка
- **Push notifications в extension** — нет реального use-case
- **Webhooks в extension** — только запросы по действию юзера, не background polling
- **Оффлайн режим с local SQLite** — foundation в chrome.storage достаточно, SQLite overkill

---

## Выводы

Extension **технически полностью готов** к публикации в Chrome Web Store. Работа осталась исключительно operational: регистрация dev-аккаунта, подготовка материалов, ручное заполнение CWS listing, ожидание review. Весь код, CI/CD, privacy policy, промо-кнопка в SPA — уже готовы и автоматически активируются как только подменим один URL-placeholder после CWS approve.

**Оценка времени от `git push main` до публичной доступности через CWS**:
- Deploy backend+SPA: ~10 минут (CI)
- Screenshots + promo tile: ~30 минут (могу сгенерировать)
- CWS dev account + upload + listing: ~1 час ручной работы
- **CWS review: 1-3 недели** (блокер — не наш)
- Финальная подмена URL после approve: ~1 минута

**Итого work time: ~2 часа + 1-3 недели review**.
