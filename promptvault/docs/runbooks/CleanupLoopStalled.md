# Runbook: CleanupLoopStalled

**Severity:** P1 (warning)
**Pager:** Telegram only

## Symptom

Counter `analytics_cleanup_runs_total` не вырос за 25 часов — retention cleanup loop **не сделал ни одной итерации**. Counter инкрементится в начале каждого тика, независимо от того, удалено что-то или нет, поэтому отсутствие роста = реально стоящий loop.

> Старая версия алерта (до 2026-05-04) смотрела на `analytics_cleanup_deleted_total` и часто давала false positive на свежем проде или при низкой активности — когда loop работал, но удалять было нечего. Сейчас этот scenario считается нормой и не fire'ится.

## Impact

Средний: данные накапливаются дольше планового retention (90d). Если loop простоит долго — рост размера БД и потенциально PII в `team_activity` старше политики хранения.

## Investigation

```bash
# 1. Контейнер api жив?
ssh root@85.239.39.45 "docker ps --format '{{.Names}}: {{.Status}}' | grep promptvault-api"

# 2. Loop делает что-то в логах?
ssh root@85.239.39.45 "docker logs promptvault-api-1 --since 25h 2>&1 | grep -E 'analytics.cleanup'"

# 3. Метрика напрямую
ssh root@85.239.39.45 'docker exec promptvault-api-1 wget -qO- http://localhost:8080/metrics | grep analytics_cleanup'

# 4. Объёмы таблиц (для оценки impact)
ssh root@85.239.39.45 'docker run --rm -e PGPASSWORD=$(grep ^DATABASE_PASSWORD= /root/promtlab/promptvault/.env.prod | cut -d= -f2-) -v /root/promtlab/promptvault/ca.crt:/ca.crt:ro postgres:18-alpine psql "host=$(grep ^DATABASE_HOST= /root/promtlab/promptvault/.env.prod | cut -d= -f2-) port=5432 user=gen_user dbname=promtlab sslmode=verify-full sslrootcert=/ca.crt" -c "SELECT pg_size_pretty(pg_total_relation_size('"'"'team_activity'"'"')) as activity, pg_size_pretty(pg_total_relation_size('"'"'share_views'"'"')) as views, pg_size_pretty(pg_total_relation_size('"'"'prompt_usage'"'"')) as usage;"'
```

## Mitigation

1. Перезапустить api контейнер:
   ```bash
   ssh root@85.239.39.45 "cd /root/promtlab/promptvault && docker compose -f docker-compose.prod.yml restart api"
   ```
2. Через 5-10 минут проверить, что `analytics_cleanup_runs_total` вырос:
   ```bash
   ssh root@85.239.39.45 'docker exec promptvault-api-1 wget -qO- http://localhost:8080/metrics | grep analytics_cleanup_runs_total'
   ```

Если метрика всё ещё не растёт — loop падает на старте или goroutine скрытно умерла. Check `slog.Error` события вокруг `analytics.cleanup.*.failed`.

## Resolution

Investigate `internal/usecases/analytics/cleanup.go` — особенно если есть panic в repository методах (`activity.CleanupByRetention`, `analytics.CleanupShareViewsByRetention`, `analytics.CleanupPromptUsageByRetention`). Эти ошибки логируются, но если recovery упало в goroutine выше — loop умолкает.

## Post-mortem

Требуется только если loop стоял > 48 часов или повторяется > 2 раз в неделю.
