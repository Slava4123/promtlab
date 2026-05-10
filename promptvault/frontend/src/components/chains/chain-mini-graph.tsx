// Phase 16 UI polish: компактная SVG-схема структуры цепочки для filled-state
// карточки. Рендерит первые ChainStepsPreviewLimit (5) шагов как boxes/rhombus
// со стрелками, для длинных цепочек добавляет "+N" badge.
//
// Не интерактивно (не клики, не focus) — просто визуальный hint. Для полного
// просмотра граф'а юзер идёт в /chains/{id}/canvas.

import type { ChainStepPreview } from "@/api/types"
import { pluralizeRu } from "@/lib/pluralize"

interface ChainMiniGraphProps {
  /** Первые N шагов цепочки (backend ограничивает до 5). */
  stepsPreview: ChainStepPreview[]
  /** Total шагов в цепочке. Если > stepsPreview.length — рисуем "+N" badge. */
  totalSteps: number
}

export function ChainMiniGraph({ stepsPreview, totalSteps }: ChainMiniGraphProps) {
  if (stepsPreview.length === 0) {
    return (
      <div className="flex h-10 items-center justify-center rounded-md border border-dashed border-border/60 text-[0.7rem] text-muted-foreground">
        Нет шагов
      </div>
    )
  }

  const nodeSize = 16
  const arrowWidth = 14
  const padding = 4
  const remaining = totalSteps - stepsPreview.length
  const showBadge = remaining > 0

  // Высчитываем total width: nodeSize * N + arrowWidth * (N-1) + (showBadge ? badgeW : 0).
  const badgeWidth = showBadge ? 32 : 0
  const totalWidth =
    stepsPreview.length * nodeSize +
    Math.max(stepsPreview.length - 1, 0) * arrowWidth +
    (showBadge ? arrowWidth + badgeWidth : 0) +
    padding * 2

  return (
    <svg
      viewBox={`0 0 ${totalWidth} 24`}
      className="h-6 w-full text-muted-foreground"
      aria-label={`Структура цепочки: ${totalSteps} ${pluralizeRu(totalSteps, "шаг", "шага", "шагов")}`}
    >
      {stepsPreview.map((step, i) => {
        const x = padding + i * (nodeSize + arrowWidth)
        const cx = x + nodeSize / 2
        const cy = 12
        const isFork = step.step_type === "fork"
        return (
          <g key={step.position}>
            {isFork ? (
              <polygon
                points={`${cx},${cy - 6} ${cx + 7},${cy} ${cx},${cy + 6} ${cx - 7},${cy}`}
                fill="currentColor"
                fillOpacity="0.2"
                stroke="currentColor"
                strokeOpacity="0.6"
                strokeWidth="1"
              />
            ) : (
              <rect
                x={x}
                y={cy - 6}
                width={nodeSize - 4}
                height="12"
                rx="2"
                fill="currentColor"
                fillOpacity="0.15"
                stroke="currentColor"
                strokeOpacity="0.5"
                strokeWidth="1"
              />
            )}
            {i < stepsPreview.length - 1 && (
              <ArrowSegment x={x + nodeSize - 4} y={cy} width={arrowWidth + 4} />
            )}
          </g>
        )
      })}
      {showBadge && (
        <g>
          <ArrowSegment
            x={padding + stepsPreview.length * nodeSize + (stepsPreview.length - 1) * arrowWidth - 4}
            y={12}
            width={arrowWidth + 4}
          />
          <rect
            x={totalWidth - badgeWidth - padding}
            y="6"
            width={badgeWidth}
            height="12"
            rx="6"
            fill="currentColor"
            fillOpacity="0.1"
            stroke="currentColor"
            strokeOpacity="0.4"
          />
          <text
            x={totalWidth - badgeWidth / 2 - padding}
            y="15"
            fontSize="9"
            fontWeight="500"
            textAnchor="middle"
            fill="currentColor"
            fillOpacity="0.7"
          >
            +{remaining}
          </text>
        </g>
      )}
    </svg>
  )
}

function ArrowSegment({ x, y, width }: { x: number; y: number; width: number }) {
  const endX = x + width - 3
  return (
    <g stroke="currentColor" strokeOpacity="0.5" strokeWidth="1" fill="none">
      <line x1={x} y1={y} x2={endX} y2={y} />
      <polyline points={`${endX - 3},${y - 2} ${endX},${y} ${endX - 3},${y + 2}`} fill="currentColor" fillOpacity="0.5" />
    </g>
  )
}
