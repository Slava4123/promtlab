# Runbook: InsightsComputeLoopStalled

**Severity:** P1 (warning)
**Pager:** Telegram only

## Symptom

Smart Insights в `/dashboard` для Max-юзеров stale (старее 24 часов). Юзер видит панель «Insights» с устаревшими данными или сообщением "обновлено вчера".

## Impact

Только Max-юзеры. Pro и Free не используют Insights. Бизнес-impact низкий — фича работает, но без свежих данных. **Не платный SLO breach** (Insights не входят в core SLO availability).

## Investigation

```bash
# 1. Проверить логи insights_loop
ssh root@85.239.39.45 "docker logs promptvault-api-1 --since 25h 2>&1 | grep insights_loop"

# 2. Prometheus query — последний successful pereсчёт
# analytics_insights_refresh_total{result="success"} - rate(...) per 5m

# 3. Loki query (Grafana → Explore → Loki):
# {container="promptvault-api-1"} |= "analytics.insights_loop"

# 4. Проверить миграцию pg_trgm (для PossibleDuplicates feature)
docker exec promptvault-api-1 wget -qO- http://api:8080/metrics | grep insights_loop
```

## Mitigation

1. Перезапустить api контейнер: `docker compose -f docker-compose.prod.yml restart api`
2. Через 5 мин проверить counter `analytics_insights_refresh_total{result="success"}` — должен инкрементиться.

## Resolution

- Если повторяется: investigate root cause в коде `internal/usecases/analytics/insights_loop.go`.
- Возможные причины: OpenRouter rate-limit hit, миграция pg_trgm не применилась, panic в loop.
- Code change → PR → merge → deploy.

## Post-mortem

Не требуется (P1, низкий impact).
