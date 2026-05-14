import { useAuthStore } from "@/stores/auth-store"
import { useWorkspaceStore, type TeamContext } from "@/stores/workspace-store"

/**
 * useCurrentTeam — effective workspace team для текущего юзера.
 *
 * Возвращает null если:
 *   - team в store отсутствует (personal workspace явно выбран),
 *   - user ещё не загружен (restoreSession in flight),
 *   - persisted team принадлежит другому юзеру (browser crash без logout,
 *     account switch на shared computer).
 *
 * Это derived state — пересчитывается синхронно при render'е, в отличие от
 * async syncOwner() в auth-store. Без этого хука Sidebar успевает сделать
 * `?team_id=99` запросов до того, как syncOwner успеет очистить store, и
 * backend возвращает 403 (юзер не состоит в чужой команде).
 */
export function useCurrentTeam(): TeamContext | null {
  const team = useWorkspaceStore((s) => s.team)
  const ownerUserId = useWorkspaceStore((s) => s.ownerUserId)
  const userId = useAuthStore((s) => s.user?.id ?? null)

  if (!team) return null
  if (userId == null) return null
  if (ownerUserId !== null && ownerUserId !== userId) return null
  return team
}

/** Удобный шорткат — возвращает teamId или null. */
export function useCurrentTeamId(): number | null {
  return useCurrentTeam()?.teamId ?? null
}
