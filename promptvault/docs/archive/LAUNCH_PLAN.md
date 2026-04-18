# План запуска ПромтЛаб

## Контекст

Продукт feature-complete: CRUD промптов, команды, AI (Claude Sonnet 4), MCP (24 tools), Chrome extension, версионирование, шаринг, бейджи, стрики, онбординг. Домен `promtlabs.ru` настроен на VPS (Timeweb Cloud). Заявка на T-Bank для оплаты подана.

---

## Фаза 1: Локальное тестирование и polish

### 1.1 Smoke-тест всех фич
- [x] Регистрация (email + пароль)
- [x] Верификация email (код приходит)
- [x] Логин / логаут
- [ ] OAuth (GitHub, Google, Yandex) — нужны реальные redirect URI
- [x] Забыл пароль → сброс через email
- [x] CRUD промптов (создание, редактирование, удаление, восстановление из корзины)
- [x] Версионирование (update → get_versions → revert)
- [x] Коллекции + теги (создание, присвоение, фильтрация)
- [x] Поиск + автодополнение
- [x] Избранное, закрепление, стрики
- [x] AI-панель (Enhance, Rewrite, Analyze, Variations) — SSE работает
- [ ] Template variables `{{переменная}}` — подстановка через форму (нужно руками в UI)
- [x] Команды: создание, приглашение, роли
- [x] Шаринг ссылкой — открытие без авторизации
- [x] Онбординг (34 стартовых шаблона)
- [x] Настройки: профиль, смена пароля, API-ключи
- [x] Корзина: удаление → просмотр → восстановление
- [x] MCP: 24 tools протестированы через live HTTP (E2E)
- [ ] Extension: загрузить в Chrome, вставить промпт в ChatGPT (нужно руками)

### 1.2 Баги и polish
- [ ] Проверить мобильную вёрстку (375px, 768px)
- [ ] Консоль без ошибок на каждой странице
- [ ] Все тексты на русском (нет english placeholder-ов)
- [ ] Dark mode не ломает читаемость
- [x] Landing page: полный редизайн (9 секций, dark mode, glassmorphism, анимации)

### 1.3 Legal
- [x] Пользовательское соглашение (`/legal/terms`)
- [x] Политика конфиденциальности (`/legal/privacy`)
- [x] Политика конфиденциальности расширения (`/legal/extension-privacy`)
- [x] Ссылки на legal в footer лендинга, sidebar, sign-up

### 1.4 Production-конфиг
- [ ] `.env.prod` на VPS — все секреты заполнены (JWT_SECRET не dev)
- [ ] CORS origins = `https://promtlabs.ru`
- [ ] SMTP настроен на production email
- [ ] Sentry DSN (GlitchTip) для production
- [ ] MCP_ENABLED=true

---

## Фаза 2: Деплой на production

### 2.1 Коммит + push
- [x] Закоммитить MCP 24 tools + legal pages + landing (18c8e4b)
- [ ] Push в main → CI/CD pipeline запускается

### 2.2 CI/CD pipeline
- [ ] lint → зелёный
- [ ] test-backend → зелёный
- [ ] test-frontend → зелёный
- [ ] test-extension → зелёный
- [ ] build-push → Docker images в GHCR
- [ ] deploy → SSH на VPS, docker compose up

### 2.3 Верификация production
- [ ] `https://promtlabs.ru` открывается
- [ ] SSL сертификат валидный (certbot)
- [ ] Регистрация + логин работает
- [ ] Email доходит (не в спам)
- [ ] OAuth callback URL-ы обновлены на провайдерах (GitHub/Google/Yandex)
- [ ] MCP доступен: `https://promtlabs.ru/mcp`
- [ ] GlitchTip ловит ошибки

---

## Фаза 3: Подключение оплаты через T-Bank

### 3.1 Инфраструктура (после одобрения T-Bank)
- [ ] Модель `Subscription` в БД (user_id, plan, status, starts_at, expires_at, payment_id)
- [ ] Миграция: `000019_create_subscriptions.{up,down}.sql`
- [ ] T-Bank HTTP API интеграция (`internal/infrastructure/tbank/`)
- [ ] Webhook handler для уведомлений об оплате (`/api/webhooks/tbank`)

### 3.2 Backend
- [ ] `usecases/subscription/` — Create, Cancel, Check, Webhook
- [ ] `delivery/http/subscription/` — endpoints
- [ ] Middleware/проверка тира в usecases
- [ ] Enforce лимиты: промпты (50/500/unlim), коллекции (3/unlim), AI (5/100/unlim), команды (1/5/unlim)

### 3.3 Frontend
- [ ] Pricing page — реальные кнопки "Купить" вместо "Скоро"
- [ ] Checkout flow: выбор тарифа → редирект на T-Bank → callback → активация
- [ ] Страница управления подпиской (текущий план, дата продления, отмена)
- [ ] Upgrade prompts в UI когда лимит достигнут ("Перейди на Pro")

### 3.4 Тестирование оплаты
- [ ] Тестовый режим T-Bank (sandbox)
- [ ] Успешная оплата → подписка активна
- [ ] Отмена → доступ до конца оплаченного периода
- [ ] Webhook replay — идемпотентность
- [ ] Expired subscription → fallback на Free лимиты

---

## Фаза 4: Запуск и привлечение

### 4.1 SEO / Meta
- [ ] Meta title/description на каждой public-странице
- [ ] Open Graph теги (og:title, og:image) для шаринга
- [ ] robots.txt, sitemap.xml
- [ ] Structured data (JSON-LD) для Google

### 4.2 Публикация расширения
- [ ] Chrome Web Store ($5 dev account, скриншоты, описание) — ревью 1-3 недели
- [ ] Firefox Add-ons (бесплатно) — ревью 1-3 дня

### 4.3 Публикация MCP
- [ ] Official MCP Registry (mcp-publisher CLI)
- [ ] Smithery.ai
- [ ] README с инструкцией подключения

### 4.4 Маркетинг
- [ ] Пост на Habr ("Как я сделал self-hosted менеджер промптов")
- [ ] Telegram-каналы про AI (русскоязычные)
- [ ] ProductHunt launch
- [ ] Twitter/X пост
- [ ] GitHub README с бейджами и скриншотами

---

## Порядок действий

```
Сейчас:
  ├─ Заявка T-Bank (подана)
  └─ Фаза 1: локальное тестирование (в процессе)

После тестирования:
  └─ Фаза 2: деплой на VPS (1 день)

Параллельно с ожиданием T-Bank:
  └─ Фаза 4.1-4.3: SEO, extension в CWS, MCP в реестры

После одобрения T-Bank:
  └─ Фаза 3: оплата (3-5 дней)

После оплаты:
  └─ Фаза 4.4: маркетинг
```

---

## Что уже сделано в этой сессии

1. **MCP-сервер расширен с 12 до 24 tools** — favorite, pin, pinned/recent, revert, share, collection update/get, tag delete, search suggest, increment usage
2. **72 unit теста** для MCP (все зелёные)
3. **E2E тестирование** всех 24 tools через live HTTP
4. **Legal pages** — /legal/terms, /legal/privacy (+ ссылки в footer, sidebar, sign-up)
5. **Лендинг** — полный редизайн (9 секций, dark mode, glow, анимации, pricing)
6. **Email** обновлён на slava0gpt@gmail.com во всех legal pages
7. **Коммит**: `18c8e4b` — 17 файлов, +1756 строк

---

## Верификация готовности к запуску

- [ ] Приложение работает на `https://promtlabs.ru`
- [ ] Регистрация → логин → создание промпта — работает
- [ ] Оплата Pro → лимиты расширяются
- [ ] Extension в Chrome Web Store
- [ ] MCP в реестре
- [ ] Legal pages на месте
- [ ] Мониторинг работает (GlitchTip)
