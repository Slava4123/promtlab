// useCurrentTeamRole — возвращает роль текущего юзера в активной команде
// (из workspace-store) или null для personal-пространства.
//
// Используется для RBAC-проверок UI: viewer не должен видеть кнопки create/edit/delete.

import { useTeams } from "@/hooks/use-teams"
import { useCurrentTeam } from "@/hooks/use-current-team"
import type { TeamRole } from "@/api/types"

export interface TeamRoleInfo {
  /** Роль в текущей команде. null = personal-пространство (юзер сам owner всего). */
  role: TeamRole | null
  /** true для personal-пространства — все права у юзера. */
  isPersonal: boolean
  /** true если юзер может писать (create/update/delete) — owner или editor или personal. */
  canWrite: boolean
  /** true если юзер может только читать и запускать execution (viewer-runner). */
  isViewer: boolean
}

export function useCurrentTeamRole(): TeamRoleInfo {
  const team = useCurrentTeam()
  const { data: teams } = useTeams()

  if (!team) {
    return { role: null, isPersonal: true, canWrite: true, isViewer: false }
  }
  const found = teams?.find((t) => t.id === team.teamId)
  const role = (found?.role ?? null) as TeamRole | null
  return {
    role,
    isPersonal: false,
    canWrite: role === "owner" || role === "editor",
    isViewer: role === "viewer",
  }
}
