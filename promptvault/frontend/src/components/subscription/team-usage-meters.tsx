import { Users } from "lucide-react"

import { useTeamUsage } from "@/hooks/use-subscription"
import type { QuotaInfo } from "@/api/types"
import { Skeleton } from "@/components/ui/skeleton"

interface TeamUsageMetersProps {
  /** slug команды для GET /api/teams/{slug}/usage. */
  slug: string
  /** Опционально — название команды для заголовка. Если не передано, берётся
   * из ответа API (team_name). Передавать имеет смысл когда название
   * уже известно от родителя — это избегает «прыжка» имени при загрузке. */
  teamName?: string
  className?: string
}

/**
 * TeamUsageMeters — progress-bars для team-pool лимитов одной команды.
 * Pack TU. Показывает 3 ресурса: промпты, коллекции, цепочки.
 *
 * Лимиты из плана owner'а команды — для всех участников применяется
 * одно и то же значение (Pack T model).
 */
export function TeamUsageMeters({ slug, teamName, className = "" }: TeamUsageMetersProps) {
  const { data, isLoading, error } = useTeamUsage(slug)

  const displayName = teamName ?? data?.team_name ?? "Команда"

  return (
    <div className={`rounded-lg border border-border bg-card/50 p-4 ${className}`}>
      <div className="mb-3 flex items-center gap-2">
        <Users className="h-4 w-4 text-muted-foreground" />
        <h3 className="text-sm font-medium text-foreground">{displayName}</h3>
        {data && (
          <span className="ml-auto rounded-full border border-border/60 px-2 py-0.5 text-[0.65rem] uppercase tracking-wide text-muted-foreground">
            план владельца: {data.owner_plan_id}
          </span>
        )}
      </div>

      {isLoading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-6" />
          ))}
        </div>
      )}

      {error && (
        <p className="text-xs text-destructive">Не удалось загрузить использование команды.</p>
      )}

      {data && (
        <div className="space-y-3">
          <Meter label="Промпты" info={data.prompts} />
          <Meter label="Коллекции" info={data.collections} />
          {data.chains.limit > 0 && <Meter label="Цепочки" info={data.chains} />}
        </div>
      )}
    </div>
  )
}

function Meter({ label, info }: { label: string; info: QuotaInfo }) {
  if (info.limit <= 0) return null
  const pct = Math.min((info.used / info.limit) * 100, 100)
  const color =
    pct >= 90 ? "bg-red-500" : pct >= 75 ? "bg-amber-500" : "bg-emerald-500"

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium tabular-nums">
          {info.used.toLocaleString("ru-RU")} / {info.limit.toLocaleString("ru-RU")}
        </span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-muted/40">
        <div className={`h-full rounded-full ${color} transition-all`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}
