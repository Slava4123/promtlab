import type { TeamRole } from "@/api/types"

const ROLE_CONFIG: Record<TeamRole, { label: string; className: string }> = {
  owner: { label: "Владелец", className: "bg-violet-500/15 text-violet-400 ring-violet-500/20" },
  editor: { label: "Редактор", className: "bg-blue-500/15 text-blue-400 ring-blue-500/20" },
  viewer: { label: "Читатель", className: "bg-zinc-500/15 text-zinc-400 ring-zinc-500/20" },
}

export function RoleBadge({ role }: { role: TeamRole }) {
  const config = ROLE_CONFIG[role] ?? { label: role, className: "bg-zinc-500/15 text-zinc-400 ring-zinc-500/20" }
  return (
    <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-[0.7rem] font-medium ring-1 ring-inset ${config.className}`}>
      {config.label}
    </span>
  )
}
