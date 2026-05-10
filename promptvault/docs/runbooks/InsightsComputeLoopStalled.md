# Runbook: InsightsComputeLoopStalled

**Severity:** P1 (warning)
**Pager:** Telegram only

## Symptom

Counter `analytics_insights_loop_runs_total` не вырос за 25 часов — Smart Insights compute loop **не сделал ни одной итерации**. Counter инкрементится в начале каждого тика, независимо от наличия Max-юзеров, поэтому отсутствие роста = реально стоящий loop.

> Старая версия алерта (до 2026-05-04) смотрела на `analytics_insights_refresh_total{result="success"}` и давала false positive если на проде нет ни одного Max-юзера (цикл по пустому списку → success counter не растёт). Сейчас этот scenario не fire'ится — алерт триггерится только если loop реально мёртв.

## Impact

Низкий-средний: Max-юзеры видят stale Insights. Pro и Free не используют. **Не платный SLO breach** (Insights не входят в core SLO availability).

Связанный алерт **`InsightsComputeLoopAllErrors`** (severity warning) триггерится если loop запускается, Max-юзеры есть, но **все** запуски `ComputeInsights` падают. Это уже реальная проблема (БД, OpenRouter rate-limit, миграция pg_trgm не применилась).

## Investigation

```bash
# 1. Контейнер api жив?
ssh root@85.239.39.45 "docker ps --format '{{.Names}}: {{.Status}}' | grep promptvault-api"

# 2. Loop тикает в логах?
ssh root@85.239.39.45 "docker logs promptvault-api-1 --since 25h 2>&1 | grep 'analytics.insights_loop'"
# Ожидаем: analytics.insights_loop.run ok=N failed=M total=K (даже total=0 — это OK для пустого prod)

# 3. Метрики напрямую
ssh root@85.239.39.45 'docker exec promptvault-api-1 wget -qO- http://localhost:8080/metrics | grep -E "^analytics_insights"'

# 4. Loki query (если есть Grafana доступ)
# {container="promptvault-api-1"} |= "analytics.insights_loop"

# 5. Если AllErrors fire'ится — проверить миграцию pg_trgm (нужна для PossibleDuplicates)
ssh root@85.239.39.45 'docker run --rm -e PGPASSWORD=... postgres:18-alpine psql "..." -c "SELECT extname FROM pg_extension WHERE extname='"'"'pg_trgm'"'"';"'
```

## Mitigation

1. Перезапустить api контейнер:
   ```bash
   ssh root@85.239.39.45 "cd /root/promtlab/promptvault && docker compose -f docker-compose.prod.yml restart api"
   ```
2. Через 5-10 минут проверить, что `analytics_insights_loop_runs_total` вырос.

## Resolution

- Если повторяется: investigate root cause в `internal/usecases/analytics/insights_loop.go`. Проверить что goroutine не падает с panic'ом, и что `Stop()` не вызывается раньше времени из app shutdown.
- Возможные причины полной остановки: panic в `compute()`, который не recovery'ится; неправильный `interval` (= 0); закрытый `stopCh` при старте.
- Code change → PR → merge → deploy.

## Post-mortem

Не требуется (P1, низкий impact). Если стоял > 48h — добавить в incident log.
