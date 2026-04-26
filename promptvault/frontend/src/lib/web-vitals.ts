/**
 * Web Vitals (Core + extras) — Real User Monitoring через Sentry/GlitchTip.
 * Phase 16 Этап 5.
 *
 * Метрики:
 * - LCP (Largest Contentful Paint) — < 2.5s good, > 4.0s poor
 * - INP (Interaction to Next Paint) — < 200ms good, > 500ms poor
 * - CLS (Cumulative Layout Shift) — < 0.1 good, > 0.25 poor
 * - FCP (First Contentful Paint) — < 1.8s good, > 3.0s poor
 * - TTFB (Time to First Byte) — < 800ms good, > 1800ms poor
 *
 * Отправка через Sentry.captureMessage с level: 'info' и tags: { rum: true }.
 * GlitchTip принимает события — фильтруются в Issues по tag 'rum'.
 *
 * Session Replay не активируем — GlitchTip OSS имеет partial support
 * (events приходят, но UI replay viewer не работает). Web Vitals + breadcrumbs
 * (через browserTracingIntegration) покрывают 80% debug сценариев.
 */
import * as Sentry from "@sentry/react"
import {
  onCLS,
  onFCP,
  onINP,
  onLCP,
  onTTFB,
  type Metric,
} from "web-vitals"

/**
 * Web Vitals rating thresholds — Google Core Web Vitals defaults.
 * https://web.dev/articles/vitals
 */
const RATING_LEVELS: Record<Metric["rating"], "info" | "warning" | "error"> = {
  good: "info",
  "needs-improvement": "warning",
  poor: "error",
}

function reportVital(metric: Metric): void {
  // Sentry.captureMessage с structured data для GlitchTip Issues filter.
  Sentry.captureMessage(`Web Vital: ${metric.name}`, {
    level: RATING_LEVELS[metric.rating] ?? "info",
    tags: {
      rum: "true",
      metric: metric.name,
      rating: metric.rating,
    },
    contexts: {
      vital: {
        name: metric.name,
        value: metric.value,
        rating: metric.rating,
        delta: metric.delta,
        id: metric.id,
        navigationType: metric.navigationType,
      },
    },
  })
}

/**
 * Регистрирует callbacks для всех Core Web Vitals.
 * Вызывается ОДИН раз в main.tsx после initSentry().
 *
 * web-vitals SDK автоматически submits финальные значения когда страница
 * скрывается (visibilitychange) или unloads — отправка через
 * navigator.sendBeacon() гарантирует доставку даже на закрытии вкладки.
 */
export function initWebVitals(): void {
  if (typeof window === "undefined") {
    return
  }

  onLCP(reportVital)
  onINP(reportVital)
  onCLS(reportVital)
  onFCP(reportVital)
  onTTFB(reportVital)
}
