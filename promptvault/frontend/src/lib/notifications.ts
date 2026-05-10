// Derived notifications для bell-popup. Объединяет 2 источника:
//   - team invitations (server data — useMyInvitations)
//   - quota over-limit warnings (derived из useUsage)
//
// Read-state хранится в localStorage по stable id уведомления. Для quota
// id содержит used+limit — если usage изменилось (юзер удалил лишний промпт
// или, наоборот, добавил), id меняется и уведомление снова unread.
// Для invitation id = `invitation-<id>` — приглашение «исчезает» из счётчика
// после accept/decline на стороне сервера (мы просто не получаем его в списке).

import type { TeamInvitation } from "@/api/types"
import type { UsageSummary, QuotaInfo } from "@/api/types"

export type NotifKind = "invitation" | "quota_over"

export interface Notification {
  id: string
  kind: NotifKind
  title: string
  body: string
  /** Для invitation — данные для accept/decline. */
  invitation?: TeamInvitation
  /** Для quota_over — куда вести по основной кнопке. */
  cta?: { label: string; href: string }
}

const READ_STORAGE_KEY = "pv_notifications_read"

function loadReadSet(): Set<string> {
  try {
    const raw = localStorage.getItem(READ_STORAGE_KEY)
    if (!raw) return new Set()
    const arr = JSON.parse(raw) as string[]
    return new Set(arr)
  } catch {
    return new Set()
  }
}

function saveReadSet(set: Set<string>) {
  try {
    localStorage.setItem(READ_STORAGE_KEY, JSON.stringify(Array.from(set)))
  } catch {
    /* private mode / quota — read-state не персистится, но не критично */
  }
}

export function markRead(id: string) {
  const set = loadReadSet()
  set.add(id)
  saveReadSet(set)
}

export function markAllRead(ids: string[]) {
  const set = loadReadSet()
  ids.forEach((id) => set.add(id))
  saveReadSet(set)
}

export function clearOldReads(currentIds: Set<string>) {
  // Чистим записи которых уже нет в живых notifications — иначе localStorage
  // растёт без границ за счёт старых quota-id (used/limit меняются).
  const set = loadReadSet()
  const trimmed = new Set<string>()
  set.forEach((id) => {
    if (currentIds.has(id)) trimmed.add(id)
  })
  if (trimmed.size !== set.size) {
    saveReadSet(trimmed)
  }
}

export function isRead(id: string): boolean {
  return loadReadSet().has(id)
}

// Per-quota meta для текста уведомления. href ведёт на список где юзер
// может удалить лишние записи.
// Phase 16-Y: share_links убран — на share больше нет квот (TTL вместо).
const QUOTA_META: Record<keyof Pick<UsageSummary, "prompts" | "collections" | "teams" | "chains">, {
  noun: (n: number) => string
  cta: string
  href: string
}> = {
  prompts: {
    noun: (n) => plur(n, "промпт", "промпта", "промптов"),
    cta: "Перейти к промптам",
    href: "/dashboard",
  },
  collections: {
    noun: (n) => plur(n, "коллекция", "коллекции", "коллекций"),
    cta: "Перейти к коллекциям",
    href: "/collections",
  },
  teams: {
    noun: (n) => plur(n, "команда", "команды", "команд"),
    cta: "Перейти к командам",
    href: "/teams",
  },
  chains: {
    noun: (n) => plur(n, "цепочка", "цепочки", "цепочек"),
    cta: "Перейти к цепочкам",
    href: "/chains",
  },
}

function plur(n: number, one: string, few: string, many: string): string {
  const m10 = n % 10
  const m100 = n % 100
  if (m10 === 1 && m100 !== 11) return one
  if (m10 >= 2 && m10 <= 4 && (m100 < 10 || m100 >= 20)) return few
  return many
}

/** Соберём все уведомления из источников. Read-state применяется в notification-center
 *  (на этом уровне возвращаем все, чтобы UI мог показывать badge с unread count). */
export function buildNotifications(
  invitations: TeamInvitation[] | undefined,
  usage: UsageSummary | undefined,
): Notification[] {
  const out: Notification[] = []

  // 1. Приглашения — каждое отдельно. id уникален по invitation.id, статус read
  // не нужен (invitation после accept/decline пропадает из API-ответа).
  for (const inv of invitations ?? []) {
    out.push({
      id: `invitation-${inv.id}`,
      kind: "invitation",
      title: `Приглашение в «${inv.team_name}»`,
      body: `${inv.inviter_name} приглашает вас как ${roleRu(inv.role)}.`,
      invitation: inv,
    })
  }

  // 2. Quota over-limit — отдельное уведомление на каждую категорию.
  if (usage) {
    for (const [key, meta] of Object.entries(QUOTA_META) as Array<[
      keyof typeof QUOTA_META,
      typeof QUOTA_META[keyof typeof QUOTA_META],
    ]>) {
      const info = usage[key] as QuotaInfo | undefined
      if (!info || info.limit <= 0 || info.used <= info.limit) continue
      const over = info.used - info.limit
      out.push({
        id: `quota-${key}-${info.used}-${info.limit}`,
        kind: "quota_over",
        title: `Превышен лимит — ${over} ${meta.noun(over)} сверх плана`,
        body: `Сейчас у вас ${info.used}, лимит — ${info.limit}. Удалите лишние, чтобы создавать новые.`,
        cta: { label: meta.cta, href: meta.href },
      })
    }
  }

  return out
}

function roleRu(role: string): string {
  switch (role) {
    case "owner": return "владельца"
    case "editor": return "редактора"
    case "viewer": return "читателя"
    default: return role
  }
}
