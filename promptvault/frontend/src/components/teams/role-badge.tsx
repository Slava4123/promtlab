import { Link } from "react-router-dom"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { RolePermissionsTable } from "./role-permissions-table"
import type { TeamRole } from "@/api/types"

const ROLE_CONFIG: Record<TeamRole, { label: string; className: string }> = {
  owner: { label: "Владелец", className: "bg-violet-500/15 text-violet-400 ring-violet-500/20" },
  editor: { label: "Редактор", className: "bg-blue-500/15 text-blue-400 ring-blue-500/20" },
  viewer: { label: "Читатель", className: "bg-zinc-500/15 text-zinc-400 ring-zinc-500/20" },
}

const ROLE_SUMMARY: Record<TeamRole, string> = {
  owner: "Полный доступ: приглашать участников, менять роли, удалить команду.",
  editor: "Управляет промптами, коллекциями и тегами команды. Не управляет участниками.",
  viewer: "Просматривает и использует промпты. Не создаёт и не редактирует.",
}

interface RoleBadgeProps {
  role: TeamRole
  // interactive=false — отключает popover (используется в местах где клик
  // перекрывается другим действием, например внутри кнопки).
  interactive?: boolean
}

/**
 * RoleBadge показывает роль участника команды. По клику открывает popover
 * с таблицей прав всех ролей — контекстная документация без ухода со страницы
 * (best practice 2025: popover для справочного контента, не tooltip/modal).
 */
export function RoleBadge({ role, interactive = true }: RoleBadgeProps) {
  const config = ROLE_CONFIG[role] ?? { label: role, className: "bg-zinc-500/15 text-zinc-400 ring-zinc-500/20" }
  const badgeCls = `inline-flex items-center rounded-md px-2 py-0.5 text-[0.7rem] font-medium ring-1 ring-inset ${config.className}`

  if (!interactive) {
    return <span className={badgeCls}>{config.label}</span>
  }

  return (
    <Popover>
      <PopoverTrigger
        className={`${badgeCls} cursor-help transition-opacity hover:opacity-80 focus:outline-none focus-visible:ring-2 focus-visible:ring-violet-500/40`}
        aria-label={`Роль: ${config.label}. Нажмите чтобы посмотреть список прав`}
      >
        {config.label}
      </PopoverTrigger>
      <PopoverContent className="w-[min(540px,calc(100vw-24px))]">
        <div className="mb-2 flex items-baseline justify-between gap-3">
          <h3 className="text-sm font-semibold text-foreground">{config.label}</h3>
          <Link
            to="/help#team-roles"
            className="text-[0.72rem] text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
          >
            Подробнее в Помощи
          </Link>
        </div>
        <p className="mb-3 text-[0.78rem] leading-relaxed text-muted-foreground">
          {ROLE_SUMMARY[role]}
        </p>
        <RolePermissionsTable highlight={role} compact />
      </PopoverContent>
    </Popover>
  )
}
