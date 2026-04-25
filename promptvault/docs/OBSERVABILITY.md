# Observability (Phase 14.3 + Phase 15 prep)

Документ фиксирует Prometheus counters, Sentry breadcrumbs и alert rules
для prod-инфраструктуры promtlabs.ru.

## Текущий статус (Phase 15)

- [x] Phase 14.3: counters в `metrics.go`, alert rules в `infra/prometheus/alerts.yaml`.
- [x] Phase 15 Шаг A: IP-allowlist middleware на `/metrics`, fix `for: 1h` у `InsightsComputeLoopDead`.
- [x] Phase 15 Шаг B: `infra/prometheus/prometheus.yml`, `infra/alertmanager/alertmanager.yml`, Grafana datasource provisioning.
- [ ] Phase 15 Шаг C: Prometheus в `docker-compose.prod.yml` — после upgrade VPS до 4 GB.
- [ ] Phase 15 Шаг D: Alertmanager + Telegram receiver — после создания бота. На VPS:
  - `bot_token` — файл `infra/alertmanager/secrets/bot_token` (gitignored), монтируется в контейнер.
  - `chat_id` — заменить placeholder `1` в `alertmanager.yml`: `sed -i "s/chat_id: 1$/chat_id: $TELEGRAM_CHAT_ID/" infra/alertmanager/alertmanager.yml`. (Alertmanager v0.27.0 не поддерживает `chat_id_file` — появилось в v0.28+.)
- [ ] Phase 15 Шаг E: Grafana + nginx vhost `grafana.promtlabs.ru` — после A-записи DNS + `htpasswd`.
- [ ] Phase 15 Шаг F: Runbook по alert'ам + memory budget update в `DEPLOY.md §11.9`.

Полный план: `docs/PHASE15_OBSERVABILITY_PLAN.md`.

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
