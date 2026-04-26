# Runbook: ShareQuotaIncrementLeak

**Severity:** P0 (critical) — **REVENUE AT RISK**
**Pager:** Telegram + Email immediate

## Symptom

Counter `share_quota_increment_failed_total` инкрементится. Это значит share-ссылка **создана успешно** в БД, но `IncrementDailyUsage` (счётчик квоты на тарифе) **упал**.

Юзер обходит дневной лимит — может создавать неограниченные share-ссылки бесплатно.

## Impact

**Прямой revenue impact** для платных тарифов. Free юзеры с лимитом 2 share/день могут создать 100+ ссылок без ограничения. Pro/Max аналогично.

При rate > 0.001/sec для 10m → **alert**.

## Investigation

```bash
# 1. Логи api — найти точный exception
ssh root@85.239.39.45 "docker logs promptvault-api-1 --since 15m 2>&1 | grep 'share.quota.increment_failed'"

# 2. GlitchTip Issues
# Открыть https://sentry.promtlabs.ru → проект promptvault-backend
# → filter share_link_create

# 3. БД — counter table состояние
# SELECT * FROM subscription_quota WHERE quota_kind='share_link_daily' ORDER BY updated_at DESC LIMIT 10;

# 4. Связанные share-ссылки за последние 15 мин
# SELECT id, owner_id, created_at FROM share_links WHERE created_at > NOW() - INTERVAL '15 minutes';
```

## Mitigation

**HOTFIX immediately:**
1. Если виновник — конкретный user (один специфический ID повторяется в логах) → `UPDATE users SET frozen_at=NOW() WHERE id=X` через admin panel.
2. Если bug в коде (concurrency, transaction issue) — rollback к предыдущему image:
   ```bash
   ssh root@85.239.39.45 "cd /root/promtlab/promptvault && \
     docker tag ghcr.io/slava4123/promtlab-api:prev ghcr.io/slava4123/promtlab-api:latest && \
     docker compose -f docker-compose.prod.yml up -d api"
   ```

## Resolution

- Code review `internal/usecases/share/share.go` IncrementDailyUsage path.
- Test coverage добавить — concurrency / GORM transaction edge cases.
- Manual reconciliation: пересчитать `subscription_quota.amount` для затронутых юзеров.

## Post-mortem

**ОБЯЗАТЕЛЬНО** — это revenue-impacting incident.
