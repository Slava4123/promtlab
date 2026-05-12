import { useState, useEffect } from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Loader2,
  Send,
  Copy,
  Play,
  GitBranch,
} from "lucide-react"
import { Button } from "../../components/ui/button"
import { Label } from "../../components/ui/label"
import { Textarea } from "../../components/ui/textarea"
import { useToast } from "../../components/ui/toaster"
import { useChain, useStartExecution, useExecution, useAdvanceStep } from "../../hooks/use-chains"
import { useInsertPrompt } from "../../hooks/use-insert-prompt"
import { renderTemplate, extractVariables } from "../../lib/template"
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
  const [initialVars, setInitialVars] = useState<Record<string, string>>({})

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
      }
    })
  }, [chainId])

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

  // Pre-flight: вытащим chain-level переменные из всех шагов с type=manual.
  const initialVarNames = collectChainVars(chain.steps)

  async function handleStart() {
    if (!chainId) return
    try {
      const exec = await startMut.mutateAsync({ chainId, initialVars })
      saveActiveExec({ execId: exec.id, chainId, startedAt: Date.now() })
      setExecId(exec.id)
    } catch (err) {
      toast({
        title: "Не удалось запустить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  // ----- Pre-run: ввод initial vars -----
  if (execId === null) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
          <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h2 className="flex-1 truncate text-sm font-semibold">{chain.name}</h2>
        </div>
        <div className="flex-1 overflow-y-auto p-3 space-y-3">
          <p className="text-xs text-(--color-muted-foreground)">
            Цепочка из {chain.steps.length} шагов. Заполните начальные переменные и запустите.
          </p>
          {initialVarNames.length > 0 ? (
            <div className="space-y-2.5">
              <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                Начальные переменные
              </div>
              {initialVarNames.map((name) => (
                <div key={name} className="space-y-1">
                  <Label
                    htmlFor={`var-${name}`}
                    className="font-mono text-xs text-(--color-primary)"
                  >
                    {`{{${name}}}`}
                  </Label>
                  <Textarea
                    id={`var-${name}`}
                    value={initialVars[name] ?? ""}
                    onChange={(e) =>
                      setInitialVars((prev) => ({ ...prev, [name]: e.target.value }))
                    }
                    rows={2}
                    placeholder={`Значение для ${name}`}
                  />
                </div>
              ))}
            </div>
          ) : (
            <p className="text-[10px] text-(--color-muted-foreground)">
              В цепочке нет начальных переменных — можно запустить сразу.
            </p>
          )}
        </div>
        <div className="border-t border-(--color-border) p-2">
          <Button
            type="button"
            onClick={handleStart}
            disabled={startMut.isPending}
            className="w-full gap-1.5"
          >
            <Play className="h-3.5 w-3.5" />
            {startMut.isPending ? "Запускаю…" : "Запустить цепочку"}
          </Button>
        </div>
      </div>
    )
  }

  // ----- In-run state -----
  if (execQuery.isPending || !execQuery.data) {
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

  const handleAdvance = async (output: string, branchIndex?: number) => {
    try {
      const next = await advanceMut.mutateAsync({ stepOutput: output, chosenBranchIndex: branchIndex })
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

// Собирает имена переменных, не предоставляемых output'ами шагов
// (т.е. type=manual). Для шага type=fork — пропускаем.
function collectChainVars(steps: ChainStep[]): string[] {
  const seen = new Set<string>()
  for (const step of steps) {
    if (step.step_type !== "prompt") continue
    const mapping = step.variable_mapping
    for (const [, src] of Object.entries(mapping)) {
      if (src.type === "manual" && src.var_name) {
        seen.add(src.var_name)
      }
    }
  }
  return Array.from(seen)
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
  const [output, setOutput] = useState("")
  const [manualVars, setManualVars] = useState<Record<string, string>>({})

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

  async function captureAndCopy() {
    // Try to capture AI response from current tab. Placeholder for Phase 3 polish.
    try {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true })
      if (!tab?.id) return
      const resp = await chrome.tabs
        .sendMessage(tab.id, { type: "content.captureLastAIResponse" })
        .catch(() => null)
      if (resp && resp.type === "content.captured" && resp.text) {
        setOutput(resp.text)
        toast({ title: "AI-ответ захвачен", variant: "success", durationMs: 1500 })
      } else {
        toast({ title: "Не удалось захватить ответ", variant: "info" })
      }
    } catch {
      // ignore
    }
  }

  // Подсветка переменных
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
                      <Label className="font-mono text-xs text-(--color-primary)">
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
                <Label>Промпт</Label>
                <span className="text-[10px] text-(--color-muted-foreground)">
                  {resolvedContent.length.toLocaleString("ru-RU")} симв
                </span>
              </div>
              <div className="whitespace-pre-wrap rounded-md border border-(--color-border) bg-(--color-muted)/30 p-3 text-xs max-h-48 overflow-y-auto">
                {resolvedContent}
              </div>
            </div>

            {/* Output field */}
            <div className="space-y-1">
              <div className="flex items-center justify-between">
                <Label htmlFor="step-output">Ответ AI (для следующего шага)</Label>
                <button
                  type="button"
                  onClick={captureAndCopy}
                  className="text-[10px] text-(--color-primary) hover:underline"
                >
                  Захватить из вкладки
                </button>
              </div>
              <Textarea
                id="step-output"
                value={output}
                onChange={(e) => setOutput(e.target.value)}
                rows={4}
                placeholder="Вставьте ответ AI сюда (или оставьте пустым)"
              />
            </div>
          </>
        )}
      </div>

      {step.step_type !== "fork" && (
        <div className="flex items-center gap-2 border-t border-(--color-border) p-2">
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={() => insertFn(resolvedContent)}
            className="gap-1.5"
            disabled={advancing}
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
            size="sm"
            onClick={() => onAdvance(output)}
            disabled={advancing}
            className="gap-1.5"
          >
            {advancing ? <Loader2 className="h-3 w-3 animate-spin" /> : <ArrowRight className="h-3.5 w-3.5" />}
            Далее
          </Button>
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
        <GitBranch className="h-4 w-4 text-purple-500" />
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
            <span className="rounded bg-purple-500/15 px-1.5 py-0.5 font-mono text-[10px]">
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
  const stepsCount = Object.keys(exec.step_outputs).length
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
      <CheckCircle2 className="h-12 w-12 text-emerald-500" />
      <div>
        <h3 className="text-sm font-semibold">Цепочка завершена</h3>
        <p className="mt-1 text-xs text-(--color-muted-foreground)">
          {chainName}: пройдено {stepsCount} шагов
        </p>
      </div>
      <Button type="button" onClick={onClose} size="sm">
        К списку цепочек
      </Button>
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
