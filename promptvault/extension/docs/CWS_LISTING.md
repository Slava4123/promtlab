# Chrome Web Store + AMO — листинг расширения «ПромтЛаб»

Тексты, метаданные и ассеты для submission в Chrome Web Store (Chrome) и addons.mozilla.org (Firefox).

---

## 1. Базовая информация

| Field | Value |
|-------|-------|
| **Name** | ПромтЛаб — библиотека AI-промптов |
| **Short name** | ПромтЛаб |
| **Version** | 1.0.0 |
| **Category** | Productivity |
| **Language** | Russian (primary), English (planned) |
| **Pricing** | Free (functionality requires promtlabs.ru account, paid tiers available) |
| **Privacy Policy URL** | https://promtlabs.ru/legal/extension-privacy |
| **Homepage URL** | https://promtlabs.ru |
| **Support email** | support@promtlabs.ru |

---

## 2. Описание для CWS / AMO

### 2.1. Short description (RU, ≤132 символа)

> Менеджер AI-промптов. Вставляйте библиотеку в ChatGPT, Claude, Gemini, Yandex GPT, GigaChat и другие — одним кликом.

### 2.2. Detailed description (RU, ~3000 символов)

```
ПромтЛаб — это полноценная библиотека AI-промптов прямо в боковой панели браузера. Создавайте, организуйте и вставляйте свои промпты в любой популярный AI-сервис без копи-паста.

🚀 ПОДДЕРЖИВАЕМЫЕ AI-СЕРВИСЫ
• ChatGPT (chatgpt.com)
• Claude (claude.ai)
• Gemini (gemini.google.com)
• Perplexity (perplexity.ai)
• Yandex GPT (Алиса, ya.ru, yandex.ru/alice)
• GigaChat (giga.chat, developers.sber.ru)
• DeepSeek Chat
• Mistral Le Chat
• Qwen Chat

⚡ ОСНОВНЫЕ ВОЗМОЖНОСТИ
• Библиотека промптов с тегами, коллекциями и поиском
• Шаблоны с переменными {{var}} — заполняйте перед вставкой
• Вставка в активную вкладку одним кликом или Cmd+K
• Broadcast — вставить во все открытые AI-вкладки сразу
• Undo последней вставки (5 секунд)
• Quick-save: выделите текст на любой странице → ПКМ → «Сохранить как промпт»
• Версионирование промптов с diff и восстановлением
• Цепочки промптов (Chains) — многошаговые workflow

👥 КОМАНДНАЯ РАБОТА
• Создавайте команды с ролями owner/editor/viewer
• Общие промпты, коллекции, теги
• Брендинг команды (логотип, цвет)
• Лента активности и аналитика

🔒 БЕЗОПАСНОСТЬ И ПРИВАТНОСТЬ
• Self-hosted в России (promtlabs.ru)
• Никаких западных SaaS-зависимостей
• Содержимое AI-чатов НЕ передаётся на сервер
• OAuth через GitHub / Google / Yandex
• HTTPS, JWT, опциональная 2FA

💰 ТАРИФЫ
• Free — 15 промптов, 3 коллекции, базовые функции
• Pro 599₽/мес — 500 промптов, цепочки, аналитика
• Max 1299₽/мес — безлимитные промпты, Smart Insights, team-features

🎯 ТРЕБОВАНИЯ
Расширение требует аккаунт на promtlabs.ru. Регистрация бесплатна и доступна прямо из расширения.

🛡️ ПРАВА
• sidePanel / sidebarAction — открытие панели
• storage — локальный кэш и API-ключ
• activeTab — определение AI-вкладки
• scripting — обновление content-scripts
• contextMenus — меню «Сохранить как промпт»

Доступ только к 13 явно указанным доменам — никаких <all_urls>. Подробности в Политике конфиденциальности.

🐛 ПОДДЕРЖКА
support@promtlabs.ru
```

### 2.3. Short description (EN, ≤132 chars, future)

> AI prompt manager. Insert your library into ChatGPT, Claude, Gemini, Yandex GPT, GigaChat and more — with one click.

---

## 3. Категории и теги

**Primary category:** Productivity
**Secondary tags:** AI, ChatGPT, Claude, Gemini, productivity, prompt engineering, library, workflow, chains, Russian

---

## 4. Скриншоты (5 штук, 1280×800)

Подготовить следующие скриншоты (порядок = importance в CWS gallery):

1. **Dashboard на ChatGPT** — открытая sidepanel рядом с chatgpt.com, виден список промптов + Quota indicator + Notifications bell.
2. **Insert in action** — выбран промпт с {{var}}, форма заполнения переменных, кнопка «Вставить».
3. **Chains run-wizard** — текущий шаг цепочки с отрендеренным промптом, кнопками «Скопировать» и «Далее».
4. **Team workspace** — workspace switcher, командные промпты, бейдж роли «Редактор».
5. **Settings: Integrations** — API-keys management с подсветкой «текущий ключ» + MCP setup hint.

**Технические требования:**
- Размер: 1280×800 PNG (CWS) / 1200×1000 PNG (AMO)
- DPI: 72 (стандарт)
- Без водяных знаков и логотипов сторонних брендов в крупном плане
- Скриншоты должны соответствовать реальной функциональности (не моки)

**Инструмент для генерации:** Chrome DevTools → Device Toolbar → Custom 1280×800 viewport → Cmd+Shift+P → "Capture screenshot".

---

## 5. Promo tile (440×280)

Минимальный дизайн:
- Фон: brand-gradient (фиолетовый primary)
- Иконка ПромтЛаб 128×128 слева
- Текст справа: «ПромтЛаб» (24px bold) + «AI-промпты для 9 чат-ботов» (14px)
- Без сложной графики (CWS rejection risk)

**Generator template:** Figma frame 440×280, экспорт PNG @1x.

---

## 6. Иконки

- 128×128, 48×48, 32×32, 16×16 PNG (уже есть в `public/icon/`)
- Стиль: simple, recognizable at small sizes
- Не использовать тонкие линии или мелкий текст
- Фон: прозрачный или solid color (не gradient — портит recognition в favicon-zone)

---

## 7. Permissions justification (CWS form)

CWS требует пояснение для каждого permission:

| Permission | Justification |
|------------|---------------|
| `sidePanel` | Открытие боковой панели как основного UI расширения. |
| `storage` | Сохранение API-ключа, темы, локального кэша промптов для офлайн-доступа. |
| `activeTab` | Определение активной AI-вкладки для проверки совместимости перед вставкой промпта. |
| `scripting` | Re-injection content-scripts при обновлении расширения (необходимо для MV3 совместимости). |
| `contextMenus` | Меню «Сохранить выделение как промпт» для быстрого захвата текста с любой страницы. |
| `host_permissions` (13 доменов) | Вставка промптов в AI-чаты и синхронизация с serverом promtlabs.ru. Все 13 доменов явно перечислены, никакого `<all_urls>`. |

---

## 8. Single purpose statement (CWS требование)

> **Single purpose:** Расширение «ПромтЛаб» имеет единственное назначение — управление личной библиотекой AI-промптов и их вставка в окна чат-ботов поддерживаемых AI-сервисов одним кликом.

---

## 9. Pre-submission checklist

- [ ] Сборка через `npm run zip` (создаст `.output/promptvault-extension-1.0.0-chrome.zip`)
- [ ] Manifest версия совпадает с `package.json` (1.0.0) и `wxt.config.ts`
- [ ] Privacy policy опубликована на `https://promtlabs.ru/legal/extension-privacy`
- [ ] Скриншоты (5 шт) загружены в `extension/store-screenshots/`
- [ ] Promo tile (440×280) создан
- [ ] Описание ru/en проверено на typos
- [ ] Tested на свежем профиле Chrome (clean state) — basic flow работает
- [ ] **B-18 закрыто** — селекторы для 5 новых LLM проверены на живых сайтах
- [ ] Отдельный публичный YouTube-видео-демо (опционально, повышает conversion)

---

## 10. CWS submission steps (Developer Dashboard)

1. https://chrome.google.com/webstore/devconsole/ → New item
2. Upload `.output/promptvault-extension-1.0.0-chrome.zip`
3. Заполнить Store listing (тексты из секции 2 выше)
4. Загрузить скриншоты + promo tile + иконки
5. Privacy practices → честно проставить data usage:
   - **Personally identifiable info** ✓ (email, password при login)
   - **Authentication info** ✓ (API key, JWT)
   - **Web history** ✗ (мы не собираем browsing history)
   - **User content** ✓ (промпты, которые юзер создаёт)
   - **Location** ✗
   - **Health** ✗
   - **Financial** ✗ (платежи на promtlabs.ru, не в расширении)
6. Sumbit for review (≈1-3 рабочих дня)

---

## 11. AMO submission (Firefox)

1. `npm run zip:firefox` → `.output/promptvault-extension-1.0.0-firefox.zip`
2. https://addons.mozilla.org/developers/ → Submit New Add-on
3. Listed → On this site
4. Заполнить:
   - Name, summary (краткое описание, английское)
   - Categories: Productivity, Web Development
   - Tags: ai, chatgpt, claude, prompts, productivity
   - License: Mozilla Public License 2.0 (рекомендуется) или All Rights Reserved
   - Privacy policy URL
5. Submit for review (1-5 рабочих дней)

**AMO специфика:**
- Sources code review более строгий чем CWS — на минифицированном коде reviewer может попросить sources
- WXT уже выдаёт unminified в dev mode — для AMO submission подойдёт `wxt build --mode production` с source maps

---

## 12. Post-submission

- Отслеживать статус через email от Google/Mozilla
- При rejection — читать reasons, исправлять, resubmit
- После approval — расширение появится в каталоге через 1-2 часа
- Promotional URL: `https://chromewebstore.google.com/detail/<extension-id>` — добавить в landing на promtlabs.ru
