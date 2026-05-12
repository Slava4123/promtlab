import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  FileText,
  GitBranch,
  Loader2,
  Play,
  ArrowDown,
} from "lucide-react"
import { Button } from "../../components/ui/button"
import { useChain } from "../../hooks/use-chains"

// Vertical timeline для DAG-структуры цепочки. В узком sidepanel нет места для
// классического canvas (@xyflow/react) — поэтому показываем шаги сверху-вниз
// с visualизацией fork-branches как боковых веток.
export function ChainCanvasPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const chainId = id ? Number(id) : null
  const chainQuery = useChain(chainId)

  if (chainQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const chain = chainQuery.data
  if (!chain) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-(--color-muted-foreground)">
        Цепочка не найдена
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">{chain.name}</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {chain.steps.length === 0 ? (
          <p className="py-12 text-center text-[10px] text-(--color-muted-foreground)">
            В цепочке пока нет шагов
          </p>
        ) : (
          <div className="space-y-0">
            {chain.steps.map((step, idx) => (
              <div key={step.id} className="flex flex-col items-center">
                {/* Step node */}
                <div
                  className="w-full rounded-lg border-2 bg-(--color-card) p-3"
                  style={{
                    borderColor:
                      step.step_type === "fork"
                        ? "#a855f7"
                        : "var(--color-primary)",
                  }}
                >
                  <div className="flex items-center gap-2">
                    <span
                      className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-bold"
                      style={{
                        background:
                          step.step_type === "fork"
                            ? "#a855f7"
                            : "var(--color-primary)",
                        color: "#fff",
                      }}
                    >
                      {idx + 1}
                    </span>
                    {step.step_type === "fork" ? (
                      <GitBranch className="h-3.5 w-3.5 text-purple-500" />
                    ) : (
                      <FileText className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
                    )}
                    <span className="flex-1 truncate text-xs font-medium">
                      {step.name || step.prompt?.title || `Шаг ${idx + 1}`}
                    </span>
                  </div>
                  {step.prompt?.title && step.name && (
                    <div className="mt-1 ml-8 truncate text-[10px] text-(--color-muted-foreground)">
                      {step.prompt.title}
                    </div>
                  )}
                  {step.step_type === "fork" && step.conditions?.branches && (
                    <div className="mt-2 ml-8 space-y-1">
                      {step.conditions.branches.map((b, i) => (
                        <div
                          key={i}
                          className="flex items-center gap-1.5 text-[10px]"
                        >
                          <span className="font-mono rounded bg-purple-500/15 px-1.5 py-0.5 text-purple-500">
                            {i + 1}
                          </span>
                          <span className="text-(--color-muted-foreground)">{b.label}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
                {/* Connector */}
                {idx < chain.steps.length - 1 && (
                  <ArrowDown className="my-1 h-4 w-4 text-(--color-muted-foreground)" />
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="border-t border-(--color-border) p-2">
        <Button
          type="button"
          size="sm"
          onClick={() => navigate(`/chains/${chain.id}/run`)}
          disabled={chain.steps.length === 0}
          className="w-full gap-1.5"
        >
          <Play className="h-3.5 w-3.5" />
          Запустить цепочку
        </Button>
      </div>
    </div>
  )
}
