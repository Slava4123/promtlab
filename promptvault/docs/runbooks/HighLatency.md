# Runbook: HighLatencyP99

**Severity:** P1 (warning)
**Pager:** Telegram

## Symptom

p99 latency запросов > 500ms за последние 10 минут. Юзеры видят медленную работу UI / API delays.

## Impact

UX deterioration. Не downtime, но бракосочетание производительности.

## Investigation

```bash
# 1. Какие endpoints медленные?
# Grafana → ПромтЛаб — Производительность приложения → Top 10 slowest endpoints
# Запомнить top-3 worst.

# 2. Где задержка — backend код или DB?
# Grafana → Tempo → search service="promptvault-api" + latency > 500ms
# → click trace → waterfall view spans
# Если SQL spans занимают 80% времени → DB медленная.

# 3. DB состояние
# Grafana → ПромтЛаб — База данных → connections, slow queries

# 4. CPU/RAM пресс — VPS перегружен?
# Grafana → ПромтЛаб — Инфраструктура VPS → Free RAM / Load average
```

## Mitigation

1. Если slow query → `EXPLAIN ANALYZE` + добавить index (отдельный PR).
2. Если RAM/CPU exhaustion → restart api для cleanup memory leaks: `docker compose restart api`. Если повторяется — upgrade VPS.
3. Если external API (OpenRouter) медленная → ждать или сменить модель temporary.

## Resolution

- Index missing → migration с `CREATE INDEX CONCURRENTLY ...`.
- N+1 queries → fix через GORM Preload.
- Memory leak → profile через pprof (dev mode).

## Post-mortem

Не требуется (P1).
