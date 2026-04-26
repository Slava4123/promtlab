# Service Level Objectives (SLO) — PromtLab

Документ фиксирует SLO targets и формулы расчёта error budget. Phase 16
Этап 4. Базируется на Google SRE Workbook Chapter 5 (Alerting on SLOs).

## SLO Targets

| SLI | Target | Window | Budget |
|---|---|---|---|
| **Availability** | 99.9% (success ratio) | 30 дней (rolling) | 43 мин downtime / месяц |
| **Latency p99** | < 500 ms | 28 дней (rolling) | — |

## Формулы

### Availability SLI

```
SLI = 1 - (rate(http_requests_total{status=~"5.."}[30d]) / rate(http_requests_total[30d]))
SLO_target = 0.999  (99.9%)
Error budget = 1 - 0.999 = 0.001 (0.1% errors допустимо)
```

При 100k запросов/мес → допустимо 100 ошибок 5xx суммарно.
При 10M запросов/мес → допустимо 10000 ошибок 5xx.

### Latency SLI

```
SLI = histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[28d])) < 0.5
```

99-й перцентиль за 28 дней должен быть менее 500 ms.

## Multi-burn-rate alerts

Стандартная формула (Google SRE Workbook Ch. 5, Table 5-2):

| Window | Burn rate threshold | Budget consumed | Severity | for |
|---|---|---|---|---|
| 1h | **14.4×** | 2% | critical (P0) | 2 min |
| 6h | **6×** | 5% | critical (P0) | 15 min |
| 3d | **1×** | 10% | warning (P1) | 1h |

Burn rate = `current_error_rate / (1 - SLO_target)`

Для SLO 99.9% (= 0.001 budget):
- 14.4× burn = 1.44% error rate
- 6× burn = 0.6% error rate
- 1× burn = 0.1% error rate

## Где видно

- **Grafana → ПромтЛаб — SLO / SLA** dashboard:
  - Availability stat (target 99.9% — green/yellow/red)
  - Error Budget остаток (% remaining)
  - Latency p99 stat (target < 500ms)
  - Burn rate windows (multi-window timeseries)
  - Latency p99 во времени с SLO threshold линией
  - Error Budget consumed cumulative

- **Telegram alerts:**
  - `HighErrorBurnRateFast` (1h × 14.4) → P0
  - `HighErrorBurnRateSlow` (6h × 6) → P0
  - `HighErrorBurnRateMild` (3d × 1) → P1
  - `HighLatencyP99` → P1
  - `InstanceDown` (blackbox external probe) → P0

## Decision framework

### Если budget > 50%
- Можно делать рискованные deploys (рефакторинг, схема изменения).
- Tolerance к short blips.

### Если budget 20-50%
- Замедлить рискованные изменения, фокус на stability.
- Code reviews более тщательные.

### Если budget < 20%
- **Feature freeze** (не deploy'им новое).
- Все ресурсы на stabilization.
- Post-mortem каждого incident.

### Если budget = 0% (полностью израсходован)
- **Полный rollback** последнего deploy если он триггернул.
- Только critical bug fixes идут в prod.
- Эскалация — выпустим новый release только когда budget восстановится.

## Источники

- [Google SRE Workbook Ch. 5 — Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/)
- [Google SRE Book Ch. 4 — Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
- Recording rules: `infra/prometheus/slo_rules.yaml`
- Dashboard: `infra/grafana/provisioning/dashboards/slo.json`
