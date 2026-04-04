import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

export type TeamContext = {
  teamSlug: string
  teamId: number
  teamName: string
}

interface WorkspaceState {
  team: TeamContext | null
  setTeam: (slug: string, id: number, name: string) => void
  clearTeam: () => void
}

export const useWorkspaceStore = create<WorkspaceState>()(
  devtools(
    persist(
      (set) => ({
        team: null,
        setTeam: (slug, id, name) => set({ team: { teamSlug: slug, teamId: id, teamName: name } }),
        clearTeam: () => set({ team: null }),
      }),
      { name: "workspace-store" },
    ),
    { name: "workspace-store" },
  ),
)
