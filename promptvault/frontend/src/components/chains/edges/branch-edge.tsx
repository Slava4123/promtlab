// BranchEdge — кастомное ребро с пилюлей-label.
//
// Почему НЕ SVG <rect>+<text> внутри того же <g>: при множественных fork-edges
// прямоугольник одного edge перекрывает только path того же edge. Линии
// СОСЕДНИХ edges (которые идут позже в массиве React Flow) рисуются поверх
// прямоугольника, потому что React Flow выводит каждый edge как отдельный
// `<g class="react-flow__edge">` и порядок отрисовки = порядок в массиве.
//
// Почему `<EdgeLabelRenderer>`: это portal, который рендерит DIV-ы ПОВЕРХ
// всего SVG-слоя через отдельный `.react-flow__edgelabel-renderer` контейнер.
// HTML-elements в этом DOM-overlay гарантированно лежат выше любых SVG paths,
// независимо от их порядка. Источник: https://reactflow.dev/learn/customization/edge-labels
//
// Pill ставим у TARGET (group-узла), а не в середине bezier — у середины
// разные edges того же fork лежат рядом и pill'ы могут наезжать друг на друга.
// У target каждая стрелка одинока.

import { memo, useMemo, type CSSProperties } from "react"
import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from "@xyflow/react"

export interface BranchEdgeData extends Record<string, unknown> {
  label?: string
  color?: string
}

const TARGET_OFFSET = 24

// MN-44: вынесены const styles за render — не пересоздаются между rerender'ами.
// Динамика только в edgeStroke/labelStyle через useMemo.
const labelBaseStyle: CSSProperties = {
  position: "absolute",
  background: "rgb(255, 255, 255)",
  borderRadius: "999px",
  padding: "3px 10px",
  fontSize: "12px",
  fontWeight: 600,
  maxWidth: "240px",
  whiteSpace: "normal",
  wordBreak: "break-word",
  display: "-webkit-box",
  WebkitLineClamp: 3,
  WebkitBoxOrient: "vertical",
  overflow: "hidden",
  textOverflow: "ellipsis",
  lineHeight: "1.25",
  pointerEvents: "none",
  boxShadow: "0 1px 2px rgba(0,0,0,0.08)",
  zIndex: 10,
}

function BranchEdgeBase(props: EdgeProps) {
  const { id, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, markerEnd, data } =
    props
  const d = (data as BranchEdgeData) ?? {}
  const color = d.color ?? "#94a3b8"
  const [path] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  })

  // MN-44: useMemo сохраняет ссылку на style между rerender'ами при тех же
  // (color, targetX, targetY) — позволяет React.memo на BaseEdge/div работать.
  const edgeStyle = useMemo<CSSProperties>(
    () => ({ stroke: color, strokeWidth: 2.5 }),
    [color],
  )
  const labelStyle = useMemo<CSSProperties>(
    () => ({
      ...labelBaseStyle,
      transform: `translate(-50%, -50%) translate(${targetX}px, ${targetY - TARGET_OFFSET}px)`,
      border: `2px solid ${color}`,
      color,
    }),
    [color, targetX, targetY],
  )

  return (
    <>
      <BaseEdge id={id} path={path} style={edgeStyle} markerEnd={markerEnd} />
      {d.label && (
        <EdgeLabelRenderer>
          <div
            // nodrag/nopan — чтобы pill не ломал ReactFlow gestures.
            //
            // Контейнер `.react-flow__edgelabel-renderer` имеет z-index: auto,
            // и в DOM-порядке он ПЕРЕД `.react-flow__nodes`. Это значит nodes
            // (включая group-узел = target) лежат поверх labels по умолчанию.
            // Ставим явный z-index: 10 на pill, чтобы он гарантированно был
            // выше всех соседей в stacking context viewport'а — иначе pill
            // прячется под group-узлом и линия SVG просвечивает.
            className="nodrag nopan"
            style={labelStyle}
          >
            {d.label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  )
}

export const BranchEdge = memo(BranchEdgeBase)
