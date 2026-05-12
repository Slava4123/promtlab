# Extension TODO — Phases & Placeholders

Сборная таблица всех placeholder-страниц и polish-фаз, которые сейчас не реализованы в extension. Указано почему отложено, что нужно сделать, и из каких файлов брать референс из Web.

---

## Phase 1 polish

### Prompt detail для `/prompts/:id` (отдельная страница уже сделана)
- ✅ Готово (см. `pages/prompts/detail-page.tsx`)

### Prompt editor: file import drop-zone
- **Где**: `pages/prompts/editor-page.tsx`
- **Что**: drag-drop файлов (.md, .txt, .docx) → парсинг → контент промпта
- **Reference**: `frontend/src/pages/prompt-editor.tsx` — `FileImportDropZone` компонент с `mammoth` (docx), `pdfjs-dist` (PDF), `turndown` (HTML → markdown)
- **Зависимости**: `mammoth`, `pdfjs-dist`, `turndown`, `jschardet` (encoding detection)

### Prompt editor: split-view (Edit | Preview одновременно)
- **Где**: `pages/prompts/editor-page.tsx`
- **Что**: на широких screenshot — split-view (две колонки), на узком — tab-switcher (как сейчас)
- **Reference**: `frontend/src/components/prompts/prompt-split-editor.tsx`

---

## Phase 2 polish (organization)

### `/collections/:id` — drill-down коллекции
- **Где**: `routes.tsx` line ~64 (placeholder)
- **Что**: страница «промпты внутри коллекции» — header + список промптов с фильтром по collection_id
- **Reference**: `frontend/src/pages/collections.tsx` (drill-down inline + dialog)
- **API**: `GET /api/prompts?collection_id=:id` — уже есть

### `/tags/:id` — drill-down тега
- **Где**: `routes.tsx` (placeholder)
- **Что**: страница «промпты с этим тегом»
- **API**: `GET /api/prompts?tag_ids=:id`

### `/teams/:slug/analytics` — аналитика команды
- **Где**: `routes.tsx` (placeholder)
- **Что**: метрики команды (top contributors, top prompts)
- **Reference**: `frontend/src/pages/team-analytics.tsx`
- **API**: `GET /api/analytics/teams/:id`

### `/teams/:slug/activity` — лента активности команды
- **Где**: `routes.tsx` (placeholder)
- **Что**: virtualized timeline событий (prompt.created, member.invited, role.changed)
- **Reference**: `frontend/src/pages/team-activity.tsx` + `frontend/src/components/activity/activity-timeline.tsx`
- **API**: `GET /api/teams/:slug/activity` (cursor-paginated)

### Dashboard: stat cards (Total / Favorites / Usage / Collections)
- **Где**: `components/home.tsx` — добавить cards в шапку
- **Reference**: `frontend/src/pages/dashboard.tsx`

---

## Phase 3 polish (chains)

### `/chains/new` — создание цепочки в extension
- **Где**: `routes.tsx` (placeholder «Phase 3 polish»)
- **Что**: форма с name/description → POST /api/chains → редирект на `/chains/:id/edit`
- **Reference**: `frontend/src/pages/chains/index.tsx` (dialog)

### `/chains/:id` — детали цепочки
- **Где**: `routes.tsx` (placeholder)
- **Что**: read-only обзор цепочки + список steps + actions (Edit, Run, Runs, Delete, Duplicate)
- **Reference**: `frontend/src/pages/chains/editor.tsx` (header section)

### `/chains/:id/edit` — редактор цепочки (inline-tree)
- **Где**: `routes.tsx` (placeholder)
- **Что**: древовидный редактор шагов: добавить prompt-шаг, fork-шаг, reorder, edit variable_mapping, conditions
- **Reference**: `frontend/src/pages/chains/editor.tsx` (35KB файл)
- **Сложность**: высокая — требует tree-walking logic, prompt picker, fork conditions builder
- **Зависимости**: уже установлен `@dnd-kit/sortable` (но не используется ещё)

### `/chains/:id/canvas` — DAG-визуализация
- **Где**: `routes.tsx` (placeholder)
- **Что**: graph через `@xyflow/react` + `elkjs` (auto-layout)
- **Reference**: `frontend/src/pages/chains/canvas.tsx`
- **Зависимости**: `@xyflow/react`, `elkjs` — нужно установить
- **Адаптация под sidepanel**: на узком viewport — vertical timeline вместо canvas

### Chain run: AI response autocapture
- **Где**: `pages/chains/run-page.tsx` — кнопка «Захватить из вкладки» уже есть
- **Reference**: `lib/chat-selectors.ts` (best-effort селекторы для 12 хостов)
- **Что доработать**: на новых сайтах селекторы могут отвалиться → нужно live recon + Sentry breadcrumbs

---

## Phase 4 polish (teams)

### `/teams/:slug/branding` — редактирование брендинга
- **Где**: `routes.tsx` (placeholder «Phase 4 polish»)
- **Что**: текущая страница только display; нужно: upload logo (bytea), color palette picker, save
- **Reference**: `frontend/src/components/teams/branding-form.tsx`, `logo-uploader.tsx`, `color-palette-picker.tsx`
- **API**: `POST /api/teams/:slug/branding/logo` (multipart), `PUT /api/teams/:slug/branding` — уже есть
- **Сложность**: medium — bytea upload требует multipart/form-data в bg-client

### Team invitations: мои приглашения
- **Где**: ещё нет route
- **Что**: страница «Меня пригласили в команды» → принять/отклонить
- **Reference**: `frontend/src/pages/invitations.tsx`
- **API**: `GET /api/invitations`, `POST /api/invitations/:id/accept`, `/decline`

---

## Phase 5 polish (settings)

### `/settings/security` — 2FA, sessions, password change beyond profile
- **Где**: `routes.tsx` (placeholder)
- **Что**: TOTP enroll (QR-code), список активных sessions, revoke session
- **Reference**: `frontend/src/pages/settings/security.tsx`
- **API**: `POST /api/auth/2fa/{enroll,verify,disable}`, `GET /api/auth/sessions`, `DELETE /api/auth/sessions/:id`
- **Зависимости**: `qrcode.react` (для QR display)

### `/settings/accounts` — Connected accounts (OAuth)
- **Где**: `routes.tsx` (placeholder)
- **Что**: список linked accounts (Google, GitHub, Yandex), link/unlink
- **Reference**: `frontend/src/pages/settings/accounts.tsx`
- **API**: `GET /api/auth/linked-accounts`, `POST /api/auth/link/:provider`, `DELETE /api/auth/unlink/:provider`

### `/settings/notifications` — email preferences
- **Где**: `routes.tsx` (placeholder)
- **Что**: toggle insight_emails_enabled, streak_reminders, weekly_digest
- **Reference**: `frontend/src/pages/settings/notifications.tsx`
- **API**: `PATCH /api/auth/notifications/insights`

### `/settings/referral` — реферальная программа
- **Где**: `routes.tsx` (placeholder)
- **Что**: реферальный код, copy-link, список приглашённых
- **Reference**: `frontend/src/pages/settings/referral.tsx`
- **API**: `GET /api/auth/referral`

---

## Phase 6 polish (auth)

### Полный OAuth flow в extension
- **Где**: сейчас deep-link на Web
- **Что**: `chrome.identity.launchWebAuthFlow` → OAuth provider → callback → access_token
- **Reference**: `frontend/src/stores/auth-store.ts` (OAuth обработка через redirect)
- **Зависимости**: `permissions: ["identity"]` в manifest

### Email/password sign-in в extension
- **Где**: сейчас только API-key setup
- **Что**: форма email/password → POST /api/auth/login → JWT → сохранить и работать как frontend
- **Reference**: `frontend/src/pages/sign-in.tsx`
- **Изменения**: переход с API-key auth на JWT auth, refresh cookie handling, или dual-mode

### TOTP support
- **Где**: после email/password sign-in
- **Что**: при `totp_required` → форма с code input → POST /api/auth/verify-totp
- **API**: уже есть

### Forgot/Reset password в extension
- **Что**: deep-link → email → token → reset form
- **API**: `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`

---

## Phase 7 polish

### Conditional chains visual builder
- **Где**: сейчас в Web JSON-textarea для conditions
- **Что**: drag-drop builder для branches: label + next_step_id selector
- **Reference**: Phase 16-C в backend (отложено, требует Max-only)

### Downgrade preview dialog
- **Где**: `pages/settings/subscription-page.tsx` — cancel/downgrade без preview
- **Что**: перед cancel показать warning «потеряете X промптов, Y коллекций»
- **API**: `GET /api/subscription/downgrade-preview?planId=free`
- **Reference**: `frontend/src/components/subscription/downgrade-preview-dialog.tsx`

### Pricing: yearly toggle с −10%
- **Где**: `pages/pricing-page.tsx` — уже частично есть
- **Что**: проверить корректность отображения yearly-планов

---

## Phase 8 polish

### Firefox build
- **Где**: `wxt.config.ts` — `build:firefox` script есть, но не настроен
- **Что**:
  - `webextension-polyfill` dep
  - `sidebar_action` entrypoint вместо `side_panel` для Firefox
  - browser-specific manifest через `manifest: ({ browser }) => ...`
  - CI matrix chrome + firefox в `.github/workflows/extension.yml`
- **Дистрибуция**: addons.mozilla.org (AMO) review 1-2 недели

### Programmatic content script injection
- ✅ Реализовано (см. `background.ts` → `reinjectContentScripts`)

### Selector reliability monitoring
- **Где**: `lib/chat-selectors.ts` — селекторы best-effort
- **Что**: Sentry breadcrumb `selector.miss` + telemetry counter
- **Reference**: `WXT_SENTRY_DSN` → GlitchTip
- **Зависимости**: sentry уже есть

### Visual builder для template variables
- **Где**: сейчас просто {{var}} в textarea
- **Что**: UI с autocomplete переменных, validation, типизация
- **Reference**: `frontend/src/components/prompts/prompt-split-editor.tsx`

---

## Phase 9 polish (нереализованные)

### Onboarding flow
- **Где**: `components/onboarding-overlay.tsx` — есть basic, нужно расширить
- **Что**: 5-step tour: sign-in, browse prompts, search Cmd+K, insert на claude.ai, попробовать chain

### Changelog popup
- **Где**: `pages/changelog-page.tsx` — есть страница, но не popup
- **Что**: при release новой версии → toast «Что нового»

### Code-splitting analysis
- **Где**: editor-page chunk 716KB (CodeMirror), versions-page 118KB (diff-viewer)
- **Что**: lazy load `@uiw/react-codemirror` только в editor; `react-diff-viewer-continued` только в versions

### i18n (RU primary, EN fallback)
- **Где**: пока только RU hardcoded
- **Что**: `react-i18next` + `locales/{ru,en}.json`

### E2E tests (Playwright + chrome-devtools-mcp)
- **Где**: нет
- **Что**: critical flows: auth, create prompt, run chain, share

### CWS + AMO release
- **Где**: пока только локальный unpacked
- **Что**: zip + upload в Chrome Web Store + AMO

---

## Bugs / improvements (отдельный треккер)

### Home component имеет свой top-header (дубликат с AppShell)
- **Где**: `components/home.tsx`
- **Что**: убрать свой header, использовать global AppHeader
- **Зависимости**: рефактор AppShell + Home

### Prompt card: hover-actions (Edit/Delete/Share)
- **Где**: `components/prompt-card.tsx` — есть только Pin/Favorite
- **Что**: добавить Edit/Delete/Share через dropdown menu

### Drag-drop reorder в chain editor
- **Где**: chain editor вообще placeholder
- **Зависимости**: `@dnd-kit/sortable` (нужно установить)

### History page (этот файл документирует — ниже реализован)
- ✅ Перешло в реализованное в текущей итерации

### Analytics page
- ✅ Перешло в реализованное в текущей итерации

### Coverage новых сайтов
- **Где**: `lib/selectors.ts` — Yandex/GigaChat/DeepSeek/Mistral/Qwen селекторы best-effort
- **Что**: live recon на каждом сайте после login → актуализация input selectors + response selectors

---

## Архитектурные вопросы (для обсуждения)

1. **Frontend ↔ Extension drift**: shared package есть только для types/template/utils. Hooks дублированы. Долгосрочно стоит вынести API client + hooks в `@pv/shared/api`.

2. **chain_var format**: backend пишет в `step_outputs[step_<id>]`, extension в `resolveStepContent` ищет `step_${src.var_name}`. Несоответствие может ломать chains с chain_var переменными. Нужна верификация через тест end-to-end.

3. **CWS submission**: privacy policy URL, иконки 128px, screenshots, описание — `extension/store-screenshots/` есть только promo-small.html.
