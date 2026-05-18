# Pricing Iteration v3 — Design Doc

**Дата:** 2026-05-17
**Owner:** Slava Kovalchuk
**Статус:** утверждён, готов к implementation plan

---

## Контекст

**Фича.** Pricing iteration v3 — четыре изменения цен/лимитов без повышения базовых цен Pro/Max:

1. Free `max_prompts`: 15 → **25** (откат Pack E к более мягкому Free, активация-friendly).
2. Annual discount: 10% → **20%** (миграция `pro_yearly` 6 490 → 5 750 ₽, `max_yearly` 13 990 → 12 470 ₽).
3. Pro Smart Insights teaser: **2 типа** из 7 (`unused_prompts` + `possible_duplicates`), остальные 5 остаются Max-only.
4. Referral reward: **+30 дней Pro пригласившему** через delayed trigger (через 14 дней после первого платежа реферри, защита от refund-арбитража).

**Зачем.**

- Free 15 → 25: устранить риск убитой активации; Stanford 80/3-rule остаётся соблюдён.
- Annual −20%: укрепить annual-cohort (LTV × 2), привести к норме РФ-рынка (Битрикс −30%, Аспро −40%).
- Pro Insights: создать reason-to-upgrade Free → Pro и одновременно teaser Pro → Max.
- Referral: запустить виральный loop, инфра в БД готова (миграция 000032), reward не реализован.

**Жёсткие ограничения.**

- Backward compatibility: grandfather через `users.legacy_quotas` сохраняем.
- T-Bank rebillId использует amount из первой транзакции — существующие подписки продолжают платить старую цену до конца периода, новая цена применяется на renewal.
- Принцип «без AI на нашей стороне» сохраняется (Smart Insights — детерминированные SQL).
- Self-hosted: без новых внешних сервисов.

**Аудитория плана.** Исполнитель (Slava — сам имплементит) и PR-ревьюер (он же). Глубина — implementation-ready.

**Отложено в backlog.** Lifetime deal, Edu-тариф 299 ₽, Team-тариф 2 999 ₽, Power-фичи Max (batch ops, API key advanced scopes).

---

## Карта существующего кода

**Слои.** Clean Architecture: `delivery/http/<feature>` → `usecases/<feature>` → `interface/repository` → `infrastructure/postgres/repository`. Domain models — один пакет `internal/models`.

**Эталонные файлы.**

- `backend/internal/usecases/quota/quota.go:73 effectiveLimit(user, field, planValue)` — паттерн `max(legacy, plan)`. Переиспользуем для миграции Free 25 (повышение → grandfather не нужен).
- `backend/internal/usecases/analytics/service.go:111 GetInsightsGated` — master-gate `subscription.IsMax(planID)` через `lookupPlanID:86` (JWT-claims fast-path, M9). Это место разветвляем на per-type filter.
- `backend/internal/usecases/subscription/renewal.go:134 plans.GetByID(ctx, sub.PlanID)` — renewal берёт актуальную цену. **Подтверждает: Annual −20% не требует кода, только миграция.**
- `backend/internal/pkg/safeloop/safeloop.go RunWithRecover` — паттерн background-loops с Prometheus counter паник. Используем для `ReferralRewardLoop`.
- `frontend/src/components/analytics/upgrade-gate.tsx:12 UpgradeGate({title, description, targetPlan})` — готовый компонент для locked-секций.

**Тесты-эталоны.**

- `backend/internal/usecases/quota/quota_test.go:13-135` — fake in-memory repos с трейсом `incrementLog`, ассерты через `errors.As(&qe)`.
- `backend/internal/usecases/subscription/renewal_test.go` — fake `payment.PaymentProvider` + fake repos + fake `RenewalNotifier`.

**Свежий git log.** Subscription/quota/analytics в фазе стабилизации: `4cc52f2 Pack T team-pool quotas`, `aff9448 PlanID в JWT (M9)`, `d3e34b2 silent error swallow fix`. Активного refactor в наших директориях нет. Pricing-блок завершён `f2daa95 Phase 14.2`.

**Конвенции.** slog со структурными атрибутами; доменные ошибки в `usecases/<feature>/errors.go`; миграции `000NNN_description.{up,down}.sql` idempotent (`IF NOT EXISTS`); `promauto.CounterVec` с zero-init label combinations в `init()`; TanStack Query для всех frontend API; sonner для toast'ов.

---

## 1. Резюме

Pricing iteration v3 — четыре атомарные тарифные правки без повышения цен Pro/Max. Реализуются как 3 миграции (Free 25, Annual −20%, новая таблица `referral_pending_rewards`) + код-изменения в двух местах (analytics gate per-type, новый `ReferralRewardLoop`). Существующие grandfather (`legacy_quotas`) и T-Bank renewal pricing работают «бесплатно» благодаря `effectiveLimit` и `plans.GetByID()` на renewal.

**Ключевые технические решения.**

1. **Pro Insights teaser** через константу `proAllowedInsights []string` в `usecases/analytics`, единая точка фильтрации в `ComputeInsights` и `GetInsightsGated`. YAGNI alternative от JSONB-колонки.
2. **Referral reward** через delayed cron-loop (новая таблица + 14-day eligibility check) — защита от refund-арбитража, идемпотентность через UNIQUE constraint.
3. **Reward для Free-юзера** — синтетическая `Subscription{status: 'active', plan_id: 'pro', auto_renew: false, rebill_id: ''}` на 30 дней; auto-downgrade использует существующий `expirationLoop`.

**Аудитория плана.** Исполнитель + PR-ревьюер. Implementation-ready: для каждого шага указан критерий готовности и эталонный паттерн.

---

## 2. Архитектурные решения

### Решение 1: Pro Insights gate — константа в коде vs JSONB-колонка vs feature flag

- **Решение:** Hardcoded slice `proAllowedInsights = []string{models.InsightUnusedPrompts, models.InsightPossibleDuplicates}` в `usecases/analytics`. Гейт переезжает с master `IsMax` (`service.go:111`) на per-type filter. Loop `insights_loop.go` начинает обрабатывать Pro+Max через переименование repo-метода `ListMaxUsers` → `ListPaidUsers`.
- **Альтернативы:**
  - (A) JSONB-колонка `subscription_plans.allowed_insight_types TEXT[]`: гибкость без миграций кода при добавлении 4-го tier'а.
  - (B) Feature flag env `PRO_INSIGHTS_TYPES=unused,duplicates`: можно менять без релиза.
  - (C) Константа в коде **(выбран)**: 1 место правды, легко тестировать.
- **Trade-offs:**
  - ✅ Простой, типобезопасный (нет JSON unmarshalling в hot-path).
  - ✅ Тестируется table-driven case'ами.
  - ❌ Добавление 4-го tier'а потребует кодовых правок. Acceptable — у нас 3 tier'а уже 1.5 года.
- **Источник.** Эталон `analytics.experimentalInsights` boolean kill-switch (Phase 15) — тот же подход с константой в Service struct.

### Решение 2: Referral reward trigger — delayed cron vs synchronous webhook vs lazy

- **Решение:** Delayed cron-loop (паттерн `subscription.RenewalLoop`). Новая таблица `referral_pending_rewards(id, referrer_id, referee_id, payment_id, eligible_at TIMESTAMPTZ, created_at)`. На webhook `payment.succeeded` в `subscription.HandleWebhook` — если у юзера есть `referred_by` и `referrer.referral_rewarded_at IS NULL`, INSERT в `referral_pending_rewards` с `eligible_at = now() + 14 дней`. Loop `ReferralRewardLoop` через `safeloop.RunWithRecover` каждый час: SELECT WHERE `eligible_at < now()`, проверяет что payment всё ещё `succeeded` (не refunded), grant + UPDATE `users.referral_rewarded_at`, DELETE row.
- **Альтернативы:**
  - (A) Synchronous on webhook + revoke on refund: race conditions, реверс сложен.
  - (B) Lazy при следующем webhook (на 2-м платеже того же юзера): зависит от 2-го платежа, может никогда не произойти.
  - (C) Delayed cron-loop **(выбран)**.
- **Trade-offs:**
  - ✅ Защита от refund-арбитража (14d > T-Bank refund window).
  - ✅ Идемпотентность через UNIQUE `(referee_id)`.
  - ✅ Reuse safeloop паттерна.
  - ❌ Новая таблица.
- **Источник.** `subscription/renewal.go:RenewalLoop` (паттерн lookahead + retry); UNIQUE constraint idempotency — `payments.idempotency_key`.

### Решение 3: Reward для Free-юзера — trial Subscription vs legacy_quotas vs skip

- **Решение:** При grant'е, если referrer на Free, создаётся `Subscription{user_id: referrer, plan_id: 'pro', status: 'active', current_period_end: now()+30d, rebill_id: '', auto_renew: false}`. Без T-Bank payment. По истечении 30 дней — existing `expirationLoop` переводит в `expired` + downgrade в Free. Если у юзера уже есть active Pro/Max subscription — продлеваем `current_period_end +30d`.
- **Альтернативы:**
  - (A) Временно поднять `users.legacy_quotas` через колонку `legacy_quotas_expires_at`: нет attribution в БД.
  - (B) Пропускать reward если рефер на Free: убивает виральный loop (90% юзеров — Free).
  - (C) Trial Subscription **(выбран)**.
- **Trade-offs:**
  - ✅ Reuse существующего expiration механизма.
  - ✅ Auth/quota/insights автоматически видят Pro (через `users.plan_id` обновляемый записью).
  - ❌ Поле `rebill_id=""` — на renewal loop может попытаться продлить. Митигация: `auto_renew=false` — `renewal.go:117` пропускает такие записи. [ДОПУЩЕНИЕ: проверить SELECT-фильтр в `ListReadyForRenewal`].
- **Источник.** `models/subscription.go:117 NewSubscription` — конструктор; для referral переопределяем `AutoRenew=false` явно.

---

## 3. Изменения в коде

### Создаём

- `backend/internal/infrastructure/postgres/migrations/000072_free_prompts_25.up.sql` + `.down.sql`.
- `backend/internal/infrastructure/postgres/migrations/000073_annual_discount_20pct.up.sql` + `.down.sql`.
- `backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.up.sql` + `.down.sql`.
- `backend/internal/usecases/referral/reward.go` — `Service.GrantReward`, `ReferralRewardLoop`.
- `backend/internal/usecases/referral/types.go` — `RewardSummary{Granted, Skipped, Errors}`.
- `backend/internal/usecases/referral/errors.go` — доменные ошибки.
- `backend/internal/usecases/referral/reward_test.go`.
- `backend/internal/interface/repository/referral_reward.go` — interface `ReferralRewardRepository{Create, ListEligible, Delete, FindByReferee}`.
- `backend/internal/infrastructure/postgres/repository/referral_reward_repo.go` — GORM реализация.
- `backend/internal/models/referral_pending_reward.go` — GORM модель.
- `docs/adr/0008-pertype-insights-gate.md`.
- `docs/adr/0009-delayed-referral-reward.md`.
- `docs/runbooks/ReferralRewardLoopStalled.md`.

### Меняем

- `backend/internal/usecases/analytics/service.go` — `GetInsightsGated` ветвится: если plan ∈ {pro, pro_yearly} → filter insights по `proAllowedInsights`; если ∈ {max, max_yearly} → все; иначе → `ErrPaidRequired` (переименование `ErrMaxRequired`).
- `backend/internal/usecases/analytics/insights.go:ComputeInsights` — параметризовать `allowedTypes []string`; перебор 7 типов фильтрует через `slices.Contains`.
- `backend/internal/usecases/analytics/insights_loop.go:78` — `ListMaxUsers` → `ListPaidUsers`; передавать актуальный `allowedTypes` для plan'а.
- `backend/internal/interface/repository/user.go` + `infrastructure/postgres/repository/user_repo.go` — переименовать `ListMaxUsers` → `ListPaidUsers` с фильтром `plan_id IN ('pro','pro_yearly','max','max_yearly')`.
- `backend/internal/usecases/subscription/subscription.go:HandleWebhook` — на `payment.succeeded` с `referee.referred_by != ""` и `referrer.referral_rewarded_at IS NULL` → INSERT в `referral_pending_rewards`. Под flag `REFERRAL_REWARD_ENABLED`.
- `backend/internal/app/app.go` — wire-up `ReferralRewardLoop` + repos; start/stop в `Run/Shutdown`.
- `backend/internal/infrastructure/config/` — добавить поля `AnalyticsConfig.ProInsightsTeaserEnabled bool`, `ReferralConfig.RewardEnabled bool`.
- `frontend/src/pages/analytics.tsx:169-183` — три состояния: Free → `<UpgradeGate targetPlan="Pro">`, Pro → `<InsightsPanel insights={proInsights}/>` + locked-card для 5 типов, Max → как сейчас.
- `frontend/src/hooks/use-analytics.ts:useInsights` — `enabled` flag triggers на Pro+Max.
- `frontend/src/pages/pricing.tsx:299` — динамика badge `−10%` → расчёт `Math.round((monthly*12 - yearly)/(monthly*12)*100)` из реальных цен.

### Сущности

```go
// backend/internal/models/referral_pending_reward.go
type ReferralPendingReward struct {
    ID         uint      `gorm:"primaryKey"`
    ReferrerID uint      `gorm:"not null"`
    RefereeID  uint      `gorm:"not null;uniqueIndex"`
    PaymentID  uint      `gorm:"not null"`
    EligibleAt time.Time `gorm:"not null;index"`
    CreatedAt  time.Time
}
```

```go
// backend/internal/usecases/analytics
var proAllowedInsights = []string{
    models.InsightUnusedPrompts,
    models.InsightPossibleDuplicates,
}

func (s *Service) ComputeInsights(ctx context.Context, userID uint, teamID *uint, allowed []string) error
```

### Контракты слоёв

- **HTTP → Usecase (analytics):** `GET /api/analytics/insights` → `service.GetInsightsGated(userID, teamID)` без изменения сигнатуры. Внутри сервиса теперь filter, не master gate.
- **Usecase (subscription) → Repo (referral_pending_rewards):** subscription пишет напрямую в repo, без `referral.Service` dependency в subscription package (loose coupling).
- **Loop → Repo:** `ReferralRewardLoop.tick()` → `ListEligible(now)` → проверка `payments.Status == succeeded` → grant → DELETE row.

---

## 4. Модель данных

### Миграция 000072 — Free max_prompts 15→25

```sql
-- up
UPDATE subscription_plans SET max_prompts = 25, updated_at = NOW() WHERE id = 'free';
-- down
UPDATE subscription_plans SET max_prompts = 15, updated_at = NOW() WHERE id = 'free';
```

**Влияние на данные.** Backfill не нужен. Юзеры с `legacy_quotas.max_prompts=50` (Pack E) сохраняют 50 через `effectiveLimit = max(legacy, plan)`. Юзеры с legacy_quotas={} — апгрейд 15→25. **Никто не теряет лимит.**

**Обратимость rollback.** Полная.

### Миграция 000073 — Annual discount 10%→20%

```sql
-- up
UPDATE subscription_plans SET price_kop = 575000,  updated_at = NOW() WHERE id = 'pro_yearly';
UPDATE subscription_plans SET price_kop = 1247000, updated_at = NOW() WHERE id = 'max_yearly';
-- down
UPDATE subscription_plans SET price_kop = 649000  WHERE id = 'pro_yearly';
UPDATE subscription_plans SET price_kop = 1399000 WHERE id = 'max_yearly';
```

**Влияние на данные.** Существующие `subscriptions` НЕ затрагиваются — `current_period_end` уже зафиксирован. На renewal `renewal.go:134 plans.GetByID()` прочтёт новую цену → `pay.AmountKop = plan.PriceKop` → T-Bank Charge на новую сумму.

**Обратимость rollback.** Полная.

**[ДОПУЩЕНИЕ]:** T-Bank Charge на rebillId принимает произвольный Amount в Init request. Проверить на staging перед prod.

### Миграция 000074 — Referral pending rewards

```sql
-- up
CREATE TABLE referral_pending_rewards (
    id          BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referee_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_id  BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    eligible_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_referral_pending_unique_referee
    ON referral_pending_rewards (referee_id);
CREATE INDEX idx_referral_pending_eligible_at
    ON referral_pending_rewards (eligible_at);

-- down
DROP INDEX IF EXISTS idx_referral_pending_eligible_at;
DROP INDEX IF EXISTS idx_referral_pending_unique_referee;
DROP TABLE IF EXISTS referral_pending_rewards;
```

**Влияние на данные.** Новая таблица, пустая. Backfill не делаем — существующие рефералы остаются без награды (acceptable для MVP).

**Обратимость rollback.** DROP TABLE удаляет pending'и (теряются grant'ы которые не успели произойти). Webhook-логика должна быть guarded `REFERRAL_REWARD_ENABLED` flag.

---

## 5. API контракт

`GET /api/analytics/insights` — схема ответа не меняется (`Insights[]`), но **семантика** изменяется:

| Plan | Старое поведение | Новое поведение |
|---|---|---|
| free | HTTP 402 `ErrMaxRequired` | HTTP 402 `ErrPaidRequired` |
| pro / pro_yearly | HTTP 402 `ErrMaxRequired` | HTTP 200 — массив только `unused_prompts` и `possible_duplicates` |
| max / max_yearly | HTTP 200 — все 7 типов | без изменений |

Доменная ошибка `analytics.ErrMaxRequired` → `analytics.ErrPaidRequired`. HTTP-маппинг сохраняет 402.

`GET /api/auth/referral` — без изменения схемы (`ReferralInfo{code, invited_count, referred_by, reward_granted}`). После реализации reward'а `reward_granted=true` для юзеров с непустым `referral_rewarded_at` — фронт уже умеет.

`POST /api/webhooks/tbank` — внешний контракт не меняется. Меняется только internal handler.

---

## 6. Зависимости

- **Внешние сервисы:** нет новых.
- **Новые библиотеки:** нет. `slices.Contains` в stdlib Go 1.21+ (проект на Go 1.25).
- **Кросс-командные блокеры:** нет.
- **Supply-chain:** lockfile не меняется.

---

## 7. План тестирования

### Unit

- `quota_test.go` (новый case): после 000072, юзер с `legacy_quotas={"max_prompts":50}` → `effectiveLimit=50`; новый юзер (legacy={}) → `effectiveLimit=25`.
- `analytics/service_test.go` (table-driven): `GetInsightsGated` возвращает правильный subset для plan ∈ {free, pro, pro_yearly, max, max_yearly}. Free → `ErrPaidRequired`. Pro → 2 типа. Max → 7.
- `analytics/insights_test.go`: `ComputeInsights(allowed=[unused,duplicates])` пишет только эти 2 типа.
- `referral/reward_test.go`:
  - `GrantReward(referrer на Pro)` → `current_period_end += 30d`, `referral_rewarded_at = now`.
  - `GrantReward(referrer на Free)` → создаёт trial Subscription{pro, 30d, auto_renew=false}.
  - `GrantReward(referrer на Max)` → продление Max на 30d.
  - Идемпотентность: повторный grant → no-op.
  - `ListEligible` фильтрует `eligible_at < now`.
- `subscription/webhook_scenarios_test.go`: payment.succeeded + referred_by → INSERT pending.

### Integration (testcontainers)

- Миграция 000074 forward+rollback.
- E2E referral: B регнулся с ref=A, оплатил Pro → INSERT pending → симуляция «прошло 14d» → loop tick → A получает trial Pro, pending удалён.
- Refund-protection: payment.refunded до 14d → loop пропускает grant.

### E2E (Playwright)

- Free → `UpgradeGate targetPlan="Pro"` на `/analytics`.
- Pro → 2 insight типа + locked-карточки для 5.
- Max → все 7.

### Применимо

- Contract testing — N/A (нет new endpoints).
- Property-based — N/A.
- Load — N/A на этом этапе.
- Security — manual review на referral arbitrage (см. §12).

### Что НЕ тестируем

- T-Bank rebillId charge на изменённую цену — manual smoke на staging.
- Existing `expirationLoop` для trial subscription — уже покрыт.

---

## 8. Наблюдаемость

### Метрики

| Имя | Тип | Labels | Назначение |
|---|---|---|---|
| `referral_rewards_pending_total` | CounterVec | `result` (`recorded`/`skipped_already_rewarded`/`skipped_no_referrer`) | Создание pending'ов |
| `referral_rewards_granted_total` | CounterVec | `referrer_plan` (`free`/`pro`/`max`) | Успешные grant'ы |
| `referral_rewards_skipped_total` | CounterVec | `reason` (`refunded`/`referee_cancelled`/`referrer_deleted`) | Не дали reward |
| `analytics_insights_gated_total` | CounterVec | `plan`, `result` (`full`/`partial`/`blocked`) | Эффективность teaser'а |

**Cardinality.** Pending=3, Granted=3, Skipped=3, Gated=9 = 18 series. Безопасно.

### Логи

```
slog.Info("referral.pending.inserted", "referrer_id", ..., "referee_id", ..., "payment_id", ..., "eligible_at", ...)
slog.Info("referral.reward.granted", "referrer_id", ..., "from_plan", ..., "to_plan", "pro", "new_period_end", ..., "is_trial", true|false)
slog.Warn("referral.reward.skipped", "referrer_id", ..., "referee_id", ..., "reason", ...)
slog.Info("analytics.insights.gated", "user_id", ..., "plan", ..., "allowed_types_count", ..., "total_types", 7)
```

### Трейсы

`ReferralRewardLoop.tick` — span с атрибутами `referral.eligible_count`, `referral.granted_count`, `referral.skipped_count`.

### Sentry fingerprints

- `referral.grant.failed.{db|payment_lookup}`.
- `analytics.gate.unknown_plan_id`.

### Алерты Grafana

| Алерт | Порог | Severity | Runbook |
|---|---|---|---|
| `referral_pending_backlog_growing` | `pending - granted - skipped > 100` за 24h | warning | `ReferralRewardLoopStalled.md` |
| `referral_grant_error_rate_high` | `rate(referral.reward.failed[15m]) > 0.1/min` | critical | `ReferralRewardLoopStalled.md` |
| `referral_refund_arbitrage` | `rate(skipped{reason="refunded"}[1h]) > 5/h` | warning | Investigate в Sentry |

### SLO / SLI

- **SLO:** 99% pending'ов процессятся в течение 24 часов после `eligible_at`.
- **Error budget:** 1% pending'ов в месяц могут зависать > 24h.

---

## 9. План внедрения

| ID | Шаг | Owner | Критерий готовности | Зависит |
|---|---|---|---|---|
| **S1** | Миграция 000072 (Free 25). Frontend `pricing.tsx:75` уже динамический. | backend | `psql -c "SELECT max_prompts FROM subscription_plans WHERE id='free'"` = 25; quota integration-test зелёный | — |
| **S2** | Миграция 000073 (Annual −20%). Frontend `pricing.tsx:299` динамика badge через `Math.round((monthly*12-yearly)/(monthly*12)*100)`. | backend + frontend | psql показывает 575000/1247000; pricing-page показывает «−20%» | — |
| **S3** | Backend Pro Insights gate refactor: `proAllowedInsights` const, `GetInsightsGated` per-type, `ComputeInsights(allowed)`, `ListMaxUsers→ListPaidUsers`. Env flag `PRO_INSIGHTS_TEASER_ENABLED` (default false). | backend | `analytics_test.go` 5 plan-case'ов зелёные; CI зелёный | — |
| **S4** | Frontend insights UI: три состояния. `useInsights(enabled=isPaid)`. Locked-карточки с tooltip «Доступно в Max». | frontend | Manual: Free→UpgradeGate, Pro→2+5locked, Max→7 | S3 |
| **S5** | Миграция 000074 + модель + repo. | backend | `migrate up/down/up` зелёный; repo unit-тест зелёный | — |
| **S6** | `usecases/referral/reward.go`: `Service.GrantReward`. 3 ветки (Free/Pro/Max). Env flag `REFERRAL_REWARD_ENABLED` (default false). | backend | `reward_test.go` 6+ case'ов зелёные | S5 |
| **S7** | `ReferralRewardLoop` (паттерн `RenewalLoop`). Interval 1h, batch ≤100, `safeloop.RunWithRecover("referral_reward")`. | backend | Smoke dev: INSERT pending eligible_at=now-1m → 1 тик → grant + DELETE | S6 |
| **S8** | Webhook integration в `HandleWebhook`. | backend | `webhook_scenarios_test.go` новый case зелёный | S5 |
| **S9** | Wire-up в `app.go`: создание loop + repos, start/stop. | backend | Сервер стартует с `REFERRAL_REWARD_ENABLED=true`, лог `referral.reward.loop_started` | S7, S8 |
| **S10** | ADR-0008, ADR-0009, Runbook, CLAUDE.md upd. | docs | Файлы существуют | S3, S7 |
| **S11** | Включить `PRO_INSIGHTS_TEASER_ENABLED=true` на prod после 1 недели observability. | DevOps | `analytics_insights_gated_total{plan="pro"}` > 0 за 7d | S4 (deployed) + 7d |
| **S12** | Включить `REFERRAL_REWARD_ENABLED=true` на prod после 1 недели QA. | DevOps | `referral_rewards_pending_total{result="recorded"}` > 0 после первых платежей с referred_by | S9 (deployed) + 7d |

**Atomicity.** S1-S2 = pt.1 (pricing migrations). S3-S4 = pt.2 (Insights). S5-S9 = pt.3 (Referral). S11-S12 = flip switches.

Размер диффов:

- S1, S2: ≤50 строк.
- S3: ~200 строк.
- S4: ~100 строк.
- S5: ~80 строк.
- S6: ~150 строк.
- S7: ~250 строк.
- S8: ~50 строк.

---

## 10. Rollout и kill-switch

### Стратегия

- **Wave 1 (S1+S2): direct prod.** Низкий риск, обратимо через миграцию down. Без feature flag.
- **Wave 2 (S3+S4) под `PRO_INSIGHTS_TEASER_ENABLED`.** Default false → деплой → 1 неделя observability → enable.
- **Wave 3 (S5-S9) под `REFERRAL_REWARD_ENABLED`.** Default false → деплой → 1 неделя QA → enable.

### Feature flags

| Имя | Owner | Default | Где читается | При false |
|---|---|---|---|---|
| `PRO_INSIGHTS_TEASER_ENABLED` | backend | false | `usecases/analytics/service.go` через `config.AnalyticsConfig` | `GetInsightsGated` Max-only (как сейчас) |
| `VITE_PRO_INSIGHTS_TEASER_ENABLED` | frontend | false | `analytics.tsx` | UpgradeGate для всех не-Max |
| `REFERRAL_REWARD_ENABLED` | backend | false | `subscription.go` + `app.go` | webhook не INSERT'ит pending; loop не стартует |

### Kill-switch RTO

30-60 сек через env-flag + Docker restart. Frontend — rebuild 5-10 мин.

### Communication

Changelog entries при release:

- «Бесплатный тариф теперь даёт 25 промптов».
- «Годовая подписка теперь выгоднее: −20% от месячной».
- «Подсказки на Pro: забытые промпты и дубликаты».
- «Реферальная программа: пригласите друга → +30 дней Pro».

In-app banner для Pro юзеров: «У вас новые подсказки!» с ссылкой на `/analytics` (опционально, можно ограничиться changelog).

---

## 11. Документация

- **README** — N/A.
- **ADR-0008** «Per-type Smart Insights gating» — почему const, не JSONB.
- **ADR-0009** «Delayed referral reward через cron-loop» — почему 14-day window, не synchronous.
- **Runbook `docs/runbooks/ReferralRewardLoopStalled.md`** — симптомы, диагностика, шаги восстановления, escalation.
- **CLAUDE.md upd** (раздел «Ключевые решения») — 2-3 строки про per-type insights и referral pending pattern.
- **OpenAPI** — N/A (behavior change у `/api/analytics/insights` без change схемы).

---

## 12. Риски и митигации

### Технические риски

- **T-Bank Charge на изменённую цену.** На renewal `payment.Init` отправляет новый Amount. Если T-Bank требует фиксированную сумму для rebillId — Charge упадёт. **Митигация:** staging-тест на pro_yearly до S2 prod-deploy. Текущий `renewal.go:179` уже принимает динамический amount. [ДОПУЩЕНИЕ].
- **Pro юзер видит 5 locked-карточек → раздражение.** **Митигация:** позитивный tooltip («4 типа аналитики в Max»), reuse `UpgradeGate` (визуально знаком).
- **Referral abuse: множественные fake-аккаунты с одной картой.** Один человек регистрирует 5 аккаунтов, оплачивает с одной карты → 150 дней Pro. **Митигация:** [ДОПУЩЕНИЕ] T-Bank возвращает CardId в webhook → UNIQUE check на `(referrer_id, card_fingerprint)`. Если поле недоступно — отложить на v2. На MVP риск низкий: реально оплачивать 5×599₽=2995₽ за 5 мес Pro = no economic gain.
- **Edge case: referrer на Max когда grant.** Решение: extend Max на 30d (его tier выше). Логика в `GrantReward` switch'ем по plan'у.
- **Trial Subscription `auto_renew=false` попадает в renewal cycle.** [ДОПУЩЕНИЕ]: `ListReadyForRenewal` фильтрует по `auto_renew=true`. Проверить в S6 при чтении repo SQL.
- **Plan cache TTL 5 мин** означает до 5 минут разные числа на разных репликах. Acceptable для тарифной миграции.

### Pre-mortem: «через 6 месяцев это сломалось»

1. **Pro Insights не конвертит в Max.** Юзеры довольны 2 типами. Сигнал: `gated_total{plan="pro"}` высокий, Pro→Max conversion <3%. **Митигация**: A/B тест 1-тип vs 2-тип cohort.
2. **Referral arbitrage** через двунаправленные рефералы (2 друга реферят друг друга). **Митигация**: CHECK constraint или check в коде. Добавить если метрика `bidirectional > 5%`.
3. **Annual −20% съел маржу** при росте инфра-cost. **Митигация**: SLA с Timeweb на фиксированный $RUB.
4. **Free 25 слишком мягкий**, конверсия <2%. **Митигация**: миграция 000075 откатывает на 20.
5. **Cron loop отстаёт при росте**. 10k платящих × 5% reffed = 500/день. Loop 1h × 100 batch = 2400/день. OK до 20k платящих. **Митигация**: batch ↑ или частота тика.
6. **`expirationLoop` неправильно downgrade'ит trial subscription**. Trial expired → возврат на Free → платёж за обычный Pro создаёт второй subscription → unique conflict. **Митигация**: integration-тест.

### Известные ограничения

- T-Bank rebillId не возвращает card fingerprint в текущей реализации `payment.PaymentProvider`.
- Grafana дашборд — ручной JSON-import.
- Frontend feature flag через Vite env — рабочее окружение требует rebuild.

---

## 13. Метрики успеха

### Бизнес

| Метрика | Baseline | Цель | Срок |
|---|---|---|---|
| Free → Pro conversion (cohort 30d) | ~1.5% [ДОПУЩЕНИЕ из BUSINESS_RESEARCH] | 3-4% | 90 дней после S1+S2 |
| Pro → Max conversion (cohort 60d) | ~5% [ДОПУЩЕНИЕ] | 12-15% | 90 дней после S11 |
| Annual share of paying users | 0% | 25-30% | 60 дней после S2 |
| K-factor (referral viral) | 0 | 0.15-0.30 | 90 дней после S12 |

### Технические

| Метрика | Цель |
|---|---|
| `/api/analytics/insights` p95 latency | не вырастет > +5ms |
| `/api/webhooks/tbank` p95 latency | не вырастет > +20ms |
| `ReferralRewardLoop` tick p95 | < 5 sec |
| `analytics_insights_gated_total{result="partial"}` | > 100/день для plan=pro в первую неделю после S11 |
| `referral.reward.granted` error rate | < 0.5% |

### Срок измерения

- S1+S2: 1 неделя observability.
- S3+S4: 60 дней (Pro→Max lag).
- S5-S9: 90 дней (viral циклы медленные).

---

## 14. Открытые вопросы

1. **T-Bank Charge accept изменённый Amount?** Кому: исполнитель (staging webhook test). Блокирует: нет — deploy S2 после успешного теста.
2. **`ListReadyForRenewal` фильтрует `auto_renew=true`?** Кому: исполнитель (читает SQL при S6). Блокирует: нет — если не фильтрует, добавить фильтр.
3. **Card fingerprint в T-Bank webhook?** Кому: исполнитель (читает payload). Блокирует: нет — anti-abuse v2 отложен.
4. **Email-нотификация при expired trial subscription** — релевантно для referral-trial? Кому: исполнитель. Блокирует: нет — guard в notifier по `sub.rebill_id == ""`.
5. **Bidirectional referral protection в БД** — сейчас или после первого случая? Кому: пользователь. Блокирует: нет.
6. **In-app banner для Pro юзеров «у вас новые подсказки»?** Кому: пользователь. Блокирует: нет.

---

## Self-check

- [x] **Инвентаризация инструментов проведена.** Evidence: TaskCreate/TaskUpdate (#1-#8), AskUserQuestion (5 вопросов), Read (3 файла напрямую), Glob (3 паттерна), Bash (git log), Agent subagent (1 вызов 5 областей), WebSearch (4 запроса).
- [x] **Прочитан релевантный код.** Evidence: `models/subscription.go:14-263`, `usecases/subscription/renewal.go:1-275`, `usecases/auth/referral.go:1-115`; через subagent: `quota.go:73`, `analytics/service.go:111`, `insights_loop.go:78`, `analytics.tsx:169-183`, `plan_repo.go:20`, миграции 000068/000069.
- [x] **Внешняя документация.** Evidence: WebSearch — Битрикс24 −30%, Яндекс 360 от 309₽, BotHub/SYNTX/F5AI context. [НЕДОСТУПНО без верификации: T-Bank Acquiring API spec про rebillId+Amount — отложено в Открытые вопросы #1].
- [x] **Решения с ≥2 альтернативами.** Evidence: §2 Решение 1 (3: const/JSONB/env), Решение 2 (3: delayed/sync/lazy), Решение 3 (3: trial-sub/legacy_quotas/skip).
- [x] **Нет over-engineering.** Evidence: 1 новая таблица с 1 use-case; 1 новый usecase (`referral.RewardService`) с 1 use-case; 1 константа в analytics. Никаких «на будущее».
- [x] **Edge cases / errors / security.** Evidence: §7 refund-protection integration test, §8 Sentry fingerprints, §12 6 pre-mortem сценариев + 6 technical risks. Referral arbitrage явно обсужден.
- [x] **Консистентность с существующими паттернами.** Evidence: `safeloop.RunWithRecover` (`renewal.go:91`), `effectiveLimit` (`quota.go:73`), `UpgradeGate` (`upgrade-gate.tsx:12`), миграции `IF NOT EXISTS` (000068/000069).
- [x] **Критерии готовности конкретны.** Evidence: §9 — S1: psql=25 + integration test green; S3: 5 plan-case unit tests green; S7: smoke INSERT eligible_at=now-1m → grant; и т.д.
- [x] **Допущения помечены.** Evidence: 4 [ДОПУЩЕНИЕ] (T-Bank Charge, auto_renew filter, card fingerprint, baseline conversion data) — все в Открытых вопросах.
- [x] **Scope discipline.** Evidence: отложены Lifetime, Edu, Team-тариф, Power-Features, Referral anti-abuse v2. Single PR thread = single feature group.
