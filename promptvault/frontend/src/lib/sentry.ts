/**
 * Sentry SDK обёртка для PromptLab frontend.
 *
 * Использует @sentry/react v10+ с GlitchTip self-hosted backend в качестве
 * Sentry-compatible endpoint (API совместим с Sentry.io).
 *
 * Feature flag: initSentry ничего не делает если VITE_SENTRY_ENABLED !== "true".
 * Это позволяет build и deploy кода в prod без фактической активации
 * мониторинга — gradual rollout friendly.
 *
 * Не поддерживается в GlitchTip (в отличие от Sentry.io):
 * - Session Replay (replayIntegration НЕ подключается)
 * - Profiling (browserProfilingIntegration НЕ подключается)
 * - React Router v7 instrumentation (используем generic browserTracingIntegration)
 */

import * as Sentry from "@sentry/react"

/** Публичный интерфейс юзера для Sentry. Минимум PII. */
export interface SentryUser {
  id: string | number
  email?: string
  username?: string
}

/**
 * Инициализирует Sentry SDK. Вызывать ОДИН раз, до createRoot, в main.tsx.
 *
 * Noop если VITE_SENTRY_ENABLED !== "true" или VITE_SENTRY_DSN пустой.
 * Это важно: в dev и при gradual rollout (PR 3 → PR 4) init не вызывается,
 * нулевой runtime overhead, нулевая сеть.
 */
export function initSentry(): void {
  const enabled = import.meta.env.VITE_SENTRY_ENABLED === "true"
  const dsn = import.meta.env.VITE_SENTRY_DSN as string | undefined

  if (!enabled) {
    // Намеренно silent в prod, warn только в dev — чтобы не пугать юзеров.
    if (import.meta.env.DEV) {
      console.info("[Sentry] init skipped: VITE_SENTRY_ENABLED !== 'true'")
    }
    return
  }

  if (!dsn) {
    console.warn("[Sentry] init skipped: VITE_SENTRY_DSN is empty")
    return
  }

  const environment = (import.meta.env.VITE_SENTRY_ENVIRONMENT as string) || "production"
  const release = (import.meta.env.VITE_SENTRY_RELEASE as string) || "dev"
  const tracesSampleRate = parseFloat(
    (import.meta.env.VITE_SENTRY_TRACES_SAMPLE_RATE as string) || "0.0",
  )

  Sentry.init({
    dsn,
    environment,
    release,
    // Generic browserTracingIntegration — ловит fetch/XHR transactions
    // автоматически. НЕ используем reactRouterV7BrowserTracingIntegration,
    // т.к. официально не задокументирована в Sentry v10 для React Router v7.
    integrations: [Sentry.browserTracingIntegration()],
    tracesSampleRate: Number.isFinite(tracesSampleRate) ? tracesSampleRate : 0.0,
    // Sample всех events при низком трафике — не нужно semplinг когда юзеров мало.
    sampleRate: 1.0,
    // Скраббинг PII в events — убираем Authorization и Cookie из headers.
    // GlitchTip получит events без JWT токенов и session cookies.
    beforeSend(event) {
      if (event.request?.headers) {
        delete event.request.headers["Authorization"]
        delete event.request.headers["authorization"]
        delete event.request.headers["Cookie"]
        delete event.request.headers["cookie"]
      }
      return event
    },
  })
}

/**
 * Устанавливает user context на Sentry scope. Вызывается в auth store после
 * успешного login / restoreSession. Ошибки после этого будут атрибутироваться
 * конкретному юзеру в GlitchTip UI.
 *
 * Noop если Sentry не инициализирован.
 */
export function setSentryUser(user: SentryUser): void {
  Sentry.setUser({
    id: String(user.id),
    email: user.email,
    username: user.username,
  })
}

/**
 * Очищает user context. Вызывается в auth store при logout, чтобы последующие
 * ошибки не атрибутировались вышедшему юзеру.
 *
 * Noop если Sentry не инициализирован.
 */
export function clearSentryUser(): void {
  Sentry.setUser(null)
}

/**
 * Повторный экспорт Sentry.captureException для удобства в api/client.ts,
 * error-boundary, и других местах, где нужен прямой capture.
 *
 * Noop если Sentry не инициализирован — SDK внутри проверяет client существование.
 */
export const captureException = Sentry.captureException
