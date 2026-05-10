// Run-mode цепочки: пошаговый wizard. Юзер копирует отрендеренный промпт,
// выполняет его в Claude/ChatGPT и нажимает «Далее». На fork-шагах вместо «Далее»
// отображаются кнопки выбора ветки. После последнего шага показывается итоговая
// карточка «Цепочка завершена».
//
// Output модели обратно в PromptVault не возвращается — юзер работает с ответом
// прямо в чате LLM. Поле «результат шага» намеренно убрано как лишний шаг.
// Бэк по-прежнему пишет step_outputs JSONB (пустыми строками) — миграция drop
// column не делается ради backward compat с прошлыми executions.

import { useEffect, useMemo, useRef, useState } from "react"
import { Link, useParams, useSearchParams } from "react-router-dom"
import { ArrowLeft, ArrowRight, CheckCircle2, ClipboardCopy, GitBranch, Info } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { useAdvanceStep, useExecution, useStartExecution } from "@/hooks/use-chains"
import { useAuthStore } from "@/stores/auth-store"
import { extractVariables, renderTemplate } from "@/lib/template/parse"
import type { ChainExecution, ChainStep, VariableMapping } from "@/api/types"

export default function ChainRunPage() {
  const { id } = useParams<{ id: string }>()
  const [params] = useSearchParams()
  const chainID = id ? Number(id) : 0
  const resumeID = params.get("resume")

  const [execID, setExecID] = useState<number>(resumeID ? Number(resumeID) : 0)

  const start = useStartExecution(chainID)
  const startedRef = useRef(false)
  // MN-63: capture mutateAsync через ref, чтобы useEffect не пересоздавался
  // при каждом ре-рендере хука useStartExecution. Раньше eslint-disable
  // подавлял react-hooks/exhaustive-deps без объяснения, как избежать
  // infinite re-run — теперь deps complete и линтер доволен.
  const startMutateRef = useRef(start.mutateAsync)
  useEffect(() => {
    startMutateRef.current = start.mutateAsync
  }, [start.mutateAsync])

  useEffect(() => {
    if (execID || startedRef.current || !chainID) return
    startedRef.current = true
    startMutateRef.current({}).then((exec) => setExecID(exec.id)).catch(() => {
      startedRef.current = false
    })
  }, [chainID, execID])

  const { data: exec, isLoading } = useExecution(execID)
  const advance = useAdvanceStep(execID)

  const isCompleted = exec?.status === "completed"
  const snapshot = exec?.chain_snapshot
  const currentStep = snapshot?.steps.find((s) => s.position === exec?.current_step)
  const isFork = currentStep?.step_type === "fork"
  // Порядковый номер prompt-шага в пройденном пути. Forks номер не получают —
  // это маршрутизация, а не работа юзера. «Из N» не показываем: N зависит от
  // выбранной ветки, заранее предсказать честно нельзя.
  const promptStepNumber = countPromptStepsToHere(exec, currentStep)
  // fork-шаги не имеют своего промпта — это контейнер с ветками. Пропускаем
  // resolve/render промпта, и в UI показываем только кнопки выбора ветки.
  const currentPromptContent =
    currentStep && !isFork && currentStep.prompt_id
      ? snapshot?.prompt_contents[currentStep.prompt_id] ?? ""
      : ""

  // Manual var inputs: имена и значения. Сбрасываются при переходе на следующий
  // шаг через useEffect ниже.
  const [manualVars, setManualVars] = useState<Record<string, string>>({})

  const { resolvedValues, manualVarNames } = useMemo(
    () => resolveStepVariables(currentStep, exec, currentPromptContent, manualVars),
    [currentStep, exec, currentPromptContent, manualVars],
  )
  const renderedPrompt = useMemo(
    () => renderTemplate(currentPromptContent, resolvedValues),
    [currentPromptContent, resolvedValues],
  )

  const planId = useAuthStore((s) => s.user?.plan_id ?? "free")
  const isFreeTier = planId === "free"

  const onAdvance = async () => {
    await advance.mutateAsync({ stepOutput: "" })
    setManualVars({})
  }

  const onChooseBranch = async (branchIndex: number) => {
    await advance.mutateAsync({ stepOutput: "", chosenBranchIndex: branchIndex })
    setManualVars({})
  }

  return (
    <div className="container mx-auto max-w-3xl p-6">
      <div className="mb-6 flex items-center gap-3">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/chains">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <h1 className="text-2xl font-semibold">{snapshot?.chain.name ?? "Запуск цепочки"}</h1>
      </div>

      {isFreeTier && exec && !isCompleted && (
        <div className="mb-4 flex gap-3 rounded-md border border-blue-500/30 bg-blue-500/5 p-3 text-sm">
          <Info className="mt-0.5 h-4 w-4 shrink-0 text-blue-600 dark:text-blue-400" />
          <div className="space-y-1">
            <p className="text-foreground">
              На Free сохраняются 3 последних запуска цепочки.
            </p>
            <p className="text-muted-foreground">
              На Pro — история из 50 запусков.{" "}
              <Link to="/pricing" className="font-medium text-blue-600 hover:underline dark:text-blue-400">
                Сравнить тарифы
              </Link>
            </p>
          </div>
        </div>
      )}

      {(isLoading || !exec) && <Skeleton className="h-64" />}

      {exec && !isCompleted && currentStep && (
        <Card className={isFork ? "border-amber-500/40" : undefined}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              {isFork && <GitBranch className="h-4 w-4 text-amber-600" />}
              {isFork
                ? `Развилка${currentStep.name ? `: ${currentStep.name}` : ""}`
                : `Шаг ${promptStepNumber}${currentStep.name ? `: ${currentStep.name}` : ""}`}
              {isFork && (
                <span className="rounded bg-amber-500/10 px-1.5 py-0.5 text-xs text-amber-700 dark:text-amber-400">
                  развилка
                </span>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isFork ? (
              <>
                <p className="text-sm text-muted-foreground">
                  Выберите ветку — в неё перейдёт выполнение цепочки.
                </p>
                <div className="space-y-2">
                  {(currentStep.conditions?.branches ?? []).map((branch, idx) => (
                    <Button
                      key={idx}
                      variant="outline"
                      className="w-full justify-start"
                      onClick={() => onChooseBranch(idx)}
                      disabled={advance.isPending}
                    >
                      <GitBranch className="mr-2 h-4 w-4 text-amber-600" />
                      {branch.label}
                    </Button>
                  ))}
                </div>
              </>
            ) : (
              <>
                {manualVarNames.length > 0 && (
                  <div className="space-y-2 rounded-md border bg-muted/30 p-3">
                    <p className="text-sm font-medium">Заполните переменные шага</p>
                    {manualVarNames.map((varName) => (
                      <div key={varName} className="space-y-1">
                        <Label htmlFor={`var-${varName}`} className="text-xs">{`{{${varName}}}`}</Label>
                        <Input
                          id={`var-${varName}`}
                          value={manualVars[varName] ?? ""}
                          onChange={(e) =>
                            setManualVars((prev) => ({ ...prev, [varName]: e.target.value }))
                          }
                        />
                      </div>
                    ))}
                  </div>
                )}

                <div>
                  <div className="mb-2 flex items-center justify-between">
                    <p className="text-sm font-medium">Промпт для этого шага</p>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => navigator.clipboard?.writeText(renderedPrompt)}
                    >
                      <ClipboardCopy className="mr-2 h-4 w-4" />
                      Скопировать
                    </Button>
                  </div>
                  <pre className="max-h-64 overflow-auto whitespace-pre-wrap rounded-md border bg-muted p-3 text-sm">
                    {renderedPrompt}
                  </pre>
                  <p className="mt-2 text-xs text-muted-foreground">
                    Скопируйте промпт, выполните его в Claude/ChatGPT и вернитесь сюда.
                  </p>
                </div>

                <div className="flex justify-end">
                  <Button onClick={onAdvance} disabled={advance.isPending}>
                    {currentStep.next_step_id ? "Далее" : "Завершить"}
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      )}

      {exec && isCompleted && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <CheckCircle2 className="h-5 w-5 text-green-600" />
              Цепочка завершена
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Все шаги выполнены.</p>
            <div className="mt-4 flex justify-end">
              <Button asChild>
                <Link to="/chains">К списку</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

interface ResolvedStepVariables {
  /** Значения для renderTemplate: имя переменной → подставленное значение. */
  resolvedValues: Record<string, string>
  /** Имена переменных, которые юзер должен ввести вручную (manual или
   *  отсутствующие в variable_mapping → fallback на manual). */
  manualVarNames: string[]
}

/**
 * Считает порядковый номер prompt-шага в текущем пройденном пути. Fork-шаги
 * не учитываются — это маршрутизация, не «работа». Для текущего prompt-шага
 * прибавляем +1 (он ещё не записан в step_outputs до AdvanceStep).
 */
function countPromptStepsToHere(
  exec: ChainExecution | undefined,
  currentStep: ChainStep | undefined,
): number {
  if (!exec || !currentStep) return 0
  const snap = exec.chain_snapshot
  if (!snap) return 0
  const byID = new Map(snap.steps.map((s) => [s.id, s]))
  let n = 0
  for (const key of Object.keys(exec.step_outputs ?? {})) {
    const idStr = key.startsWith("step_") ? key.slice(5) : key
    const id = Number(idStr)
    if (!Number.isFinite(id)) continue
    const step = byID.get(id)
    if (step?.step_type === "prompt") n++
  }
  if (currentStep.step_type === "prompt") n++
  return n
}

/**
 * Resolve переменные текущего шага. Источники:
 *   - variable_mapping шага type=chain_var → exec.variables[var_name]
 *   - manual values из локального state run-mode (включая variables без mapping)
 *
 * step_output как источник убран вместе с UI поля «результат шага».
 */
function resolveStepVariables(
  step: ChainStep | undefined,
  exec: ChainExecution | undefined,
  promptContent: string,
  manualValues: Record<string, string>,
): ResolvedStepVariables {
  if (!step || !exec) {
    return { resolvedValues: {}, manualVarNames: [] }
  }
  const allVarsInTemplate = extractVariables(promptContent)
  const mapping: VariableMapping = step.variable_mapping ?? {}
  const resolved: Record<string, string> = {}
  const manualNames: string[] = []

  for (const varName of allVarsInTemplate) {
    const source = mapping[varName]
    if (!source || source.type === "manual") {
      manualNames.push(varName)
      resolved[varName] = manualValues[varName] ?? ""
      continue
    }
    if (source.type === "chain_var" && source.var_name) {
      resolved[varName] = exec.variables[source.var_name] ?? ""
      continue
    }
    // Unknown source type — fallback на пустую строку.
    resolved[varName] = ""
  }
  return { resolvedValues: resolved, manualVarNames: manualNames }
}
