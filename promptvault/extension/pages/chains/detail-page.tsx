import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  Play,
  History,
  Edit3,
  Trash2,
  Loader2,
  GitBranch,
  FileText,
  LayoutGrid,
} from "lucide-react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import { useToast } from "../../components/ui/toaster"
import { useChain } from "../../hooks/use-chains"
import { sendBg } from "../../lib/bg-client"

export function ChainDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()
  const chainId = id ? Number(id) : null
  const chainQuery = useChain(chainId)
  const [deleteOpen, setDeleteOpen] = useState(false)

  const deleteMut = useMutation({
    mutationFn: () => sendBg({ type: "api.deleteChain", id: chainId! }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["chains"] })
      toast({ title: "Цепочка удалена", variant: "info" })
      navigate("/chains", { replace: true })
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось удалить", description: err.message, variant: "error" }),
  })

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
      <div className="flex items-center gap-1 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate("/chains")} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">{chain.name}</h2>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => navigate(`/chains/${chain.id}/edit`)}
          aria-label="Редактировать"
        >
          <Edit3 className="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => setDeleteOpen(true)}
          aria-label="Удалить"
          className="text-(--color-destructive)"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {chain.description && (
          <p className="text-xs text-(--color-muted-foreground)">{chain.description}</p>
        )}

        <div className="flex items-center gap-3 text-[10px] text-(--color-muted-foreground)">
          <span>{chain.steps.length} шагов</span>
          <span>•</span>
          <span>{chain.team_id ? "командная" : "личная"}</span>
        </div>

        <section className="space-y-1.5">
          <h3 className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
            Шаги
          </h3>
          {chain.steps.length === 0 ? (
            <p className="rounded-md border border-dashed border-(--color-border) p-4 text-center text-[10px] text-(--color-muted-foreground)">
              Шагов ещё нет. Откройте редактор, чтобы добавить.
            </p>
          ) : (
            <ol className="space-y-1.5">
              {chain.steps.map((step, idx) => (
                <li
                  key={step.id}
                  className="flex items-start gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-xs"
                >
                  <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded text-[10px] font-medium"
                    style={{
                      background: step.step_type === "fork" ? "rgba(168, 85, 247, 0.15)" : "rgba(124, 58, 237, 0.15)",
                      color: step.step_type === "fork" ? "#a855f7" : "var(--color-primary)",
                    }}
                  >
                    {idx + 1}
                  </span>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5">
                      {step.step_type === "fork" ? (
                        <GitBranch className="h-3 w-3 shrink-0 text-purple-500" />
                      ) : (
                        <FileText className="h-3 w-3 shrink-0 text-(--color-muted-foreground)" />
                      )}
                      <span className="truncate font-medium">
                        {step.name || step.prompt?.title || `Шаг ${idx + 1}`}
                      </span>
                    </div>
                    {step.step_type === "fork" && step.conditions && (
                      <div className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                        {step.conditions.branches?.length ?? 0} ветки
                      </div>
                    )}
                  </div>
                </li>
              ))}
            </ol>
          )}
        </section>
      </div>

      <div className="flex items-center gap-2 border-t border-(--color-border) p-2">
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => navigate(`/chains/${chain.id}/runs`)}
          className="gap-1.5"
        >
          <History className="h-3.5 w-3.5" />
          История
        </Button>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => navigate(`/chains/${chain.id}/canvas`)}
          className="gap-1.5"
          aria-label="Граф"
        >
          <LayoutGrid className="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          size="sm"
          onClick={() => navigate(`/chains/${chain.id}/run`)}
          disabled={chain.steps.length === 0}
          className="flex-1 gap-1.5"
        >
          <Play className="h-3.5 w-3.5" />
          Запустить
        </Button>
      </div>

      <ConfirmDialog
        open={deleteOpen}
        title="Удалить цепочку?"
        description="Цепочка будет удалена безвозвратно. История запусков сохранится."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setDeleteOpen(false)}
      />
    </div>
  )
}
