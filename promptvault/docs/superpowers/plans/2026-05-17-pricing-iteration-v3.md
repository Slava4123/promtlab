# Pricing Iteration v3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Реализовать pricing iteration v3 — Free 15→25, Annual −10%→−20%, Pro Smart Insights teaser (2 типа), Referral reward +30 дней Pro через delayed cron.

**Architecture:** 3 миграции (000072/000073/000074) + per-type insights gate в `usecases/analytics` через hardcoded `proAllowedInsights` const + новый `ReferralRewardLoop` (паттерн `subscription.RenewalLoop`) с таблицей `referral_pending_rewards` и 14-day eligibility. Существующий `effectiveLimit = max(legacy, plan)` (quota.go:73) и T-Bank renewal pricing через `plans.GetByID()` (renewal.go:134) работают без code changes для Free 25 и Annual −20%.

**Tech Stack:** Go 1.25, Chi, GORM v2, PostgreSQL 18, koanf v2 config, slog, promauto Prometheus, React 19.2 + Vite 8 + TanStack Query + Vitest, testcontainers-go v0.41.

**Дизайн-спека:** [docs/superpowers/specs/2026-05-17-pricing-iteration-v3-design.md](../specs/2026-05-17-pricing-iteration-v3-design.md)

---

## File Structure

### Создаём

| Путь | Назначение |
|---|---|
| `backend/internal/infrastructure/postgres/migrations/000072_free_prompts_25.{up,down}.sql` | Free max_prompts: 15→25 |
| `backend/internal/infrastructure/postgres/migrations/000073_annual_discount_20pct.{up,down}.sql` | pro_yearly 5750₽, max_yearly 12470₽ |
| `backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.{up,down}.sql` | Таблица + 2 индекса |
| `backend/internal/models/referral_pending_reward.go` | GORM модель `ReferralPendingReward` |
| `backend/internal/interface/repository/referral_reward.go` | Interface `ReferralRewardRepository` |
| `backend/internal/infrastructure/postgres/repository/referral_reward_repo.go` | GORM implementation |
| `backend/internal/usecases/referral/reward.go` | `Service.GrantReward` (3 ветки: Free/Pro/Max) |
| `backend/internal/usecases/referral/reward_loop.go` | `ReferralRewardLoop` (паттерн renewal.go) |
| `backend/internal/usecases/referral/types.go` | `RewardSummary{Granted, Skipped, Errors}` |
| `backend/internal/usecases/referral/errors.go` | Доменные ошибки |
| `backend/internal/usecases/referral/reward_test.go` | Table-driven unit tests |
| `backend/internal/infrastructure/postgres/repository/referral_reward_repo_test.go` | Integration test (testcontainers) |
| `docs/adr/0008-pertype-insights-gate.md` | ADR для решения 1 |
| `docs/adr/0009-delayed-referral-reward.md` | ADR для решения 2 |
| `docs/runbooks/ReferralRewardLoopStalled.md` | On-call runbook |

### Меняем

| Путь | Что меняем |
|---|---|
| `backend/internal/usecases/analytics/insights.go` | Добавить `proAllowedInsights` const + `insightsForPlan()`; `ComputeInsights(allowed []string)` |
| `backend/internal/usecases/analytics/service.go` | `GetInsightsGated` per-type filter (использует `ErrProRequired`) |
| `backend/internal/usecases/analytics/insights_loop.go` | `ListMaxUsers→ListPaidUsers`, dispatch по plan'у |
| `backend/internal/interface/repository/user.go` | Переименовать метод |
| `backend/internal/infrastructure/postgres/repository/user_repo.go` | Обновить query |
| `backend/internal/delivery/http/analytics/errors.go` | Mapping `ErrProRequired` для insights endpoint |
| `backend/internal/infrastructure/config/analytics.go` | Добавить `ProInsightsTeaserEnabled bool` |
| `backend/internal/infrastructure/config/config.go` | Добавить секцию `Referral` |
| `backend/internal/usecases/subscription/subscription.go` | В `activateSubscription` — INSERT в `referral_pending_rewards` |
| `backend/internal/app/app.go` | Wire-up `ReferralRewardLoop` + repos |
| `backend/internal/app/lifecycle.go` | Start/Stop loop |
| `backend/internal/infrastructure/metrics/metrics.go` | Новые counters |
| `frontend/src/pages/analytics.tsx` | Three-state UI: Free/Pro/Max |
| `frontend/src/pages/pricing.tsx` | Dynamic yearly badge `−{savedPct}%` |
| `promptvault/CLAUDE.md` | 2-3 строки про новые паттерны |

---

## Phase 1: Pricing Migrations (S1+S2)

### Task 1: Migration 000072 — Free max_prompts 25

**Files:**
- Create: `backend/internal/infrastructure/postgres/migrations/000072_free_prompts_25.up.sql`
- Create: `backend/internal/infrastructure/postgres/migrations/000072_free_prompts_25.down.sql`
- Test: `backend/internal/usecases/quota/quota_test.go` (новый case)

- [ ] **Step 1: Write up migration**

```sql
-- 000072_free_prompts_25.up.sql
-- Pack G: повышение Free max_prompts с 15 → 25.
-- Grandfather от Pack E (миграция 000068) сохраняется автоматически:
-- effectiveLimit = max(legacy_quotas.max_prompts, plan.MaxPrompts).
-- Юзеры с legacy={max_prompts:50} остаются на 50; новые юзеры — 25.
UPDATE subscription_plans
   SET max_prompts = 25, updated_at = NOW()
 WHERE id = 'free';
```

- [ ] **Step 2: Write down migration**

```sql
-- 000072_free_prompts_25.down.sql
UPDATE subscription_plans
   SET max_prompts = 15, updated_at = NOW()
 WHERE id = 'free';
```

- [ ] **Step 3: Write unit test для grandfather behavior**

Открой `backend/internal/usecases/quota/quota_test.go`, найди существующие test cases (паттерн `TestService_CheckPromptQuota_*`), добавь новый:

```go
func TestService_CheckPromptQuota_Free25WithLegacyGrandfather(t *testing.T) {
    ctx := context.Background()
    cases := []struct {
        name         string
        legacyJSON   string
        wantLimit    int
    }{
        {"new user — no legacy", `{}`, 25},
        {"Pack E grandfather", `{"max_prompts":50}`, 50},
        {"higher legacy stays", `{"max_prompts":100}`, 100},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            u := &models.User{ID: 1, PlanID: "free", LegacyQuotas: json.RawMessage(tc.legacyJSON)}
            users := &fakeUserRepo{user: u}
            plans := &fakePlanRepo{plan: &models.SubscriptionPlan{ID: "free", MaxPrompts: 25}}
            prompts := &fakePromptRepo{count: 0}
            svc := NewService(users, plans, prompts, /* остальные fakes */)

            err := svc.CheckPromptQuota(ctx, 1)
            require.NoError(t, err)

            // Создаём (tc.wantLimit) промптов — должно пройти
            prompts.count = int64(tc.wantLimit - 1)
            require.NoError(t, svc.CheckPromptQuota(ctx, 1))

            // На потолок — fail
            prompts.count = int64(tc.wantLimit)
            var qe *QuotaExceededError
            err = svc.CheckPromptQuota(ctx, 1)
            require.ErrorAs(t, err, &qe)
            require.Equal(t, tc.wantLimit, qe.Limit)
        })
    }
}
```

- [ ] **Step 4: Run test — должен пройти (effectiveLimit уже существует)**

Run: `go test -run TestService_CheckPromptQuota_Free25WithLegacyGrandfather ./internal/usecases/quota/ -v`
Expected: PASS (миграция не нужна для unit-test'а — он на уровне service).

- [ ] **Step 5: Integration check через psql**

После применения миграции (`docker compose -f docker-compose.dev.yml up -d --build`):

```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault -c "SELECT id, max_prompts FROM subscription_plans WHERE id='free';"
```

Expected output:
```
 id  | max_prompts
-----+-------------
 free|          25
```

- [ ] **Step 6: Test down migration**

```bash
# В docker compose контейнере (используя migrate CLI или ручной SQL)
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault -c "UPDATE subscription_plans SET max_prompts=15 WHERE id='free';"
# затем снова up:
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault -c "UPDATE subscription_plans SET max_prompts=25 WHERE id='free';"
```

Expected: оба раза без ошибок, финальное значение 25.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/infrastructure/postgres/migrations/000072_free_prompts_25.up.sql \
        backend/internal/infrastructure/postgres/migrations/000072_free_prompts_25.down.sql \
        backend/internal/usecases/quota/quota_test.go
git commit -m "feat(quotas): Pack G — Free max_prompts 15 → 25"
```

---

### Task 2: Migration 000073 — Annual discount −20%

**Files:**
- Create: `backend/internal/infrastructure/postgres/migrations/000073_annual_discount_20pct.up.sql`
- Create: `backend/internal/infrastructure/postgres/migrations/000073_annual_discount_20pct.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- 000073_annual_discount_20pct.up.sql
-- Annual discount: 10% → 20%.
-- Существующие подписки НЕ затрагиваются (T-Bank rebillId identifies card, not amount).
-- На renewal `plans.GetByID()` прочтёт новую цену → T-Bank Charge на новую сумму.
--   pro_yearly:  6490 → 5750 ₽ (-20% от monthly×12 = 7188)
--   max_yearly: 13990 → 12470 ₽ (-20% от monthly×12 = 15588)
UPDATE subscription_plans
   SET price_kop = 575000, updated_at = NOW()
 WHERE id = 'pro_yearly';

UPDATE subscription_plans
   SET price_kop = 1247000, updated_at = NOW()
 WHERE id = 'max_yearly';
```

- [ ] **Step 2: Write down migration**

```sql
-- 000073_annual_discount_20pct.down.sql
UPDATE subscription_plans
   SET price_kop = 649000, updated_at = NOW()
 WHERE id = 'pro_yearly';

UPDATE subscription_plans
   SET price_kop = 1399000, updated_at = NOW()
 WHERE id = 'max_yearly';
```

- [ ] **Step 3: Integration check через psql**

После применения миграции:

```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault -c \
  "SELECT id, price_kop FROM subscription_plans WHERE id LIKE '%_yearly' ORDER BY id;"
```

Expected:
```
     id     | price_kop
------------+-----------
 max_yearly |   1247000
 pro_yearly |    575000
```

- [ ] **Step 4: Smoke staging — T-Bank Charge на новую цену**

**До prod deploy.** На staging environment имеется тестовая pro_yearly подписка с активным `rebill_id`. Дождись/триггери renewal cycle (либо через `lookahead=48h` и подкрутить `current_period_end=now+1h` на staging БД).

Проверь в T-Bank dashboard и в `payments` таблице:

```sql
SELECT id, amount_kop, status, created_at FROM payments
 WHERE subscription_id=<staging_sub_id>
 ORDER BY created_at DESC LIMIT 1;
```

Expected: `amount_kop = 575000`, `status = 'succeeded'`. Если ошибка — открыть [Открытый вопрос #1 из spec'и] и НЕ деплоить на prod.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infrastructure/postgres/migrations/000073_annual_discount_20pct.up.sql \
        backend/internal/infrastructure/postgres/migrations/000073_annual_discount_20pct.down.sql
git commit -m "feat(billing): annual discount 10% → 20% (pro_yearly 5750₽, max_yearly 12470₽)"
```

---

### Task 3: Frontend dynamic yearly badge

**Files:**
- Modify: `frontend/src/pages/pricing.tsx:269-303` (yearly toggle badge)

- [ ] **Step 1: Write Vitest test**

Создай `frontend/src/pages/__tests__/pricing-yearly-badge.test.tsx`:

```tsx
import { render } from "@testing-library/react"
import { describe, it, expect } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router-dom"
import Pricing from "../pricing"

// Mock plans с новыми ценами −20%
const mockPlans = [
  { id: "free", name: "Free", price_kop: 0, period_days: 0, /* ... остальные fields */ },
  { id: "pro", name: "Pro", price_kop: 59900, period_days: 30 },
  { id: "pro_yearly", name: "Pro (год)", price_kop: 575000, period_days: 365 },
  { id: "max", name: "Max", price_kop: 129900, period_days: 30 },
  { id: "max_yearly", name: "Max (год)", price_kop: 1247000, period_days: 365 },
]

describe("Pricing yearly badge", () => {
  it("показывает динамический процент скидки", () => {
    // Mock usePlans hook к mockPlans
    // Кликнуть на "Ежегодно" tab
    // assert: badge text content === "−20%"
  })
})
```

(Полная реализация теста зависит от настройки моков `usePlans` — см. эталон в `frontend/src/pages/settings/__tests__/layout.test.tsx`.)

- [ ] **Step 2: Run test — FAIL**

Run: `npx vitest run pricing-yearly-badge -- --reporter=verbose`
Expected: FAIL — badge всё ещё хардкод `−10%`.

- [ ] **Step 3: Implement dynamic badge**

В `frontend/src/pages/pricing.tsx`, заменить хардкод `−10%` (строка ~299) на динамический расчёт:

```tsx
// Вынести helper рядом с yearlyAnchor (после line 167):
function yearlyDiscountPct(plans: Plan[]): number {
  const pro = plans.find((p) => p.id === "pro")
  const proYearly = plans.find((p) => p.id === "pro_yearly")
  if (!pro || !proYearly) return 0
  const expected = pro.price_kop * 12
  if (expected <= 0) return 0
  return Math.round(((expected - proYearly.price_kop) / expected) * 100)
}

// В компоненте, до return (после строки 229):
const yearlyPct = plans ? yearlyDiscountPct(plans) : 0

// В JSX, заменить хардкод "−10%" (строка 299) на:
<span className="rounded-full bg-emerald-500/15 px-2 py-0.5 text-[0.65rem] font-semibold text-emerald-600 dark:text-emerald-400">
  −{yearlyPct}%
</span>
```

- [ ] **Step 4: Run test — PASS**

Run: `npx vitest run pricing-yearly-badge -- --reporter=verbose`
Expected: PASS.

- [ ] **Step 5: Manual smoke в браузере**

```bash
docker compose -f docker-compose.dev.yml up -d --build
# Открыть http://localhost:5173/pricing
```

Expected: badge на yearly-tab показывает «−20%», карточки `Pro (год)` показывают 5 750 ₽ с зачёркнутой 7 188 ₽ и «экономия 20%».

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/pricing.tsx frontend/src/pages/__tests__/pricing-yearly-badge.test.tsx
git commit -m "feat(pricing): dynamic yearly discount badge (−20%)"
```

---

## Phase 2: Pro Insights Teaser (S3+S4)

### Task 4: proAllowedInsights const + insightsForPlan helper

**Files:**
- Modify: `backend/internal/usecases/analytics/insights.go` (add const + helper)
- Test: `backend/internal/usecases/analytics/insights_test.go` (new test case)

- [ ] **Step 1: Write failing test**

В `backend/internal/usecases/analytics/insights_test.go` добавь:

```go
func TestInsightsForPlan(t *testing.T) {
    cases := []struct {
        plan string
        want []string
    }{
        {"free", nil},
        {"pro", []string{models.InsightUnusedPrompts, models.InsightPossibleDuplicates}},
        {"pro_yearly", []string{models.InsightUnusedPrompts, models.InsightPossibleDuplicates}},
        {"max", allInsightTypes()},
        {"max_yearly", allInsightTypes()},
        {"unknown", nil},
    }
    for _, tc := range cases {
        t.Run(tc.plan, func(t *testing.T) {
            got := insightsForPlan(tc.plan)
            require.ElementsMatch(t, tc.want, got)
        })
    }
}

// helper для теста — все 7 типов
func allInsightTypes() []string {
    return []string{
        models.InsightUnusedPrompts,
        models.InsightTrending,
        models.InsightDeclining,
        models.InsightMostEdited,
        models.InsightPossibleDuplicates,
        models.InsightOrphanTags,
        models.InsightEmptyCollections,
    }
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestInsightsForPlan ./internal/usecases/analytics/ -v`
Expected: FAIL — `insightsForPlan` undefined.

- [ ] **Step 3: Implement helper**

В `backend/internal/usecases/analytics/insights.go` (после line 9 imports):

```go
// proAllowedInsights — типы Smart Insights, доступные на Pro тарифе.
// Pro получает teaser из 2 housekeeping-типов; остальные 5 (trending,
// declining, most_edited, orphan_tags, empty_collections) — Max-only.
// Решение принято в ADR-0008. Изменение состава тиров требует:
//   1) обновить этот список
//   2) обновить ADR-0008
//   3) проверить frontend `analytics.tsx` (зеркало lock-карточек)
var proAllowedInsights = []string{
    models.InsightUnusedPrompts,
    models.InsightPossibleDuplicates,
}

// maxAllInsights — все 7 типов, доступные на Max.
var maxAllInsights = []string{
    models.InsightUnusedPrompts,
    models.InsightTrending,
    models.InsightDeclining,
    models.InsightMostEdited,
    models.InsightPossibleDuplicates,
    models.InsightOrphanTags,
    models.InsightEmptyCollections,
}

// insightsForPlan возвращает список разрешённых insight типов для plan'а.
// nil — план не имеет доступа (Free / unknown). Pro имеет 2 типа,
// Max — все 7. Используется в GetInsightsGated и ComputeInsights.
func insightsForPlan(planID string) []string {
    switch planID {
    case "pro", "pro_yearly":
        return proAllowedInsights
    case "max", "max_yearly":
        return maxAllInsights
    default:
        return nil
    }
}
```

- [ ] **Step 4: Run test — PASS**

Run: `go test -run TestInsightsForPlan ./internal/usecases/analytics/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/analytics/insights.go \
        backend/internal/usecases/analytics/insights_test.go
git commit -m "feat(analytics): proAllowedInsights const + insightsForPlan helper"
```

---

### Task 5: Refactor ComputeInsights(allowed []string)

**Files:**
- Modify: `backend/internal/usecases/analytics/insights.go:ComputeInsights` (add parameter)
- Test: `backend/internal/usecases/analytics/insights_test.go` (new case)

- [ ] **Step 1: Write failing test**

Добавь в `insights_test.go`:

```go
func TestComputeInsights_FiltersByAllowed(t *testing.T) {
    ctx := context.Background()
    repo := &fakeAnalyticsRepo{
        unusedReturns:     []models.UnusedPromptRow{{ID: 1}},
        trendingReturns:   []models.TrendRow{{ID: 2}},
        duplicatesReturns: []models.DupRow{{ID: 3}},
    }
    svc := NewService(repo, nil, nil, nil, nil)
    svc.SetExperimentalInsights(true)
    svc.SetTrgmAvailable(true)
    svc.SetNowFn(func() time.Time { return time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC) })

    // Allowed только unused + duplicates (Pro набор)
    err := svc.ComputeInsights(ctx, 42, nil, []string{
        models.InsightUnusedPrompts,
        models.InsightPossibleDuplicates,
    })
    require.NoError(t, err)

    upserted := repo.upsertedTypes()
    require.ElementsMatch(t, []string{
        models.InsightUnusedPrompts,
        models.InsightPossibleDuplicates,
    }, upserted, "trending должен быть skip'нут — не в allowed list")
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestComputeInsights_FiltersByAllowed ./internal/usecases/analytics/ -v`
Expected: FAIL — `ComputeInsights` signature mismatch (3 args vs новые 4).

- [ ] **Step 3: Refactor ComputeInsights**

В `backend/internal/usecases/analytics/insights.go`, изменить сигнатуру `ComputeInsights`:

```go
// ComputeInsights — пересчёт детерминистических инсайтов для юзера.
// allowed — список разрешённых типов (см. insightsForPlan); если содержит
// тип — он считается и upsert'ится, иначе skip без SQL-запроса.
//
// Pro: allowed = proAllowedInsights (2 типа).
// Max: allowed = maxAllInsights (7 типов).
// Free: allowed = nil → no-op (loop фильтрует Free через ListPaidUsers).
func (s *Service) ComputeInsights(ctx context.Context, userID uint, teamID *uint, allowed []string) error {
    if len(allowed) == 0 {
        return nil
    }
    now := s.nowFn()
    isAllowed := func(t string) bool { return slices.Contains(allowed, t) }

    // 1. UNUSED PROMPTS
    if isAllowed(models.InsightUnusedPrompts) {
        unused, err := s.analytics.UnusedPrompts(ctx, userID, teamID, now.AddDate(0, 0, -30), 20)
        if err != nil {
            slog.WarnContext(ctx, "analytics.insights.unused_failed", "err", err, "user_id", userID, "team_id", teamID)
        } else if len(unused) > 0 {
            s.upsertSafe(ctx, userID, teamID, models.InsightUnusedPrompts, unused)
        }
    }

    // 2. TRENDING
    if isAllowed(models.InsightTrending) {
        trending, err := s.analytics.GetTrendingPrompts(ctx, userID, teamID, 2.0, true, 5)
        if err != nil {
            slog.WarnContext(ctx, "analytics.insights.trending_failed", "err", err, "user_id", userID, "team_id", teamID)
        } else if len(trending) > 0 {
            s.upsertSafe(ctx, userID, teamID, models.InsightTrending, trending)
        }
    }

    // 3. DECLINING
    if isAllowed(models.InsightDeclining) {
        declining, err := s.analytics.GetTrendingPrompts(ctx, userID, teamID, 0.5, false, 5)
        if err != nil {
            slog.WarnContext(ctx, "analytics.insights.declining_failed", "err", err, "user_id", userID, "team_id", teamID)
        } else if len(declining) > 0 {
            s.upsertSafe(ctx, userID, teamID, models.InsightDeclining, declining)
        }
    }

    // 4-7. Experimental (kill-switch + pg_trgm probe).
    if s.experimentalInsights {
        if isAllowed(models.InsightMostEdited) {
            edited, err := s.analytics.MostEditedPrompts(ctx, userID, teamID, 5)
            if err != nil {
                slog.WarnContext(ctx, "analytics.insights.most_edited_failed", "err", err, "user_id", userID, "team_id", teamID)
            } else if len(edited) > 0 {
                s.upsertSafe(ctx, userID, teamID, models.InsightMostEdited, edited)
            }
        }

        if isAllowed(models.InsightPossibleDuplicates) && s.trgmAvailable {
            dups, err := s.analytics.PossibleDuplicates(ctx, userID, teamID, 0.8, 10)
            if err != nil {
                slog.WarnContext(ctx, "analytics.insights.duplicates_failed", "err", err, "user_id", userID, "team_id", teamID)
            } else if len(dups) > 0 {
                s.upsertSafe(ctx, userID, teamID, models.InsightPossibleDuplicates, dups)
            }
        }

        if isAllowed(models.InsightOrphanTags) {
            orphans, err := s.analytics.OrphanTags(ctx, userID, teamID, 10)
            if err != nil {
                slog.WarnContext(ctx, "analytics.insights.orphan_tags_failed", "err", err, "user_id", userID, "team_id", teamID)
            } else if len(orphans) > 0 {
                s.upsertSafe(ctx, userID, teamID, models.InsightOrphanTags, orphans)
            }
        }

        if isAllowed(models.InsightEmptyCollections) {
            empties, err := s.analytics.EmptyCollections(ctx, userID, teamID, 10)
            if err != nil {
                slog.WarnContext(ctx, "analytics.insights.empty_collections_failed", "err", err, "user_id", userID, "team_id", teamID)
            } else if len(empties) > 0 {
                s.upsertSafe(ctx, userID, teamID, models.InsightEmptyCollections, empties)
            }
        }
    }

    return nil
}
```

Не забудь добавить `"slices"` в imports.

- [ ] **Step 4: Run test — PASS**

Run: `go test -run TestComputeInsights_FiltersByAllowed ./internal/usecases/analytics/ -v`
Expected: PASS.

- [ ] **Step 5: Run existing tests — все должны быть PASS**

Run: `go test -short ./internal/usecases/analytics/...`
Expected: PASS. Если `ComputeInsights` где-то ещё вызывается — обновим в Task 7.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/usecases/analytics/insights.go \
        backend/internal/usecases/analytics/insights_test.go
git commit -m "refactor(analytics): ComputeInsights принимает allowed-список типов"
```

---

### Task 6: Refactor GetInsightsGated per-type + ErrProRequired

**Files:**
- Modify: `backend/internal/usecases/analytics/service.go:GetInsightsGated`
- Test: `backend/internal/usecases/analytics/service_test.go` (new cases)

- [ ] **Step 1: Write failing test**

Создай `backend/internal/usecases/analytics/service_test.go` (или добавь в существующий):

```go
func TestService_GetInsightsGated_PerTypeFilter(t *testing.T) {
    ctx := context.Background()
    allTypes := []string{
        models.InsightUnusedPrompts,
        models.InsightTrending,
        models.InsightDeclining,
        models.InsightMostEdited,
        models.InsightPossibleDuplicates,
        models.InsightOrphanTags,
        models.InsightEmptyCollections,
    }
    cases := []struct {
        plan          string
        wantErr       error
        wantTypeCount int
    }{
        {"free", ErrProRequired, 0},
        {"pro", nil, 2},        // unused + duplicates
        {"pro_yearly", nil, 2},
        {"max", nil, 7},
        {"max_yearly", nil, 7},
    }
    for _, tc := range cases {
        t.Run(tc.plan, func(t *testing.T) {
            // Repo возвращает все 7 типов из БД; service фильтрует
            repo := &fakeAnalyticsRepo{insightsAllTypes: allTypes}
            svc := NewService(repo, nil, nil, &fakeUserRepo{user: &models.User{ID: 1, PlanID: tc.plan}}, nil)
            insights, err := svc.GetInsightsGated(ctx, 1, nil)
            if tc.wantErr != nil {
                require.ErrorIs(t, err, tc.wantErr)
                return
            }
            require.NoError(t, err)
            require.Len(t, insights, tc.wantTypeCount)
        })
    }
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestService_GetInsightsGated_PerTypeFilter ./internal/usecases/analytics/ -v`
Expected: FAIL — текущий gate возвращает `ErrMaxRequired` для Pro вместо filter'а.

- [ ] **Step 3: Refactor GetInsightsGated**

В `backend/internal/usecases/analytics/service.go`, заменить тело `GetInsightsGated`:

```go
// GetInsightsGated — публичный endpoint /api/analytics/insights.
// Pro получает 2 типа (unused + duplicates) как teaser.
// Max — все 7 типов.
// Free → ErrProRequired (HTTP 402, upgrade prompt).
func (s *Service) GetInsightsGated(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
    planID, err := s.lookupPlanID(ctx, userID)
    if err != nil {
        return nil, err
    }
    allowed := insightsForPlan(planID)
    if len(allowed) == 0 {
        return nil, ErrProRequired
    }
    all, err := s.analytics.GetInsights(ctx, userID, teamID)
    if err != nil {
        return nil, err
    }
    // Filter в памяти — repo читает все типы, service отдаёт только разрешённые.
    filtered := make([]models.SmartInsight, 0, len(all))
    for _, ins := range all {
        if slices.Contains(allowed, ins.InsightType) {
            filtered = append(filtered, ins)
        }
    }
    return filtered, nil
}
```

Добавить `"slices"` в imports если ещё нет.

- [ ] **Step 4: Run test — PASS**

Run: `go test -run TestService_GetInsightsGated_PerTypeFilter ./internal/usecases/analytics/ -v`
Expected: PASS.

- [ ] **Step 5: Run all analytics tests — PASS**

Run: `go test -short ./internal/usecases/analytics/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/usecases/analytics/service.go \
        backend/internal/usecases/analytics/service_test.go
git commit -m "feat(analytics): GetInsightsGated per-type filter (Pro 2 types, Max 7)"
```

---

### Task 7: Rename ListMaxUsers → ListPaidUsers + insights_loop dispatch

**Files:**
- Modify: `backend/internal/interface/repository/user.go` (rename method)
- Modify: `backend/internal/infrastructure/postgres/repository/user_repo.go` (update query)
- Modify: `backend/internal/usecases/analytics/insights_loop.go` (use new name + dispatch)
- Test: `backend/internal/infrastructure/postgres/repository/user_repo_test.go` (integration)

- [ ] **Step 1: Write failing test (integration через testcontainers)**

В `backend/internal/infrastructure/postgres/repository/user_repo_test.go`:

```go
func TestUserRepo_ListPaidUsers(t *testing.T) {
    if testing.Short() {
        t.Skip("requires postgres testcontainer")
    }
    ctx := context.Background()
    db := setupTestDB(t) // testcontainers helper
    repo := NewUserRepository(db)

    // Сидим 4 юзеров с разными планами
    users := []models.User{
        {Email: "a@t.local", PlanID: "free"},
        {Email: "b@t.local", PlanID: "pro", Status: "active"},
        {Email: "c@t.local", PlanID: "pro_yearly", Status: "active"},
        {Email: "d@t.local", PlanID: "max", Status: "active"},
    }
    for i := range users {
        require.NoError(t, db.Create(&users[i]).Error)
    }

    ids, err := repo.ListPaidUsers(ctx)
    require.NoError(t, err)
    // a (free) исключён; b, c, d — included
    require.Len(t, ids, 3)
    require.ElementsMatch(t, []uint{users[1].ID, users[2].ID, users[3].ID}, ids)
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestUserRepo_ListPaidUsers ./internal/infrastructure/postgres/repository/ -v`
Expected: FAIL — `ListPaidUsers` undefined.

- [ ] **Step 3: Update interface**

В `backend/internal/interface/repository/user.go`, переименовать (line 60):

```go
// Было:
// ListMaxUsers(ctx context.Context) ([]uint, error)

// Стало:
// ListPaidUsers возвращает ID активных юзеров на Pro/Max (с любой periodicity).
// Используется в analytics insights loop — Pro имеет teaser, Max — полный набор.
ListPaidUsers(ctx context.Context) ([]uint, error)
```

- [ ] **Step 4: Update implementation**

В `backend/internal/infrastructure/postgres/repository/user_repo.go`, найти `ListMaxUsers` (line ~29) и заменить:

```go
// Было:
// func (r *userRepo) ListMaxUsers(ctx context.Context) ([]uint, error) {
//     var ids []uint
//     err := r.db.WithContext(ctx).Model(&models.User{}).
//         Where("plan_id LIKE ? AND status = ?", "max%", "active").
//         Pluck("id", &ids).Error
//     return ids, err
// }

// Стало:
func (r *userRepo) ListPaidUsers(ctx context.Context) ([]uint, error) {
    var ids []uint
    err := r.db.WithContext(ctx).Model(&models.User{}).
        Where("plan_id IN ? AND status = ?",
            []string{"pro", "pro_yearly", "max", "max_yearly"}, "active").
        Pluck("id", &ids).Error
    return ids, err
}
```

- [ ] **Step 5: Update insights_loop**

В `backend/internal/usecases/analytics/insights_loop.go` (line ~78), заменить:

```go
// Было:
// ids, err := l.users.ListMaxUsers(ctx)

// Стало:
ids, err := l.users.ListPaidUsers(ctx)
if err != nil {
    slog.ErrorContext(ctx, "insights.loop.list_paid_users_failed", "err", err)
    return
}

// Внутри errgroup'а в Goroutine — нужен plan-aware dispatch.
// Loop теперь делает один extra-roundtrip для лукапа плана юзера. Это
// приемлемо: ListPaidUsers возвращает limited set (paying users),
// для каждого один SELECT по PK кеш-friendly. Альтернатива — JOIN
// при ListPaidUsers, возвращать (id, plan_id) — оставим под YAGNI
// если loop станет узким местом.
for _, uid := range ids {
    uid := uid
    g.Go(func() error {
        user, err := l.users.GetByID(gctx, uid)
        if err != nil {
            slog.WarnContext(gctx, "insights.loop.user_lookup_failed", "user_id", uid, "err", err)
            return nil // не валим всю партию
        }
        allowed := insightsForPlan(user.PlanID)
        if len(allowed) == 0 {
            return nil // skip: юзер не на платном плане (возможно изменился между ListPaidUsers и Get)
        }
        if err := l.svc.ComputeInsights(gctx, uid, nil, allowed); err != nil {
            slog.WarnContext(gctx, "insights.loop.compute_failed", "user_id", uid, "err", err)
        }
        // Team-scope (ListOwnedTeams pass) — также с allowed
        // ... (существующий код team-loop оставить, добавить allowed parameter)
        return nil
    })
}
```

**Внимание:** `insightsForPlan` — package-local function в analytics. Если `insights_loop.go` в том же пакете — доступна напрямую. Если в подпакете — экспортировать.

- [ ] **Step 6: Build + test**

```bash
go build ./...
go test -short ./internal/usecases/analytics/...
go test ./internal/infrastructure/postgres/repository/ -run TestUserRepo_ListPaidUsers
```

Expected: всё PASS.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/interface/repository/user.go \
        backend/internal/infrastructure/postgres/repository/user_repo.go \
        backend/internal/infrastructure/postgres/repository/user_repo_test.go \
        backend/internal/usecases/analytics/insights_loop.go
git commit -m "refactor(analytics): ListMaxUsers→ListPaidUsers + plan-aware loop dispatch"
```

---

### Task 8: HTTP analytics error mapping — ErrProRequired для insights

**Files:**
- Modify: `backend/internal/delivery/http/analytics/errors.go`

- [ ] **Step 1: Write failing test**

В `backend/internal/delivery/http/analytics/handler_test.go` (создать если нет):

```go
func TestRespondError_InsightsProRequired(t *testing.T) {
    rec := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/analytics/insights", nil)
    respondError(rec, req, analyticsuc.ErrProRequired)

    require.Equal(t, http.StatusPaymentRequired, rec.Code)
    var body map[string]any
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
    require.Equal(t, "Pro", body["plan"])
    require.Equal(t, "/pricing", body["upgrade_url"])
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestRespondError_InsightsProRequired ./internal/delivery/http/analytics/ -v`
Expected: FAIL — текущий маппинг `ErrProRequired` → `respondTierRequired(w, "export", "Pro")`. Для insights endpoint нужен другой feature label.

- [ ] **Step 3: Update mapping**

В `backend/internal/delivery/http/analytics/errors.go`:

```go
// respondError — единый мост между доменными ошибками analytics и HTTP.
// ErrMaxRequired (legacy) и ErrProRequired (insights teaser) — оба 402,
// но с разным `plan` в body для корректного upgrade-prompt'а.
func respondError(w http.ResponseWriter, r *http.Request, err error) {
    switch {
    case errors.Is(err, analyticsuc.ErrForbidden):
        httperr.Respond(w, httperr.Forbidden("Нет доступа"))
    case errors.Is(err, analyticsuc.ErrNotFound):
        httperr.Respond(w, httperr.NotFound("Не найдено"))
    case errors.Is(err, analyticsuc.ErrMaxRequired):
        respondTierRequired(w, "insights", "Max")
    case errors.Is(err, analyticsuc.ErrProRequired):
        // С Pricing iteration v3 — insights требуют Pro (teaser). Остальные Pro-only
        // фичи (CSV export) тоже маппятся сюда, feature label берётся из контекста endpoint'а.
        // На уровне error mapping мы не знаем какой именно endpoint — но respondTierRequired
        // не использует feature label для логики, только для error message клиенту.
        // → один универсальный label "premium_feature" покрывает оба случая.
        respondTierRequired(w, "premium_feature", "Pro")
    default:
        httperr.RespondWithRequest(w, r, httperr.Internal(err))
    }
}
```

**Если** existing handler-ы дёргают `respondError` для разных endpoint'ов (export vs insights), и feature label важен — добавь wrapper'ы `respondInsightsError(w,r,err)` и `respondExportError(w,r,err)` каждый с правильным label. Проверь use-sites через `grep -rn "respondError(" backend/internal/delivery/http/analytics/`.

- [ ] **Step 4: Run test — PASS**

Run: `go test -run TestRespondError_InsightsProRequired ./internal/delivery/http/analytics/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/analytics/errors.go \
        backend/internal/delivery/http/analytics/handler_test.go
git commit -m "feat(analytics): HTTP mapping ErrProRequired → 402 для insights teaser"
```

---

### Task 9: Feature flag PRO_INSIGHTS_TEASER_ENABLED + config wiring

**Files:**
- Modify: `backend/internal/infrastructure/config/analytics.go` (add field)
- Modify: `backend/internal/usecases/analytics/service.go` (Service struct + setter)
- Modify: `backend/internal/usecases/analytics/insights.go` (guard insightsForPlan)
- Modify: `backend/internal/app/app.go` (wire flag)
- Modify: `promptvault/.env.example` (document flag)

- [ ] **Step 1: Add config field**

В `backend/internal/infrastructure/config/analytics.go`:

```go
type AnalyticsConfig struct {
    ExperimentalInsights    bool `koanf:"experimental_insights"`
    // ProInsightsTeaserEnabled — Pricing iteration v3 (ADR-0008).
    // При false: GetInsightsGated отдаёт ErrMaxRequired для Pro (legacy поведение).
    // При true: Pro получает teaser из 2 типов (unused + duplicates).
    // Loop insights_loop соответственно либо обрабатывает Max-only, либо Pro+Max.
    // Default false — включить после 1 недели observability после Wave 2 deploy.
    ProInsightsTeaserEnabled bool `koanf:"pro_insights_teaser_enabled"`
}
```

- [ ] **Step 2: Add Service field + setter**

В `backend/internal/usecases/analytics/service.go` — добавить поле в Service:

```go
type Service struct {
    // ... existing fields
    proInsightsTeaserEnabled bool
}

// SetProInsightsTeaserEnabled — Pricing iteration v3 kill-switch.
func (s *Service) SetProInsightsTeaserEnabled(v bool) {
    s.proInsightsTeaserEnabled = v
}
```

- [ ] **Step 3: Guard insightsForPlan**

В `backend/internal/usecases/analytics/insights.go`, изменить `insightsForPlan` на метод Service:

```go
// insightsForPlan возвращает разрешённые типы для plan'а.
// При выключенном teaser-flag Pro обрабатывается как Free (nil → ErrProRequired).
func (s *Service) insightsForPlan(planID string) []string {
    switch planID {
    case "pro", "pro_yearly":
        if !s.proInsightsTeaserEnabled {
            return nil
        }
        return proAllowedInsights
    case "max", "max_yearly":
        return maxAllInsights
    default:
        return nil
    }
}
```

Затем обновить use-sites — в `service.go:GetInsightsGated` и `insights_loop.go` заменить `insightsForPlan(planID)` на `s.insightsForPlan(planID)` / `l.svc.insightsForPlan(...)` (доступно через анличный package).

В `insights_test.go` обновить `TestInsightsForPlan` — теперь это метод Service:

```go
func TestService_insightsForPlan(t *testing.T) {
    cases := []struct {
        plan           string
        teaserEnabled  bool
        want           []string
    }{
        {"pro", true,  []string{models.InsightUnusedPrompts, models.InsightPossibleDuplicates}},
        {"pro", false, nil}, // teaser off → free-like
        {"max", true,  allInsightTypes()},
        {"max", false, allInsightTypes()}, // Max не зависит от flag'а
        {"free", true, nil},
    }
    for _, tc := range cases {
        t.Run(fmt.Sprintf("%s/%v", tc.plan, tc.teaserEnabled), func(t *testing.T) {
            svc := &Service{proInsightsTeaserEnabled: tc.teaserEnabled}
            got := svc.insightsForPlan(tc.plan)
            require.ElementsMatch(t, tc.want, got)
        })
    }
}
```

- [ ] **Step 4: Wire flag в app.go**

В `backend/internal/app/app.go` — после создания `analytics.Service`, добавить:

```go
analyticsService := analyticsuc.NewService(/* ... */)
analyticsService.SetExperimentalInsights(cfg.Analytics.ExperimentalInsights)
analyticsService.SetTrgmAvailable(trgmAvailable)
analyticsService.SetPlanFromCtx(planFromCtxCallback)
analyticsService.SetProInsightsTeaserEnabled(cfg.Analytics.ProInsightsTeaserEnabled)  // NEW
```

- [ ] **Step 5: Update .env.example**

В `promptvault/.env.example` добавить:

```bash
# Pricing iteration v3: включает Smart Insights teaser на Pro (2 типа из 7).
# Default false — включить после 1 недели observability после backend deploy.
ANALYTICS_PRO_INSIGHTS_TEASER_ENABLED=false
```

- [ ] **Step 6: Run tests + build**

```bash
go build ./...
go test -short ./internal/usecases/analytics/... ./internal/infrastructure/config/...
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/infrastructure/config/analytics.go \
        backend/internal/usecases/analytics/service.go \
        backend/internal/usecases/analytics/insights.go \
        backend/internal/usecases/analytics/insights_test.go \
        backend/internal/app/app.go \
        promptvault/.env.example
git commit -m "feat(analytics): feature flag PRO_INSIGHTS_TEASER_ENABLED"
```

---

### Task 10: Frontend three-state insights UI

**Files:**
- Modify: `frontend/src/pages/analytics.tsx:169-183`
- Modify: `frontend/src/hooks/use-analytics.ts:useInsights` (enabled flag)
- Create: `frontend/src/components/analytics/insights-locked-card.tsx` (новый компонент)
- Test: `frontend/src/pages/__tests__/analytics-insights-states.test.tsx`

- [ ] **Step 1: Write failing test**

Создай `frontend/src/pages/__tests__/analytics-insights-states.test.tsx`:

```tsx
import { render, screen } from "@testing-library/react"
import { describe, it, expect, vi } from "vitest"

describe("Analytics insights — three states", () => {
  it("Free → UpgradeGate(Pro)", () => {
    // Mock useAuthStore: { user: { plan_id: "free" } }
    // Render <Analytics />
    expect(screen.getByText(/Подсказки — на тарифе Pro/i)).toBeInTheDocument()
  })

  it("Pro → 2 insights + 5 locked", () => {
    // Mock useAuthStore: { user: { plan_id: "pro" } }
    // Mock useInsights: returns 2 insights (unused + duplicates)
    // Render <Analytics />
    expect(screen.getByText(/Забытые промпты/i)).toBeInTheDocument()
    expect(screen.getByText(/Возможные дубликаты/i)).toBeInTheDocument()
    // Locked cards
    const lockedCards = screen.getAllByText(/Доступно в Max/i)
    expect(lockedCards.length).toBe(5)
  })

  it("Max → 7 insights, нет lock'ов", () => {
    // Mock useAuthStore: { user: { plan_id: "max" } }
    // Mock useInsights: returns 7 insights
    expect(screen.queryByText(/Доступно в Max/i)).not.toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test — FAIL**

Run: `npx vitest run analytics-insights-states`
Expected: FAIL — компонент `InsightsLockedCard` undefined, UI single-state.

- [ ] **Step 3: Create InsightsLockedCard component**

```tsx
// frontend/src/components/analytics/insights-locked-card.tsx
import { Lock } from "lucide-react"
import { Link } from "react-router-dom"

interface Props {
  title: string
  description: string
}

export function InsightsLockedCard({ title, description }: Props) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-muted/20 p-4">
      <div className="mb-2 flex items-center gap-2">
        <Lock className="h-4 w-4 text-muted-foreground" />
        <h3 className="text-sm font-medium text-muted-foreground">{title}</h3>
      </div>
      <p className="mb-3 text-xs text-muted-foreground">{description}</p>
      <Link
        to="/pricing"
        className="text-xs font-medium text-violet-600 hover:underline dark:text-violet-400"
      >
        Доступно в Max →
      </Link>
    </div>
  )
}
```

- [ ] **Step 4: Update useInsights hook**

В `frontend/src/hooks/use-analytics.ts:useInsights`, проверь enabled flag. Поменять с `isMax` на `isPaid`:

```ts
// Было: useInsights(isMax: boolean)
// Стало:
export function useInsights(isPaid: boolean) {
  return useQuery({
    queryKey: ["analytics", "insights"],
    queryFn: () => api<SmartInsight[]>("/analytics/insights"),
    enabled: isPaid,
    staleTime: 5 * 60 * 1000,
  })
}
```

Use-site в `analytics.tsx` (line ~32-34):

```tsx
const planId = (user?.plan_id ?? "free") as PlanID
const isPaid = planId === "pro" || planId === "pro_yearly" || planId === "max" || planId === "max_yearly"
const isMax = planId === "max" || planId === "max_yearly"
const insightsQuery = useInsights(isPaid)
```

- [ ] **Step 5: Replace single-state with three-state UI**

В `frontend/src/pages/analytics.tsx`, заменить блок (line 169-183):

```tsx
// Smart Insights section
{!isPaid && (
  <UpgradeGate
    title="Подсказки — на тарифе Pro"
    description="Забытые промпты и дубликаты помогут навести порядок. Полный набор — в Max."
    targetPlan="Pro"
  />
)}

{isPaid && insightsQuery.isLoading && <Loader2 className="h-6 w-6 animate-spin" />}

{isPaid && insightsQuery.data && (
  <div className="space-y-4">
    <InsightsPanel insights={insightsQuery.data} />

    {/* Pro юзер видит locked-карточки для 5 Max-only типов */}
    {!isMax && (
      <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
        <InsightsLockedCard
          title="Растущая популярность"
          description="Промпты, использование которых выросло за 7 дней."
        />
        <InsightsLockedCard
          title="Падающая популярность"
          description="Промпты, которые перестали активно использоваться."
        />
        <InsightsLockedCard
          title="Самые редактируемые"
          description="Топ промптов по количеству версий."
        />
        <InsightsLockedCard
          title="Теги без промптов"
          description="Orphan-теги для уборки."
        />
        <InsightsLockedCard
          title="Пустые коллекции"
          description="Коллекции без промптов."
        />
      </div>
    )}
  </div>
)}
```

Добавь `import { InsightsLockedCard } from "@/components/analytics/insights-locked-card"` в начало файла.

- [ ] **Step 6: Run test — PASS**

Run: `npx vitest run analytics-insights-states`
Expected: PASS.

- [ ] **Step 7: Manual smoke в браузере**

```bash
docker compose -f docker-compose.dev.yml up -d --build
# Логин как e2e-free@test.local → /analytics → UpgradeGate Pro
# Логин как e2e-pro@test.local → /analytics → 2 insights + 5 locked
# Логин как e2e-max@test.local → /analytics → 7 insights, нет lock'ов
```

(Требуется `ANALYTICS_PRO_INSIGHTS_TEASER_ENABLED=true` на backend.)

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/analytics.tsx \
        frontend/src/hooks/use-analytics.ts \
        frontend/src/components/analytics/insights-locked-card.tsx \
        frontend/src/pages/__tests__/analytics-insights-states.test.tsx
git commit -m "feat(analytics): three-state insights UI (Free/Pro/Max)"
```

---

## Phase 3: Referral Reward (S5-S9)

### Task 11: Migration 000074 — referral_pending_rewards table

**Files:**
- Create: `backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.up.sql`
- Create: `backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- 000074_referral_pending_rewards.up.sql
-- Pricing iteration v3 (ADR-0009): delayed referral reward.
-- На webhook payment.succeeded → INSERT с eligible_at = now() + 14 дней.
-- ReferralRewardLoop ежечасно: SELECT WHERE eligible_at < now → grant + DELETE.
--
-- UNIQUE на referee_id — защита от double-INSERT при retry webhook'ов.
-- Один рефери → одна награда → одна запись в pending (до grant'а).

CREATE TABLE IF NOT EXISTS referral_pending_rewards (
    id          BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referee_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_id  BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    eligible_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_referral_pending_unique_referee
    ON referral_pending_rewards (referee_id);

CREATE INDEX IF NOT EXISTS idx_referral_pending_eligible_at
    ON referral_pending_rewards (eligible_at);
```

- [ ] **Step 2: Write down migration**

```sql
-- 000074_referral_pending_rewards.down.sql
DROP INDEX IF EXISTS idx_referral_pending_eligible_at;
DROP INDEX IF EXISTS idx_referral_pending_unique_referee;
DROP TABLE IF EXISTS referral_pending_rewards;
```

- [ ] **Step 3: Apply migration**

```bash
docker compose -f docker-compose.dev.yml up -d --build
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault -c "\d referral_pending_rewards"
```

Expected: таблица существует со всеми колонками + 2 индекса.

- [ ] **Step 4: Test rollback**

```bash
# Apply down (через migrate CLI или ручной SQL)
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault -c "DROP TABLE referral_pending_rewards CASCADE;"
# Re-apply up:
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U postgres -d promptvault < backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.up.sql
```

Expected: оба раза без ошибок.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.up.sql \
        backend/internal/infrastructure/postgres/migrations/000074_referral_pending_rewards.down.sql
git commit -m "feat(referral): миграция 000074 — referral_pending_rewards"
```

---

### Task 12: ReferralPendingReward model + repo interface + GORM impl

**Files:**
- Create: `backend/internal/models/referral_pending_reward.go`
- Create: `backend/internal/interface/repository/referral_reward.go`
- Create: `backend/internal/infrastructure/postgres/repository/referral_reward_repo.go`
- Create: `backend/internal/infrastructure/postgres/repository/referral_reward_repo_test.go`

- [ ] **Step 1: Write failing integration test**

```go
// backend/internal/infrastructure/postgres/repository/referral_reward_repo_test.go
package repository

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "promptvault/internal/models"
)

func TestReferralRewardRepo_CRUD(t *testing.T) {
    if testing.Short() {
        t.Skip("requires postgres testcontainer")
    }
    ctx := context.Background()
    db := setupTestDB(t)
    repo := NewReferralRewardRepository(db)

    // Сидим 2 юзеров + payment
    referrer := &models.User{Email: "r@t.local", PlanID: "pro", ReferralCode: "REFREF01"}
    referee := &models.User{Email: "e@t.local", PlanID: "pro"}
    require.NoError(t, db.Create(referrer).Error)
    require.NoError(t, db.Create(referee).Error)
    payment := &models.Payment{UserID: referee.ID, AmountKop: 59900, Status: models.PaymentSucceeded,
                                Currency: "RUB", Provider: "tbank", ExternalID: "ext-1", IdempotencyKey: "idem-1"}
    require.NoError(t, db.Create(payment).Error)

    now := time.Now()
    eligibleAt := now.Add(14 * 24 * time.Hour)

    // Create
    pending := &models.ReferralPendingReward{
        ReferrerID: referrer.ID,
        RefereeID:  referee.ID,
        PaymentID:  payment.ID,
        EligibleAt: eligibleAt,
    }
    require.NoError(t, repo.Create(ctx, pending))
    require.NotZero(t, pending.ID)

    // Create idempotent (UNIQUE on referee_id) — повторный INSERT должен вернуть error
    dup := &models.ReferralPendingReward{
        ReferrerID: referrer.ID, RefereeID: referee.ID, PaymentID: payment.ID,
        EligibleAt: eligibleAt,
    }
    err := repo.Create(ctx, dup)
    require.Error(t, err, "должен быть UNIQUE violation на referee_id")

    // FindByReferee
    found, err := repo.FindByReferee(ctx, referee.ID)
    require.NoError(t, err)
    require.Equal(t, pending.ID, found.ID)

    // ListEligible — eligibleAt > now → пусто
    eligible, err := repo.ListEligible(ctx, now, 10)
    require.NoError(t, err)
    require.Empty(t, eligible)

    // Симулируем «прошло 14 дней»
    require.NoError(t, db.Model(pending).Update("eligible_at", now.Add(-1*time.Minute)).Error)

    eligible, err = repo.ListEligible(ctx, now, 10)
    require.NoError(t, err)
    require.Len(t, eligible, 1)

    // Delete
    require.NoError(t, repo.Delete(ctx, pending.ID))
    found, err = repo.FindByReferee(ctx, referee.ID)
    require.Nil(t, found)
    require.NoError(t, err)
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestReferralRewardRepo_CRUD ./internal/infrastructure/postgres/repository/ -v`
Expected: FAIL — `NewReferralRewardRepository` undefined.

- [ ] **Step 3: Create model**

```go
// backend/internal/models/referral_pending_reward.go
package models

import "time"

// ReferralPendingReward — отложенный grant'у пригласившего, ожидающий refund-окна.
// На webhook payment.succeeded субscription создаёт row с eligible_at = now + 14d.
// ReferralRewardLoop через час делает SELECT WHERE eligible_at < now → grant + DELETE.
// UNIQUE на referee_id (одна награда на одного реферри).
type ReferralPendingReward struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    ReferrerID uint      `gorm:"not null" json:"referrer_id"`
    RefereeID  uint      `gorm:"not null;uniqueIndex" json:"referee_id"`
    PaymentID  uint      `gorm:"not null" json:"payment_id"`
    EligibleAt time.Time `gorm:"not null;index" json:"eligible_at"`
    CreatedAt  time.Time `json:"created_at"`
}

func (ReferralPendingReward) TableName() string { return "referral_pending_rewards" }
```

- [ ] **Step 4: Create interface**

```go
// backend/internal/interface/repository/referral_reward.go
package repository

import (
    "context"
    "time"

    "promptvault/internal/models"
)

// ReferralRewardRepository — pending'и для отложенного grant'а реферальной награды.
// Записи живут от webhook'а payment.succeeded до grant'а через 14 дней.
type ReferralRewardRepository interface {
    // Create — INSERT pending. Возвращает error на UNIQUE violation (referee_id).
    Create(ctx context.Context, pending *models.ReferralPendingReward) error
    // ListEligible — SELECT WHERE eligible_at < ts ORDER BY eligible_at LIMIT N.
    ListEligible(ctx context.Context, ts time.Time, limit int) ([]models.ReferralPendingReward, error)
    // FindByReferee — для idempotency check. nil + nil если не найдено.
    FindByReferee(ctx context.Context, refereeID uint) (*models.ReferralPendingReward, error)
    // Delete — после успешного grant'а.
    Delete(ctx context.Context, id uint) error
}
```

- [ ] **Step 5: Create GORM implementation**

```go
// backend/internal/infrastructure/postgres/repository/referral_reward_repo.go
package repository

import (
    "context"
    "errors"
    "time"

    "gorm.io/gorm"

    repo "promptvault/internal/interface/repository"
    "promptvault/internal/models"
)

type referralRewardRepo struct {
    db *gorm.DB
}

func NewReferralRewardRepository(db *gorm.DB) repo.ReferralRewardRepository {
    return &referralRewardRepo{db: db}
}

func (r *referralRewardRepo) Create(ctx context.Context, pending *models.ReferralPendingReward) error {
    return r.db.WithContext(ctx).Create(pending).Error
}

func (r *referralRewardRepo) ListEligible(ctx context.Context, ts time.Time, limit int) ([]models.ReferralPendingReward, error) {
    var rows []models.ReferralPendingReward
    err := r.db.WithContext(ctx).
        Where("eligible_at < ?", ts).
        Order("eligible_at ASC").
        Limit(limit).
        Find(&rows).Error
    return rows, err
}

func (r *referralRewardRepo) FindByReferee(ctx context.Context, refereeID uint) (*models.ReferralPendingReward, error) {
    var row models.ReferralPendingReward
    err := r.db.WithContext(ctx).Where("referee_id = ?", refereeID).First(&row).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &row, nil
}

func (r *referralRewardRepo) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&models.ReferralPendingReward{}, id).Error
}
```

- [ ] **Step 6: Run test — PASS**

Run: `go test -run TestReferralRewardRepo_CRUD ./internal/infrastructure/postgres/repository/ -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/models/referral_pending_reward.go \
        backend/internal/interface/repository/referral_reward.go \
        backend/internal/infrastructure/postgres/repository/referral_reward_repo.go \
        backend/internal/infrastructure/postgres/repository/referral_reward_repo_test.go
git commit -m "feat(referral): ReferralPendingReward model + repository"
```

---

### Task 13: Referral usecase — types + errors

**Files:**
- Create: `backend/internal/usecases/referral/types.go`
- Create: `backend/internal/usecases/referral/errors.go`

- [ ] **Step 1: Create types.go**

```go
// backend/internal/usecases/referral/types.go
package referral

// RewardSummary — результат одного тика ReferralRewardLoop.
// Используется для observability (logging + metrics).
type RewardSummary struct {
    Granted        int
    SkippedRefund  int // payment.refunded к моменту eligibility
    SkippedActive  int // referee subscription уже не active
    SkippedDeleted int // referrer был удалён
    Errors         int // grant fail'нул по другой причине
}

func (s RewardSummary) Total() int {
    return s.Granted + s.SkippedRefund + s.SkippedActive + s.SkippedDeleted + s.Errors
}

// Constants — длительности reward'а и eligibility window.
const (
    // RewardDays — длительность Pro-периода, которым награждаем пригласившего.
    RewardDays = 30
    // EligibilityDays — задержка между первым платежом реферри и grant'ом.
    // Должна превышать T-Bank refund-окно (14 дней), иначе arbitrage-риск.
    EligibilityDays = 14
)
```

- [ ] **Step 2: Create errors.go**

```go
// backend/internal/usecases/referral/errors.go
package referral

import "errors"

// Доменные ошибки. HTTP-маппинг — в delivery/http/referral/errors.go
// (если фича будет иметь endpoint'ы; на MVP — только background grant).
var (
    // ErrAlreadyRewarded — referrer уже получил награду за этого referee
    // (или раньше за другого — referral_rewarded_at != NULL).
    ErrAlreadyRewarded = errors.New("referral: уже награждено")
    // ErrPaymentRefunded — pending был создан, но payment refunded до eligibility.
    ErrPaymentRefunded = errors.New("referral: payment refunded")
    // ErrRefereeInactive — referee subscription больше не active (cancelled/expired).
    ErrRefereeInactive = errors.New("referral: referee inactive")
    // ErrReferrerMissing — referrer был удалён (ON DELETE CASCADE pending удалит, но edge cases).
    ErrReferrerMissing = errors.New("referral: referrer missing")
)
```

- [ ] **Step 3: Build check**

Run: `go build ./internal/usecases/referral/...`
Expected: успешно (package пока без real logic, только types/errors).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/usecases/referral/types.go \
        backend/internal/usecases/referral/errors.go
git commit -m "feat(referral): доменные ошибки + RewardSummary types"
```

---

### Task 14: Service.GrantReward (Pro/Max/Free branches)

**Files:**
- Create: `backend/internal/usecases/referral/reward.go`
- Create: `backend/internal/usecases/referral/reward_test.go`

- [ ] **Step 1: Write failing test (table-driven)**

```go
// backend/internal/usecases/referral/reward_test.go
package referral

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "promptvault/internal/models"
)

func TestService_GrantReward_ProReferrer(t *testing.T) {
    ctx := context.Background()
    now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
    existingEnd := now.Add(15 * 24 * time.Hour)

    sub := &models.Subscription{
        ID: 100, UserID: 10, PlanID: "pro", Status: models.SubStatusActive,
        CurrentPeriodEnd: existingEnd,
    }
    subs := &fakeSubRepo{active: sub}
    users := &fakeUserRepo{user: &models.User{ID: 10, PlanID: "pro", ReferralRewardedAt: nil}}
    pays := &fakePayRepo{paymentStatus: models.PaymentSucceeded}

    svc := NewService(subs, users, pays, nil)
    svc.SetNowFn(func() time.Time { return now })

    err := svc.GrantReward(ctx, 10 /*referrerID*/, 20 /*refereeID*/, 200 /*paymentID*/)
    require.NoError(t, err)

    // Pro-период продлился на 30 дней.
    require.Equal(t, existingEnd.Add(30*24*time.Hour), subs.active.CurrentPeriodEnd)
    // referral_rewarded_at установлен.
    require.True(t, users.markRewardedCalled)
}

func TestService_GrantReward_FreeReferrer(t *testing.T) {
    ctx := context.Background()
    now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

    users := &fakeUserRepo{user: &models.User{ID: 10, PlanID: "free"}}
    subs := &fakeSubRepo{active: nil} // no existing subscription
    pays := &fakePayRepo{paymentStatus: models.PaymentSucceeded}

    svc := NewService(subs, users, pays, nil)
    svc.SetNowFn(func() time.Time { return now })

    err := svc.GrantReward(ctx, 10, 20, 200)
    require.NoError(t, err)

    // Создалась trial subscription.
    require.NotNil(t, subs.createdSub)
    require.Equal(t, "pro", subs.createdSub.PlanID)
    require.False(t, subs.createdSub.AutoRenew)
    require.Equal(t, "", subs.createdSub.RebillId)
    require.Equal(t, now.Add(30*24*time.Hour), subs.createdSub.CurrentPeriodEnd)
    // users.plan_id обновлён на pro.
    require.True(t, users.setPlanCalled)
    require.Equal(t, "pro", users.setPlanID)
}

func TestService_GrantReward_MaxReferrer(t *testing.T) {
    ctx := context.Background()
    now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
    existingEnd := now.Add(15 * 24 * time.Hour)

    sub := &models.Subscription{
        ID: 100, UserID: 10, PlanID: "max", Status: models.SubStatusActive,
        CurrentPeriodEnd: existingEnd,
    }
    subs := &fakeSubRepo{active: sub}
    users := &fakeUserRepo{user: &models.User{ID: 10, PlanID: "max"}}
    pays := &fakePayRepo{paymentStatus: models.PaymentSucceeded}

    svc := NewService(subs, users, pays, nil)
    svc.SetNowFn(func() time.Time { return now })

    err := svc.GrantReward(ctx, 10, 20, 200)
    require.NoError(t, err)

    // Max-период продлился на 30 дней (не downgrade в Pro!).
    require.Equal(t, existingEnd.Add(30*24*time.Hour), subs.active.CurrentPeriodEnd)
    require.Equal(t, "max", subs.active.PlanID)
}

func TestService_GrantReward_Idempotent(t *testing.T) {
    ctx := context.Background()
    rewardedAt := time.Now()
    users := &fakeUserRepo{user: &models.User{ID: 10, PlanID: "pro", ReferralRewardedAt: &rewardedAt}}
    svc := NewService(nil, users, nil, nil)

    err := svc.GrantReward(ctx, 10, 20, 200)
    require.ErrorIs(t, err, ErrAlreadyRewarded)
}

func TestService_GrantReward_PaymentRefunded(t *testing.T) {
    ctx := context.Background()
    users := &fakeUserRepo{user: &models.User{ID: 10, PlanID: "pro"}}
    pays := &fakePayRepo{paymentStatus: models.PaymentRefunded}
    svc := NewService(nil, users, pays, nil)

    err := svc.GrantReward(ctx, 10, 20, 200)
    require.ErrorIs(t, err, ErrPaymentRefunded)
}

// fakes — простейшие реализации; полные fakes можно вынести в test helpers
type fakeSubRepo struct {
    active       *models.Subscription
    createdSub   *models.Subscription
}

func (f *fakeSubRepo) GetActiveByUserID(ctx context.Context, userID uint) (*models.Subscription, error) {
    if f.active == nil {
        return nil, nil
    }
    return f.active, nil
}
func (f *fakeSubRepo) Create(ctx context.Context, sub *models.Subscription) error {
    f.createdSub = sub
    return nil
}
func (f *fakeSubRepo) UpdatePeriodEnd(ctx context.Context, subID uint, periodEnd time.Time) error {
    f.active.CurrentPeriodEnd = periodEnd
    return nil
}

type fakeUserRepo struct {
    user                *models.User
    markRewardedCalled  bool
    setPlanCalled       bool
    setPlanID           string
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id uint) (*models.User, error) { return f.user, nil }
func (f *fakeUserRepo) MarkReferralRewarded(ctx context.Context, userID uint) (bool, error) {
    f.markRewardedCalled = true
    return true, nil
}
func (f *fakeUserRepo) SetPlan(ctx context.Context, userID uint, planID string) error {
    f.setPlanCalled = true
    f.setPlanID = planID
    return nil
}

type fakePayRepo struct {
    paymentStatus models.PaymentStatus
}

func (f *fakePayRepo) GetByID(ctx context.Context, id uint) (*models.Payment, error) {
    return &models.Payment{ID: id, Status: f.paymentStatus}, nil
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestService_GrantReward ./internal/usecases/referral/ -v`
Expected: FAIL — `NewService` undefined.

- [ ] **Step 3: Implement Service + GrantReward**

```go
// backend/internal/usecases/referral/reward.go
package referral

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    repo "promptvault/internal/interface/repository"
    "promptvault/internal/models"
)

// Service — grant'ит реферальные награды. Вызывается из ReferralRewardLoop
// (для pending'ов с истёкшим eligible_at) и потенциально вручную из admin-инструмента.
type Service struct {
    subs    repo.SubscriptionRepository
    users   repo.UserRepository
    pays    repo.PaymentRepository
    pending repo.ReferralRewardRepository
    nowFn   func() time.Time
}

func NewService(
    subs repo.SubscriptionRepository,
    users repo.UserRepository,
    pays repo.PaymentRepository,
    pending repo.ReferralRewardRepository,
) *Service {
    return &Service{
        subs:    subs,
        users:   users,
        pays:    pays,
        pending: pending,
        nowFn:   time.Now,
    }
}

// SetNowFn — для unit-тестов.
func (s *Service) SetNowFn(fn func() time.Time) { s.nowFn = fn }

// GrantReward выдаёт +30 дней Pro пригласившему. Идемпотентно через
// users.referral_rewarded_at (атомарный CAS через MarkReferralRewarded).
//
// Ветки:
//   - referrer на Pro/Pro_yearly — продлеваем current_period_end на 30 дней.
//   - referrer на Max/Max_yearly — продлеваем current_period_end Max на 30 дней (его tier выше).
//   - referrer на Free — создаём trial Subscription{plan:pro, 30d, auto_renew:false, rebill_id:""}.
//
// Препроверки: refunded payment → ErrPaymentRefunded; already rewarded → ErrAlreadyRewarded.
func (s *Service) GrantReward(ctx context.Context, referrerID, refereeID, paymentID uint) error {
    referrer, err := s.users.GetByID(ctx, referrerID)
    if err != nil {
        return fmt.Errorf("get referrer: %w", err)
    }
    if referrer == nil {
        return ErrReferrerMissing
    }
    if referrer.ReferralRewardedAt != nil {
        return ErrAlreadyRewarded
    }

    // Refund check — payment больше не valid?
    payment, err := s.pays.GetByID(ctx, paymentID)
    if err != nil {
        return fmt.Errorf("get payment: %w", err)
    }
    if payment == nil || payment.Status != models.PaymentSucceeded {
        return ErrPaymentRefunded
    }

    now := s.nowFn()
    rewardDuration := time.Duration(RewardDays) * 24 * time.Hour

    switch referrer.PlanID {
    case "max", "max_yearly":
        // Extend Max.
        if err := s.extendActiveSubscription(ctx, referrerID, rewardDuration); err != nil {
            return fmt.Errorf("extend max: %w", err)
        }
    case "pro", "pro_yearly":
        // Extend Pro.
        if err := s.extendActiveSubscription(ctx, referrerID, rewardDuration); err != nil {
            return fmt.Errorf("extend pro: %w", err)
        }
    default:
        // Free → create trial Pro Subscription.
        if err := s.createTrialPro(ctx, referrerID, now.Add(rewardDuration)); err != nil {
            return fmt.Errorf("create trial pro: %w", err)
        }
    }

    // Mark rewarded (atomic CAS).
    ok, err := s.users.MarkReferralRewarded(ctx, referrerID)
    if err != nil {
        return fmt.Errorf("mark rewarded: %w", err)
    }
    if !ok {
        // Race: кто-то другой уже наградил → не fatal, но логируем.
        slog.WarnContext(ctx, "referral.reward.race", "referrer_id", referrerID)
        return ErrAlreadyRewarded
    }

    slog.InfoContext(ctx, "referral.reward.granted",
        "referrer_id", referrerID, "referee_id", refereeID,
        "from_plan", referrer.PlanID, "reward_days", RewardDays,
        "is_trial", referrer.PlanID == "free")

    return nil
}

func (s *Service) extendActiveSubscription(ctx context.Context, userID uint, duration time.Duration) error {
    sub, err := s.subs.GetActiveByUserID(ctx, userID)
    if err != nil {
        return err
    }
    if sub == nil {
        return fmt.Errorf("no active subscription for user %d", userID)
    }
    newEnd := sub.CurrentPeriodEnd.Add(duration)
    return s.subs.UpdatePeriodEnd(ctx, sub.ID, newEnd)
}

func (s *Service) createTrialPro(ctx context.Context, userID uint, periodEnd time.Time) error {
    sub := &models.Subscription{
        UserID:             userID,
        PlanID:             "pro",
        Status:             models.SubStatusActive,
        CurrentPeriodStart: s.nowFn(),
        CurrentPeriodEnd:   periodEnd,
        RebillId:           "",     // trial без T-Bank
        AutoRenew:          false,  // critical: renewal_loop skip
    }
    if err := s.subs.Create(ctx, sub); err != nil {
        return err
    }
    return s.users.SetPlan(ctx, userID, "pro")
}
```

**Внимание:** `SubscriptionRepository` должен иметь методы `GetActiveByUserID`, `UpdatePeriodEnd`, `Create`. Если каких-то нет — добавь в interface перед коммитом. `PaymentRepository.GetByID(ctx, uint) (*models.Payment, error)` — проверить существует ли. Если нет — добавь.

- [ ] **Step 4: Run test — PASS**

Run: `go test -run TestService_GrantReward ./internal/usecases/referral/ -v`
Expected: все 5 case'ов PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/referral/reward.go \
        backend/internal/usecases/referral/reward_test.go
git commit -m "feat(referral): Service.GrantReward — Pro/Max/Free branches + idempotency"
```

---

### Task 15: ReferralRewardLoop

**Files:**
- Create: `backend/internal/usecases/referral/reward_loop.go`

- [ ] **Step 1: Write failing test**

В `reward_test.go` добавь:

```go
func TestRewardLoop_TickGrantsEligible(t *testing.T) {
    ctx := context.Background()
    now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

    pending := &models.ReferralPendingReward{
        ID: 1, ReferrerID: 10, RefereeID: 20, PaymentID: 200,
        EligibleAt: now.Add(-1 * time.Hour),
    }
    pendingRepo := &fakePendingRepo{eligible: []models.ReferralPendingReward{*pending}}
    users := &fakeUserRepo{user: &models.User{ID: 10, PlanID: "free"}}
    subs := &fakeSubRepo{active: nil}
    pays := &fakePayRepo{paymentStatus: models.PaymentSucceeded}

    svc := NewService(subs, users, pays, pendingRepo)
    svc.SetNowFn(func() time.Time { return now })
    loop := NewRewardLoop(svc, pendingRepo, time.Minute, 100)
    loop.SetNowFn(func() time.Time { return now })

    summary := loop.tickOnce(ctx)
    require.Equal(t, 1, summary.Granted)
    require.True(t, pendingRepo.deleted)
}

type fakePendingRepo struct {
    eligible []models.ReferralPendingReward
    deleted  bool
}

func (f *fakePendingRepo) Create(ctx context.Context, p *models.ReferralPendingReward) error { return nil }
func (f *fakePendingRepo) ListEligible(ctx context.Context, ts time.Time, limit int) ([]models.ReferralPendingReward, error) {
    return f.eligible, nil
}
func (f *fakePendingRepo) FindByReferee(ctx context.Context, refereeID uint) (*models.ReferralPendingReward, error) {
    return nil, nil
}
func (f *fakePendingRepo) Delete(ctx context.Context, id uint) error { f.deleted = true; return nil }
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestRewardLoop_TickGrantsEligible ./internal/usecases/referral/ -v`
Expected: FAIL — `NewRewardLoop` undefined.

- [ ] **Step 3: Implement loop**

```go
// backend/internal/usecases/referral/reward_loop.go
package referral

import (
    "context"
    "errors"
    "log/slog"
    "time"

    repo "promptvault/internal/interface/repository"
    "promptvault/internal/pkg/safeloop"
)

// RewardLoop — background обработчик pending'ов. Ежечасно SELECT'ит
// eligible_at < now → вызывает GrantReward → удаляет row.
//
// Паттерн скопирован с subscription.RenewalLoop (safeloop + ticker + stop chan).
type RewardLoop struct {
    svc      *Service
    pending  repo.ReferralRewardRepository
    interval time.Duration
    batch    int
    nowFn    func() time.Time
    stopCh   chan struct{}
}

func NewRewardLoop(svc *Service, pending repo.ReferralRewardRepository, interval time.Duration, batch int) *RewardLoop {
    return &RewardLoop{
        svc:      svc,
        pending:  pending,
        interval: interval,
        batch:    batch,
        nowFn:    time.Now,
        stopCh:   make(chan struct{}),
    }
}

func (l *RewardLoop) SetNowFn(fn func() time.Time) { l.nowFn = fn }

func (l *RewardLoop) Start() {
    slog.Info("referral.reward.loop_started", "interval", l.interval, "batch", l.batch)
    go l.run()
}

func (l *RewardLoop) Stop() { close(l.stopCh) }

func (l *RewardLoop) run() {
    ticker := time.NewTicker(l.interval)
    defer ticker.Stop()
    safeloop.RunWithRecover("referral_reward", func() { _ = l.tickOnce(context.Background()) })
    for {
        select {
        case <-ticker.C:
            safeloop.RunWithRecover("referral_reward", func() { _ = l.tickOnce(context.Background()) })
        case <-l.stopCh:
            slog.Info("referral.reward.loop_stopped")
            return
        }
    }
}

// tickOnce — обрабатывает один batch eligible pending'ов. Вынесена для тестов.
func (l *RewardLoop) tickOnce(ctx context.Context) RewardSummary {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    var summary RewardSummary
    now := l.nowFn()
    pendings, err := l.pending.ListEligible(ctx, now, l.batch)
    if err != nil {
        slog.ErrorContext(ctx, "referral.reward.list_failed", "err", err)
        return summary
    }
    for _, p := range pendings {
        if err := l.svc.GrantReward(ctx, p.ReferrerID, p.RefereeID, p.PaymentID); err != nil {
            switch {
            case errors.Is(err, ErrAlreadyRewarded):
                summary.SkippedActive++
                slog.InfoContext(ctx, "referral.reward.skipped_already_rewarded",
                    "referrer_id", p.ReferrerID, "referee_id", p.RefereeID)
            case errors.Is(err, ErrPaymentRefunded):
                summary.SkippedRefund++
                slog.InfoContext(ctx, "referral.reward.skipped_refunded",
                    "referrer_id", p.ReferrerID, "referee_id", p.RefereeID, "payment_id", p.PaymentID)
            case errors.Is(err, ErrReferrerMissing):
                summary.SkippedDeleted++
            default:
                summary.Errors++
                slog.ErrorContext(ctx, "referral.reward.grant_failed",
                    "err", err, "referrer_id", p.ReferrerID, "referee_id", p.RefereeID)
                continue // не удаляем row — retry на следующем тике
            }
        } else {
            summary.Granted++
        }
        // Delete pending row на success ИЛИ на терминальном skip (already_rewarded/refunded/deleted).
        if err := l.pending.Delete(ctx, p.ID); err != nil {
            slog.ErrorContext(ctx, "referral.reward.delete_failed", "err", err, "pending_id", p.ID)
        }
    }
    if summary.Total() > 0 {
        slog.InfoContext(ctx, "referral.reward.tick_summary",
            "granted", summary.Granted, "skipped_refund", summary.SkippedRefund,
            "skipped_active", summary.SkippedActive, "skipped_deleted", summary.SkippedDeleted,
            "errors", summary.Errors)
    }
    return summary
}
```

- [ ] **Step 4: Run test — PASS**

Run: `go test ./internal/usecases/referral/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecases/referral/reward_loop.go \
        backend/internal/usecases/referral/reward_test.go
git commit -m "feat(referral): ReferralRewardLoop — 1h tick + safeloop + batch processing"
```

---

### Task 16: Subscription webhook — INSERT pending в activateSubscription

**Files:**
- Modify: `backend/internal/usecases/subscription/subscription.go` (insert после ActivateWithPlanUpdate)
- Modify: `backend/internal/usecases/subscription/types.go` (или где Service struct) — добавить поля
- Test: `backend/internal/usecases/subscription/webhook_scenarios_test.go` (новый case)

- [ ] **Step 1: Write failing test**

В `backend/internal/usecases/subscription/webhook_scenarios_test.go` добавь:

```go
func TestHandleWebhook_FirstPayment_InsertsReferralPending(t *testing.T) {
    ctx := context.Background()
    referrer := &models.User{ID: 100, Email: "r@t.local", PlanID: "pro", ReferralCode: "REFREF01", ReferralRewardedAt: nil}
    referee := &models.User{ID: 200, Email: "e@t.local", PlanID: "free", ReferredBy: "REFREF01"}
    plan := &models.SubscriptionPlan{ID: "pro", PriceKop: 59900, PeriodDays: 30}
    payment := &models.Payment{ID: 1, UserID: 200, AmountKop: 59900, Status: models.PaymentPending,
                                ExternalID: "ext-1", IdempotencyKey: "idem-1", Provider: "tbank"}

    subs := &fakeSubRepoFull{}
    plans := &fakePlanRepo{plan: plan}
    pays := &fakePayRepo{payment: payment}
    users := &fakeUserRepoWithReferral{referrer: referrer, referee: referee}
    pending := &fakePendingRepo{}

    svc := NewService(subs, plans, pays, users, /* payment provider */ nil, &config.PaymentConfig{Enabled: true})
    svc.SetReferralPendingRepo(pending)
    svc.SetReferralRewardEnabled(true)
    svc.SetNowFn(func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) })

    // Симулируем CONFIRMED webhook от T-Bank.
    err := svc.HandleWebhook(ctx, "tbank", map[string]string{
        "Status":     "CONFIRMED",
        "PaymentId":  "ext-1",
        "OrderId":    "ord-1",
        "Amount":     "59900",
        "RebillId":   "rebill-xyz",
    })
    require.NoError(t, err)

    // Pending row создан с eligible_at = now + 14d
    require.NotNil(t, pending.created)
    require.Equal(t, uint(100), pending.created.ReferrerID)
    require.Equal(t, uint(200), pending.created.RefereeID)
    require.Equal(t, uint(1), pending.created.PaymentID)
    require.Equal(t,
        time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
        pending.created.EligibleAt)
}

func TestHandleWebhook_NoReferredBy_NoPending(t *testing.T) {
    // referee без referred_by → не INSERT'им
}

func TestHandleWebhook_ReferrerAlreadyRewarded_NoPending(t *testing.T) {
    // referrer.ReferralRewardedAt != nil → skip INSERT
}

func TestHandleWebhook_FlagDisabled_NoPending(t *testing.T) {
    // ReferralRewardEnabled=false → skip
}
```

- [ ] **Step 2: Run test — FAIL**

Run: `go test -run TestHandleWebhook_FirstPayment_InsertsReferralPending ./internal/usecases/subscription/ -v`
Expected: FAIL — `SetReferralPendingRepo` undefined.

- [ ] **Step 3: Add fields to subscription.Service**

В `backend/internal/usecases/subscription/subscription.go` (или types.go где Service struct):

```go
type Service struct {
    // ... existing fields
    referralPending       repo.ReferralRewardRepository
    referralRewardEnabled bool
}

// SetReferralPendingRepo + SetReferralRewardEnabled wired в app.go.
func (s *Service) SetReferralPendingRepo(r repo.ReferralRewardRepository) {
    s.referralPending = r
}
func (s *Service) SetReferralRewardEnabled(v bool) {
    s.referralRewardEnabled = v
}
```

- [ ] **Step 4: Implement INSERT в activateSubscription**

В `subscription.go` (после успешной `s.subs.ActivateWithPlanUpdate(...)`, line ~493):

```go
// После activate, до return:
if s.referralRewardEnabled && s.referralPending != nil {
    s.tryRecordReferralPending(ctx, pay)
}
```

Добавь helper:

```go
// tryRecordReferralPending — на первом успешном платеже referee создаёт pending
// для отложенного grant'а пригласившему. Errors логируются, не возвращаются —
// фейл pending'а не должен ломать активацию подписки.
func (s *Service) tryRecordReferralPending(ctx context.Context, pay *models.Payment) {
    referee, err := s.users.GetByID(ctx, pay.UserID)
    if err != nil || referee == nil {
        return
    }
    if referee.ReferredBy == "" {
        return // не реферал
    }

    referrer, err := s.users.GetByReferralCode(ctx, referee.ReferredBy)
    if err != nil || referrer == nil {
        return
    }
    if referrer.ReferralRewardedAt != nil {
        return // уже наградили
    }

    // Idempotency — может ли быть уже pending row на этого referee?
    existing, err := s.referralPending.FindByReferee(ctx, referee.ID)
    if err != nil {
        slog.WarnContext(ctx, "referral.pending.find_failed", "err", err, "referee_id", referee.ID)
        return
    }
    if existing != nil {
        return // already recorded
    }

    eligibleAt := s.nowFn().Add(time.Duration(referral.EligibilityDays) * 24 * time.Hour)
    pending := &models.ReferralPendingReward{
        ReferrerID: referrer.ID,
        RefereeID:  referee.ID,
        PaymentID:  pay.ID,
        EligibleAt: eligibleAt,
    }
    if err := s.referralPending.Create(ctx, pending); err != nil {
        // UNIQUE violation — race с concurrent webhook'ом; не фатально.
        slog.WarnContext(ctx, "referral.pending.insert_failed",
            "err", err, "referrer_id", referrer.ID, "referee_id", referee.ID)
        return
    }
    slog.InfoContext(ctx, "referral.pending.inserted",
        "referrer_id", referrer.ID, "referee_id", referee.ID,
        "payment_id", pay.ID, "eligible_at", eligibleAt)
}
```

Добавь imports: `"promptvault/internal/usecases/referral"` (для константы `EligibilityDays`).

- [ ] **Step 5: Run test — PASS**

Run: `go test -run TestHandleWebhook_FirstPayment_InsertsReferralPending ./internal/usecases/subscription/ -v`
Expected: PASS. Также добавь остальные 3 case'а (no referred_by / already rewarded / flag disabled) и убедись что они PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/usecases/subscription/subscription.go \
        backend/internal/usecases/subscription/webhook_scenarios_test.go
git commit -m "feat(referral): subscription webhook → INSERT pending reward на первом платеже"
```

---

### Task 17: App wire-up + REFERRAL_REWARD_ENABLED config

**Files:**
- Modify: `backend/internal/infrastructure/config/config.go` (добавить `Referral` секцию)
- Create: `backend/internal/infrastructure/config/referral.go`
- Modify: `backend/internal/app/app.go` (wire repos + loop)
- Modify: `backend/internal/app/lifecycle.go` (start/stop)
- Modify: `promptvault/.env.example`

- [ ] **Step 1: Create ReferralConfig**

```go
// backend/internal/infrastructure/config/referral.go
package config

// ReferralConfig — Pricing iteration v3 (ADR-0009).
// RewardEnabled включает:
//   1) webhook subscription → INSERT в referral_pending_rewards
//   2) ReferralRewardLoop ежечасно SELECT'ит eligible_at < now и grant'ит +30 дней Pro
// Default false — включить после 1 недели QA после backend deploy.
type ReferralConfig struct {
    RewardEnabled bool `koanf:"reward_enabled"`
}
```

- [ ] **Step 2: Wire в Config struct**

В `config.go`:

```go
type Config struct {
    // ... existing
    Referral ReferralConfig `koanf:"referral"`
}
```

- [ ] **Step 3: Wire repos + loop в app.go**

В `backend/internal/app/app.go` после создания других repos:

```go
referralPendingRepo := postgresrepo.NewReferralRewardRepository(db)
```

После создания `subscription.Service`:

```go
subscriptionService.SetReferralPendingRepo(referralPendingRepo)
subscriptionService.SetReferralRewardEnabled(cfg.Referral.RewardEnabled)
```

Создание Loop:

```go
referralRewardService := referraluc.NewService(
    subscriptionRepo, userRepo, paymentRepo, referralPendingRepo,
)
referralRewardLoop := referraluc.NewRewardLoop(
    referralRewardService,
    referralPendingRepo,
    pick(1*time.Hour, 1*time.Minute), // interval (1h prod, 1m dev)
    100, // batch
)
```

Сохрани `referralRewardLoop *referraluc.RewardLoop` в App struct.

- [ ] **Step 4: Start/Stop в lifecycle.go**

В `StartBackground()`:

```go
if a.cfg.Referral.RewardEnabled {
    a.referralRewardLoop.Start()
}
```

В `Shutdown(timeout)`:

```go
if a.cfg.Referral.RewardEnabled {
    a.referralRewardLoop.Stop()
}
```

- [ ] **Step 5: Update .env.example**

```bash
# Pricing iteration v3: реферальная награда +30 дней Pro пригласившему.
# Включает webhook INSERT + ReferralRewardLoop. Default false — включить
# после 1 недели QA после deploy.
REFERRAL_REWARD_ENABLED=false
```

- [ ] **Step 6: Build + smoke**

```bash
go build ./...
docker compose -f docker-compose.dev.yml up -d --build
docker compose -f docker-compose.dev.yml logs api | grep "referral.reward.loop"
```

С `REFERRAL_REWARD_ENABLED=true` ожидаем `referral.reward.loop_started`. С `false` — отсутствие лога.

- [ ] **Step 7: Manual smoke с INSERT**

```bash
# Активируй loop с быстрым interval (1m) для теста
# В БД: ручной INSERT eligible_at = NOW() - 1 minute
docker compose -f docker-compose.dev.yml exec postgres psql -U postgres -d promptvault -c "
INSERT INTO referral_pending_rewards (referrer_id, referee_id, payment_id, eligible_at)
VALUES (1, 2, 1, NOW() - INTERVAL '1 minute');"

# Жди 1 минуту, проверь:
docker compose -f docker-compose.dev.yml exec postgres psql -U postgres -d promptvault -c "
SELECT * FROM referral_pending_rewards;
SELECT id, plan_id, referral_rewarded_at FROM users WHERE id=1;"
```

Expected: pending row deleted, users.referral_rewarded_at != NULL.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/infrastructure/config/config.go \
        backend/internal/infrastructure/config/referral.go \
        backend/internal/app/app.go \
        backend/internal/app/lifecycle.go \
        promptvault/.env.example
git commit -m "feat(referral): wire-up ReferralRewardLoop + REFERRAL_REWARD_ENABLED flag"
```

---

## Phase 4: Observability + Docs

### Task 18: Prometheus metrics

**Files:**
- Modify: `backend/internal/infrastructure/metrics/metrics.go` (add counters + zero-init)
- Modify: `backend/internal/usecases/referral/reward_loop.go` (увеличивать)
- Modify: `backend/internal/usecases/referral/reward.go` (увеличивать на grant)
- Modify: `backend/internal/usecases/subscription/subscription.go` (увеличивать на INSERT pending)
- Modify: `backend/internal/usecases/analytics/service.go` (увеличивать на gated call)

- [ ] **Step 1: Add counters в metrics.go**

В `backend/internal/infrastructure/metrics/metrics.go`:

```go
var (
    // ... existing metrics

    ReferralRewardsPendingTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "referral_rewards_pending_total",
        Help: "Сколько pending'ов создаётся на webhook payment.succeeded",
    }, []string{"result"}) // recorded | skipped_already_rewarded | skipped_no_referrer

    ReferralRewardsGrantedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "referral_rewards_granted_total",
        Help: "Успешные grant'ы реферальных наград",
    }, []string{"referrer_plan"}) // free | pro | max

    ReferralRewardsSkippedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "referral_rewards_skipped_total",
        Help: "Pending'и которые не дали reward'а",
    }, []string{"reason"}) // refunded | already_rewarded | referrer_deleted

    AnalyticsInsightsGatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "analytics_insights_gated_total",
        Help: "Эффективность teaser'а на /api/analytics/insights",
    }, []string{"plan", "result"}) // result: full|partial|blocked
)

func init() {
    // ... existing zero-inits

    // Zero-init для всех expected combinations (защита от absent_over_time false-positives).
    for _, r := range []string{"recorded", "skipped_already_rewarded", "skipped_no_referrer"} {
        ReferralRewardsPendingTotal.WithLabelValues(r).Add(0)
    }
    for _, p := range []string{"free", "pro", "max"} {
        ReferralRewardsGrantedTotal.WithLabelValues(p).Add(0)
    }
    for _, r := range []string{"refunded", "already_rewarded", "referrer_deleted"} {
        ReferralRewardsSkippedTotal.WithLabelValues(r).Add(0)
    }
    for _, p := range []string{"free", "pro", "max"} {
        for _, r := range []string{"full", "partial", "blocked"} {
            AnalyticsInsightsGatedTotal.WithLabelValues(p, r).Add(0)
        }
    }
}
```

- [ ] **Step 2: Wire в GrantReward (reward.go)**

После успешного grant'а:

```go
// В конце GrantReward, перед return nil:
plan := normalizePlanLabel(referrer.PlanID) // "free" / "pro" / "max"
metrics.ReferralRewardsGrantedTotal.WithLabelValues(plan).Inc()
```

`normalizePlanLabel` — helper маппит `pro_yearly→pro`, `max_yearly→max`.

В error-branches:

```go
case errors.Is(err, ErrAlreadyRewarded):
    metrics.ReferralRewardsSkippedTotal.WithLabelValues("already_rewarded").Inc()
case errors.Is(err, ErrPaymentRefunded):
    metrics.ReferralRewardsSkippedTotal.WithLabelValues("refunded").Inc()
case errors.Is(err, ErrReferrerMissing):
    metrics.ReferralRewardsSkippedTotal.WithLabelValues("referrer_deleted").Inc()
```

- [ ] **Step 3: Wire в subscription tryRecordReferralPending**

```go
// recorded:
metrics.ReferralRewardsPendingTotal.WithLabelValues("recorded").Inc()
// skipped_already_rewarded:
if referrer.ReferralRewardedAt != nil {
    metrics.ReferralRewardsPendingTotal.WithLabelValues("skipped_already_rewarded").Inc()
    return
}
// skipped_no_referrer:
if referee.ReferredBy == "" {
    metrics.ReferralRewardsPendingTotal.WithLabelValues("skipped_no_referrer").Inc()
    return
}
```

- [ ] **Step 4: Wire в GetInsightsGated**

```go
// В service.go:
allowed := s.insightsForPlan(planID)
plan := normalizePlanLabel(planID)
if len(allowed) == 0 {
    metrics.AnalyticsInsightsGatedTotal.WithLabelValues(plan, "blocked").Inc()
    return nil, ErrProRequired
}
result := "full"
if len(allowed) < len(maxAllInsights) {
    result = "partial"
}
metrics.AnalyticsInsightsGatedTotal.WithLabelValues(plan, result).Inc()
// ... continue with filter
```

- [ ] **Step 5: Verify через /metrics endpoint**

```bash
docker compose -f docker-compose.dev.yml up -d --build
curl -s http://localhost:8080/metrics | grep -E "referral_rewards|analytics_insights_gated"
```

Expected: все zero-init series видны.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/infrastructure/metrics/metrics.go \
        backend/internal/usecases/referral/reward.go \
        backend/internal/usecases/referral/reward_loop.go \
        backend/internal/usecases/subscription/subscription.go \
        backend/internal/usecases/analytics/service.go
git commit -m "feat(observability): метрики referral_rewards_* + analytics_insights_gated_total"
```

---

### Task 19: ADR-0008, ADR-0009, Runbook, CLAUDE.md upd

**Files:**
- Create: `docs/adr/0008-pertype-insights-gate.md`
- Create: `docs/adr/0009-delayed-referral-reward.md`
- Create: `docs/runbooks/ReferralRewardLoopStalled.md`
- Modify: `promptvault/CLAUDE.md` (Ключевые решения секция)

- [ ] **Step 1: ADR-0008**

```markdown
<!-- docs/adr/0008-pertype-insights-gate.md -->
# ADR-0008: Per-type Smart Insights gating (Pro teaser)

**Дата:** 2026-05-17
**Статус:** Принято

## Контекст
Pricing iteration v3 требует чтобы Pro юзеры видели **2 типа** Smart Insights
(unused + duplicates), а Max — все 7. До этой итерации использовался master-gate
`IsMax(planID)` (analytics/service.go:111) — Free/Pro получали 402, Max — всё.

## Решение
Per-type filter через hardcoded `proAllowedInsights []string` константу в
`usecases/analytics`. `insightsForPlan(planID)` возвращает разрешённый набор;
`GetInsightsGated` filter'ит результат `GetInsights` в памяти; `ComputeInsights`
принимает `allowed []string` параметр и пропускает SQL для не-allowed типов.

## Альтернативы
- **JSONB-колонка `subscription_plans.allowed_insight_types TEXT[]`** — гибкость,
  можно менять без релиза. Минус: overengineering для 3 tier'ов.
- **Env feature flag `PRO_INSIGHTS_TYPES=unused,duplicates`** — можно менять без
  релиза. Минус: конфигурация state в env vs БД — антипаттерн.
- **Const в коде (выбран)** — 1 место правды, типобезопасный.

## Trade-offs
✅ Простой, читаемый.
✅ Тестируется table-driven case'ами.
❌ Добавление 4-го tier'а требует кодовых правок (acceptable — 3 tier'а 1.5 года).

## Импликации
- Feature flag `PRO_INSIGHTS_TEASER_ENABLED` — для kill-switch в первую неделю
  после deploy.
- `ListMaxUsers` репозиторий переименован в `ListPaidUsers` (включает Pro).
- Loop делает extra-roundtrip в `users.GetByID(uid)` для определения `allowed`.
```

- [ ] **Step 2: ADR-0009**

```markdown
<!-- docs/adr/0009-delayed-referral-reward.md -->
# ADR-0009: Delayed referral reward через cron-loop

**Дата:** 2026-05-17
**Статус:** Принято

## Контекст
Pricing iteration v3 вводит реферальную награду: +30 дней Pro пригласившему
после **первого платежа** реферри. Существуют 3 паттерна реализации.

## Решение
**Delayed cron-loop** (паттерн `subscription.RenewalLoop`). На webhook
`payment.succeeded` → INSERT в `referral_pending_rewards` с
`eligible_at = now() + 14 дней`. Ежечасный `ReferralRewardLoop` SELECT'ит
`eligible_at < now()`, проверяет что payment всё ещё `succeeded` (не refunded),
вызывает `GrantReward`, удаляет row.

Защита от refund-арбитража: 14 дней превышает T-Bank refund window.
Идемпотентность: UNIQUE constraint на `(referee_id)` + `users.referral_rewarded_at`
атомарный CAS через `MarkReferralRewarded`.

## Альтернативы
- **Synchronous on webhook + revoke on refund** — race conditions, реверс сложен
  (если рефер уже потратил бонусные дни).
- **Lazy при следующем webhook** — зависит от 2-го платежа того же юзера,
  может никогда не произойти.

## Trade-offs
✅ Защита от refund-арбитража.
✅ Идемпотентность через UNIQUE.
✅ Reuse `safeloop.RunWithRecover` паттерна.
❌ Новая таблица (acceptable).

## Импликации
- Reward для Free-referrer'а — создаётся trial Subscription{plan:pro, 30d,
  auto_renew:false, rebill_id:""}; auto-downgrade через existing `expirationLoop`.
- Reward для Max-referrer'а — продление Max-периода на 30d (не downgrade в Pro).
- Anti-abuse v2 (card fingerprint) отложен.
```

- [ ] **Step 3: Runbook**

```markdown
<!-- docs/runbooks/ReferralRewardLoopStalled.md -->
# Runbook: ReferralRewardLoop stalled / backlog растёт

**Алерт:** `referral_pending_backlog_growing` (Grafana)
**Severity:** warning
**Owner:** Slava (он же on-call)

## Симптомы
- `referral_rewards_pending_total{result="recorded"} - granted - skipped > 100`
  за 24h.
- Лог `referral.reward.tick_summary` отсутствует в течение >1h.

## Диагностика

```bash
# 1. Размер backlog'а в БД
docker compose exec postgres psql -U postgres -d promptvault -c "
SELECT count(*) as backlog,
       min(eligible_at) as oldest_eligible
  FROM referral_pending_rewards
 WHERE eligible_at < NOW();"

# 2. Логи loop'а
docker compose logs api | grep "referral.reward" | tail -50

# 3. Panic counter
curl -s http://localhost:8080/metrics | grep loop_panics_total
```

## Возможные причины и фикс

### A. Loop не стартовал
Лог `referral.reward.loop_started` отсутствует. → Проверить
`REFERRAL_REWARD_ENABLED=true` в env, перезапустить контейнер.

### B. Loop panic'ит
`promptvault_loop_panics_total{loop="referral_reward"}` > 0. → Грепнуть Sentry
по `referral.reward.grant_failed`, проверить stack trace.

### C. БД timeout
`referral.reward.list_failed` ошибки. → Проверить нагрузку на PostgreSQL,
размер таблицы, индекс `idx_referral_pending_eligible_at`.

### D. GrantReward валится на конкретной записи
`referral.reward.grant_failed` со специфичным `referrer_id`. → Проверить
есть ли corrupted user/subscription для этого id. Manual SQL-cleanup pending.

## Recovery
Восстановление автоматическое — loop возобновится после устранения root cause.
Если backlog огромный (>1000) — увеличить `batch` в `NewRewardLoop(...)` в
`app.go` (default 100 → 500) и перезапустить.

## Escalation
Если backlog растёт >24h после фикса — manual SQL grant'ов через admin tool
(`go run ./cmd/grant-pending-rewards` — TBD создать при необходимости).
```

- [ ] **Step 4: CLAUDE.md upd**

В `promptvault/CLAUDE.md` секция «Ключевые решения», добавить:

```markdown
- **Pricing iteration v3 (2026-05-17):**
  - Free max_prompts 15→25 (миграция 000072); grandfather от Pack E через `effectiveLimit = max(legacy, plan)` сохраняется автоматически.
  - Annual −20% (миграция 000073); существующие подписки продолжают платить старую цену до renewal, T-Bank Charge берёт цену из `plans.GetByID()` на renewal.
  - Pro Smart Insights teaser: `proAllowedInsights = [unused, duplicates]` константа в `usecases/analytics`, per-type filter в `GetInsightsGated`/`ComputeInsights`. Master-gate `IsMax` заменён на `insightsForPlan(planID)`. См. ADR-0008.
  - Referral reward: +30 дней Pro пригласившему через `ReferralRewardLoop` (паттерн `RenewalLoop`). Pending row создаётся на webhook payment.succeeded с `eligible_at = now() + 14d` (UNIQUE на referee_id для idempotency), grant через час после eligibility. Free-referrer получает trial `Subscription{auto_renew:false, rebill_id:""}`. См. ADR-0009.
  - Feature flags: `PRO_INSIGHTS_TEASER_ENABLED`, `REFERRAL_REWARD_ENABLED` (оба default false, включаются после observability недели).
```

- [ ] **Step 5: Commit**

```bash
git add docs/adr/0008-pertype-insights-gate.md \
        docs/adr/0009-delayed-referral-reward.md \
        docs/runbooks/ReferralRewardLoopStalled.md \
        promptvault/CLAUDE.md
git commit -m "docs: ADR-0008, ADR-0009, runbook + CLAUDE.md upd для pricing v3"
```

---

## Финальный smoke test перед prod-flip flags

После всех 19 tasks merged в main:

- [ ] **Final 1: Backend stable**

```bash
go test -short -race -count=1 -timeout=5m ./...
golangci-lint run
```

Expected: всё зелёное.

- [ ] **Final 2: Frontend stable**

```bash
cd frontend
npm run lint
npx vitest run
npm run build
```

Expected: всё зелёное.

- [ ] **Final 3: Staging deploy + 1 неделя observability**

С `PRO_INSIGHTS_TEASER_ENABLED=false`, `REFERRAL_REWARD_ENABLED=false`.

Проверь:
- `/metrics` показывает все новые series (zero-инициализированы).
- Существующие insights endpoint'ы работают как раньше (Max получает 7 типов, Free/Pro — 402 `ErrMaxRequired`).
- Webhook subscription работает (миграции 000072/000073 уже применены, цены обновились).

- [ ] **Final 4: Flip PRO_INSIGHTS_TEASER_ENABLED=true**

```bash
# На VPS:
sed -i 's/ANALYTICS_PRO_INSIGHTS_TEASER_ENABLED=false/ANALYTICS_PRO_INSIGHTS_TEASER_ENABLED=true/' .env.prod
docker compose -f docker-compose.prod.yml restart api
```

Monitor `analytics_insights_gated_total{plan="pro", result="partial"}` — должно расти.

- [ ] **Final 5: Flip REFERRAL_REWARD_ENABLED=true**

```bash
sed -i 's/REFERRAL_REWARD_ENABLED=false/REFERRAL_REWARD_ENABLED=true/' .env.prod
docker compose -f docker-compose.prod.yml restart api
```

Monitor `referral_rewards_pending_total{result="recorded"}` после первых платежей с `referred_by`.

---

## Self-Review

**1. Spec coverage.**

| Spec секция | Задача |
|---|---|
| §3 Создаём миграция 000072 | Task 1 |
| §3 Создаём миграция 000073 | Task 2 |
| §3 Меняем frontend pricing.tsx | Task 3 |
| §3 Создаём миграция 000074 | Task 11 |
| §3 Создаём model + repo referral | Task 12 |
| §3 Создаём types + errors referral | Task 13 |
| §3 Создаём GrantReward + Service | Task 14 |
| §3 Создаём ReferralRewardLoop | Task 15 |
| §3 Меняем insights.go (proAllowedInsights, ComputeInsights) | Tasks 4, 5 |
| §3 Меняем analytics/service.go GetInsightsGated | Task 6 |
| §3 Меняем insights_loop ListPaidUsers | Task 7 |
| §3 Меняем repository user.go | Task 7 |
| §3 Меняем delivery/http/analytics/errors.go | Task 8 |
| §3 Меняем subscription/subscription.go HandleWebhook | Task 16 |
| §3 Меняем app.go | Task 17 |
| §3 Меняем config (AnalyticsConfig + ReferralConfig) | Tasks 9, 17 |
| §3 Меняем frontend analytics.tsx | Task 10 |
| §7 Unit tests | Tasks 1, 4, 5, 6, 14, 15 |
| §7 Integration tests | Tasks 7 (ListPaidUsers), 12 (referral repo) |
| §7 E2E Playwright | **GAP** — см. ниже |
| §8 Метрики | Task 18 |
| §10 Feature flags | Tasks 9, 17 |
| §11 ADR + Runbook + CLAUDE.md | Task 19 |

**GAP:** Playwright E2E (§7) не покрыт отдельным task'ом — spec помечает как optional «опциональная Phase 2 после реализации». Не добавляю в план — это backlog item.

**2. Placeholder scan.** Просмотрено — все шаги содержат либо точный код, либо точную команду + expected output. Test helpers (`fakeSubRepo`, `fakeUserRepo`) показаны полностью где впервые используются. SQL миграции — целиком.

**3. Type consistency.**
- `insightsForPlan` — везде метод Service (`s.insightsForPlan(planID)`) после Task 9. Tasks 4-6 могут вначале использовать функцию, в Task 9 переписывается в метод — explicit step.
- `ListPaidUsers` — везде после Task 7 (rename).
- `ReferralRewardLoop` vs `RewardLoop` — везде `RewardLoop` (внутри package `referral`), внешне доступ через `referral.RewardLoop`. Конструктор `NewRewardLoop`.
- `GrantReward(ctx, referrerID, refereeID, paymentID uint)` — единая сигнатура в Tasks 14, 15.
- Константы `RewardDays = 30`, `EligibilityDays = 14` — введены в Task 13, используются в Tasks 14, 16.

Готов к execution.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-17-pricing-iteration-v3.md`.

**Two execution options:**

1. **Subagent-Driven (recommended)** — Fresh subagent per task + two-stage review между задачами. Подходит для длинных планов, защищает основной контекст.

2. **Inline Execution** — Batch execution в этой сессии через executing-plans skill, checkpoints для review.

Какой подход?
