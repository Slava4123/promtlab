// ForkNode — узел развилки. Янтарный border, иконка GitBranch, читаемый список
// branch-pills с метками. HoverCard показывает full-список branches с детально.
//
// Один общий source handle снизу — чтобы Dagre корректно расставил branches
// в дочерние узлы без визуального шума от множества handle-точек.

import { memo } from "react"
import { Handle, Position, useStore, type NodeProps } from "@xyflow/react"
import { CheckCircle2, GitBranch } from "lucide-react"

import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card"

export type ForkNodeRunState = "pending" | "current" | "completed" | "idle"

export interface ForkBranchView {
  /** Уникальный handle id для xyflow connector (формат "branch-N"). */
  handleId: string
  label: string
  /** Описание ветки — куда ведёт (имя следующего шага или «Конец»). */
  targetName?: string
}

export interface ForkNodeData extends Record<string, unknown> {
  stepID: number
  position: number
  name: string
  /** Заголовок промпта (роутер) — отображается в hover. */
  promptTitle?: string
  promptContent?: string
  branches: ForkBranchView[]
  runState?: ForkNodeRunState
  /** Подсветить выбранную ветку (run-mode). */
  chosenHandleID?: string | null
}

// Contextual Zoom threshold: на zoom-out скрываем подробный список branches —
// он превращается в нечитаемую кашу — оставляем только header+название и
// счётчик «N веток». Symmetric с prompt-node.
const COMPACT_ZOOM_THRESHOLD = 0.5

function ForkNodeBase({ data, selected }: NodeProps) {
  const d = data as ForkNodeData
  const state = d.runState ?? "idle"
  const isCompleted = state === "completed"
  const isCurrent = state === "current"
  const compact = useStore((s) => s.transform[2] < COMPACT_ZOOM_THRESHOLD)

  return (
    <HoverCard openDelay={200} closeDelay={100}>
      <HoverCardTrigger asChild>
        <div>
          <Card
            className={[
              "w-[300px] cursor-pointer border-amber-500/50 transition-all",
              selected ? "ring-2 ring-amber-500" : "",
              isCurrent ? "shadow-[0_0_28px_rgba(245,158,11,0.55)] animate-pulse" : "",
              isCompleted ? "opacity-60" : "",
              state === "pending" ? "opacity-50" : "",
            ].join(" ")}
          >
            <Handle type="target" position={Position.Top} className="!bg-amber-500 !w-2 !h-2" />
            <CardHeader className="flex flex-row items-center gap-2 p-3 pb-2">
              <div className="rounded bg-amber-500/15 p-1 text-amber-600 dark:text-amber-400">
                {isCompleted ? <CheckCircle2 className="h-4 w-4 text-green-600" /> : <GitBranch className="h-4 w-4" />}
              </div>
              <div className="flex min-w-0 flex-col">
                <span className="text-xs text-muted-foreground">развилка</span>
                <p className="line-clamp-2 text-sm font-medium text-foreground">{d.name || "Развилка"}</p>
                {d.promptTitle && (
                  <p className="mt-0.5 line-clamp-1 text-xs text-muted-foreground">{d.promptTitle}</p>
                )}
              </div>
            </CardHeader>
            <CardContent className="space-y-1.5 px-3 pb-3 pt-0">
              {compact ? (
                <p className="text-center text-sm font-semibold text-amber-600 dark:text-amber-400">
                  {d.branches.length} {pluralBranches(d.branches.length)}
                </p>
              ) : (
                d.branches.map((b) => {
                  const chosen = d.chosenHandleID === b.handleId
                  return (
                    <div
                      key={b.handleId}
                      className={[
                        "rounded-md border px-2.5 py-1.5 text-xs leading-tight",
                        chosen
                          ? "border-amber-500 bg-amber-500/15 font-semibold text-foreground"
                          : "border-border bg-muted/30 text-foreground",
                      ].join(" ")}
                    >
                      <p className="line-clamp-1 font-medium">{b.label}</p>
                      {b.targetName && (
                        <p className="line-clamp-1 text-[10px] text-muted-foreground">→ {b.targetName}</p>
                      )}
                    </div>
                  )
                })
              )}
            </CardContent>
            {/* Один общий source-handle снизу — Dagre расставит branches горизонтально по target'ам. */}
            <Handle type="source" position={Position.Bottom} className="!bg-amber-500 !w-2 !h-2" />
          </Card>
        </div>
      </HoverCardTrigger>
      <HoverCardContent side="right" className="w-96">
        <div className="space-y-2">
          <p className="text-sm font-semibold">{d.name || "Развилка"}</p>
          {d.promptTitle && (
            <p className="text-xs text-muted-foreground">Промпт-роутер: {d.promptTitle}</p>
          )}
          <div>
            <p className="mb-1 text-xs font-medium uppercase text-muted-foreground">Варианты ({d.branches.length})</p>
            <ul className="space-y-1 text-xs">
              {d.branches.map((b, i) => (
                <li key={b.handleId} className="rounded border border-border/60 px-2 py-1">
                  <span className="font-medium">{i + 1}.</span> {b.label}
                  {b.targetName && (
                    <span className="ml-1 text-muted-foreground">→ {b.targetName}</span>
                  )}
                </li>
              ))}
            </ul>
          </div>
          {d.promptContent && (
            <details>
              <summary className="cursor-pointer text-xs text-muted-foreground">
                Содержимое промпта-роутера
              </summary>
              <pre className="mt-1 max-h-40 overflow-auto whitespace-pre-wrap rounded bg-muted/50 p-2 text-xs">
                {d.promptContent}
              </pre>
            </details>
          )}
        </div>
      </HoverCardContent>
    </HoverCard>
  )
}

export const ForkNode = memo(ForkNodeBase)

function pluralBranches(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return "ветка"
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return "ветки"
  return "веток"
}
