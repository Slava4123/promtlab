# Известные баги и недореализованные места

Найденные в текущей итерации QA. Сгруппировано по приоритету.

---

## 🔴 Critical — UI ломается / endpoints возвращают 4xx

### B-1. Trash: URL type plural vs singular ✅ ИСПРАВЛЕНО
- **Симптом**: «Не удалось восстановить» / «Не удалось удалить» в Корзине
- **Где**: `lib/api.ts` методы `restoreTrashPrompt`, `restoreTrashCollection`, `permanentDeleteTrashPrompt`, `permanentDeleteTrashCollection`
- **Причина**: я слал `/api/trash/prompts/N/restore` (plural). Backend (`trash/handler.go::parseTypeAndID`) принимает только `prompt` / `collection` (singular)
- **Fix**: заменил `prompts` → `prompt`, `collections` → `collection` в 4 URL-ах
- **Verify**: после reload extension — Корзина → «Вернуть» / «Удалить навсегда» должны работать

### B-2. History: пустые названия промптов ✅ ИСПРАВЛЕНО
- **Симптом**: на странице «История использования» видно только время (18:43, 18:42…) без названий
- **Где**: `pages/history-page.tsx` использует `item.prompt_title`
- **Причина**: backend возвращает `prompt: { id, title, model, tags }` (nested), а я угадал `prompt_title` (плоское поле)
- **Fix**: обновил `UsageHistoryItem` type под backend shape + рендер `item.prompt?.title ?? \`Промпт #${item.prompt_id}\``
- **Verify**: после reload — на /history видны названия + model badge

---

## 🟠 High — фичи отсутствуют

### B-3. Collection drill-down (`/collections/:id`) ✅ ИСПРАВЛЕНО
- **Симптом**: при клике на коллекцию открывается placeholder «PHASE 2»
- **Что нужно**:
  - Page header с иконкой/цветом + имя коллекции + count
  - Список промптов в этой коллекции через `GET /api/prompts?collection_id=:id`
  - Карточки промптов (как на Dashboard)
  - Action button «Открыть в редакторе»
  - Edit/Delete коллекции прямо со страницы
- **Reference**: `frontend/src/pages/collections.tsx` — drill-down logic inline
- **API**: `GET /api/prompts?collection_id=:id` + `GET /api/collections/:id`

### B-4. Tag drill-down (`/tags/:id`) ✅ ИСПРАВЛЕНО
- **Симптом**: placeholder при клике на тег
- **Что нужно**: аналогично коллекциям, фильтр `tag_ids=:id`
- **API**: `GET /api/prompts?tag_ids=:id`

### B-5. Notifications settings (`/notifications` и `/settings/notifications`) ✅ ИСПРАВЛЕНО (частично)
- Реализован Smart Insights toggle. Streak/weekly — backend endpoints отсутствуют, помечены как «Скоро».
- **Симптом**: placeholder «PHASE 5»
- **Что нужно**:
  - Toggle `insight_emails_enabled` (Smart Insights digest)
  - Toggle streak reminders
  - Toggle weekly digest
- **API**:
  - `PATCH /api/auth/notifications/insights` — есть
  - Остальные нужно уточнить в backend (могут не существовать)

### B-6. Security settings (`/settings/security`) ⏸ ОТЛОЖЕНО
- Backend endpoints для обычных юзеров (`/api/auth/2fa/*`, `/api/auth/sessions`) не существуют. Остаётся placeholder с deep-link.
- **Симптом**: placeholder
- **Что нужно**:
  - 2FA enroll (QR code через `qrcode.react`)
  - Active sessions list + revoke
  - Password change уже есть в Profile
- **API**:
  - `POST /api/auth/2fa/{enroll,verify,disable}` — статус неизвестен
  - `GET /api/auth/sessions`, `DELETE /api/auth/sessions/:id` — статус неизвестен

### B-7. Connected accounts (`/settings/accounts`) ✅ ИСПРАВЛЕНО
- Список linked accounts + unlink. Link через OAuth flow — deep-link на web.
- **Симптом**: placeholder
- **Что нужно**:
  - Список linked accounts (Google/GitHub/Yandex) + статус
  - Link new account → OAuth flow
  - Unlink existing
- **API**:
  - `GET /api/auth/linked-accounts`
  - `POST /api/auth/link/:provider`
  - `DELETE /api/auth/unlink/:provider`

### B-8. Referral (`/settings/referral`) ✅ ИСПРАВЛЕНО
- Код, copy-link, stats invited_count + reward_granted.
- **Симптом**: placeholder
- **Что нужно**: код, copy-link, список приглашённых
- **API**: `GET /api/auth/referral`

---

## 🟡 Medium — chains UX

### B-9. Chains: editor, detail, canvas (`/chains/new`, `/chains/:id`, `/chains/:id/edit`, `/chains/:id/canvas`) ✅ ИСПРАВЛЕНО
- Все 4 страницы реализованы. Canvas — vertical timeline вместо @xyflow (для узкого sidepanel).
  Тонкое редактирование variable_mapping и conditions — deep-link на веб.
- **Симптом**: 4 placeholders «Phase 3 polish»
- **Сложность**: высокая
- **Что нужно**:
  - New: dialog с name/description → POST /api/chains
  - Detail: read-only обзор + actions
  - Edit: inline-tree editor (35KB в frontend), prompt picker, fork conditions
  - Canvas: DAG через `@xyflow/react` + `elkjs` (нужно установить, или vertical timeline для узкого sidepanel)

### B-10. Chain run: chain_var resolution (известный edge-case) ✅ ИСПРАВЛЕНО
- Убрал fallback `exec.step_outputs[step_<var_name>]` — backend ключ `step_<id>` (uint),
  не совпадает с `var_name` (string). chain_var теперь только из `exec.variables`
  (initial-level), как frontend run.tsx после Phase 16-C.
- **Симптом**: при использовании `type: "chain_var"` переменных значение из step_output может не подставиться
- **Где**: `pages/chains/run-page.tsx::resolveStepContent`
- **Причина**: backend пишет `step_outputs[step_<step.id>]`, я ищу по `step_${src.var_name}`. Несовпадение ключа
- **Что нужно**: уточнить актуальный формат от `chain.Service.AdvanceStep` в backend и привести client-side resolution в соответствие

---

## 🟢 Low — polish

### B-11. Team branding edit (`/teams/:slug/branding`) ✅ ИСПРАВЛЕНО
- Logo upload (multipart direct fetch), 12 brand-чипов + color picker, tagline/website.
- **Симптом**: placeholder, только read-only display
- **Что нужно**: upload logo (multipart bytea), color palette picker
- **API**: `POST /api/teams/:slug/branding/logo`, `PUT /api/teams/:slug/branding`

### B-12. Team analytics (`/teams/:slug/analytics`) ✅ ИСПРАВЛЕНО
- Range selector, metric cards с delta, top prompts, contributors leaderboard, model segmentation.
- **Симптом**: placeholder
- **Что нужно**: аналогично personal analytics но team-scoped
- **API**: `GET /api/analytics/teams/:id`

### B-13. Team activity (`/teams/:slug/activity`) ✅ ИСПРАВЛЕНО
- Infinite-scroll лента событий с event_type → label/icon маппингом.
- **Симптом**: placeholder
- **Что нужно**: virtualized timeline событий
- **API**: `GET /api/teams/:slug/activity` (cursor-paginated)

### B-14. Onboarding flow расширение
- **Что есть**: базовый `OnboardingOverlay`
- **Что нужно**: 5-step tour для новых юзеров

### B-15. Changelog popup
- **Что есть**: `ChangelogPage` (страница)
- **Что нужно**: popup-toast при первом запуске после release с has_unread=true

---

## 🔵 Infrastructure

### B-16. Firefox build
- WXT поддерживает, `build:firefox` уже в package.json
- Нужно: `webextension-polyfill`, sidebar entrypoint вместо side_panel, AMO submission

### B-17. Selector reliability monitoring
- На 5 новых сайтах селекторы best-effort
- Нужно: Sentry breadcrumb `selector.miss` + counter в telemetry

### B-18. Live recon для 5 новых сайтов
- Yandex GPT (`yandex.ru/alice` новый редизайн), GigaChat, DeepSeek, Mistral, Qwen
- Селекторы input/response best-effort без живой валидации
- Нужно: на каждом сайте — открыть DevTools, проверить актуальные селекторы, обновить `lib/selectors.ts` + `lib/chat-selectors.ts`

### B-19. CWS + AMO release
- ✅ Privacy policy URL: `https://promtlabs.ru/legal/extension-privacy`
- ✅ Screenshots: 5×1280×800 в `extension/store-screenshots/`
- ✅ Promo tile (large 1400×560 + small 920×680 в той же папке;
  440×280 small marketing tile — опционально, пока нет)
- ✅ Описание: `docs/CWS_LISTING.md` §2.2 (3000 символов RU)
- ✅ Single-purpose statement: `CWS_LISTING.md` §8
- ✅ Чек-лист подачи: `docs/PUBLISHING_CHECKLIST.md`
- AMO (Firefox) — отложен на следующую итерацию

---

## 📊 Метрики реализации

- **Routes**: 30+
- **Полностью реализованы**: ~25
- **Placeholder'ы**: ~5 (см. B-3..B-8, B-9 cluster)
- **Известные баги**: 2 critical (фикснуты в текущей итерации) + ~17 todo

## Стратегия дальше

1. **Текущий приоритет** — фиксы B-1, B-2 (✅ сделано) + B-3, B-4, B-5 (drill-down + notifications)
2. **Phase Chains polish** — B-9 cluster (большая работа)
3. **Settings polish** — B-6, B-7, B-8
4. **Release prep** — B-16, B-19

Каждый bug фиксится в отдельном PR, но в текущем dev-цикле — батчем.
