# Runbook: CleanupLoopStalled

**Severity:** P1 (warning)
**Pager:** Telegram only

## Symptom

Retention cleanup не удалил ни одной строки за 25 часов из `team_activity` / `share_views` / `prompt_usage`. Метрика `analytics_cleanup_deleted_total` плоская.

## Impact

Низкий: данные накапливаются дольше планового retention (90d). На малых объёмах — незаметно. На больших — рост размера БД.

**На свежем prod / низкая активность — это нормальный false positive** (нет старых данных для удаления). Можно silence.

## Investigation

```bash
# 1. Объёмы таблиц
ssh root@85.239.39.45 'docker run --rm -e PGPASSWORD=$(grep ^DATABASE_PASSWORD= /root/promtlab/promptvault/.env.prod | cut -d= -f2-) -v /root/promtlab/promptvault/ca.crt:/ca.crt:ro postgres:18-alpine psql "host=$(grep ^DATABASE_HOST= /root/promtlab/promptvault/.env.prod | cut -d= -f2-) port=5432 user=gen_user dbname=promtlab sslmode=verify-full sslrootcert=/ca.crt" -c "SELECT pg_size_pretty(pg_total_relation_size('"'"'team_activity'"'"')) as activity, pg_size_pretty(pg_total_relation_size('"'"'share_views'"'"')) as views, pg_size_pretty(pg_total_relation_size('"'"'prompt_usage'"'"')) as usage;"'

# 2. Cron последний run в логах
ssh root@85.239.39.45 "docker logs promptvault-api-1 --since 25h 2>&1 | grep cleanup_loop"

# 3. Самый старый запись в каждой таблице
# SELECT MAX(created_at) FROM team_activity;
```

## Mitigation

Если объёмы небольшие — silence через amtool на 25 часов:
```bash
ssh -L 9093:127.0.0.1:9093 root@85.239.39.45 &
amtool silence add alertname=CleanupLoopStalled --duration=25h --comment="low activity prod"
```

Если объёмы большие и loop сломан:
1. `docker compose restart api`
2. Через 1ч проверить deleted counter.

## Resolution

Investigate `internal/usecases/analytics/cleanup_loop.go` если повторяется на активном prod.

## Post-mortem

Не требуется.
