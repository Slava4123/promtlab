import { Check, X } from "lucide-react"
import type { TeamRole } from "@/api/types"

/**
 * Single source of truth для team-role permission matrix.
 *
 * Синхронизировано с backend-логикой:
 *   usecases/team + middleware team-access checker.
 * Порядок столбцов по нарастанию прав: Читатель < Редактор < Владелец.
 * Если на backend изменятся permission-правила — обязательно обновить
 * этот массив (поиск по "role-permissions-table").
 */
interface Permission {
  label: string
  allowed: Record<TeamRole, boolean>
}

const PERMISSIONS: Permission[] = [
  {
    label: "Просматривать промпты, коллекции, теги команды",
    allowed: { viewer: true, editor: true, owner: true },
  },
  {
    label: "Использовать промпты (копировать, вставлять через расширение)",
    allowed: { viewer: true, editor: true, owner: true },
  },
  {
    label: "Создавать и редактировать промпты",
    allowed: { viewer: false, editor: true, owner: true },
  },
  {
    label: "Откатывать версии промптов",
    allowed: { viewer: false, editor: true, owner: true },
  },
  {
    label: "Создавать и удалять коллекции и теги",
    allowed: { viewer: false, editor: true, owner: true },
  },
  {
    label: "Создавать публичные и share-ссылки",
    allowed: { viewer: false, editor: true, owner: true },
  },
  {
    label: "Приглашать новых участников",
    allowed: { viewer: false, editor: false, owner: true },
  },
  {
    label: "Менять роли участников",
    allowed: { viewer: false, editor: false, owner: true },
  },
  {
    label: "Удалить команду",
    allowed: { viewer: false, editor: false, owner: true },
  },
]

const ROLE_ORDER: TeamRole[] = ["viewer", "editor", "owner"]
const ROLE_LABELS: Record<TeamRole, string> = {
  viewer: "Читатель",
  editor: "Редактор",
  owner: "Владелец",
}

interface RolePermissionsTableProps {
  // highlight — подсветить столбец текущей роли юзера (контекст в командной странице).
  highlight?: TeamRole
  // compact — компактная версия для popover (меньше padding'а и шрифт).
  compact?: boolean
}

export function RolePermissionsTable({ highlight, compact = false }: RolePermissionsTableProps) {
  const headerCls = compact ? "text-[0.7rem]" : "text-[0.75rem]"
  const rowCls = compact ? "py-1.5 text-[0.75rem]" : "py-2 text-[0.82rem]"

  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <table className="w-full min-w-[360px] border-collapse">
        <thead className="bg-muted/40">
          <tr>
            <th
              scope="col"
              className={`px-3 text-left font-medium text-muted-foreground ${headerCls} ${compact ? "py-1.5" : "py-2"}`}
            >
              Действие
            </th>
            {ROLE_ORDER.map((r) => (
              <th
                key={r}
                scope="col"
                className={`px-2 text-center font-medium ${headerCls} ${compact ? "py-1.5" : "py-2"} ${
                  highlight === r ? "text-violet-400" : "text-muted-foreground"
                }`}
              >
                {ROLE_LABELS[r]}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {PERMISSIONS.map((p, i) => (
            <tr key={p.label} className={i % 2 === 0 ? "bg-background" : "bg-muted/10"}>
              <td className={`px-3 text-foreground ${rowCls}`}>{p.label}</td>
              {ROLE_ORDER.map((r) => (
                <td
                  key={r}
                  className={`px-2 text-center ${rowCls} ${highlight === r ? "bg-violet-500/[0.06]" : ""}`}
                >
                  {p.allowed[r] ? (
                    <Check className="mx-auto h-3.5 w-3.5 text-emerald-500" aria-label="разрешено" />
                  ) : (
                    <X className="mx-auto h-3.5 w-3.5 text-muted-foreground/40" aria-label="запрещено" />
                  )}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
