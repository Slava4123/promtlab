# Runbook: InsightsComputeLoopDead

**Severity:** P0 (critical)
**Pager:** Telegram + Email

## Symptom

Метрика `analytics_insights_refresh_total{result="success"}` отсутствует **48+ часов**. Это значит либо api контейнер мёртв, либо `MetricsEnabled=false`, либо loop никогда не стартовал в этой версии кода.

## Impact

**Если api мёртв** — это критично, frontend не работает (P0).
**Если только loop сломан** — Max-юзеры теряют Insights (P1), но api работает.

## Investigation

```bash
# 1. API живой?
curl -I https://promtlabs.ru/api/health  # ожидаем 200

# 2. Контейнер up?
ssh root@85.239.39.45 "docker ps | grep api"

# 3. Метрика реально отсутствует или Prometheus отвалился?
# В Prometheus UI: query `up{job="promptvault-api"}` — должен быть 1.

# 4. Логи api — есть ли panic / startup error?
ssh root@85.239.39.45 "docker logs promptvault-api-1 --tail 100"

# 5. .env.prod — SERVER_METRICS_ENABLED=true?
ssh root@85.239.39.45 "grep ^SERVER_METRICS_ENABLED /root/promtlab/promptvault/.env.prod"
```

## Mitigation

**Если api crash-loop:**
1. `docker logs promptvault-api-1 --since 10m` — найти panic / fatal error.
2. Если миграция падает — rollback через `docker tag ghcr.io/slava4123/promtlab-api:prev :latest && docker compose up -d api`.
3. Если из-за env — поправить `.env.prod`, restart.

**Если api UP но loop dead:**
1. `docker compose restart api` — перезапустит fresh.
2. Через 1ч проверить counter.

## Resolution

- Investigate root cause через GlitchTip Issues (есть ли exception?).
- Fix code → PR → deploy.

## Post-mortem

**ОБЯЗАТЕЛЬНО** если impact > 5 мин downtime для api.
