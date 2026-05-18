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
