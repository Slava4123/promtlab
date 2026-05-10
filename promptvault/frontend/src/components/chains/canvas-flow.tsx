// CanvasFlow — read-only визуализатор графа цепочки на @xyflow/react.
//
// Графовая модель v3 (Phase 16 v3):
//   prompt-шаг → next_step_id (NULL → end-node)
//   fork-шаг   → conditions.branches[].next_step_id
//
// UX-приоритет: юзер должен сразу понимать, какой шаг к какой ветке принадлежит.
// Реализовано через swimlane-контейнеры (xyflow `type: 'group'` = `BranchGroupNode`):
// каждая ветка fork-шага оборачивается в тонированный контейнер с заголовком
// «Ветка: <label>». Все шаги ветки получают `parentId` контейнера и физически
// рисуются внутри. Цвет контейнера + label на стрелке от fork — accent, не
// единственный сигнал (для color-blind accessibility текстовое имя ветки есть
// и в стрелке, и в заголовке swimlane).
//
// Layout — ELKjs `layered` с `INCLUDE_CHILDREN` (Dagre плохо умеет parent/child
// группировку). Реализация в `use-chain-layout.ts`.

import { useMemo } from "react"
import {
  Background,
  BackgroundVariant,
  Controls,
  MarkerType,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  type Edge,
  type EdgeTypes,
  type Node,
  type NodeTypes,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"

import { Skeleton } from "@/components/ui/skeleton"
import { useELKLayout } from "./use-chain-layout"
import { PromptNode, type PromptNodeData } from "./nodes/prompt-node"
import { ForkNode, type ForkNodeData, type ForkBranchView } from "./nodes/fork-node"
import { EndNode } from "./nodes/end-node"
import { BranchGroupNode, type BranchGroupNodeData } from "./nodes/branch-group-node"
import { BranchEdge, type BranchEdgeData } from "./edges/branch-edge"
import type { ChainConditions, ChainDetail, ChainStep } from "@/api/types"

const nodeTypes: NodeTypes = {
  prompt: PromptNode,
  fork: ForkNode,
  end: EndNode,
  group: BranchGroupNode,
}

const edgeTypes: EdgeTypes = {
  branch: BranchEdge,
}

// Okabe-Ito-friendly палитра. Цвет назначается по индексу ветки внутри одного
// fork — у каждого fork палитра считается заново, межвыходного шеринга нет.
const BRANCH_PALETTE = [
  "#0072B2", // blue
  "#009E73", // bluish-green
  "#D55E00", // vermillion
  "#CC79A7", // pink
  "#56B4E9", // sky blue
  "#E69F00", // orange-yellow
  "#F0E442", // yellow
  "#999999", // grey
]

interface CanvasFlowProps {
  chain: ChainDetail
  /** В fullscreen-режиме контейнер растягивается на 100% родителя
   *  вместо фиксированной высоты viewport-12rem. */
  fillParent?: boolean
}

export function CanvasFlow(props: CanvasFlowProps) {
  return (
    <ReactFlowProvider>
      <CanvasFlowInner {...props} />
    </ReactFlowProvider>
  )
}

function CanvasFlowInner({ chain, fillParent = false }: CanvasFlowProps) {
  const { rawNodes, rawEdges } = useMemo(() => buildFlow(chain), [chain])
  const { nodes, edges } = useELKLayout(rawNodes, rawEdges)

  if (nodes.length === 0) {
    return <Skeleton className={fillParent ? "h-full w-full" : "h-[calc(100vh-12rem)] w-full"} />
  }

  return (
    <div className={fillParent ? "h-full w-full" : "h-[calc(100vh-12rem)] w-full"}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
        zoomOnDoubleClick={false}
        onlyRenderVisibleElements
        minZoom={0.1}
        maxZoom={2}
        fitView
        fitViewOptions={{ padding: 0.2, minZoom: 0.1, maxZoom: 1 }}
        proOptions={{ hideAttribution: true }}
      >
        <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
        <Controls className="!bg-card !border-border" showInteractive={false} />
        <MiniMap
          className="!bg-card !border-border"
          nodeColor={(n) => {
            if (n.type === "fork") return "#f59e0b"
            if (n.type === "end") return "#22c55e"
            if (n.type === "group") return (n.data as BranchGroupNodeData)?.color ?? "#94a3b8"
            return "#8b5cf6"
          }}
          pannable
          zoomable
        />
      </ReactFlow>
    </div>
  )
}

// --- Преобразование chain → nodes/edges ---

interface BuildResult {
  rawNodes: Node[]
  rawEdges: Edge[]
}

interface BranchGroupRef {
  id: string
  color: string
  label: string
}

function buildFlow(chain: ChainDetail): BuildResult {
  const rawNodes: Node[] = []
  const rawEdges: Edge[] = []
  const stepsByID = new Map<number, ChainStep>()
  for (const s of chain.steps) {
    stepsByID.set(s.id, s)
  }

  // Назначаем каждой ветке каждого fork-шага группу-контейнер. Заодно собираем
  // ownership: stepID → groupID. Это нужно для расстановки parentId на узлах
  // и для определения parentId самих fork-узлов (вложенные fork внутри ветки
  // другого fork должны быть внутри его группы).
  const groupRefByForkBranch = new Map<string, BranchGroupRef>() // ключ: `${forkID}:${branchIdx}`
  const ownershipGroupID = new Map<number, string>() // stepID → groupID

  for (const step of chain.steps) {
    if (step.step_type !== "fork") continue
    const branches = parseBranches(step.conditions)
    branches.forEach((b, idx) => {
      const groupID = `group-${step.id}-${idx}`
      const color = BRANCH_PALETTE[idx % BRANCH_PALETTE.length]
      groupRefByForkBranch.set(`${step.id}:${idx}`, { id: groupID, color, label: b.label })
      // Trace подцепочки ветки: помечаем все шаги до следующего fork как
      // принадлежащие этой группе. Сам встретившийся fork (вложенный) тоже
      // принадлежит группе — его parentId = текущая группа.
      if (b.next_step_id == null) return
      let cur = stepsByID.get(b.next_step_id)
      while (cur) {
        if (ownershipGroupID.has(cur.id)) break
        ownershipGroupID.set(cur.id, groupID)
        if (cur.step_type === "fork") break // его branches — собственные группы
        if (cur.next_step_id == null) break
        const next = stepsByID.get(cur.next_step_id)
        if (!next) break
        cur = next
      }
    })
  }

  // Создаём group-узлы для каждой ветки. Их parentId = группа, к которой
  // принадлежит fork-шаг (вложенная группировка). Для root fork — без parentId.
  for (const step of chain.steps) {
    if (step.step_type !== "fork") continue
    const branches = parseBranches(step.conditions)
    const forkParent = ownershipGroupID.get(step.id) // вложен в чью-то ветку?
    branches.forEach((_, idx) => {
      const ref = groupRefByForkBranch.get(`${step.id}:${idx}`)!
      const data: BranchGroupNodeData = { label: ref.label, color: ref.color }
      rawNodes.push({
        id: ref.id,
        type: "group",
        data,
        // Initial x не задаём — порядок children в layered ELK задаётся через
        // `elk.position '(idx,0)'` (вычисляется по индексу в массиве rawNodes).
        // См. use-chain-layout.ts.
        position: { x: 0, y: 0 },
        ...(forkParent ? { parentId: forkParent, extent: "parent" as const } : {}),
        style: { pointerEvents: "none" },
      })
    })
  }

  // Шаги (prompt + fork) и их ребра.
  for (const step of chain.steps) {
    const parent = ownershipGroupID.get(step.id)
    const parentProps = parent ? { parentId: parent, extent: "parent" as const } : {}

    if (step.step_type === "fork") {
      const branches = parseBranches(step.conditions)
      const branchViews: ForkBranchView[] = branches.map((b, idx) => {
        const target = b.next_step_id != null ? stepsByID.get(b.next_step_id) : undefined
        const targetName = target
          ? target.name?.trim() || target.prompt?.title || `Шаг ${target.position}`
          : "Конец цепочки"
        return { handleId: `branch-${idx}`, label: b.label, targetName }
      })
      const data: ForkNodeData = {
        stepID: step.id,
        position: step.position,
        name: step.name,
        branches: branchViews,
        runState: "idle",
        chosenHandleID: null,
      } as ForkNodeData
      rawNodes.push({
        id: nodeIDForStep(step.id),
        type: "fork",
        data,
        position: { x: 0, y: 0 },
        ...parentProps,
      })

      branches.forEach((b, idx) => {
        const ref = groupRefByForkBranch.get(`${step.id}:${idx}`)!
        // Ребро от fork → группа-контейнер ветки. Цель = group-узел, не первый
        // шаг — это даёт ELK правильную геометрию (вход в swimlane сверху).
        // Label оставляем: заголовок swimlane дублирует, но на стрелке имя ветки
        // — основной сигнал «куда идёт эта стрелка». При большом ELK spacing
        // (см. use-chain-layout) стрелки достаточно длинные, чтобы пилюли
        // разносились по X и не слипались.
        rawEdges.push(
          makeBranchEdge(`e-${step.id}-${idx}-group`, nodeIDForStep(step.id), ref.id, b.label, ref.color),
        )
        // Если ветка пустая — рисуем end-node внутри группы (видим, что ветка
        // ведёт в никуда).
        if (b.next_step_id == null) {
          const endID = `end-${step.id}-${idx}`
          rawNodes.push({
            id: endID,
            type: "end",
            data: { label: "Конец" },
            position: { x: 0, y: 0 },
            parentId: ref.id,
            extent: "parent" as const,
          })
        }
      })
      continue
    }

    // prompt-шаг
    const data: PromptNodeData = {
      stepID: step.id,
      position: step.position,
      name: step.name,
      promptID: step.prompt_id ?? 0,
      promptTitle: step.prompt?.title,
      promptContent: step.prompt?.content,
      runState: "idle",
    } as PromptNodeData
    rawNodes.push({
      id: nodeIDForStep(step.id),
      type: "prompt",
      data,
      position: { x: 0, y: 0 },
      ...parentProps,
    })

    if (step.next_step_id != null) {
      const target = stepsByID.get(step.next_step_id)
      if (target) {
        rawEdges.push(makeNeutralEdge(`e-${step.id}-${target.id}`, nodeIDForStep(step.id), nodeIDForStep(target.id)))
      }
      continue
    }

    // prompt-шаг без next_step_id — лист (конец цепочки или конец ветки).
    const endID = `end-${step.id}`
    rawNodes.push({
      id: endID,
      type: "end",
      data: {},
      position: { x: 0, y: 0 },
      ...parentProps,
    })
    rawEdges.push(makeNeutralEdge(`e-${step.id}-end`, nodeIDForStep(step.id), endID))
  }

  return { rawNodes, rawEdges }
}

function makeBranchEdge(id: string, source: string, target: string, label: string | undefined, color: string): Edge {
  const data: BranchEdgeData = { label, color }
  return {
    id,
    source,
    target,
    type: "branch",
    data,
    markerEnd: { type: MarkerType.ArrowClosed, color, width: 18, height: 18 },
  }
}

function makeNeutralEdge(id: string, source: string, target: string): Edge {
  return {
    id,
    source,
    target,
    type: "branch",
    data: {} satisfies BranchEdgeData,
    markerEnd: { type: MarkerType.ArrowClosed, color: "#94a3b8", width: 16, height: 16 },
  }
}

function nodeIDForStep(stepID: number): string {
  return `step-${stepID}`
}

function parseBranches(conditions: ChainConditions | undefined): Array<{ label: string; next_step_id: number | null | undefined }> {
  if (!conditions || !Array.isArray(conditions.branches)) return []
  return conditions.branches.map((b) => ({ label: b.label, next_step_id: b.next_step_id }))
}
