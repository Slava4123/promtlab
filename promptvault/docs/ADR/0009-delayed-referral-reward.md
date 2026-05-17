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
