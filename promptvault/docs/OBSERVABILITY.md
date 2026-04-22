# Observability (Phase 14.3)

Документ фиксирует Prometheus counters, Sentry breadcrumbs и alert rules
для prod-инфраструктуры promtlabs.ru.

## Prometheus `/metrics`

- **Endpoint:** `GET /metrics` (backend).
- **Feature flag:** `SERVER_METRICS_ENABLED=true` в `.env.prod`.
- **Защита:** IP-allowlist на nginx (internal ingress).
- **Формат:** Prometheus text exposition (через `promhttp.Handler()`).

### Counters

| Имя | Тип | Labels | Назначение |
|---|---|---|---|
| `share_quota_increment_failed_total` | Counter | — | Revenue-leak: share-ссылка создана, daily-quota счётчик не увеличился |
| `analytics_insights_refresh_total` | CounterVec | `result` = `success` \| `error` \| `rate_limited` | Итерации InsightsComputeLoop и `/api/analytics/insights/refresh` |
| `analytics_cleanup_deleted_total` | CounterVec | `table` = `team_activity` \| `share_views` \| `prompt_usage` | Удалённые строки retention cleanup'ом |

Источник регистрации — `backend/internal/infrastructure/metrics/metrics.go`.

## Sentry breadcrumbs

Добавлены в `delivery/http/analytics/handler.go`:

- `analytics/insights.refresh.trigger` — user_id.
- `analytics/export.trigger` — format, scope, range.

Данные не содержат PII (email, prompt content). Включать через
`SENTRY_ENABLED=true`.

## Alert rules (Prometheus)

Хранятся в `infra/prometheus/alerts.yaml` (см. файл рядом с этим документом).

### Группа `analytics-loops`

```yaml
groups:
  - name: analytics-loops
    interval: 5m
    rules:
      - alert: InsightsComputeLoopStalled
        expr: increase(analytics_insights_refresh_total{result="success"}[25h]) == 0
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "InsightsComputeLoop не проработал 25+ часов"
          description: "Нет успешных пересчётов Smart Insights за последние 25 часов. Проверить logs (analytics.insights_loop.*) и состояние cron-а."

      - alert: InsightsComputeLoopDead
        expr: absent_over_time(analytics_insights_refresh_total{result="success"}[48h])
        labels:
          severity: critical
        annotations:
          summary: "InsightsComputeLoop мёртв > 48 часов"

      - alert: CleanupLoopStalled
        expr: increase(analytics_cleanup_deleted_total[25h]) == 0
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "CleanupLoop не проработал 25+ часов"
          description: "Retention cleanup (team_activity/share_views/prompt_usage) не удалил ни одной записи за 25 часов. Нормально для тестовых инстансов, в prod — проверить loop."

      - alert: ShareQuotaIncrementLeak
        expr: rate(share_quota_increment_failed_total[5m]) > 0
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Revenue-leak: share quota increments fail"
          description: "Share-ссылки создаются, но daily quota counter не инкрементится. Revenue at risk. Логи: share.quota.increment_failed."
```

Проверка: `promtool check rules infra/prometheus/alerts.yaml`.
