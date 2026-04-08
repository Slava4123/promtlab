import { create } from "zustand"
import { devtools } from "zustand/middleware"
import { api, apiVoid, setTokens, clearTokens, ensureFreshToken } from "@/api/client"
import type { User, AuthResponse } from "@/api/types"
import { setSentryUser, clearSentryUser } from "@/lib/sentry"

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean

  login: (email: string, password: string) => Promise<void>
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

      login: async (email, password) => {
        const data = await api<AuthResponse>("/auth/login", {
          method: "POST",
          body: JSON.stringify({ email, password }),
        })
        setTokens(data.tokens)
        set({ user: data.user, isAuthenticated: true })
        setSentryUser({
          id: data.user.id,
          email: data.user.email,
          username: data.user.username,
        })
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
        clearSentryUser()
        set({ user: null, isAuthenticated: false })
      },

      fetchMe: async () => {
        const user = await api<User>("/auth/me")
        set({ user, isAuthenticated: true })
      },

      restoreSession: async () => {
        // Пробуем refresh через HttpOnly cookie
        try {
          await ensureFreshToken()
          const user = await api<User>("/auth/me")
          set({ user, isAuthenticated: true, isLoading: false })
          setSentryUser({
            id: user.id,
            email: user.email,
            username: user.username,
          })
        } catch (err) {
          const msg = err instanceof Error ? err.message : ""
          const isAuthError = msg.includes("Сессия истекла") || msg.includes("unauthorized") || msg.includes("refresh failed") || msg.includes("invalid") || msg.includes("expired")
          if (isAuthError) {
            clearTokens()
            set({ user: null, isAuthenticated: false, isLoading: false })
          } else {
            // Transient error — retry on next navigation
            set({ isLoading: false })
          }
        }
      },
    }),
    { name: "auth-store" },
  ),
)
