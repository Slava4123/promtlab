import { useState, useEffect, useRef } from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Loader2,
  Send,
  Copy,
  GitBranch,
} from "lucide-react"
import { Button } from "../../components/ui/button"
import { Label } from "../../components/ui/label"
import { Textarea } from "../../components/ui/textarea"
import { useToast } from "../../components/ui/toaster"
import { pluralAfterDo } from "@pv/shared/utils/plural"
import { useChain, useStartExecution, useExecution, useAdvanceStep } from "../../hooks/use-chains"
import { useInsertPrompt } from "../../hooks/use-insert-prompt"
import { useActiveTab } from "../../hooks/use-active-tab"
import { hostLabel } from "../../lib/messages"
import { renderTemplate, extractVariables } from "@pv/shared/template"
import type {
  ChainStep,
  VariableMapping,
  ChainExecution,
  Prompt,
  ChainConditions,
} from "../../lib/types"

const STORAGE_KEY = "pv.activeExec"

interface PersistedExec {
  execId: number
  chainId: number
  startedAt: number
}

// Сохраняет activeExec в chrome.storage.session для recovery при reload.
function saveActiveExec(p: PersistedExec) {
  chrome.storage.session?.set({ [STORAGE_KEY]: p })
}
function clearActiveExec() {
  chrome.storage.session?.remove(STORAGE_KEY)
}

export function ChainRunPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const chainId = id ? Number(id) : null
  const chainQuery = useChain(chainId)
  const startMut = useStartExecution()
  const [execId, setExecId] = useState<number | null>(null)
  // Capture mutateAsync через ref, чтобы auto-start useEffect не пересоздавался
  // при ре-рендере хука useStartExecution (MN-63 паттерн из frontend).
  const startMutateRef = useRef(startMut.mutateAsync)
  useEffect(() => {
    startMutateRef.current = startMut.mutateAsync
  }, [startMut.mutateAsync])
  const startedRef = useRef(false)

  const execQuery = useExecution(execId)
  const advanceMut = useAdvanceStep(execId ?? 0)
  const insert = useInsertPrompt()

  // Восстанавливаем active exec при первом mount, если есть.
  useEffect(() => {
    if (!chainId) return
    chrome.storage.session?.get(STORAGE_KEY).then((data) => {
      const p = data?.[STORAGE_KEY] as PersistedExec | undefined
      if (p && p.chainId === chainId) {
        setExecId(p.execId)
        startedRef.current = true
      }
    })
  }, [chainId])

  const chain = chainQuery.data

  // Auto-start execution с пустыми initial_vars (как frontend run.tsx).
  // Manual-переменные шага запрашиваются на каждом шаге, где они нужны,
  // не на отдельном pre-run экране — это просто и не путает юзера.
  useEffect(() => {
    if (execId || startedRef.current) return
    if (!chainId || !chain) return
    if (chain.steps.length === 0) return // empty chain handled below
    startedRef.current = true
    startMutateRef
      .current({ chainId, initialVars: {} })
      .then((exec) => {
        saveActiveExec({ execId: exec.id, chainId, startedAt: Date.now() })
        setExecId(exec.id)
      })
      .catch((err: unknown) => {
        startedRef.current = false
        toast({
          title: "Не удалось запустить",
          description: err instanceof Error ? err.message : undefined,
          variant: "error",
        })
      })
  }, [chainId, chain, execId, toast])

  if (chainQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  if (!chain) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-(--color-muted-foreground)">
        Цепочка не найдена
      </div>
    )
  }

  // Empty chain — show empty state without trying to start.
  if (chain.steps.length === 0) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
          <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h2 className="flex-1 truncate text-sm font-semibold">{chain.name}</h2>
        </div>
        <div className="flex flex-1 flex-col items-center justify-center gap-3 p-6 text-center">
          <GitBranch className="h-10 w-10 text-(--color-muted-foreground)/40" />
          <div>
            <p className="text-sm font-medium">В цепочке нет шагов</p>
            <p className="mt-1 text-[10px] text-(--color-muted-foreground)">
              Добавьте хотя бы один шаг чтобы запустить.
            </p>
          </div>
          <Button
            type="button"
            variant="brand"
            size="sm"
            onClick={() => navigate(`/chains/${chain.id}/edit`)}
            className="gap-1.5"
          >
            Открыть редактор
          </Button>
        </div>
      </div>
    )
  }

  // Старт цепочки или загрузка существующего execution.
  if (execId === null || execQuery.isPending || !execQuery.data) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const exec = execQuery.data
  if (exec.status === "completed") {
    return (
      <ChainCompletionView
        exec={exec}
        chainName={chain.name}
        onClose={() => {
          clearActiveExec()
          navigate(`/chains`)
        }}
      />
    )
  }

  const currentStep = exec.chain_snapshot.steps.find((s) => s.position === exec.current_step)
  if (!currentStep) {
    return (
      <div className="flex h-full items-center justify-center p-4 text-sm text-(--color-muted-foreground)">
        Шаг не найден. Цепочка повреждена.
      </div>
    )
  }

  // Output модели в PromptVault не возвращается — юзер работает с ответом
  // прямо в чате LLM (как frontend run.tsx после Phase 16-C). step_output
  // всегда "" — backend по-прежнему пишет в JSONB пустые строки для
  // backward-compat с прошлыми executions.
  const handleAdvance = async (_unused: string, branchIndex?: number) => {
    try {
      const next = await advanceMut.mutateAsync({ stepOutput: "", chosenBranchIndex: branchIndex })
      if (next.status === "completed") {
        clearActiveExec()
      }
    } catch (err) {
      toast({
        title: "Не удалось продвинуть",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  return (
    <ChainStepView
      exec={exec}
      step={currentStep}
      chainName={chain.name}
      onAdvance={handleAdvance}
      onCancel={() => {
        clearActiveExec()
        navigate(-1)
      }}
      insertFn={async (text) =>
        insert.insert(
          { id: 0, title: chain.name, content: text } as Prompt,
          text,
          { silent: false },
        )
      }
      advancing={advanceMut.isPending}
    />
  )
}

interface ChainStepViewProps {
  exec: ChainExecution
  step: ChainStep
  chainName: string
  onAdvance: (output: string, branchIndex?: number) => Promise<void>
  onCancel: () => void
  insertFn: (text: string) => Promise<boolean>
  advancing: boolean
}

function ChainStepView({ exec, step, chainName, onAdvance, onCancel, insertFn, advancing }: ChainStepViewProps) {
  const { toast } = useToast()
  const [manualVars, setManualVars] = useState<Record<string, string>>({})
  const activeTab = useActiveTab()
  const canInsert = activeTab.supported
  const targetLabel = hostLabel(activeTab.host)

  const prompt = step.prompt
  const promptContent = prompt?.content ?? exec.chain_snapshot.prompt_contents[step.prompt_id ?? 0] ?? ""

  // Подставляем переменные.
  const resolvedContent = resolveStepContent(step, exec, manualVars, promptContent)
  const variables = extractVariables(promptContent)

  async function copyContent() {
    try {
      await navigator.clipboard.writeText(resolvedContent)
      toast({ title: "Скопировано", variant: "success", durationMs: 1500 })
    } catch {
      toast({ title: "Не удалось скопировать", variant: "error" })
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={onCancel} aria-label="Отменить">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1 min-w-0">
          <h2 className="truncate text-sm font-semibold">{chainName}</h2>
          <div className="flex items-center gap-1 text-[10px] text-(--color-muted-foreground)">
            <span>Шаг {exec.current_step}</span>
            <span>•</span>
            <span>{step.step_type === "fork" ? "выбор ветки" : step.name || prompt?.title}</span>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {step.step_type === "fork" && step.conditions ? (
          <ForkStepUI
            conditions={step.conditions}
            onSelect={(branchIndex) => onAdvance("", branchIndex)}
            advancing={advancing}
          />
        ) : (
          <>
            {/* Manual var inputs */}
            {variables.length > 0 && (
              <div className="space-y-2">
                <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  Переменные шага
                </div>
                {variables.map((v) => {
                  // Если переменная маппится через chain_var — её значение уже resolved
                  // (см. resolveStepContent). Показываем только manual или fallback.
                  const isFromChain = isMappedFromChain(step.variable_mapping, v)
                  if (isFromChain) return null
                  return (
                    <div key={v} className="space-y-1">
                      <Label className="font-mono text-xs text-(--color-brand)">
                        {`{{${v}}}`}
                      </Label>
                      <Textarea
                        value={manualVars[v] ?? ""}
                        onChange={(e) => setManualVars((p) => ({ ...p, [v]: e.target.value }))}
                        rows={2}
                        placeholder={`Значение для ${v}`}
                      />
                    </div>
                  )
                })}
              </div>
            )}

            {/* Resolved prompt preview */}
            <div className="space-y-1">
              <div className="flex items-center justify-between">
                <Label>Промпт для этого шага</Label>
                <span className="text-[10px] text-(--color-muted-foreground)">
                  {resolvedContent.length.toLocaleString("ru-RU")} симв
                </span>
              </div>
              <div className="whitespace-pre-wrap rounded-md border border-(--color-border) bg-(--color-muted)/30 p-3 text-xs max-h-48 overflow-y-auto">
                {resolvedContent}
              </div>
              <p className="text-[10px] text-(--color-muted-foreground)">
                Скопируйте промпт, выполните его в Claude/ChatGPT и вернитесь сюда.
              </p>
            </div>
          </>
        )}
      </div>

      {step.step_type !== "fork" && (
        <div className="flex flex-col gap-1.5 border-t border-(--color-border) p-2">
          {!canInsert && (
            <p className="text-center text-[10px] text-(--color-muted-foreground)">
              Откройте ChatGPT / Claude / Gemini и т.д. чтобы вставить
            </p>
          )}
          <div className="flex items-center gap-2">
            {/* «Вставить» — primary action этого экрана (отправить промпт в чат AI).
                «Копировать» — fallback secondary. «Далее» — переход к следующему
                шагу wizard'а, тоже primary но в другой плоскости (workflow vs use). */}
            <Button
              type="button"
              variant="brand"
              size="sm"
              onClick={() => insertFn(resolvedContent)}
              className="gap-1.5"
              disabled={advancing || !canInsert}
              title={
                canInsert
                  ? targetLabel
                    ? `Вставить в ${targetLabel}`
                    : "Вставить"
                  : "Откройте поддерживаемый AI-сайт"
              }
            >
              <Send className="h-3.5 w-3.5" />
              Вставить
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={copyContent}
              className="gap-1.5"
              disabled={advancing}
            >
              <Copy className="h-3.5 w-3.5" />
              Копировать
            </Button>
            <div className="flex-1" />
            <Button
              type="button"
              variant="brand"
              size="sm"
              onClick={() => onAdvance("")}
              disabled={advancing}
              className="gap-1.5"
            >
              {advancing ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <ArrowRight className="h-3.5 w-3.5" />
              )}
              Далее
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function ForkStepUI({
  conditions,
  onSelect,
  advancing,
}: {
  conditions: ChainConditions
  onSelect: (idx: number) => void
  advancing: boolean
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <GitBranch className="h-4 w-4 text-(--color-brand)" />
        <div className="text-xs font-medium">Выберите ветку</div>
      </div>
      <p className="text-[10px] text-(--color-muted-foreground)">
        Какой путь продолжить?
      </p>
      <div className="space-y-1.5">
        {conditions.branches.map((b, i) => (
          <Button
            key={i}
            type="button"
            variant="outline"
            onClick={() => onSelect(i)}
            disabled={advancing}
            className="w-full justify-start gap-2 text-left"
          >
            <span className="rounded bg-(--color-brand-muted) px-1.5 py-0.5 font-mono text-[10px] text-(--color-brand)">
              {i + 1}
            </span>
            <span>{b.label}</span>
          </Button>
        ))}
      </div>
    </div>
  )
}

function ChainCompletionView({
  exec,
  chainName,
  onClose,
}: {
  exec: ChainExecution
  chainName: string
  onClose: () => void
}) {
  const navigate = useNavigate()
  const stepsCount = Object.keys(exec.step_outputs).length
  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 p-6 text-center">
      <div className="rounded-full bg-(--color-success)/10 p-3">
        <CheckCircle2 className="h-10 w-10 text-(--color-success)" />
      </div>
      <div className="max-w-[260px]">
        <h3 className="text-sm font-semibold">Цепочка завершена</h3>
        <p className="mt-1 text-xs text-(--color-muted-foreground)">
          <span className="font-medium text-(--color-foreground)">{chainName}</span>
          {": пройдено "}
          {pluralAfterDo(stepsCount, "шаг", "шага", "шагов")}
        </p>
      </div>
      <div className="flex w-full max-w-[260px] flex-col gap-2 pt-2">
        <Button
          type="button"
          variant="brand"
          size="sm"
          onClick={() => navigate(`/chains/${exec.chain_id}/run`, { replace: true })}
          className="w-full"
        >
          Запустить ещё раз
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => navigate(`/chains/${exec.chain_id}/runs`, { replace: true })}
          className="w-full"
        >
          История запусков
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onClose}
          className="w-full"
        >
          К списку цепочек
        </Button>
      </div>
    </div>
  )
}

// --- Helpers ---

function resolveStepContent(
  step: ChainStep,
  exec: ChainExecution,
  manualVars: Record<string, string>,
  promptContent: string,
): string {
  // chain_var → только exec.variables (initial-level переменные цепочки).
  // step_outputs как источник убран: backend пишет ключи step_<step.id> (uint),
  // а variable_mapping хранит var_name (string) — несовпадение схем ломало
  // подстановку. См. frontend pages/chains/run.tsx::resolveStepVariables.
  // Output модели не возвращается в PromptVault — юзер работает с ответом
  // прямо в чате LLM.
  const values: Record<string, string> = {}
  for (const [name, src] of Object.entries(step.variable_mapping)) {
    if (src.type === "chain_var" && src.var_name) {
      values[name] = exec.variables[src.var_name] ?? ""
    } else if (src.type === "manual" && src.var_name) {
      values[name] = exec.variables[src.var_name] ?? manualVars[name] ?? ""
    }
  }
  for (const [k, v] of Object.entries(manualVars)) {
    if (!(k in values)) values[k] = v
  }
  return renderTemplate(promptContent, values)
}

function isMappedFromChain(mapping: VariableMapping, varName: string): boolean {
  return mapping[varName]?.type === "chain_var"
}
