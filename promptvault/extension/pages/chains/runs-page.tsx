import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, CheckCircle2, XCircle, Clock, Loader2 } from "lucide-react"
import { Button } from "../../components/ui/button"
import { useChain, useExecutions } from "../../hooks/use-chains"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import { cn } from "../../lib/utils"
import type { ChainExecutionStatus } from "../../lib/types"

// Status colors через semantic tokens — синхронизируются с темой и общей
// идентичностью. Раньше hardcoded blue/emerald/amber — диссонировали с brand.
const STATUS_META: Record<
  ChainExecutionStatus,
  {
    label: string
    iconColor: string
    badgeBg: string
    cardBorder: string
    icon: React.ComponentType<{ className?: string }>
  }
> = {
  in_progress: {
    label: "В процессе",
    iconColor: "text-(--color-info)",
    badgeBg: "bg-(--color-info)/10 text-(--color-info)",
    cardBorder: "border-(--color-info)/40",
    icon: Clock,
  },
  completed: {
    label: "Завершено",
    iconColor: "text-(--color-success)",
    badgeBg: "bg-(--color-success)/10 text-(--color-success)",
    cardBorder: "border-(--color-border)",
    icon: CheckCircle2,
  },
  abandoned: {
    label: "Прервано",
    iconColor: "text-(--color-warning)",
    badgeBg: "bg-(--color-warning)/10 text-(--color-warning)",
    cardBorder: "border-(--color-border)",
    icon: XCircle,
  },
}

export function ChainRunsPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const chainId = id ? Number(id) : null
  const chainQuery = useChain(chainId)
  const execsQuery = useExecutions(chainId)

  if (chainQuery.isPending || execsQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const chain = chainQuery.data
  const execs = execsQuery.data?.items ?? []

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1 min-w-0">
          <h2 className="truncate text-sm font-semibold">История запусков</h2>
          {chain && <p className="text-[10px] text-(--color-muted-foreground)">{chain.name}</p>}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {execs.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <Clock className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Запусков пока нет</p>
            <Button
              type="button"
              variant="brand"
              size="sm"
              onClick={() => navigate(`/chains/${chainId}/run`)}
              className="mt-2"
            >
              Запустить
            </Button>
          </div>
        ) : (
          <ul className="space-y-1.5">
            {execs.map((exec) => {
              const meta = STATUS_META[exec.status]
              const Icon = meta.icon
              // Показываем relative time для завершённых (полезно — «16 ч назад»),
              // и тот же relative для in_progress (когда «начат» полезнее, чем
              // абсолютное время — «23 мин назад» сразу даёт контекст).
              const displayTime = exec.completed_at
                ? formatRelativeDate(exec.completed_at)
                : formatRelativeDate(exec.started_at)
              return (
                <li key={exec.id}>
                  <div
                    className={cn(
                      "flex items-center gap-2 rounded-md border bg-(--color-card) px-2.5 py-2",
                      meta.cardBorder,
                    )}
                  >
                    <Icon className={cn("h-4 w-4 shrink-0", meta.iconColor)} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 text-xs">
                        <span
                          className={cn(
                            "rounded px-1.5 py-0.5 text-[10px] font-medium",
                            meta.badgeBg,
                          )}
                        >
                          {meta.label}
                        </span>
                        <span className="text-(--color-muted-foreground)">
                          шаг {exec.current_step}
                        </span>
                      </div>
                      <div className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                        {displayTime}
                      </div>
                    </div>
                  </div>
                </li>
              )
            })}
          </ul>
        )}
      </div>
    </div>
  )
}
