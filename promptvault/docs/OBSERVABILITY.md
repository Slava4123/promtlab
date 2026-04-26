# Observability (Phase 14.3 + Phase 15 ✅ + Phase 16)

Документ фиксирует metrics, traces, logs и alerts для prod-инфраструктуры
promtlabs.ru.

## Текущий статус — Phase 15 ✅ closed (2026-04-26)

Phase 15 завершена и фактически расширена до Phase 16 объёма (cAdvisor,
Loki, Tempo, SLO/SLA multi-burn-rate alerts). Финальный scope ниже.

- [x] **Phase 14.3** — counters в `metrics.go`, alert rules в `alerts.yaml`.
- [x] **Phase 15 A** — IP-allowlist на `/metrics`, fix `for: 1h` у `InsightsComputeLoopDead`.
- [x] **Phase 15 B** — `prometheus.yml`, `alertmanager.yml`, Grafana datasource provisioning.
- [x] **Phase 15 C** — Prometheus в `docker-compose.prod.yml` (90d retention, 10GB cap).
- [x] **Phase 15 D** — Alertmanager + email receiver через Gmail SMTP
      (см. `infra/alertmanager/SECRETS.md`). Telegram удалён: с прод VPS Timeweb
      api.telegram.org:443 блокируется на IPv4 (РКН), пакеты не уходят.
- [x] **Phase 15 E** — Grafana + nginx vhost `grafana.promtlabs.ru`, basic auth.
- [x] **Phase 15 F** — Runbook (этот документ) + memory budget в `DEPLOY.md`.
- [x] **Phase 16** — node-exporter, postgres-exporter, cAdvisor v0.52.1 (cgroup v2),
      blackbox, Loki + Promtail (logs), Tempo (traces), SLO multi-burn-rate alerts.

Полный план: `docs/PHASE15_OBSERVABILITY_PLAN.md`.

## Stack components

| Компонент | Версия | Bind | RAM limit | Назначение |
|---|---|---|---|---|
| Prometheus | v3.0.1 | 127.0.0.1:9090 | 384M | TSDB + alert evaluation, 90d retention |
| Alertmanager | v0.27.0 | 127.0.0.1:9093 | 96M | Routing → Gmail SMTP (email-only) |
| Grafana | 11.3.0 | grafana.promtlabs.ru | 192M | Dashboards (5 шт., русские) + Explore |
| Loki | 3.6.0 | 127.0.0.1:3100 | 384M | Logs schema v13, 7d retention |
| Promtail | 3.6.0 | — | 96M | Docker SD + slog JSON pipeline |
| Tempo | 2.6.0 | 127.0.0.1:3200 | 384M | Traces OTLP gRPC :4317, 7d retention |
| node-exporter | v1.8.2 | — | 64M | Host CPU/RAM/Disk/Net/Load |
| postgres-exporter | v0.16.0 | — | 64M | pg_stat_* metrics |
| cAdvisor | v0.55.1 | — | 192M | Per-container metrics, cgroup v2 + Docker overlayfs |
| blackbox-exporter | v0.25.0 | — | 64M | External HTTP probes |

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

## Distributed tracing (Tempo)

Backend инструментирован через OpenTelemetry SDK (`internal/infrastructure/telemetry/otel.go`):

- **HTTP spans** — `otelchi.Middleware("promptvault-api", otelchi.WithChiRoutes(r))`
  нормализует route patterns (`/api/prompts/{id}` вместо raw URL) — отсутствие
  cardinality explosion даже при бурном QPS.
- **SQL spans** — `otelgorm.NewPlugin()` подключён в `postgres.go`, все repos
  передают `*gorm.DB.WithContext(ctx)` → SQL spans линкуются к parent HTTP span.
- **Propagator** — W3C TraceContext + Baggage (готовы к multi-service deployments).
- **Sampler** — `ParentBased(TraceIDRatioBased(rate))`. Уважает родительское
  решение, поэтому traces не разрываются между сервисами.

### Активация в проде

В `.env.prod`:

```env
TELEMETRY_ENABLED=true
TELEMETRY_OTLP_ENDPOINT=tempo:4317
TELEMETRY_TRACES_SAMPLE_RATE=0.1
```

`0.1` (10% sampling) — баланс между видимостью и filesystem cost Tempo
(7d retention, без object storage). При низком QPS можно поднять до `0.5`
или `1.0`, мониторить размер `tempo-data` через `du -sh` ежедневно первые
3 дня после изменения.

### Поиск spans

`Grafana → Explore → datasource Tempo`:

- **Search** — выбрать service `promptvault-api`, по последним 5 мин.
- **TraceQL** для медленных запросов:
  ```traceql
  { resource.service.name = "promptvault-api" && duration > 100ms }
  ```
- **Links to logs** — datasource Tempo сконфигурирован с derivedFields на
  trace_id из Loki, в одном клике переходишь на логи нужного запроса.

## Email-only delivery (Gmail SMTP)

С прод VPS Timeweb (РФ) Telegram Bot API недоступен: пакеты к
`api.telegram.org:443` блокируются на IPv4. Поэтому в `alertmanager.yml`
один receiver `email`, доставка через `smtp.gmail.com:587` (работает по
IPv6 через Docker bridge).

Все severity (critical + warning) уходят на `promstlab@gmail.com` с темой
`[CRITICAL]/[WARNING] <alertname>` и HTML-телом со списком alerts +
runbook ссылкой.

**Активация на VPS:** см. `infra/alertmanager/SECRETS.md` (Gmail App
Password в `infra/alertmanager/secrets/smtp_password`, владелец 65534:65534).
В будущем при появлении канала с прокси-выходом наружу (Telegram через
selfhost bot-api или Mattermost/Discord webhook) — добавить второй receiver.

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

## Runbook по alert'ам

Когда email приходит уведомление с одним из 4 alert'ов — действовать
по порядку: проверить → диагностировать → починить → подтвердить. После
получения сообщения сразу зайти в Grafana (`https://grafana.promtlabs.ru`)
или Prometheus UI (`ssh -L 9090:127.0.0.1:9090 deploy@85.239.39.45`) и
посмотреть график соответствующей метрики за последние 6 часов.

### `InsightsComputeLoopStalled` (warning)

**Условие:** `increase(analytics_insights_refresh_total{result="success"}[25h]) == 0` 1+ час подряд.
**Что значит:** Smart Insights `InsightsComputeLoop` не дал ни одного успешного пересчёта за 25 часов. Это не «мёртв» (для этого есть `InsightsComputeLoopDead`), а скорее — все попытки заканчивались `error` или loop тормозит. Юзеры с тарифом Max видят stale insights.
**Quick checks:**
1. `docker compose -f docker-compose.prod.yml logs api --since 26h | grep -i 'analytics.insights_loop'` — есть ли `error` записи?
2. Prometheus query `sum by (result) (increase(analytics_insights_refresh_total[25h]))` — что вместо `success`? Если `error` >> 0 — есть exception в loop.
3. GlitchTip → проект `promptvault-backend` → фильтр по `analytics.insights_loop` тегу — последние ошибки.
**Типовые причины:** упал OpenRouter/Anthropic API; миграция `pg_trgm` не применилась (для PossibleDuplicates); rate-limit на OpenRouter exceeded.
**Действия:** перезапустить api (`docker compose restart api`); если ошибка повторяется через 1 цикл — escalate (debug в коде loop).

### `InsightsComputeLoopDead` (critical)

**Условие:** `absent_over_time(analytics_insights_refresh_total{result="success"}[48h])` 1+ час подряд.
**Что значит:** метрики вообще нет за 48 часов — либо api контейнер мёртв, либо `MetricsEnabled=false`, либо loop никогда не стартовал.
**Quick checks:**
1. `curl -I https://promtlabs.ru/api/health` — api отвечает?
2. `docker compose -f docker-compose.prod.yml ps api` — контейнер up?
3. `docker exec promtlab-api-1 wget -qO- http://api:8080/metrics | grep analytics_insights` — counter существует?
**Типовые причины:** api контейнер crash-loop, `SERVER_METRICS_ENABLED=false`, миграция `000048_pg_trgm.up.sql` не применилась → loop падает на старте.
**Действия:** restart api → check logs → если не помогает, rollback на предыдущий image (см. `.prev_deploy_commit`).

### `CleanupLoopStalled` (warning)

**Условие:** `increase(analytics_cleanup_deleted_total[25h]) == 0` 1+ час подряд.
**Что значит:** retention cleanup за 25 часов не удалил ни одной строки из `team_activity` / `share_views` / `prompt_usage`.
**Quick checks:**
1. На свежем prod (мало данных) — это **нормально**, не паниковать. False positive когда retention окно ещё не превышено.
2. `docker exec promtlab-api-1 wget -qO- http://api:8080/metrics | grep analytics_cleanup` — счётчики растут вообще?
3. Если объёмы значительные — `psql -c "SELECT MAX(created_at) FROM team_activity;"` → если есть записи старше 90 дней, loop сломан.
**Действия:** если объёмы значительные → restart api + проверить logs `analytics.cleanup_loop`. Иначе — silence через `amtool silence add alertname=CleanupLoopStalled` на 25 часов.

### `ShareQuotaIncrementLeak` (critical)

**Условие:** `rate(share_quota_increment_failed_total[5m]) > 0` 10+ минут подряд.
**Что значит:** **revenue at risk.** Share-ссылки создаются успешно, но `IncrementDailyUsage` (квота на тарифе) падает. Юзер обходит лимит → бесплатно создаёт неограниченные share-ссылки.
**Quick checks:**
1. `docker compose -f docker-compose.prod.yml logs api --since 15m | grep 'share.quota.increment_failed'` — текст ошибки.
2. PostgreSQL: `psql -c "SELECT COUNT(*) FROM share_links WHERE created_at > NOW() - INTERVAL '15 minutes';"` — сколько успешных создано.
3. PostgreSQL: `psql -c "SELECT * FROM subscription_quota WHERE quota_kind='share_link_daily' ORDER BY updated_at DESC LIMIT 5;"` — состояние счётчика квоты.
**Типовые причины:** мигрировал `subscription_quota` row missing для нового юзера; БД lock на increment query (рассинхрон между table); GORM transactional issue.
**Действия:** **высокий приоритет**. Sentry → найти exception. Если виновник в коде — hotfix + deploy. Если data integrity — restore счётчика квоты вручную, написать post-mortem.

### Silencing alert'ов

При плановых работах (миграция, deploy, restart):
```bash
ssh -L 9093:127.0.0.1:9093 deploy@85.239.39.45 &
amtool silence add alertname=InsightsComputeLoopStalled --duration=2h --comment="planned restart"
amtool silence list
```

Или через Alertmanager UI: `http://localhost:9093/#/silences/new` после SSH-tunnel.
