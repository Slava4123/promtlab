// Phase 16 v3: Inline-tree editor.
//
// Цепочка строится как дерево из явных рёбер:
//   - prompt-шаг → next_step_id (один путь)
//   - fork-шаг   → conditions.branches[].next_step_id (несколько веток)
//
// Дерево обходится из «корневого» шага (на который никто не ссылается) рекурсивно.
// Кнопка «+ Шаг» в листе ветки добавляет шаг через after_step_id (для prompt) или
// parent_fork_id + branch_index (для пустой ветки fork-шага). «+ Развилка»
// доступна только тарифу Max и проверяется бэком (ErrForkRequiresMax).

import { useState } from "react"
import { useNavigate, useParams, Link } from "react-router-dom"
import { toast } from "sonner"
import {
  ArrowDown,
  ArrowLeft,
  ArrowUp,
  GitBranch,
  History,
  Layout,
  Lock,
  Pencil,
  PlayCircle,
  Plus,
  Save,
  Trash2,
} from "lucide-react"

import { ApiError } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  useAddStep,
  useChain,
  useCreateChain,
  useMoveStepDown,
  useMoveStepUp,
  useRemoveStep,
  useUpdateChain,
  useUpdateStep,
} from "@/hooks/use-chains"
import type { ChainStep } from "@/api/types"
import { PromptPicker } from "@/components/chains/prompt-picker"
import { useAuthStore } from "@/stores/auth-store"
import { useCurrentTeamRole } from "@/hooks/use-team-role"
import { useCurrentTeam } from "@/hooks/use-current-team"

const MAX_TIERS = new Set(["max", "max_yearly"])

// Унифицированный обработчик ошибок мутаций цепочки.
// При 402 (quota exceeded) ничего не показываем — `api/client.ts` уже открыл
// глобальный quota-exceeded-dialog с CTA «Получить Pro», как в /prompts /collections.
// Для остальных ошибок — toast.error (как в use-settings/use-api-keys паттерне).
function reportMutationError(err: unknown, prefix: string) {
  if (err instanceof ApiError && err.status === 402) return
  const message = err instanceof Error ? err.message : String(err)
  toast.error(`${prefix}: ${message}`)
}

export default function ChainEditorPage() {
  const { id } = useParams<{ id: string }>()
  const isNew = !id
  const chainID = id ? Number(id) : 0
  const { data: chain } = useChain(chainID)

  if (isNew) {
    return <ChainEditorForm key="new" mode="create" />
  }
  if (!chain) {
    return (
      <div className="container mx-auto max-w-4xl p-6">
        <div className="text-sm text-muted-foreground">Загрузка цепочки…</div>
      </div>
    )
  }
  return (
    <ChainEditorForm
      key={chain.id}
      mode="edit"
      initialName={chain.name}
      initialDescription={chain.description}
      steps={chain.steps}
      chainID={chainID}
    />
  )
}

interface ChainEditorFormProps {
  mode: "create" | "edit"
  initialName?: string
  initialDescription?: string
  steps?: ChainStep[]
  chainID?: number
}

function ChainEditorForm({
  mode,
  initialName = "",
  initialDescription = "",
  steps = [],
  chainID = 0,
}: ChainEditorFormProps) {
  const navigate = useNavigate()
  const create = useCreateChain()
  const update = useUpdateChain()
  const addStep = useAddStep(chainID)
  const removeStep = useRemoveStep(chainID)
  const moveUp = useMoveStepUp(chainID)
  const moveDown = useMoveStepDown(chainID)
  const updateStep = useUpdateStep(chainID)

  const planId = useAuthStore((s) => s.user?.plan_id ?? "free")
  const team = useCurrentTeam()
  const teamId = team?.teamId ?? null
  const { canWrite, isPersonal } = useCurrentTeamRole()
  // Fork-gate UI: для personal — по своему плану; для team — оптимистично
  // показываем enabled (бэк проверит plan team-owner и вернёт ErrForkRequiresMax
  // если ни один owner команды не на Max). Так Pro-editor в Max-команде видит
  // кнопку enabled и фактически может ей пользоваться.
  const isMax = isPersonal ? MAX_TIERS.has(planId) : true

  const [name, setName] = useState(initialName)
  const [description, setDescription] = useState(initialDescription)

  // Состояние диалогов «+ Шаг» / «+ Развилка». Открываются с привязкой к месту
  // вставки (after_step_id или parent_fork_id+branch_index). null = закрыт.
  const [addStepCtx, setAddStepCtx] = useState<InsertContext | null>(null)
  const [addForkCtx, setAddForkCtx] = useState<InsertContext | null>(null)
  // Редактирование существующего fork: храним сам шаг, чтобы prefill диалог.
  const [editForkStep, setEditForkStep] = useState<ChainStep | null>(null)
  // Подтверждение удаления шага/развилки (заменяет native confirm() для UX consistency).
  const [pendingRemoveStep, setPendingRemoveStep] = useState<ChainStep | null>(null)

  const onSubmitNew = async () => {
    if (!name.trim()) return
    const created = await create.mutateAsync({
      name: name.trim(),
      description: description.trim(),
      team_id: teamId,
    })
    navigate(`/chains/${created.id}/edit`, { replace: true })
  }

  const onSubmitUpdate = async () => {
    if (!chainID) return
    await update.mutateAsync({ id: chainID, name: name.trim(), description: description.trim() })
  }

  const handleAddStep = async (input: {
    promptId: number
    name?: string
    location: InsertContext
  }) => {
    try {
      await addStep.mutateAsync({
        prompt_id: input.promptId,
        name: input.name,
        after_step_id: input.location.after_step_id,
        parent_fork_id: input.location.parent_fork_id,
        branch_index: input.location.branch_index,
      })
      setAddStepCtx(null)
    } catch (err) {
      reportMutationError(err, "Не удалось добавить шаг")
    }
  }

  const handleAddFork = async (input: {
    name?: string
    branchLabels: string[]
    location: InsertContext
  }) => {
    try {
      await addStep.mutateAsync({
        // fork — контейнер без своего промпта; prompt_id не передаётся.
        name: input.name,
        step_type: "fork",
        conditions: {
          branches: input.branchLabels.map((label) => ({ label, next_step_id: null })),
        },
        after_step_id: input.location.after_step_id,
        parent_fork_id: input.location.parent_fork_id,
        branch_index: input.location.branch_index,
      })
      setAddForkCtx(null)
    } catch (err) {
      reportMutationError(err, "Не удалось создать развилку")
    }
  }

  const handleRemove = (step: ChainStep) => {
    setPendingRemoveStep(step)
  }

  const handleMove = async (stepID: number, dir: "up" | "down") => {
    try {
      await (dir === "up" ? moveUp.mutateAsync(stepID) : moveDown.mutateAsync(stepID))
    } catch (err) {
      reportMutationError(err, "Не удалось переместить шаг")
    }
  }

  const handleEditFork = async (input: { name?: string; branches: { label: string; next_step_id?: number | null }[] }) => {
    if (!editForkStep) return
    try {
      await updateStep.mutateAsync({
        stepId: editForkStep.id,
        name: input.name ?? "",
        step_type: "fork",
        conditions: { branches: input.branches },
      })
      setEditForkStep(null)
    } catch (err) {
      reportMutationError(err, "Не удалось обновить развилку")
    }
  }

  const tree = buildTree(steps)

  return (
    <div className="container mx-auto max-w-4xl p-6">
      <div className="mb-6 flex flex-wrap items-center gap-3">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/chains">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <h1 className="text-2xl font-semibold">
          {mode === "create" ? "Новая цепочка" : "Редактировать цепочку"}
        </h1>
        {mode === "edit" && (
          <div className="ml-auto flex items-center gap-2">
            <Button variant="outline" size="sm" asChild>
              <Link to={`/chains/${chainID}/canvas`}>
                <Layout className="mr-2 h-4 w-4" />
                Граф
              </Link>
            </Button>
            <Button variant="outline" size="sm" asChild>
              <Link to={`/chains/${chainID}/runs`}>
                <History className="mr-2 h-4 w-4" />
                История
              </Link>
            </Button>
            <Button variant="outline" size="sm" asChild>
              <Link to={`/chains/${chainID}/run`}>
                <PlayCircle className="mr-2 h-4 w-4" />
                Запустить
              </Link>
            </Button>
          </div>
        )}
      </div>

      <Card>
        <CardContent className="space-y-4 pt-6">
          <div className="space-y-2">
            <Label htmlFor="chain-name">Название</Label>
            <Input
              id="chain-name"
              placeholder="Например: PRD Generator"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="chain-description">Описание</Label>
            <Textarea
              id="chain-description"
              placeholder="Краткое описание цепочки и того, что она делает"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              maxLength={2000}
            />
          </div>
          <div className="flex justify-end">
            {mode === "create" ? (
              <Button onClick={onSubmitNew} disabled={!name.trim() || create.isPending || !canWrite}>
                <Save className="mr-2 h-4 w-4" />
                Создать
              </Button>
            ) : (
              <Button
                onClick={onSubmitUpdate}
                disabled={update.isPending || !canWrite}
                title={!canWrite ? "Изменять цепочку могут только Владелец и Редактор команды" : undefined}
              >
                <Save className="mr-2 h-4 w-4" />
                Сохранить мета-данные
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {mode === "edit" && (
        <>
          <h2 className="mb-3 mt-8 text-lg font-semibold">Шаги</h2>

          {tree ? (
            <StepNode
              node={tree}
              parent={null}
              isMax={isMax}
              canWrite={canWrite}
              onRemove={handleRemove}
              onAddStepHere={(loc) => setAddStepCtx(loc)}
              onAddForkHere={(loc) => setAddForkCtx(loc)}
              onMove={handleMove}
              onEditFork={(step) => setEditForkStep(step)}
            />
          ) : (
            <Card>
              <CardContent className="py-6 text-center text-sm text-muted-foreground">
                Цепочка пуста. Добавьте первый шаг ниже.
              </CardContent>
            </Card>
          )}

          {/* Tail-add кнопки (если цепочка пуста ИЛИ корень — root prompt-шаг
              и его хвост уже завершён — кнопки покажутся в самом конце ветки
              рекурсивно, см. StepNode). Если tree===null показываем здесь. */}
          {!tree && canWrite && (
            <div className="mt-4">
              {/* Пустая цепочка: только «+ Шаг». Развилка в пустоте бессмысленна
                  — нечего ветвить. Появится после первого добавленного шага. */}
              <AddRow
                isMax={isMax}
                showFork={false}
                onAddStep={() => setAddStepCtx({})}
                onAddFork={() => setAddForkCtx({})}
              />
            </div>
          )}
          {!tree && !canWrite && (
            <Card>
              <CardContent className="py-6 text-center text-sm text-muted-foreground">
                У вас роль читателя в этой команде. Запросите права редактора у владельца, чтобы добавлять шаги.
              </CardContent>
            </Card>
          )}

          <p className="mt-6 text-xs text-muted-foreground">
            Развилки (fork) — доступны на тарифе Max. Можно добавить любое
            количество веток на любой глубине, в каждой — свои шаги.
            {!isMax && isPersonal && (
              <>
                {" "}
                <Link to="/pricing" className="font-medium text-amber-600 underline hover:text-amber-700 dark:text-amber-500">
                  Перейти на Max →
                </Link>
              </>
            )}
            {" "}
            <Link to={`/chains/${chainID}/canvas`} className="underline">
              Посмотреть как граф
            </Link>
            .
          </p>
        </>
      )}

      {addStepCtx && (
        <AddStepDialog
          open
          onClose={() => setAddStepCtx(null)}
          onSubmit={(input) => handleAddStep({ ...input, location: addStepCtx })}
        />
      )}
      {addForkCtx && (
        <AddForkDialog
          open
          onClose={() => setAddForkCtx(null)}
          onSubmit={(input) => handleAddFork({ ...input, location: addForkCtx })}
        />
      )}
      {editForkStep && (
        <ForkDialog
          mode="edit"
          open
          initialName={editForkStep.name}
          initialBranches={editForkStep.conditions?.branches ?? []}
          onClose={() => setEditForkStep(null)}
          onSubmit={(input) => handleEditFork(input)}
        />
      )}
      <ConfirmDialog
        open={pendingRemoveStep !== null}
        onOpenChange={(v) => !v && setPendingRemoveStep(null)}
        title={pendingRemoveStep?.step_type === "fork" ? "Удалить развилку?" : "Удалить шаг?"}
        description={
          pendingRemoveStep?.step_type === "fork"
            ? "Все шаги внутри веток развилки тоже удалятся (через FK CASCADE). Это действие нельзя отменить."
            : "Все ссылки в цепочке пересвяжутся автоматически. Это действие нельзя отменить."
        }
        confirmLabel="Удалить"
        isPending={removeStep.isPending}
        onConfirm={() => {
          if (!pendingRemoveStep) return
          removeStep.mutate(pendingRemoveStep.id, { onSettled: () => setPendingRemoveStep(null) })
        }}
      />
    </div>
  )
}

// --- Tree node ---

interface InsertContext {
  after_step_id?: number
  parent_fork_id?: number
  branch_index?: number
}

interface TreeNode {
  step: ChainStep
  next?: TreeNode | null
  branches?: { label: string; index: number; head: TreeNode | null }[]
}

function buildTree(steps: ChainStep[]): TreeNode | null {
  if (steps.length === 0) return null
  const byId = new Map<number, ChainStep>()
  steps.forEach((s) => byId.set(s.id, s))

  const incoming = new Set<number>()
  steps.forEach((s) => {
    if (s.next_step_id) incoming.add(s.next_step_id)
    if (s.step_type === "fork" && s.conditions) {
      s.conditions.branches.forEach((b) => {
        if (b.next_step_id) incoming.add(b.next_step_id)
      })
    }
  })
  const roots = steps.filter((s) => !incoming.has(s.id))
  if (roots.length === 0) return null
  roots.sort((a, b) => a.position - b.position)
  const root = roots[0]

  const visited = new Set<number>()
  const build = (step: ChainStep): TreeNode => {
    if (visited.has(step.id)) return { step }
    visited.add(step.id)
    if (step.step_type === "fork") {
      const branches = (step.conditions?.branches ?? []).map((b, i) => ({
        label: b.label,
        index: i,
        head: b.next_step_id && byId.has(b.next_step_id) ? build(byId.get(b.next_step_id)!) : null,
      }))
      return { step, branches }
    }
    const next =
      step.next_step_id && byId.has(step.next_step_id) ? build(byId.get(step.next_step_id)!) : null
    return { step, next }
  }
  return build(root)
}

interface StepNodeProps {
  node: TreeNode
  /** Узел-родитель в линейной части (prompt-предшественник). Используется для
   *  определения, можно ли двигать prompt-шаг «вверх». null — текущий узел корень
   *  своей подцепочки (т.е. либо вершина дерева, либо первый в ветке fork). */
  parent: TreeNode | null
  isMax: boolean
  /** RBAC: viewer не имеет права на add/remove/move/edit. При false скрываем
   *  все mutate-кнопки и AddRow'ы — узел становится read-only картинкой. */
  canWrite: boolean
  onRemove: (step: ChainStep) => void
  onAddStepHere: (location: InsertContext) => void
  onAddForkHere: (location: InsertContext) => void
  onMove: (stepID: number, dir: "up" | "down") => void
  onEditFork: (step: ChainStep) => void
}

function StepNode({ node, parent, isMax, canWrite, onRemove, onAddStepHere, onAddForkHere, onMove, onEditFork }: StepNodeProps) {
  const step = node.step

  if (step.step_type === "fork") {
    return (
      <div className="space-y-3">
        <Card className="border-amber-500/40 bg-amber-50/30 dark:bg-amber-950/10">
          <CardContent className="flex items-center gap-3 p-4">
            <GitBranch className="h-5 w-5 text-amber-600" />
            <div className="flex-1">
              <p className="text-sm font-medium">
                Развилка{step.name ? `: ${step.name}` : ""}
              </p>
              <p className="text-xs text-muted-foreground">
                {node.branches?.length ?? 0} {pluralBranches(node.branches?.length ?? 0)}
              </p>
            </div>
            {canWrite && (
              <>
                <Button variant="ghost" size="icon" onClick={() => onEditFork(step)} aria-label="Редактировать развилку">
                  <Pencil className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" onClick={() => onRemove(step)} aria-label="Удалить развилку">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </>
            )}
          </CardContent>
        </Card>
        <div className="ml-4 space-y-4 border-l-2 border-amber-500/30 pl-4">
          {(node.branches ?? []).map((branch) => (
            <div key={branch.index} className="space-y-2">
              <div className="text-sm font-medium text-amber-700 dark:text-amber-400">
                ↳ {branch.label || `Ветка ${branch.index + 1}`}
              </div>
              <div className="ml-3 space-y-2">
                {branch.head ? (
                  <StepNode
                    node={branch.head}
                    parent={null}
                    isMax={isMax}
                    canWrite={canWrite}
                    onRemove={onRemove}
                    onAddStepHere={onAddStepHere}
                    onAddForkHere={onAddForkHere}
                    onMove={onMove}
                    onEditFork={onEditFork}
                  />
                ) : (
                  <div className="rounded-md border border-dashed bg-muted/20 px-3 py-2 text-xs text-muted-foreground">
                    Ветка пуста
                  </div>
                )}
                {!branch.head && canWrite && (
                  // Пустая ветка: только «+ Шаг». «+ Развилка» появится после
                  // первого prompt-шага в этой ветке — иначе нечего ветвить.
                  <AddRow
                    isMax={isMax}
                    showFork={false}
                    onAddStep={() =>
                      onAddStepHere({ parent_fork_id: step.id, branch_index: branch.index })
                    }
                    onAddFork={() =>
                      onAddForkHere({ parent_fork_id: step.id, branch_index: branch.index })
                    }
                  />
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  // prompt-шаг. canMoveUp/canMoveDown вычисляются по соседям в линейной части:
  // вверх — если предшественник тоже prompt; вниз — если следующий prompt.
  const canMoveUp = parent != null && parent.step.step_type === "prompt"
  const canMoveDown = node.next != null && node.next.step.step_type === "prompt"

  // Заголовок карточки = имя шага если задано, иначе заголовок промпта.
  // Номер шага не показываем — position это технический timestamp создания и
  // после reorder он не отражает фактического порядка в графе.
  const headline = step.name?.trim() || step.prompt?.title || `Промпт #${step.prompt_id}`
  const subline = step.name?.trim() ? step.prompt?.title : undefined

  return (
    <div className="space-y-2">
      <Card>
        <CardContent className="flex items-center gap-3 p-4">
          <div className="flex-1 min-w-0">
            <p className="truncate text-sm font-medium">{headline}</p>
            {subline && <p className="truncate text-xs text-muted-foreground">{subline}</p>}
          </div>
          {canWrite && (
            <>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => onMove(step.id, "up")}
                disabled={!canMoveUp}
                title={canMoveUp ? "Поднять выше" : "Уже первый в этой подцепочке"}
              >
                <ArrowUp className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => onMove(step.id, "down")}
                disabled={!canMoveDown}
                title={canMoveDown ? "Опустить ниже" : "Уже последний в линии"}
              >
                <ArrowDown className="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" onClick={() => onRemove(step)} aria-label="Удалить шаг">
                <Trash2 className="h-4 w-4" />
              </Button>
            </>
          )}
        </CardContent>
      </Card>
      {node.next ? (
        <StepNode
          node={node.next}
          parent={node}
          isMax={isMax}
          canWrite={canWrite}
          onRemove={onRemove}
          onAddStepHere={onAddStepHere}
          onAddForkHere={onAddForkHere}
          onMove={onMove}
          onEditFork={onEditFork}
        />
      ) : (
        canWrite && (
          <AddRow
            isMax={isMax}
            onAddStep={() => onAddStepHere({ after_step_id: step.id })}
            onAddFork={() => onAddForkHere({ after_step_id: step.id })}
          />
        )
      )}
    </div>
  )
}

function pluralBranches(n: number) {
  if (n === 1) return "ветка"
  if (n >= 2 && n <= 4) return "ветки"
  return "веток"
}

// --- Add row (тонкий ряд кнопок «+ Шаг» и «+ Развилка») ---

function AddRow({
  isMax,
  onAddStep,
  onAddFork,
  showFork = true,
}: {
  isMax: boolean
  onAddStep: () => void
  onAddFork: () => void
  /** Скрывает «+ Развилка» — для пустых веток/цепочек (нечего ветвить пока
   *  нет ни одного prompt-шага). */
  showFork?: boolean
}) {
  return (
    <div className="flex flex-wrap gap-2">
      <Button variant="outline" size="sm" onClick={onAddStep}>
        <Plus className="mr-1.5 h-4 w-4" />
        Шаг
      </Button>
      {showFork && (
        <Button
          variant="outline"
          size="sm"
          onClick={onAddFork}
          disabled={!isMax}
          title={isMax ? "Добавить развилку" : "Развилки доступны на тарифе Max"}
        >
          {isMax ? (
            <GitBranch className="mr-1.5 h-4 w-4" />
          ) : (
            <Lock className="mr-1.5 h-4 w-4" />
          )}
          Развилка
        </Button>
      )}
    </div>
  )
}

// --- Add step dialog ---

function AddStepDialog({
  open,
  onClose,
  onSubmit,
}: {
  open: boolean
  onClose: () => void
  onSubmit: (input: { promptId: number; name?: string }) => void
}) {
  const [promptId, setPromptId] = useState<number | null>(null)
  const [promptTitle, setPromptTitle] = useState("")
  const [name, setName] = useState("")

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Добавить шаг</DialogTitle>
          <DialogDescription>
            Выберите промпт. Ответ этого шага станет доступен последующим как переменная.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <PromptPicker
            value={promptId}
            selectedTitle={promptTitle}
            onChange={(id, prompt) => {
              setPromptId(id)
              setPromptTitle(prompt.title)
            }}
            placeholder="Найти и выбрать промпт…"
          />
          <Input
            placeholder="Название шага (опционально)"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={100}
          />
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            Отмена
          </Button>
          <Button
            onClick={() => promptId && onSubmit({ promptId, name: name.trim() || undefined })}
            disabled={!promptId}
          >
            Добавить
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// --- Add fork dialog ---

// AddForkDialog — обёртка вокруг ForkDialog в режиме создания. Принимает
// branchLabels (без next_step_id — у новой развилки веток ничего нет за собой).
function AddForkDialog({
  open,
  onClose,
  onSubmit,
}: {
  open: boolean
  onClose: () => void
  onSubmit: (input: { name?: string; branchLabels: string[] }) => void
}) {
  return (
    <ForkDialog
      mode="create"
      open={open}
      onClose={onClose}
      onSubmit={({ name, branches }) =>
        onSubmit({ name, branchLabels: branches.map((b) => b.label) })
      }
    />
  )
}

interface BranchDraft {
  /** Стабильный draft-id для React key — без него фокус прыгает при addBranch/removeBranch
   *  (key={i} переиспользовал DOM-узел соседней ветки и input терял курсор). MN-51. */
  _id: string
  label: string
  /** Существующий next_step_id ветки — сохраняется при редактировании, чтобы
   *  не оторвать подцепочку шагов от ветки. У новых веток nil. */
  next_step_id?: number | null
}

// branchDraftId — стабильный id для key={...}. Не криптографически уникальный,
// просто чтобы не повторялся в одной форме. crypto.randomUUID для prod, fallback
// для старых браузеров и jsdom.
function branchDraftId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID()
  }
  return `b-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

// ForkDialog — общий диалог для создания и редактирования развилки. В режиме
// edit берёт initialName + initialBranches; ветки могут содержать next_step_id,
// которые сохраняются при submit (чтобы не разорвать существующие подцепочки).
function ForkDialog({
  mode,
  open,
  onClose,
  onSubmit,
  initialName = "",
  initialBranches,
}: {
  mode: "create" | "edit"
  open: boolean
  onClose: () => void
  onSubmit: (input: { name?: string; branches: BranchDraft[] }) => void
  initialName?: string
  initialBranches?: { label: string; next_step_id?: number | null }[]
}) {
  const initial: BranchDraft[] = initialBranches?.length
    ? initialBranches.map((b) => ({ _id: branchDraftId(), label: b.label, next_step_id: b.next_step_id ?? null }))
    : [
        { _id: branchDraftId(), label: "Ветка 1" },
        { _id: branchDraftId(), label: "Ветка 2" },
      ]

  const [name, setName] = useState(initialName)
  const [branches, setBranches] = useState<BranchDraft[]>(initial)
  // Подтверждение удаления ветки, если в ней есть подключённые шаги (next_step_id).
  // Без подтверждения шаги останутся «висящими» — пользователю нужно явно согласиться.
  const [pendingRemoveBranchIdx, setPendingRemoveBranchIdx] = useState<number | null>(null)

  const updateLabel = (i: number, v: string) => {
    setBranches((prev) => prev.map((b, idx) => (idx === i ? { ...b, label: v } : b)))
  }
  const addBranch = () =>
    setBranches((prev) => [...prev, { _id: branchDraftId(), label: `Ветка ${prev.length + 1}`, next_step_id: null }])
  const removeBranchImmediate = (i: number) => {
    setBranches((prev) => prev.filter((_, idx) => idx !== i))
  }
  const removeBranch = (i: number) => {
    if (branches.length <= 2) return
    const target = branches[i]
    if (target.next_step_id != null) {
      // Есть подключённые шаги — спрашиваем подтверждение через ConfirmDialog.
      setPendingRemoveBranchIdx(i)
      return
    }
    removeBranchImmediate(i)
  }

  const trimmed = branches.map((b) => ({ ...b, label: b.label.trim() })).filter((b) => b.label.length > 0)
  const allUnique = new Set(trimmed.map((b) => b.label)).size === trimmed.length
  const valid = trimmed.length >= 2 && allUnique

  const title = mode === "edit" ? "Редактировать развилку" : "Добавить развилку"
  const cta = mode === "edit" ? "Сохранить" : "Создать развилку"
  const description =
    mode === "edit"
      ? "Поменяйте название развилки и ветки. Связи с шагами внутри веток сохраняются."
      : "Развилка — это пустой контейнер с ветками. После создания вы добавите свои промпты внутрь каждой ветки как обычные шаги."

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="space-y-2">
            <Label>Название развилки</Label>
            <Input
              placeholder="Как назовёте развилку — например: Тип задачи"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
            />
          </div>
          <div className="space-y-2">
            <Label>Ветки (минимум 2, без дубликатов)</Label>
            <div className="space-y-2">
              {branches.map((b, i) => (
                <div key={b._id} className="flex gap-2">
                  <Input
                    value={b.label}
                    onChange={(e) => updateLabel(i, e.target.value)}
                    maxLength={100}
                    placeholder={`Ветка ${i + 1}`}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeBranch(i)}
                    disabled={branches.length <= 2}
                    aria-label="Удалить ветку"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
              <Button type="button" variant="outline" size="sm" onClick={addBranch}>
                <Plus className="mr-1.5 h-4 w-4" />
                Добавить ветку
              </Button>
            </div>
            {!allUnique && (
              <p className="text-xs text-destructive">Названия веток должны быть уникальными.</p>
            )}
          </div>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            Отмена
          </Button>
          <Button
            onClick={() => valid && onSubmit({ name: name.trim() || undefined, branches: trimmed })}
            disabled={!valid}
          >
            {cta}
          </Button>
        </DialogFooter>
      </DialogContent>
      <ConfirmDialog
        open={pendingRemoveBranchIdx !== null}
        onOpenChange={(v) => !v && setPendingRemoveBranchIdx(null)}
        title="Удалить ветку?"
        description={
          pendingRemoveBranchIdx !== null
            ? `Ветка «${branches[pendingRemoveBranchIdx]?.label ?? ""}» содержит подключённые шаги. После удаления ветки они останутся «висящими» — увидеть и удалить их можно будет на странице /canvas.`
            : ""
        }
        confirmLabel="Удалить ветку"
        onConfirm={() => {
          if (pendingRemoveBranchIdx !== null) removeBranchImmediate(pendingRemoveBranchIdx)
          setPendingRemoveBranchIdx(null)
        }}
      />
    </Dialog>
  )
}
