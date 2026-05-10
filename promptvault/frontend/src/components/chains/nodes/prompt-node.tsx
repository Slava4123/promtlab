// PromptNode — обычный шаг цепочки. Объявлен ВНЕ родительского компонента и
// обёрнут в memo для производительности (React Flow re-render protection).
//
// HoverCard показывает превью полного содержимого промпта при наведении.

import { memo } from "react"
import { Handle, Position, useStore, type NodeProps } from "@xyflow/react"
import { CheckCircle2, FileText } from "lucide-react"

import { Card, CardContent } from "@/components/ui/card"
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card"

export type PromptNodeRunState = "pending" | "current" | "completed" | "idle"

export interface PromptNodeData extends Record<string, unknown> {
  stepID: number
  position: number
  /** Имя шага (опциональное, заданное юзером). */
  name: string
  /** ID промпта — fallback если promptTitle не загружен. */
  promptID: number
  /** Заголовок промпта — приоритетное отображение. */
  promptTitle?: string
  /** Полный контент промпта — для hover-preview. */
  promptContent?: string
  runState?: PromptNodeRunState
}

// Contextual Zoom threshold: ниже этого zoom переключаемся в compact-режим
// (только title крупнее, без subline) — иначе на zoom-out текст превращается
// в нечитаемую кашу. См. ADR / B2 в FEATURE_PROMPT_CHAINS.md.
const COMPACT_ZOOM_THRESHOLD = 0.5

function PromptNodeBase({ data, selected }: NodeProps) {
  const d = data as PromptNodeData
  const state = d.runState ?? "idle"
  const isCompleted = state === "completed"
  const isCurrent = state === "current"
  const compact = useStore((s) => s.transform[2] < COMPACT_ZOOM_THRESHOLD)

  // Заголовок = имя шага если задано, иначе заголовок промпта. Номер шага
  // (position) не показываем: это внутренний timestamp создания, а порядок
  // в графе после reorder определяется next_step_id, не position.
  const displayName = d.name?.trim() || d.promptTitle || `Промпт #${d.promptID}`
  const subline = d.name?.trim() && d.promptTitle ? d.promptTitle : null

  return (
    <HoverCard openDelay={200} closeDelay={100}>
      <HoverCardTrigger asChild>
        <div>
          <Card
            className={[
              "w-[280px] cursor-pointer transition-all",
              selected ? "ring-2 ring-primary" : "",
              isCurrent ? "border-primary shadow-[0_0_24px_rgba(139,92,246,0.4)] animate-pulse" : "",
              isCompleted ? "opacity-60" : "",
              state === "pending" ? "opacity-50" : "",
            ].join(" ")}
          >
            <Handle type="target" position={Position.Top} className="!bg-primary !w-2 !h-2" />
            <CardContent className="flex items-start gap-2 p-3">
              <div className="mt-0.5 rounded bg-blue-500/10 p-1 text-blue-600 dark:text-blue-400">
                {isCompleted ? <CheckCircle2 className="h-4 w-4 text-green-600" /> : <FileText className="h-4 w-4" />}
              </div>
              <div className="min-w-0 flex-1">
                <p className={compact ? "line-clamp-2 text-base font-semibold text-foreground" : "line-clamp-2 text-sm font-medium text-foreground"}>{displayName}</p>
                {!compact && subline && <p className="mt-0.5 line-clamp-1 text-xs text-muted-foreground">{subline}</p>}
              </div>
            </CardContent>
            <Handle type="source" position={Position.Bottom} className="!bg-primary !w-2 !h-2" />
          </Card>
        </div>
      </HoverCardTrigger>
      <HoverCardContent side="right" className="w-96">
        <div className="space-y-2">
          <p className="text-sm font-semibold">{d.promptTitle || `Промпт #${d.promptID}`}</p>
          {d.name && d.name !== d.promptTitle && (
            <p className="text-xs text-muted-foreground">Имя шага: {d.name}</p>
          )}
          {d.promptContent ? (
            <pre className="max-h-64 overflow-auto whitespace-pre-wrap rounded bg-muted/50 p-2 text-xs">
              {d.promptContent}
            </pre>
          ) : (
            <p className="text-xs italic text-muted-foreground">Контент промпта недоступен (возможно, промпт удалён).</p>
          )}
        </div>
      </HoverCardContent>
    </HoverCard>
  )
}

export const PromptNode = memo(PromptNodeBase)
