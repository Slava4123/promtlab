import { useMyInvitations } from "./use-invitations"
import { useUsageSummary } from "./use-usage-summary"
import { useNotificationsReadStore } from "../stores/notifications-read-store"
import { QUOTA_KEYS, quotaByKey } from "../lib/types"

// Считает число непрочитанных уведомлений = pending invitations +
// over-limit ресурсы, минус прочитанные (Zustand store).
// Источник правды — useNotificationsReadStore (общий с NotificationsPage),
// поэтому обновление сразу реактивно во всех потребителях без storage events.
export function useUnreadCount(): number {
  const invitations = useMyInvitations()
  const usage = useUsageSummary()
  const readIds = useNotificationsReadStore((s) => s.ids)

  const readSet = new Set(readIds)
  let count = 0

  const pending = (invitations.data ?? []).filter((i) => i.status === "pending")
  for (const inv of pending) {
    if (!readSet.has(`invitation-${inv.id}`)) count++
  }

  if (usage.data) {
    for (const key of QUOTA_KEYS) {
      const info = quotaByKey(usage.data, key)
      if (info.limit <= 0) continue
      if (info.used >= info.limit) {
        // Stable ID — только по типу ресурса, не по used/limit. Иначе при
        // refetch usage числа могут меняться → новый id → потеря read state.
        if (!readSet.has(`quota-${key}`)) count++
      }
    }
  }

  return count
}
