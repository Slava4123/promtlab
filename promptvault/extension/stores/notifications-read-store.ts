import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

// Хранит ID прочитанных/скрытых уведомлений. Общий source-of-truth для
// NotificationsPage и useUnreadCount hook'а — без него у двух потребителей
// был дрифт (storage event не fires в той же tab после localStorage.setItem).
//
// ID-схема:
//   invitation-<id>          — приглашение в команду по invitationId
//   quota-<resource_key>     — over-limit для ресурса (stable, без used/limit)

interface NotificationsReadState {
  ids: string[] // Set'ы не сериализуются persist'ом, держим как массив
  markRead: (id: string) => void
  markAllRead: (ids: string[]) => void
  isRead: (id: string) => boolean
  clear: () => void
}

export const useNotificationsReadStore = create<NotificationsReadState>()(
  devtools(
    persist(
      (set, get) => ({
        ids: [],
        markRead: (id) =>
          set((s) => (s.ids.includes(id) ? s : { ids: [...s.ids, id] })),
        markAllRead: (ids) =>
          set((s) => {
            const merged = new Set([...s.ids, ...ids])
            return { ids: Array.from(merged) }
          }),
        isRead: (id) => get().ids.includes(id),
        clear: () => set({ ids: [] }),
      }),
      { name: "pv-notifications-read" },
    ),
    { name: "notifications-read-store" },
  ),
)
