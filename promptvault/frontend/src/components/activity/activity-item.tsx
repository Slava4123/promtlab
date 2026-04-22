import { formatDistanceToNow } from "date-fns"
import { ru } from "date-fns/locale"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  FileText,
  FileEdit,
  FileX,
  RotateCcw,
  FolderPlus,
  FolderEdit,
  FolderX,
  Tag as TagIcon,
  Tags,
  Share2,
  Link2Off,
  UserPlus,
  UserMinus,
  UserCog,
} from "lucide-react"
import type { ActivityItem as ActivityItemData } from "@/api/activity"
import { cn } from "@/lib/utils"

const ICONS: Record<string, typeof FileText> = {
  "prompt.created": FileText,
  "prompt.updated": FileEdit,
  "prompt.deleted": FileX,
  "prompt.restored": RotateCcw,
  "collection.created": FolderPlus,
  "collection.updated": FolderEdit,
  "collection.deleted": FolderX,
  "tag.created": TagIcon,
  "tag.deleted": Tags,
  "share.created": Share2,
  "share.revoked": Link2Off,
  "member.added": UserPlus,
  "member.removed": UserMinus,
  "role.changed": UserCog,
}

const ACTION_TEXT: Record<string, string> = {
  "prompt.created": "создал промпт",
  "prompt.updated": "обновил промпт",
  "prompt.deleted": "удалил промпт",
  "prompt.restored": "восстановил промпт",
  "collection.created": "создал коллекцию",
  "collection.updated": "обновил коллекцию",
  "collection.deleted": "удалил коллекцию",
  "tag.created": "создал тег",
  "tag.deleted": "удалил тег",
  "share.created": "создал публичную ссылку",
  "share.revoked": "отключил публичную ссылку",
  "member.added": "добавил в команду",
  "member.removed": "удалил из команды",
  "role.changed": "изменил роль",
}

function initials(name: string, email: string): string {
  const src = name || email || "?"
  const parts = src.split(/\s+/).filter(Boolean)
  if (parts.length === 0) return "?"
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase()
  return (parts[0]![0]! + parts[1]![0]!).toUpperCase()
}

interface ActivityItemProps {
  item: ActivityItemData
  className?: string
}

export function ActivityItem({ item, className }: ActivityItemProps) {
  const Icon = ICONS[item.event_type] ?? FileText
  const action = ACTION_TEXT[item.event_type] ?? item.event_type
  const actorName = item.actor_name || item.actor_email
  const when = formatDistanceToNow(new Date(item.created_at), { addSuffix: true, locale: ru })

  // Для role.changed — показываем from/to из metadata.
  const metadata = item.metadata as { from_role?: string; to_role?: string; version_number?: number } | undefined
  const roleHint =
    item.event_type === "role.changed" && metadata?.from_role && metadata?.to_role
      ? ` (${metadata.from_role} → ${metadata.to_role})`
      : ""
  const versionHint =
    item.event_type === "prompt.updated" && metadata?.version_number
      ? ` v${metadata.version_number}`
      : ""

  return (
    <div className={cn("flex gap-3 py-3", className)}>
      <Avatar className="size-8 shrink-0">
        <AvatarFallback className="text-xs">{initials(item.actor_name ?? "", item.actor_email)}</AvatarFallback>
      </Avatar>
      <div className="flex-1 space-y-1 overflow-hidden">
        <div className="flex items-start gap-2 text-sm">
          <Icon className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
          <div className="flex-1">
            <span className="font-medium">{actorName}</span>{" "}
            <span className="text-muted-foreground">{action}</span>
            {item.target_label ? (
              <>
                {" "}
                <span className="font-medium">«{item.target_label}»</span>
              </>
            ) : null}
            {roleHint}
            {versionHint}
          </div>
        </div>
        <div className="text-xs text-muted-foreground">{when}</div>
      </div>
    </div>
  )
}
