import { create } from "zustand"
import { devtools } from "zustand/middleware"
import { api, apiVoid, setTokens, setAccessToken, clearTokens, ensureFreshToken } from "@/api/client"
import type { User, AuthResponse, AdminLoginStepResponse, VerifyTOTPResponse } from "@/api/types"
import { setSentryUser, clearSentryUser } from "@/lib/sentry"

// SESSION_HINT_KEY — флаг localStorage указывающий что у юзера (вероятно)
// есть валидный refresh cookie. Ставится при login, снимается при logout.
// Используется restoreSession чтобы НЕ делать заведомо провальный refresh
// запрос для новых посетителей landing page — это избавляет от лишнего
// 401 в browser console (Chrome DevTools всегда логирует failed requests,
// и это нельзя подавить программно).
//
// Флаг не считается источником истины — это лишь UX optimization.
// Фактическая валидность session проверяется backend через cookie.
const SESSION_HINT_KEY = "pv_has_session"

// Экспортируем чтобы OAuth-callback (и любые нестандартные auth-flow)
// могли пометить сессию после успешной аутентификации.
export function markSessionHint() {
  try { localStorage.setItem(SESSION_HINT_KEY, "1") } catch { /* quota/disabled — ignore */ }
}
function clearSessionHint() {
  try { localStorage.removeItem(SESSION_HINT_KEY) } catch { /* ignore */ }
}
function hasSessionHint(): boolean {
  try { return localStorage.getItem(SESSION_HINT_KEY) === "1" } catch { return false }
}

/**
 * LoginResult — возможные результаты login().
 * - "ok" — обычный юзер залогинился успешно.
 * - "totp_required" — admin с TOTP, UI должен показать TOTP step.
 *   Поле preAuthToken используется для последующего verify-totp вызова.
 * - "totp_enrollment_required" — admin без enrollment; залогинен, но должен
 *   пройти enroll wizard на /admin/totp.
 */
export type LoginResult =
  | { kind: "ok" }
  | { kind: "totp_required"; preAuthToken: string; email: string }
  | { kind: "totp_enrollment_required" }

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  // sessionError отличает «refresh упал из-за сети/сервера» (transient) от
  // «refresh истёк/невалиден» (auth). При transient ProtectedRoute показывает
  // offline-UI вместо редиректа на /sign-in.
  sessionError: "auth" | "transient" | null

  login: (email: string, password: string) => Promise<LoginResult>
  verifyTOTP: (preAuthToken: string, code: string) => Promise<VerifyTOTPResponse>
  register: (email: string, password: string, name: string) => Promise<void>
  logout: () => void
  fetchMe: () => Promise<void>
  restoreSession: () => Promise<void>
}

export const useAuthStore = create<AuthState>()(
  devtools(
    (set) => ({
      user: null,
      isAuthenticated: false,
      isLoading: true,
      sessionError: null,

      login: async (email, password) => {
        // Backend может вернуть два разных JSON'а: обычный AuthResponse
        // или AdminLoginStepResponse. Различаем по наличию tokens vs
        // totp_required/totp_enrollment_required флагов.
        const data = await api<AuthResponse | AdminLoginStepResponse>("/auth/login", {
          method: "POST",
          body: JSON.stringify({ email, password }),
        })

        // Admin flow 1: TOTP required → ничего не сохраняем, возвращаем токен в UI.
        if ("totp_required" in data && data.totp_required) {
          return {
            kind: "totp_required",
            preAuthToken: data.pre_auth_token ?? "",
            email,
          }
        }

        // Admin flow 2: первый login admin без enrollment → логинимся как
        // обычно, но сообщаем UI что нужно пройти enroll wizard.
        if ("totp_enrollment_required" in data && data.totp_enrollment_required) {
          const accessToken = data.access_token ?? ""
          setAccessToken(accessToken)
          markSessionHint()
          set({ user: data.user, isAuthenticated: true })
          setSentryUser({
            id: data.user.id,
            email: data.user.email,
            username: data.user.username,
          })
          return { kind: "totp_enrollment_required" }
        }

        // Обычный юзер — стандартный flow.
        const authData = data as AuthResponse
        setTokens(authData.tokens)
        markSessionHint()
        set({ user: authData.user, isAuthenticated: true })
        setSentryUser({
          id: authData.user.id,
          email: authData.user.email,
          username: authData.user.username,
        })
        // Refresh user data to get has_unread_changelog (not in login response)
        try {
          const user = await api<User>("/auth/me")
          set({ user })
        } catch {
          // Non-critical — badge will appear on next page load
        }
        return { kind: "ok" }
      },

      verifyTOTP: async (preAuthToken, code) => {
        const data = await api<VerifyTOTPResponse>("/auth/verify-totp", {
          method: "POST",
          body: JSON.stringify({ pre_auth_token: preAuthToken, code }),
        })
        setAccessToken(data.access_token)
        markSessionHint()
        set({ user: data.user, isAuthenticated: true })
        setSentryUser({
          id: data.user.id,
          email: data.user.email,
          username: data.user.username,
        })
        return data
      },

      register: async (email, password, name) => {
        await api<{ email: string; message: string }>("/auth/register", {
          method: "POST",
          body: JSON.stringify({ email, password, name }),
        })
        // Не ставим токены — пользователь должен подтвердить email
      },

      logout: async () => {
        try {
          await apiVoid("/auth/logout", { method: "POST" })
        } catch {
          // ignore — cookie очистится на сервере
        }
        clearTokens()
        clearSessionHint()
        clearSentryUser()
        set({ user: null, isAuthenticated: false, sessionError: null })
      },

      fetchMe: async () => {
        const user = await api<User>("/auth/me")
        set({ user, isAuthenticated: true })
      },

      restoreSession: async () => {
        // Optimization: если нет session hint в localStorage, значит юзер
        // никогда не логинился или уже вышел — не делаем заведомо провальный
        // refresh запрос (избавляет от 401 noise в browser console на landing).
        if (!hasSessionHint()) {
          set({ user: null, isAuthenticated: false, isLoading: false, sessionError: null })
          return
        }

        // Transient ошибки (сеть/5xx) ретраим до 3 раз с exponential backoff.
        // Auth ошибки (401/403) — сразу разлогиниваем.
        const maxAttempts = 3
        for (let attempt = 1; attempt <= maxAttempts; attempt++) {
          try {
            await ensureFreshToken()
            const user = await api<User>("/auth/me")
            set({ user, isAuthenticated: true, isLoading: false, sessionError: null })
            setSentryUser({
              id: user.id,
              email: user.email,
              username: user.username,
            })
            return
          } catch (err) {
            const msg = err instanceof Error ? err.message : ""
            const isAuthError =
              msg.includes("Сессия истекла") ||
              msg.includes("unauthorized") ||
              msg.includes("invalid") ||
              msg.includes("expired")
            if (isAuthError) {
              clearTokens()
              clearSessionHint()
              set({ user: null, isAuthenticated: false, isLoading: false, sessionError: "auth" })
              return
            }
            // Transient — ждём и ретраим.
            if (attempt < maxAttempts) {
              await new Promise((r) => setTimeout(r, attempt * 500))
              continue
            }
            // Все попытки исчерпаны. НЕ сбрасываем auth — возможно, юзер
            // был залогинен (session hint есть), но сервер/сеть лежат.
            // ProtectedRoute покажет offline UI с кнопкой "Попробовать снова".
            set({ isLoading: false, sessionError: "transient" })
          }
        }
      },
    }),
    { name: "auth-store" },
  ),
)
