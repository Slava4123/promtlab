import { useEffect, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  ChevronDown,
  ChevronUp,
  GitBranch,
  FileText,
  Loader2,
  Plus,
  Save,
  Trash2,
} from "lucide-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { Textarea } from "../../components/ui/textarea"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import { useToast } from "../../components/ui/toaster"
import { useChain } from "../../hooks/use-chains"
import { sendBg } from "../../lib/bg-client"
import { useWorkspaceStore } from "../../stores/workspace-store"
import { cn } from "../../lib/utils"
import type { Prompt } from "../../lib/types"

export function ChainEditorPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()
  const chainId = id ? Number(id) : null
  const chainQuery = useChain(chainId)
  const [addPromptOpen, setAddPromptOpen] = useState(false)
  const [removeStepId, setRemoveStepId] = useState<number | null>(null)

  // Edit name/description state
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (chainQuery.data) {
      setName(chainQuery.data.name)
      setDescription(chainQuery.data.description ?? "")
      setDirty(false)
    }
  }, [chainQuery.data])

  const updateChainMut = useMutation({
    mutationFn: () =>
      sendBg({
        type: "api.updateChain",
        id: chainId!,
        body: { name: name.trim(), description: description.trim() },
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["chains", chainId] })
      void qc.invalidateQueries({ queryKey: ["chains"] })
      toast({ title: "Сохранено", variant: "success", durationMs: 1500 })
      setDirty(false)
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось сохранить", description: err.message, variant: "error" }),
  })

  const removeStepMut = useMutation({
    mutationFn: (stepId: number) =>
      sendBg({ type: "api.removeChainStep", chainId: chainId!, stepId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["chains", chainId] })
      toast({ title: "Шаг удалён", variant: "info" })
      setRemoveStepId(null)
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось удалить", description: err.message, variant: "error" }),
  })

  const moveUpMut = useMutation({
    mutationFn: (stepId: number) =>
      sendBg({ type: "api.moveStepUp", chainId: chainId!, stepId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["chains", chainId] })
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось переместить", description: err.message, variant: "error" }),
  })

  const moveDownMut = useMutation({
    mutationFn: (stepId: number) =>
      sendBg({ type: "api.moveStepDown", chainId: chainId!, stepId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["chains", chainId] })
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось переместить", description: err.message, variant: "error" }),
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
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(`/chains/${chain.id}`)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">Редактор</h2>
        <Button
          type="button"
          size="sm"
          onClick={() => updateChainMut.mutate()}
          disabled={updateChainMut.isPending || !dirty || !name.trim()}
          className="gap-1.5"
        >
          {updateChainMut.isPending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Save className="h-3.5 w-3.5" />
          )}
          Сохранить
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {/* Name + description */}
        <section className="space-y-2">
          <div className="space-y-1">
            <Label htmlFor="chain-name">Название</Label>
            <Input
              id="chain-name"
              value={name}
              onChange={(e) => {
                setName(e.target.value)
                setDirty(true)
              }}
              maxLength={100}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="chain-desc">Описание</Label>
            <Textarea
              id="chain-desc"
              value={description}
              onChange={(e) => {
                setDescription(e.target.value)
                setDirty(true)
              }}
              rows={2}
              maxLength={2000}
            />
          </div>
        </section>

        {/* Steps */}
        <section className="space-y-2">
          <div className="flex items-center justify-between">
            <h3 className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
              Шаги
            </h3>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => setAddPromptOpen(true)}
              className="h-7 gap-1 text-[10px]"
            >
              <Plus className="h-3 w-3" />
              Добавить
            </Button>
          </div>

          {chain.steps.length === 0 ? (
            <p className="rounded-md border border-dashed border-(--color-border) p-4 text-center text-[10px] text-(--color-muted-foreground)">
              Нет шагов. Нажмите «Добавить», чтобы выбрать промпт.
            </p>
          ) : (
            <ul className="space-y-1.5">
              {chain.steps.map((step, idx) => (
                <li
                  key={step.id}
                  className="rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-xs"
                >
                  <div className="flex items-start gap-2">
                    <div className="flex flex-col gap-0.5">
                      <button
                        type="button"
                        onClick={() => moveUpMut.mutate(step.id)}
                        disabled={idx === 0 || moveUpMut.isPending}
                        className="rounded p-0.5 text-(--color-muted-foreground) hover:bg-(--color-muted) disabled:opacity-30"
                        aria-label="Выше"
                      >
                        <ChevronUp className="h-3 w-3" />
                      </button>
                      <button
                        type="button"
                        onClick={() => moveDownMut.mutate(step.id)}
                        disabled={idx === chain.steps.length - 1 || moveDownMut.isPending}
                        className="rounded p-0.5 text-(--color-muted-foreground) hover:bg-(--color-muted) disabled:opacity-30"
                        aria-label="Ниже"
                      >
                        <ChevronDown className="h-3 w-3" />
                      </button>
                    </div>
                    <span
                      className="flex h-5 w-5 shrink-0 items-center justify-center rounded text-[10px] font-medium"
                      style={{
                        background:
                          step.step_type === "fork"
                            ? "rgba(168, 85, 247, 0.15)"
                            : "rgba(124, 58, 237, 0.15)",
                        color:
                          step.step_type === "fork" ? "#a855f7" : "var(--color-primary)",
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
                      {step.prompt?.title && step.name && (
                        <div className="mt-0.5 truncate text-[10px] text-(--color-muted-foreground)">
                          {step.prompt.title}
                        </div>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={() => setRemoveStepId(step.id)}
                      className="rounded p-0.5 text-(--color-muted-foreground) hover:text-(--color-destructive)"
                      aria-label="Удалить шаг"
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>

        <p className="text-[10px] text-(--color-muted-foreground)">
          Подсказка: тонкое редактирование переменных (variable_mapping), conditions
          для forks и tree-структура — в веб-приложении.
        </p>
      </div>

      <AddPromptDialog
        open={addPromptOpen}
        chainId={chain.id}
        onClose={() => setAddPromptOpen(false)}
        onAdded={() => {
          setAddPromptOpen(false)
          void qc.invalidateQueries({ queryKey: ["chains", chainId] })
        }}
      />

      <ConfirmDialog
        open={removeStepId !== null}
        title="Удалить шаг?"
        description="Шаг будет удалён из цепочки. История выполнений сохранится."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={() => {
          if (removeStepId) removeStepMut.mutate(removeStepId)
        }}
        onClose={() => setRemoveStepId(null)}
      />
    </div>
  )
}

// Простой prompt-picker: показывает список промптов в текущем workspace,
// выбираешь → добавляем как новый шаг в конец цепочки.
function AddPromptDialog({
  open,
  chainId,
  onClose,
  onAdded,
}: {
  open: boolean
  chainId: number
  onClose: () => void
  onAdded: () => void
}) {
  const { toast } = useToast()
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const [query, setQuery] = useState("")
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [stepName, setStepName] = useState("")

  const promptsQuery = useQuery({
    queryKey: ["prompts", "picker", teamId],
    queryFn: () =>
      sendBg({
        type: "api.fetchPrompts",
        page: 1,
        pageSize: 100,
        filter: { teamId },
      }),
    enabled: open,
    staleTime: 60_000,
  })

  const addMut = useMutation({
    mutationFn: () =>
      sendBg({
        type: "api.addChainStep",
        chainId,
        body: {
          prompt_id: selectedId!,
          name: stepName.trim(),
          step_type: "prompt",
        },
      }),
    onSuccess: () => {
      toast({ title: "Шаг добавлен", variant: "success", durationMs: 1500 })
      setQuery("")
      setSelectedId(null)
      setStepName("")
      onAdded()
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось добавить", description: err.message, variant: "error" }),
  })

  if (!open) return null

  const prompts: Prompt[] = promptsQuery.data?.items ?? []
  const filtered = query.trim()
    ? prompts.filter((p) => p.title.toLowerCase().includes(query.toLowerCase()))
    : prompts

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} aria-hidden />
      <div className="relative flex h-[80vh] w-full max-w-sm flex-col rounded-lg border border-(--color-border) bg-(--color-background) shadow-xl">
        <header className="border-b border-(--color-border) p-3">
          <h3 className="text-sm font-semibold">Выбрать промпт</h3>
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Поиск…"
            autoFocus
            className="mt-2"
          />
        </header>

        <div className="flex-1 overflow-y-auto p-2">
          {promptsQuery.isPending ? (
            <div className="flex justify-center py-6">
              <Loader2 className="h-4 w-4 animate-spin text-(--color-muted-foreground)" />
            </div>
          ) : filtered.length === 0 ? (
            <p className="py-6 text-center text-[10px] text-(--color-muted-foreground)">
              Нет промптов
            </p>
          ) : (
            <ul className="space-y-1">
              {filtered.map((p) => (
                <li key={p.id}>
                  <button
                    type="button"
                    onClick={() => setSelectedId(p.id)}
                    className={cn(
                      "w-full rounded-md border px-2 py-1.5 text-left text-xs transition-colors",
                      selectedId === p.id
                        ? "border-(--color-primary) bg-(--color-primary)/10"
                        : "border-(--color-border) hover:bg-(--color-muted)/40",
                    )}
                  >
                    <div className="truncate font-medium">{p.title}</div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        {selectedId !== null && (
          <div className="border-t border-(--color-border) p-2 space-y-2">
            <div className="space-y-1">
              <Label htmlFor="step-name">Название шага (необязательно)</Label>
              <Input
                id="step-name"
                value={stepName}
                onChange={(e) => setStepName(e.target.value)}
                placeholder="Brief"
                maxLength={100}
              />
            </div>
          </div>
        )}

        <footer className="flex gap-2 border-t border-(--color-border) p-2">
          <Button type="button" variant="outline" size="sm" onClick={onClose} className="flex-1">
            Отмена
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={() => addMut.mutate()}
            disabled={selectedId === null || addMut.isPending}
            className="flex-1 gap-1.5"
          >
            {addMut.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
            Добавить
          </Button>
        </footer>
      </div>
    </div>
  )
}
