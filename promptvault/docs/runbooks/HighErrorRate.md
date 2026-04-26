# Runbook: HighErrorBurnRate (Fast / Slow / Mild)

**Severity:**
- HighErrorBurnRateFast (1h × 14.4) → **P0**
- HighErrorBurnRateSlow (6h × 6) → **P0**
- HighErrorBurnRateMild (3d × 1) → **P1**

## Symptom

5xx error rate превышает SLO budget. Юзеры видят ошибки при API-запросах.

## Impact

P0 — **acute incident**. Платящие юзеры могут не получить нужные данные / payment may fail.
P1 — slow degradation. Возможно Background задача глючит.

## Investigation

```bash
# 1. Какой endpoint падает?
# Grafana → ПромтЛаб — Производительность приложения → Запросы по status code

# 2. Bag in code или infra?
# Loki: {container="promptvault-api-1", level="ERROR"} | json
# Tempo: search by service="promptvault-api" + filter status_code >= 500

# 3. Проверить downstream — Postgres, OpenRouter, T-Bank
# Grafana → ПромтЛаб — База данных → Active connections / slow queries
# logs: "openrouter.timeout" / "tbank.error"

# 4. GlitchTip — top exceptions сейчас
# https://sentry.promtlabs.ru
```

## Mitigation

**P0 (Fast/Slow burn):**
1. Если виновник — последний deploy → rollback к prev image (см. ShareQuotaIncrementLeak runbook).
2. Если downstream (Postgres slow / OpenRouter rate-limit) → ждать или увеличить timeout / circuit breaker.
3. Если specific endpoint — temporary disable через feature flag (если есть).

**P1 (Mild burn):** низкий приоритет, fix в течение дня.

## Resolution

- Investigate exception trace в Tempo + GlitchTip.
- Fix code → PR → deploy.

## Post-mortem

**ОБЯЗАТЕЛЬНО** для P0 (Fast/Slow) — это SLO breach.
