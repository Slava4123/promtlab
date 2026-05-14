# Chrome Web Store — пошаговый чек-лист публикации

Скоуп: только Chrome Web Store (CWS). Тексты, ассеты и обоснования
permission'ов находятся в `CWS_LISTING.md`. Полная политика
конфиденциальности — в `PRIVACY_POLICY.md`. Этот файл — пошаговая
инструкция «что нажать», чтобы дойти от чистого репо до Submit.

> ℹ️ Firefox AMO / Edge Add-ons / Yandex Browser — отдельная итерация.
> Yandex Browser использует Chrome Web Store напрямую и не требует
> отдельной подачи.

---

## 0. Перед сборкой — обязательно

- [ ] Working tree чистый: `git status` → nothing to commit
- [ ] `promptvault/shared/` закоммичен (alias `@pv/shared` должен резолвиться)
- [ ] Manifest version = `1.0.0` в `wxt.config.ts:43` и `package.json:5`
- [ ] Privacy policy опубликована и доступна по
      `https://promtlabs.ru/legal/extension-privacy` (страница уже есть в
      `frontend/src/pages/legal/extension-privacy.tsx` — нужно убедиться, что
      production frontend задеплоен)
- [ ] Backend endpoints живы: `https://promtlabs.ru/api/auth/me` отвечает
      (тестовый ключ для smoke)

---

## 1. Сборка production-ZIP

В директории `promptvault/extension/`:

```powershell
npm install               # подтянет deps, в т.ч. @pv/shared через alias
npm run compile           # tsc --noEmit (strict TypeScript)
npm run lint              # ESLint 9
npm run test              # vitest run
npm run build             # wxt build → .output/chrome-mv3/
npm run zip               # wxt zip  → .output/promptvault-extension-1.0.0-chrome.zip
```

Проверки:
- [ ] Все 5 команд завершились без ошибок
- [ ] `.output/chrome-mv3/manifest.json` существует, version=1.0.0
- [ ] `.output/promptvault-extension-1.0.0-chrome.zip` < 10 MB
- [ ] Manual smoke: `npm run dev` → load unpacked `.output/chrome-mv3-dev/`
      в chrome://extensions → открыть sidepanel по `Ctrl+Shift+K` →
      проверить, что /, /prompts, /chains, /settings/notifications,
      /history рендерятся без console-ошибок

---

## 2. Developer account

- [ ] Создан аккаунт на https://chrome.google.com/webstore/devconsole/
      (одноразовая пошлина $5, оплачивается с карты которую Google примет)
- [ ] Включена 2FA на Google-аккаунте разработчика
- [ ] Заполнен Publisher profile: имя, email, страна (Russia или другая)

---

## 3. New item — загрузка пакета

1. https://chrome.google.com/webstore/devconsole/ → **+ New item**
2. **Upload Pack**:
   - [ ] Загрузить `.output/promptvault-extension-1.0.0-chrome.zip`
   - [ ] Дождаться парсинга manifest'а — CWS покажет version, permissions,
         host_permissions. Должно быть: version=1.0.0, 5 permissions,
         13 host_permissions.

---

## 4. Store listing (вкладка «Информация в магазине»)

Все тексты находятся в `CWS_LISTING.md`.

- [ ] **Detailed description**: скопировать блок из `CWS_LISTING.md` §2.2
      (RU, ~3000 символов; CWS не рендерит markdown — формат plain-text + emoji)
- [ ] **Category**: Productivity (см. §3)
- [ ] **Language**: Russian
- [ ] **Icon (128×128)**: автоматически из manifest (`public/icon/128.png`)
- [ ] **Screenshots** (минимум 1, максимум 5; 1280×800 или 640×400):
      загрузить все 5 из `extension/store-screenshots/screenshot-{1..5}.png`
      в указанном там порядке
- [ ] **Small promo tile** (440×280): _опционально_, сейчас нет — пропустить
      (CWS не блокирует submission без него)
- [ ] **Marquee promo tile** (1400×560): `extension/store-screenshots/promo-large.png`
- [ ] **YouTube video**: _опционально_, пропустить

---

## 5. Privacy practices (вкладка «Конфиденциальность»)

Формулировки — `CWS_LISTING.md` §10. Декларации:

- [ ] **Single purpose** (textarea): из `CWS_LISTING.md` §8
- [ ] **Permission justification** (textarea per permission): из §7
  - sidePanel → "Открытие боковой панели как основного UI расширения."
  - storage → "Сохранение API-ключа, темы, локального кэша промптов."
  - activeTab → "Определение активной AI-вкладки для вставки промпта."
  - scripting → "Re-inject content-scripts при обновлении расширения."
  - contextMenus → "Меню «Сохранить выделение как промпт»."
  - host_permissions → "Вставка промптов в 9 AI-чатов и связь с promtlabs.ru."
- [ ] **Remote code** (CWS вопрос): **No, I am not using Remote code**
      (всё bundled — `lib/sentry-envelope.ts` шлёт NDJSON, не загружает скрипты)
- [ ] **Data collection** form (галочки):
  - [x] Personally identifiable info (email)
  - [x] Authentication info (API key, JWT в backend)
  - [x] User activity (события вставки промптов, без содержимого)
  - [x] Website content — **частично**: контент промптов, создаваемый юзером;
        НО НЕ контент сторонних сайтов
  - [ ] Health, Financial, Location, Personal communications, Web history,
        Personal browsing history → uncheck
- [ ] **Data usage disclosures**:
  - [x] "I do not sell or transfer user data to third parties"
  - [x] "I do not use or transfer user data for purposes unrelated to the
        item's single purpose"
  - [x] "I do not use or transfer user data to determine creditworthiness"
- [ ] **Privacy policy URL**: `https://promtlabs.ru/legal/extension-privacy`

---

## 6. Distribution (вкладка «Распространение»)

- [ ] **Visibility**: Public (или Unlisted, если хочется сначала закрытое
      тестирование — потом переключим)
- [ ] **Regions**: All regions (или явный список — Россия + СНГ + страны
      без блокировок Google)
- [ ] **Pricing**: Free (платежи происходят на promtlabs.ru, не в CWS)
- [ ] **Support email**: `support@promtlabs.ru`
- [ ] **Homepage URL**: `https://promtlabs.ru`

---

## 7. Submit

- [ ] Нажать **Submit for review**
- [ ] Дождаться email от Google (≈1–7 дней)
- [ ] При rejection — прочитать `manifest` issues / store listing issues,
      исправить и переподать. Типичные блокеры:
  - Слишком общий permissions justification — переписать более точно
  - Скриншоты с водяными знаками сторонних брендов в крупном плане
  - Privacy policy URL недоступен (404) — задеплоить frontend
  - Манифест требует permission, которой нет реального использования —
      убрать из manifest

---

## 8. Post-approval

- [ ] Получен Chrome Web Store URL вида
      `https://chromewebstore.google.com/detail/<extension-id>`
- [ ] Сохранить `<extension-id>` — пригодится для:
  - Pinned ссылки в `frontend/src/pages/landing` («Установить из Chrome Web Store»)
  - `docs/MCP.md` upsell-блока
  - Email-рассылок и social-promo
- [ ] Прогнать через свежий профиль Chrome: install → sign-in →
      open ChatGPT → insert prompt → должно работать end-to-end
- [ ] Добавить link в `frontend/src/pages/landing.tsx` (или где hero-CTA)
- [ ] Анонс в Telegram-канале / лендинге

---

## Версионирование релизов

После approval любого изменения требуется новая submission:

1. Bump version в `wxt.config.ts` (manifest) **и** `package.json`
2. Дождаться, чтобы CI прогнал сборку (`.github/workflows/deploy.yml`)
3. Локально `npm run zip` или скачать artifact из CI
4. В CWS dashboard → existing item → **Package** → Upload new pack
5. Заполнить «What's new in this version» (макс 250 символов)
6. Submit for review

CWS обычно проверяет minor-update быстрее, чем первую публикацию (часто часы).

---

## Известные риски

- **Single-purpose review:** CWS может попросить пояснить, как chains и
  командные пространства относятся к «управлению промптами». Аргумент:
  цепочки = упорядоченные коллекции промптов; команды = пространство для
  обмена библиотекой. Всё подчинено единой цели — управлению библиотекой.
- **Host permissions x13:** ~~необычно много~~ оправдано: каждый домен —
  отдельный AI-чат, и для каждого нужна вставка через content-script.
  Промптить альтернативу с `optional_host_permissions` пока не готовы.
- **Remote code (Sentry):** CWS-боты могут флагнуть `glitchtip.promtlabs.ru`.
  Защита: `lib/sentry-envelope.ts` строит NDJSON-строку и шлёт через `fetch`
  — никаких eval/dynamic import. Можно прямо в justification написать:
  "Sentry NDJSON endpoint — отправка JSON через fetch, без загрузки скриптов".

---

## Связанные документы

- `CWS_LISTING.md` — тексты, метаданные, permission justifications, скриншот-specs
- `PRIVACY_POLICY.md` — полная политика, публикуется на promtlabs.ru
- `EXTENSION_TODO.md` — placeholder'ы и поздние phase'ы (не блокеры submission)
- `BUGS_TO_FIX.md` — история фиксов (release notes source)
