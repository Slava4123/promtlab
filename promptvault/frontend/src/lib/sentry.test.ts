import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

// Mock @sentry/react целиком — не хотим чтобы реальный Sentry SDK пытался
// инициализироваться в jsdom (fetch к DSN, cookies, etc.).
vi.mock("@sentry/react", () => ({
  init: vi.fn(),
  setUser: vi.fn(),
  captureException: vi.fn(),
  browserTracingIntegration: vi.fn(() => ({ name: "BrowserTracing" })),
}))

import * as Sentry from "@sentry/react"
import { initSentry, setSentryUser, clearSentryUser } from "./sentry"

describe("lib/sentry", () => {
  const originalEnv = { ...import.meta.env }

  beforeEach(() => {
    vi.clearAllMocks()
    // Vite env vars в тестах — можно мутировать через stubEnv (vitest).
    vi.stubEnv("VITE_SENTRY_ENABLED", "false")
    vi.stubEnv("VITE_SENTRY_DSN", "")
  })

  afterEach(() => {
    vi.unstubAllEnvs()
    Object.assign(import.meta.env, originalEnv)
  })

  describe("initSentry", () => {
    it("skips init when VITE_SENTRY_ENABLED !== 'true'", () => {
      vi.stubEnv("VITE_SENTRY_ENABLED", "false")
      initSentry()
      expect(Sentry.init).not.toHaveBeenCalled()
    })

    it("skips init when VITE_SENTRY_ENABLED is missing entirely", () => {
      vi.stubEnv("VITE_SENTRY_ENABLED", "")
      initSentry()
      expect(Sentry.init).not.toHaveBeenCalled()
    })

    it("skips init when VITE_SENTRY_ENABLED=true but DSN is empty", () => {
      vi.stubEnv("VITE_SENTRY_ENABLED", "true")
      vi.stubEnv("VITE_SENTRY_DSN", "")
      initSentry()
      expect(Sentry.init).not.toHaveBeenCalled()
    })

    it("calls Sentry.init with correct config when enabled with valid DSN", () => {
      vi.stubEnv("VITE_SENTRY_ENABLED", "true")
      vi.stubEnv("VITE_SENTRY_DSN", "http://key@glitchtip.local/1")
      vi.stubEnv("VITE_SENTRY_ENVIRONMENT", "production")
      vi.stubEnv("VITE_SENTRY_RELEASE", "abc123")
      vi.stubEnv("VITE_SENTRY_TRACES_SAMPLE_RATE", "0.25")

      initSentry()

      expect(Sentry.init).toHaveBeenCalledTimes(1)
      const call = (Sentry.init as ReturnType<typeof vi.fn>).mock.calls[0][0]
      expect(call.dsn).toBe("http://key@glitchtip.local/1")
      expect(call.environment).toBe("production")
      expect(call.release).toBe("abc123")
      expect(call.tracesSampleRate).toBeCloseTo(0.25)
      expect(call.sampleRate).toBe(1.0)
      expect(call.beforeSend).toBeInstanceOf(Function)
    })

    it("falls back to defaults for missing optional env vars", () => {
      vi.stubEnv("VITE_SENTRY_ENABLED", "true")
      vi.stubEnv("VITE_SENTRY_DSN", "http://key@glitchtip.local/1")
      // Не ставим остальные VITE_SENTRY_*.

      initSentry()

      const call = (Sentry.init as ReturnType<typeof vi.fn>).mock.calls[0][0]
      expect(call.environment).toBe("production")
      expect(call.release).toBe("dev")
      expect(call.tracesSampleRate).toBe(0.0)
    })

    it("beforeSend scrubs Authorization and Cookie headers", () => {
      vi.stubEnv("VITE_SENTRY_ENABLED", "true")
      vi.stubEnv("VITE_SENTRY_DSN", "http://key@glitchtip.local/1")

      initSentry()

      const beforeSend = (Sentry.init as ReturnType<typeof vi.fn>).mock.calls[0][0]
        .beforeSend as (event: unknown) => unknown

      const event = {
        request: {
          headers: {
            "Content-Type": "application/json",
            Authorization: "Bearer SECRET_JWT",
            authorization: "bearer lowercase",
            Cookie: "refresh_token=SECRET",
            cookie: "session=OTHER",
          },
        },
      }

      const scrubbed = beforeSend(event) as typeof event
      expect(scrubbed.request.headers["Authorization"]).toBeUndefined()
      expect(scrubbed.request.headers["authorization"]).toBeUndefined()
      expect(scrubbed.request.headers["Cookie"]).toBeUndefined()
      expect(scrubbed.request.headers["cookie"]).toBeUndefined()
      // Non-sensitive header остаётся нетронутым.
      expect(scrubbed.request.headers["Content-Type"]).toBe("application/json")
    })
  })

  describe("setSentryUser", () => {
    it("passes user object to Sentry.setUser with stringified id", () => {
      setSentryUser({ id: 42, email: "user@example.com", username: "alice" })
      expect(Sentry.setUser).toHaveBeenCalledWith({
        id: "42",
        email: "user@example.com",
        username: "alice",
      })
    })

    it("accepts string id without conversion issues", () => {
      setSentryUser({ id: "abc-123" })
      expect(Sentry.setUser).toHaveBeenCalledWith({
        id: "abc-123",
        email: undefined,
        username: undefined,
      })
    })
  })

  describe("clearSentryUser", () => {
    it("calls Sentry.setUser(null) to clear user context", () => {
      clearSentryUser()
      expect(Sentry.setUser).toHaveBeenCalledWith(null)
    })
  })
})
