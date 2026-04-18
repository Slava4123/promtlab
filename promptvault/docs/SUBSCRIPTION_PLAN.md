# План реализации системы подписок и тарифов PromtLabs

> Дата создания: 2026-04-13
> Обновлено: 2026-04-14 — **Статус: реализовано (sandbox), тесты зелёные**

## Статус реализации (апрель 2026)

✅ **Готово:**
- Миграции 000019-000023 (plans/subscriptions/payments/daily_feature_usage/users.plan_id)
- `usecases/subscription` + `usecases/quota` + `infrastructure/payment/tbank` + `delivery/http/subscription` + `delivery/http/webhook`
- Frontend: `/pricing`, `SubscriptionSection`, `QuotaExceededDialog`, `UsageMeters`, `useCheckout`/`useRefreshSubscription` (polling 2 мин)
- Hardening: двухфазный Payment (orphan-защита), plan_id в ProviderData (не по сумме), conditional UPDATE против race, `errors.Is` вместо string compare, timeout 30s на HTTP client, исключение Receipt/DATA из подписи webhook, fail-fast PaymentConfig валидация
- Unit-тесты: `tbank.VerifyWebhookSignature` (8 cases), `rawToSigValue` (12), `extractPlanID` (7), idempotency key

⬜ **Остаётся:**
- Production-терминал T-Bank (ждём одобрение на коммерческий терминал)
- IP allowlist T-Bank в middleware (когда диапазоны будут в доках)
- Автопродление (v2) — `SubStatusPastDue` зарезервирован
- Фискализация чеков 54-ФЗ (Receipt в Init) — если нужна для физлиц РФ
- Публичная оферта + условия возврата
- Landing pricing секция (маркетинг)

## Контекст

PromtLabs — self-hosted SaaS для управления AI-промптами (Go + React). Продукт feature-complete (12 из 13 фаз), но вся монетизация отсутствует: нет подписок, нет квот, нет оплаты. Pricing page — статический мокап с заглушками "Скоро". Все фичи доступны всем пользователям без ограничений.

**Цель:** реализовать систему тарифов (Free/Pro/Max) с enforcement квот, интеграцией T-Bank для приёма платежей и UI для управления подпиской. Это Фаза 13 — последняя перед запуском.

**Ключевые решения (подтверждены):**
- AI-лимиты: Free 5 ВСЕГО / Pro 10/день / Max 30/день
- Платёжный провайдер: T-Bank (заявка подана)
- API-доступ: для всех тарифов, ограничения через общие квоты
- Только месячная подписка для v1

---

## 1. Резюме

Строим полную систему подписок: модель данных (5 миграций), quota enforcement в usecase-слое (без нового middleware), интеграцию T-Bank для рекуррентных платежей, фронтенд с живой pricing page и upgrade-потоком.

**Ключевые технические решения:**
1. Квоты проверяются в usecase-слое (не middleware) — консистентно с существующим паттерном `ai.CheckRateLimit`
2. Планы хранятся в БД + кэшируются в памяти (обновление без редеплоя)
3. Денормализованное поле `users.plan_id` для O(1) доступа к текущему тарифу
4. `daily_feature_usage` таблица для персистентного трекинга дневных квот AI/Extension/MCP (отдельно от in-memory RPM limiter)
5. PaymentProvider интерфейс — начинаем с T-Bank, абстракция позволяет добавить ЮKassa позже

---

## 2. Финальная тарифная сетка

### 2.1 Лимиты по тарифам (различаются)

| Ресурс | Free | Pro (599₽/мес) | Max (1299₽/мес) |
|---|---|---|---|
| Промпты | 50 | 500 | Безлимит |
| Коллекции | 3 | Безлимит | Безлимит |
| AI-запросы | 5 ВСЕГО | 10/день (300/мес) | 30/день (900/мес) |
| Команды | 1 | 5 | Безлимит |
| Участники/команда | 3 | 10 | Безлимит |
| Шаринг (активные ссылки) | 2 | 10 | Безлимит |
| Вставки через расширение | 5/день | 30/день | Безлимит |
| Платные MCP-вызовы (13 из 30 tools: write/destructive) | 5/день | 30/день | Безлимит |
| Приоритетная поддержка | Нет | Да | Да |

### 2.2 Функции доступные на всех тарифах

| Функция | Free | Pro | Max |
|---|---|---|---|
| Версионирование промптов | + | + | + |
| Теги (создание, фильтрация) | + | + | + |
| Шаринг ссылкой (публичный доступ) | + | + | + |
| Поиск + автодополнение | + | + | + |
| Избранное, закрепление | + | + | + |
| Корзина (автоочистка 30 дней) | + | + | + |
| Бейджи и достижения (11 шт) | + | + | + |
| Стрики (серии дней) | + | + | + |
| Онбординг (34 стартовых шаблона) | + | + | + |
| OAuth (GitHub, Google, Яндекс) | + | + | + |
| API-ключи | + | + | + |
| MCP-сервер (30 инструментов, 13 из них едят дневную квоту) | + | + | + |
| Chrome-расширение | + | + | + |
| AI-ассистент (Enhance, Rewrite, Analyze, Variations) | + | + | + |
| SSE-стриминг ответов | + | + | + |
| Темная/светлая тема | + | + | + |
| Админ-панель (для администраторов) | + | + | + |

### 2.3 Экономика (при avg стоимости AI-запроса ~1.1₽)

- Pro 599₽: worst case (100% квоты) → 330₽ расход → **маржа 45%**
- Pro 599₽: средний юзер (40% квоты) → 132₽ расход → **маржа 78%**
- Max 1299₽: worst case → 990₽ расход → **маржа 24%**
- Max 1299₽: средний юзер → 396₽ расход → **маржа 70%**
- **Средняя маржа по базе: ~55-65%**

### 2.4 Изменения относительно текущего pricing.tsx

- AI-лимиты: 5/день→5 ВСЕГО (Free), 100/день→10/день (Pro), безлимит→30/день (Max)
- Убрать "API-доступ (скоро)" из Max
- Убрать "Оплата через ЮKassa" → "Оплата через T-Bank"
- Добавить полный список общих функций (версии, теги, шаринг, бейджи, MCP и т.д.)
- Добавить лимиты на шаринг, extension, MCP
- Оставить "Приоритетная поддержка" для Pro/Max (реализация позже)

---

## 3. Архитектурные решения

### 3.1 Quota enforcement — usecase layer (не middleware)

**Решение:** проверка квот в usecase-слое через `QuotaService`
**Альтернативы:** (a) middleware перед каждым handler, (b) в handler перед вызовом usecase
**Trade-offs:**
- Middleware: чистое cross-cutting, но требует DB-запрос на КАЖДЫЙ request + не знает о ресурсо-специфичной логике
- Handler: дублирование кода в каждом handler
- **Usecase (выбрано):** консистентно с `ai.CheckRateLimit`, тестируемо через моки, гибко (разные проверки для разных ресурсов)

```
prompt.Create()     → quotaSvc.CheckPromptQuota(userID)      → ...create...
ai.Enhance()        → quotaSvc.CheckAIQuota(userID)          → ...stream...
team.Create()       → quotaSvc.CheckTeamQuota(userID)        → ...create...
share.Create()      → quotaSvc.CheckShareLinkQuota(userID)   → ...create link...
extension (handler) → quotaSvc.CheckExtensionQuota(userID)   → ...return prompt...
mcp (tool call)     → quotaSvc.CheckMCPQuota(userID)         → ...execute tool...
```

### 3.2 Планы — DB-driven + in-memory cache

**Решение:** таблица `subscription_plans` с кэшированием (TTL 5 мин)
**Альтернативы:** (a) Go-константы, (b) .env конфиг
**Trade-offs:**
- Константы: zero-latency, но редеплой для изменения цен
- .env: тоже рестарт
- **DB + кэш (выбрано):** обновление цен без редеплоя, админ сможет менять лимиты; кэш устраняет проблему лишних запросов

### 3.3 Daily usage — persistent DB counter

**Решение:** таблица `daily_feature_usage` с `INSERT ON CONFLICT DO UPDATE SET count = count + 1`
**Альтернативы:** (a) in-memory counter (как текущий RPM limiter), (b) Redis
**Trade-offs:**
- In-memory: теряется при рестарте → юзер получает "бесплатные" запросы
- Redis: оверхед для текущего масштаба, ещё одна зависимость
- **DB (выбрано):** персистентно, атомарно через UPSERT, одна строка/юзер/день/feature_type

### 3.4 Денормализация `users.plan_id`

**Решение:** поле `plan_id` в таблице `users` (FK → subscription_plans)
**Зачем:** O(1) доступ к тарифу без JOIN через subscriptions. Обновляется транзакционно вместе с subscription.

---

## 4. Модель данных

### 4.1 Миграция 000019: subscription_plans

```sql
-- 000019_subscription_plans.up.sql
CREATE TABLE IF NOT EXISTS subscription_plans (
    id                      VARCHAR(20) PRIMARY KEY,
    name                    VARCHAR(50) NOT NULL,
    price_kop               INTEGER NOT NULL DEFAULT 0,
    period_days             INTEGER NOT NULL DEFAULT 30,
    max_prompts             INTEGER NOT NULL DEFAULT 50,
    max_collections         INTEGER NOT NULL DEFAULT 3,
    max_ai_requests_daily   INTEGER NOT NULL DEFAULT 5,
    ai_requests_is_total    BOOLEAN NOT NULL DEFAULT FALSE,
    max_teams               INTEGER NOT NULL DEFAULT 1,
    max_team_members        INTEGER NOT NULL DEFAULT 3,
    max_share_links         INTEGER NOT NULL DEFAULT 2,
    max_ext_uses_daily      INTEGER NOT NULL DEFAULT 5,
    max_mcp_uses_daily      INTEGER NOT NULL DEFAULT 5,
    features                JSONB NOT NULL DEFAULT '[]',
    sort_order              INTEGER NOT NULL DEFAULT 0,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO subscription_plans (id, name, price_kop, period_days,
    max_prompts, max_collections, max_ai_requests_daily, ai_requests_is_total,
    max_teams, max_team_members, max_share_links, max_ext_uses_daily,
    max_mcp_uses_daily, features, sort_order)
VALUES
    ('free', 'Free', 0, 0,
     50, 3, 5, TRUE,
     1, 3, 2, 5, 5,
     '[]'::jsonb, 0),
    ('pro', 'Pro', 59900, 30,
     500, -1, 10, FALSE,
     5, 10, 10, 30, 30,
     '["priority_support"]'::jsonb, 1),
    ('max', 'Max', 129900, 30,
     -1, -1, 30, FALSE,
     -1, -1, -1, -1, -1,
     '["priority_support"]'::jsonb, 2);
```

Конвенция: `-1` = безлимит. `ai_requests_is_total = TRUE` для Free (5 навсегда, не в день).

### 4.2 Миграция 000020: subscriptions

```sql
CREATE TABLE IF NOT EXISTS subscriptions (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              BIGINT NOT NULL REFERENCES users(id),
    plan_id              VARCHAR(20) NOT NULL REFERENCES subscription_plans(id),
    status               VARCHAR(20) NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','past_due','cancelled','expired')),
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end   TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    cancelled_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_subscriptions_user_active
    ON subscriptions (user_id) WHERE status IN ('active','past_due');
CREATE INDEX idx_subscriptions_expiring
    ON subscriptions (status, current_period_end);
```

Partial unique index гарантирует max 1 активную подписку на юзера.

### 4.3 Миграция 000021: payments

```sql
CREATE TABLE IF NOT EXISTS payments (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    subscription_id BIGINT REFERENCES subscriptions(id),
    external_id     VARCHAR(100) NOT NULL,
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    amount_kop      INTEGER NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','succeeded','failed','refunded')),
    provider        VARCHAR(20) NOT NULL DEFAULT 'tbank',
    provider_data   JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_payments_external ON payments (provider, external_id);
CREATE INDEX idx_payments_user ON payments (user_id, created_at DESC);
```

### 4.4 Миграция 000022: daily_feature_usage

Универсальная таблица для дневных лимитов AI, Extension, MCP (вместо трёх отдельных таблиц):

```sql
CREATE TABLE IF NOT EXISTS daily_feature_usage (
    user_id      BIGINT NOT NULL REFERENCES users(id),
    usage_date   DATE NOT NULL,
    feature_type VARCHAR(20) NOT NULL,
    count        INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date, feature_type)
);
```

`feature_type` — одно из: `'ai'`, `'extension'`, `'mcp'`. Новые типы можно добавить без миграций.

**Идентификация источника:**
- AI: запросы к `/api/ai/*` endpoints
- Extension: заголовок `X-Client-Source: extension` (добавить в extension при отправке запросов)
- MCP: запросы через `/mcp` endpoint — трекинг внутри `mcpserver/` только для 13 платных tools (write/destructive); read-only и UX-toggle (favorite/pin/increment_usage) не инкрементят счётчик

### 4.5 Миграция 000023: users.plan_id

```sql
ALTER TABLE users ADD COLUMN plan_id VARCHAR(20) NOT NULL DEFAULT 'free'
    REFERENCES subscription_plans(id);
CREATE INDEX idx_users_plan ON users (plan_id);
```

---

## 5. API контракт

### 5.1 Public — планы

| Method | Path | Описание |
|---|---|---|
| `GET` | `/api/plans` | Список активных планов (кэшируемый, без auth) |

### 5.2 Protected — подписка

| Method | Path | Описание |
|---|---|---|
| `GET` | `/api/subscription` | Текущая подписка + план |
| `GET` | `/api/subscription/usage` | Использование vs лимиты |
| `POST` | `/api/subscription/checkout` | Инициировать оплату (→ payment_url) |
| `POST` | `/api/subscription/cancel` | Отменить подписку (в конце периода) |

**POST /api/subscription/checkout — Request:**
```json
{ "plan_id": "pro" }
```

**GET /api/subscription/usage — Response 200:**
```json
{
  "plan_id": "free",
  "prompts":        { "used": 12, "limit": 50 },
  "collections":    { "used": 2,  "limit": 3 },
  "ai_requests":    { "used": 3,  "limit": 5, "is_total": true },
  "teams":          { "used": 1,  "limit": 1 },
  "share_links":    { "used": 1,  "limit": 2 },
  "ext_uses_today": { "used": 2,  "limit": 5 },
  "mcp_uses_today": { "used": 0,  "limit": 5 }
}
```

### 5.3 Webhook

| Method | Path | Описание |
|---|---|---|
| `POST` | `/api/webhooks/tbank` | Уведомление об оплате от T-Bank |

### 5.4 Коды ошибок

| HTTP | Когда | Тело |
|---|---|---|
| `402` | Квота исчерпана | `{"error":"...","quota_type":"ai_daily","used":10,"limit":10,"plan":"pro","upgrade_url":"/pricing"}` |
| `409` | Уже есть активная подписка | `{"error":"Активная подписка уже существует"}` |
| `501` | T-Bank не настроен | `{"error":"Платежи временно недоступны"}` |

---

## 6. Изменения в коде

### 6.1 Новые файлы (backend)

| Путь | Назначение |
|---|---|
| `models/subscription.go` | GORM-модели: SubscriptionPlan, Subscription, Payment, DailyFeatureUsage |
| `interface/repository/subscription.go` | Интерфейсы: PlanRepo, SubscriptionRepo, PaymentRepo, QuotaRepo |
| `infrastructure/postgres/repository/plan_repo.go` | GORM + in-memory кэш (TTL 5m) |
| `infrastructure/postgres/repository/subscription_repo.go` | GORM: CRUD подписок |
| `infrastructure/postgres/repository/payment_repo.go` | GORM: CRUD платежей |
| `infrastructure/postgres/repository/quota_repo.go` | GORM: count-запросы для квот |
| `infrastructure/payment/tbank/provider.go` | T-Bank HTTP API клиент |
| `infrastructure/payment/types.go` | PaymentProvider интерфейс |
| `infrastructure/config/payment.go` | PaymentConfig struct |
| `usecases/quota/quota.go` | QuotaService: Check*()/Increment*() |
| `usecases/quota/types.go` | UsageSummary, QuotaInfo |
| `usecases/quota/errors.go` | ErrPromptQuotaExceeded и т.д. |
| `usecases/subscription/subscription.go` | SubscriptionService: checkout/cancel/webhook |
| `usecases/subscription/types.go` | Input/Output типы |
| `usecases/subscription/errors.go` | Доменные ошибки подписки |
| `delivery/http/subscription/handler.go` | HTTP handlers |
| `delivery/http/subscription/request.go` | Request DTOs |
| `delivery/http/subscription/response.go` | Response DTOs |
| `delivery/http/subscription/errors.go` | Маппинг ошибок → HTTP |
| `delivery/http/webhook/handler.go` | Webhook handler (T-Bank) |
| `migrations/000019-000023` | 5 пар .up.sql/.down.sql |

### 6.2 Модифицируемые файлы (backend)

| Путь | Что меняется |
|---|---|
| `models/user.go` | Добавить `PlanID string` поле |
| `infrastructure/config/config.go` | Добавить `Payment PaymentConfig` |
| `infrastructure/config/loader.go` | Defaults для payment |
| `app/app.go` | Wiring новых repos/services/handlers, mount routes |
| `usecases/prompt/prompt.go` | `Create()` → вызов `quotaSvc.CheckPromptQuota()` |
| `usecases/ai/service.go` | Добавить `quotaSvc`, вызов `CheckAIQuota()` + `IncrementAIUsage()` |
| `usecases/team/team.go` | `Create()` → `CheckTeamQuota()`, `InviteMember()` → `CheckTeamMemberQuota()` |
| `usecases/admin/admin.go` | `ChangeTier()` — реализовать вместо stub |
| `delivery/http/admin/response.go` | `Tier: user.PlanID` вместо хардкода "free" |
| `delivery/http/prompt/errors.go` | Маппинг quota errors → 402 |
| `delivery/http/ai/errors.go` | Маппинг quota errors → 402 |
| `delivery/http/team/errors.go` | Маппинг quota errors → 402 |
| `delivery/http/share/handler.go` | Quota check при создании share link |
| `mcpserver/` | Трекинг tool calls + quota check MCP |
| `delivery/http/prompt/handler.go` | Трекинг extension usage (по X-Client-Source) |

### 6.3 Новые файлы (frontend)

| Путь | Назначение |
|---|---|
| `api/subscription.ts` | API-функции (getPlans, getSubscription, checkout, cancel, getUsage) |
| `hooks/use-subscription.ts` | TanStack Query хуки |
| `components/subscription/quota-exceeded-dialog.tsx` | Диалог "Лимит исчерпан → Обновить план" |
| `components/subscription/plan-badge.tsx` | Бейдж тарифа (sidebar) |
| `components/subscription/usage-meters.tsx` | Визуальные метры использования |

### 6.4 Модифицируемые файлы (frontend)

| Путь | Что меняется |
|---|---|
| `api/types.ts` | Типы Plan, SubscriptionInfo, UsageSummary, PlanID |
| `pages/pricing.tsx` | API-driven данные, рабочие кнопки оплаты, обновить лимиты |
| `stores/auth-store.ts` | `plan_id` в User type |
| `components/layout/app-sidebar.tsx` | Plan badge |
| `pages/settings.tsx` | Секция управления подпиской |

---

## 7. T-Bank интеграция

### 7.1 Flow оплаты

```
1. Юзер нажимает "Перейти на Pro" → POST /api/subscription/checkout
2. Backend вызывает T-Bank API Init (создаёт платёж)
3. T-Bank возвращает PaymentURL
4. Backend сохраняет Payment (status=pending), возвращает PaymentURL фронту
5. Фронт редиректит юзера на T-Bank форму оплаты
6. Юзер оплачивает → T-Bank вызывает webhook POST /api/webhooks/tbank
7. Backend верифицирует подпись, обновляет Payment→succeeded
8. В транзакции: создаёт/обновляет Subscription, обновляет users.plan_id
9. T-Bank редиректит юзера на success_url (/settings?payment=success)
```

### 7.2 T-Bank API endpoints (v2)

- `POST /Init` — создать платёж (Amount, OrderId, Description, SuccessURL, FailURL, NotificationURL)
- Webhook: T-Bank POST на NotificationURL с параметрами (OrderId, Status, Amount) + Token (SHA-256 HMAC)

### 7.3 Верификация webhook

T-Bank формирует `Token` = SHA-256 от конкатенации всех параметров + TerminalPassword в алфавитном порядке. Backend пересчитывает и сравнивает. Дополнительно: IP allowlist T-Bank webhook серверов.

### 7.4 Рекуррентные платежи

Для автопродления T-Bank поддерживает `Recurrent=Y` при Init и `Charge` для списания по `RebillId`. Для v1: уведомлять юзера за 3 дня до истечения, ручное продление. Автопродление — в v2.

---

## 8. План тестирования

### 8.1 Unit тесты

- `usecases/quota/` — CheckPromptQuota, CheckAIQuota и т.д. с моками репозиториев
- `usecases/subscription/` — checkout flow, cancel, webhook handling, expiration
- `infrastructure/payment/tbank/` — Init request, webhook signature verification
- `infrastructure/postgres/repository/quota_repo.go` — count queries (integration с testcontainers)

### 8.2 Integration тесты

- Полный flow: checkout → webhook → subscription activated → quota enforcement
- Expiration: subscription expires → plan downgraded to free → limits enforced
- Idempotency: duplicate webhook → no double-activation

### 8.3 E2E (ручные)

- Создать подписку Pro через T-Bank sandbox
- Убедиться что лимиты Free работают (создать 51-й промпт → 402)
- AI-квота: исчерпать лимит → получить upgrade dialog
- Отменить подписку → доступ до конца периода → downgrade

---

## 9. План внедрения (пошагово)

### Шаг 1: Модель данных и миграции (~100 LOC SQL)
- Создать 5 миграций (000019–000023)
- Добавить GORM-модели в `models/subscription.go`
- Добавить `PlanID` в `models/user.go`
- **Критерий:** `go test -short ./...` зелёный, миграции применяются при старте

### Шаг 2: Repository + QuotaService (~300 LOC Go)
- Реализовать `PlanRepo` (с кэшем), `QuotaRepo`, `SubscriptionRepo`, `PaymentRepo`
- Реализовать `usecases/quota/` с Check* методами
- Покрыть unit-тестами
- **Критерий:** тесты зелёные, QuotaService корректно считает лимиты

### Шаг 3: Интеграция квот в существующие usecases (~150 LOC Go)
- Добавить `quotaSvc` в prompt, ai, team, share usecases
- Добавить трекинг extension usage в prompt handler (по заголовку `X-Client-Source: extension`)
- Добавить трекинг MCP usage в `mcpserver/` (при каждом платном tool call — 13 типов write/destructive; read-only tools и UX-toggle не тарифицируются)
- Добавить quota error mapping в delivery/http errors
- **Критерий:** Free-юзер не может создать 51-й промпт (402), AI-квота, share, ext/MCP лимиты работают

### Шаг 4: T-Bank интеграция (~250 LOC Go)
- `infrastructure/payment/tbank/provider.go` — T-Bank API клиент
- `infrastructure/config/payment.go` — PaymentConfig
- `usecases/subscription/` — checkout, cancel, webhook handling
- `delivery/http/subscription/` и `delivery/http/webhook/`
- Wiring в `app/app.go`
- **Критерий:** checkout → T-Bank sandbox → webhook → subscription active

### Шаг 5: Admin tier change (~30 LOC Go)
- Реализовать `admin.ChangeTier()` вместо stub
- Обновить `admin/response.go` — `Tier: user.PlanID`
- **Критерий:** admin может менять тариф юзеру, аудит-лог пишется

### Шаг 6: Frontend — pricing page + checkout (~200 LOC TSX)
- Обновить `pricing.tsx`: API-driven данные, рабочие кнопки
- Добавить `api/subscription.ts` и `hooks/use-subscription.ts`
- Checkout flow: клик → redirect на T-Bank → callback
- **Критерий:** можно купить подписку через UI

### Шаг 7: Frontend — upgrade prompts + quota dialog (~150 LOC TSX)
- `quota-exceeded-dialog.tsx` — при 402 показывать upgrade
- Plan badge в sidebar
- Секция подписки в settings
- Usage meters
- **Критерий:** при исчерпании лимита юзер видит предложение upgrade

### Шаг 8: Background jobs + cleanup (~50 LOC Go)
- Cron/goroutine для expiration check (каждый час)
- Cleanup `daily_feature_usage` старше 90 дней
- Email-уведомление за 3 дня до истечения
- **Критерий:** expired подписки автоматически даунгрейдятся

---

## 10. Риски и митигации

| Риск | Вероятность | Митигация |
|---|---|---|
| T-Bank отклонит заявку | Средняя | Подготовить PaymentProvider интерфейс, быстро переключить на ЮKassa |
| Power user исчерпает маржу Max | Низкая | 30 req/день — потолок. Worst case 24% маржа, средний юзер +70% |
| Race condition при concurrent AI requests | Низкая | UPSERT атомарен, RPM limiter как первая линия |
| Webhook replay / дубли | Средняя | Idempotency key в payments, unique index |
| Юзер платит но webhook не доходит | Низкая | Success URL показывает "проверяем оплату", polling GET /subscription |

---

## 11. Метрики успеха

**Бизнес:**
- Конверсия Free → Pro (целевая: 5-10%)
- MRR (monthly recurring revenue)
- Churn rate (целевая: <10%/мес)

**Технические:**
- Webhook processing latency < 500ms
- Quota check latency < 10ms (с кэшем планов)
- Zero 5xx на payment endpoints

---

## 12. Открытые вопросы (решить ДО старта)

1. **T-Bank одобрен?** — если нет, нужно подавать на ЮKassa параллельно
2. **Коллекции для Free — 3 штуки:** сейчас юзеры могли создать больше. Рекомендация: не удалять существующие, но запретить создание новых сверх лимита (grandfather clause)
3. **Рекуррентные платежи:** T-Bank Charge API для автопродления в v1 или ручное продление?
4. **Trial period:** давать 7 дней Pro бесплатно новым юзерам или нет?
