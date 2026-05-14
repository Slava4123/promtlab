import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

export type TeamContext = {
  teamSlug: string
  teamId: number
  teamName: string
}

interface WorkspaceState {
  team: TeamContext | null
  // ownerUserId — id юзера, для которого persisted team. Привязка нужна,
  // чтобы при смене юзера (login/restoreSession другого аккаунта на той
  // же машине) старый team автоматически очищался. Без этого frontend
  // делал ?team_id=N для команды, в которой новый юзер не состоит → 403
  // на холодном старте и UX-мигание ошибок.
  ownerUserId: number | null
  setTeam: (slug: string, id: number, name: string, ownerUserId: number) => void
  clearTeam: () => void
  // Вызывается auth-store после login/restoreSession/verifyTOTP. Если текущий
  // ownerUserId не совпадает с пришедшим — workspace принадлежал другому
  // юзеру (browser crash без logout, account switch), очищаем.
  syncOwner: (userId: number) => void
}

export const useWorkspaceStore = create<WorkspaceState>()(
  devtools(
    persist(
      (set, get) => ({
        team: null,
        ownerUserId: null,
        setTeam: (slug, id, name, ownerUserId) =>
          set({ team: { teamSlug: slug, teamId: id, teamName: name }, ownerUserId }),
        clearTeam: () => set({ team: null, ownerUserId: null }),
        syncOwner: (userId) => {
          const current = get().ownerUserId
          if (current !== null && current !== userId) {
            // Workspace принадлежал прошлому юзеру — чистим, чтобы не
            // отправлять stale team_id запросы.
            set({ team: null, ownerUserId: userId })
          } else if (current === null && get().team !== null) {
            // Pre-migration data: team есть, но ownerUserId не сохранён.
            // Привязываем к текущему юзеру (one-time).
            set({ ownerUserId: userId })
          } else if (current === null) {
            set({ ownerUserId: userId })
          }
        },
      }),
      { name: "workspace-store" },
    ),
    { name: "workspace-store" },
  ),
)
