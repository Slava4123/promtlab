import { useNavigate } from "react-router-dom"
import { GitBranch, Play, History, ArrowLeft, Plus } from "lucide-react"
import { Button } from "../../components/ui/button"
import { ListSkeleton } from "../../components/list-skeleton"
import { useChains } from "../../hooks/use-chains"
import { cn } from "../../lib/utils"
import { pluralAfterDo } from "@pv/shared/utils/plural"
import type { Chain } from "../../lib/types"

export function ChainsIndexPage() {
  const navigate = useNavigate()
  const chainsQuery = useChains()

  if (chainsQuery.isPending) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
          <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h2 className="flex-1 text-sm font-semibold">Цепочки</h2>
        </div>
        <div className="flex-1 overflow-y-auto p-3">
          <ListSkeleton count={4} />
        </div>
      </div>
    )
  }

  const chains = chainsQuery.data?.items ?? []

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Цепочки промптов</h2>
        <Button
          type="button"
          variant="brand"
          size="sm"
          onClick={() => navigate("/chains/new")}
          className="gap-1"
        >
          <Plus className="h-3.5 w-3.5" />
          Новая
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {chains.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <GitBranch className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Цепочек пока нет</p>
            <p className="max-w-xs text-[10px] text-(--color-muted-foreground)">
              Многошаговые workflow для последовательных вызовов промптов.
            </p>
            <Button
              type="button"
              variant="brand"
              size="sm"
              onClick={() => navigate("/chains/new")}
              className="mt-2 gap-1.5"
            >
              <Plus className="h-3.5 w-3.5" />
              Создать первую
            </Button>
          </div>
        ) : (
          <ul className="space-y-2">
            {chains.map((chain) => (
              <ChainListCard key={chain.id} chain={chain} />
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

function ChainListCard({ chain }: { chain: Chain }) {
  const navigate = useNavigate()
  const stepCount = chain.step_count ?? chain.steps_preview?.length ?? 0
  const hasFork = chain.has_branching ?? false
  const savedRuns = chain.saved_runs_count ?? 0
  return (
    <li className="rounded-md border border-(--color-border) bg-(--color-card) p-3">
      <button
        type="button"
        onClick={() => navigate(`/chains/${chain.id}`)}
        className="flex w-full items-start gap-2 text-left"
      >
        <GitBranch className="mt-0.5 h-4 w-4 shrink-0 text-(--color-brand)" />
        <div className="flex-1 min-w-0">
          <h3 className="truncate text-sm font-medium hover:underline">{chain.name}</h3>
          {chain.description && (
            <p className="mt-0.5 line-clamp-2 text-[10px] text-(--color-muted-foreground)">
              {chain.description}
            </p>
          )}
        </div>
        {hasFork && (
          <span className="rounded bg-purple-500/10 px-1.5 py-0.5 text-[10px] text-purple-500">
            условная
          </span>
        )}
      </button>
      {/* Mini-graph */}
      {chain.steps_preview && chain.steps_preview.length > 0 && (
        <div className="mt-2 flex items-center gap-1 overflow-x-auto pb-1">
          {chain.steps_preview.map((step, i) => (
            <div key={i} className="flex items-center gap-1">
              <div
                className={cn(
                  "flex h-5 w-5 items-center justify-center rounded text-[9px]",
                  step.step_type === "fork"
                    ? "rotate-45 bg-purple-500/20 text-purple-500"
                    : "bg-(--color-brand)/20 text-(--color-brand)",
                )}
              >
                <span className={step.step_type === "fork" ? "-rotate-45" : ""}>
                  {step.position}
                </span>
              </div>
              {i < chain.steps_preview!.length - 1 && (
                <span className="text-(--color-muted-foreground)">›</span>
              )}
            </div>
          ))}
        </div>
      )}
      <div className="mt-2 flex items-center gap-3 text-[10px] text-(--color-muted-foreground)">
        <span>{pluralAfterDo(stepCount, "шаг", "шага", "шагов")}</span>
        {savedRuns > 0 && (
          <span>{pluralAfterDo(savedRuns, "запуск", "запуска", "запусков")}</span>
        )}
      </div>
      <div className="mt-2 flex gap-1.5">
        <Button
          type="button"
          variant="brand"
          size="sm"
          onClick={() => navigate(`/chains/${chain.id}/run`)}
          className="flex-1 gap-1.5 h-7 text-xs"
        >
          <Play className="h-3 w-3" />
          Запустить
        </Button>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => navigate(`/chains/${chain.id}/runs`)}
          className="h-7 px-2"
          aria-label="История"
          title="История запусков"
        >
          <History className="h-3 w-3" />
        </Button>
      </div>
    </li>
  )
}
