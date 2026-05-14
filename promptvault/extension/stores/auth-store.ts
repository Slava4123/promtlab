import { create } from "zustand"
import { devtools } from "zustand/middleware"
import type { MeResponse } from "@pv/shared/types"
import { ApiError } from "@pv/shared/types"
import { sendBg } from "../lib/bg-client"
import { setApiKey, clearApiKey, setWorkspace, setCollection } from "../lib/storage"
import { useQuotaStore } from "./quota-store"
import { useNotificationsReadStore } from "./notifications-read-store"

// Extension auth-store. Сейчас работает поверх API-key flow (storage.ts ставит
// apiKey в chrome.storage.local; background.ts читает и добавляет в Authorization).
// В Phase 6 расширим до полного OAuth + TOTP flow (frontend-equivalent).

interface AuthState {
  user: MeResponse | null
  isAuthenticated: boolean
  isLoading: boolean
  sessionError: "auth" | "transient" | null

  setUser: (user: MeResponse | null) => void
  loginWithKey: (apiKey: string) => Promise<MeResponse>
  logout: () => Promise<void>
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

      setUser: (user) => set({ user, isAuthenticated: !!user }),

      loginWithKey: async (apiKey) => {
        // validateKey проверит ключ и вернёт MeResponse; bg-client сам поставит
        // апи-ключ в Authorization header только из background.
        const me = await sendBg({ type: "api.validateKey", key: apiKey })
        await setApiKey(apiKey)
        set({ user: me, isAuthenticated: true, sessionError: null })
        return me
      },

      logout: async () => {
        await clearApiKey()
        // Чистим workspace selection в chrome.storage, иначе при re-login
        // другого юзера он увидит чужой teamId/collectionId.
        await setWorkspace(null)
        await setCollection(null)
        useQuotaStore.getState().clear()
        // Прочитанные уведомления — per-user state (id'ы invitation-<id> и
        // quota-<resource_key> могут коллидировать между аккаунтами на
        // shared computer). Чистим in-memory + persisted localStorage.
        useNotificationsReadStore.getState().clear()
        useNotificationsReadStore.persist?.clearStorage()
        set({ user: null, isAuthenticated: false, sessionError: null })
      },

      fetchMe: async () => {
        try {
          const me = await sendBg({ type: "api.getMe" })
          set({ user: me, isAuthenticated: true, sessionError: null })
        } catch (err) {
          // Используем err.code (надёжно), не err.message (зависит от локализации).
          const isAuthErr = err instanceof ApiError && err.code === "unauthorized"
          if (isAuthErr) {
            set({ user: null, isAuthenticated: false, sessionError: "auth" })
          } else {
            set({ sessionError: "transient" })
          }
          throw err
        }
      },

      restoreSession: async () => {
        try {
          const me = await sendBg({ type: "api.getMe" })
          set({
            user: me,
            isAuthenticated: true,
            isLoading: false,
            sessionError: null,
          })
        } catch (err) {
          const isAuthErr = err instanceof ApiError && err.code === "unauthorized"
          if (isAuthErr) {
            set({
              user: null,
              isAuthenticated: false,
              isLoading: false,
              sessionError: "auth",
            })
          } else {
            set({ isLoading: false, sessionError: "transient" })
          }
        }
      },
    }),
    { name: "auth-store" },
  ),
)
