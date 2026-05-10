// Auto-layout для tree-canvas через ELKjs. Hierarchical layered top-down с
// поддержкой группировки (swimlane контейнеры).
//
// Критичные правила (см. https://eclipse.dev/elk/documentation/tooldevelopers/graphdatastructure.html):
//   1. Edges должны лежать в `edges` array у LCA (lowest common ancestor)
//      их source/target — НЕ все на root уровне. Иначе ELK при INCLUDE_CHILDREN
//      воспринимает их как long-hierarchical-edges и ломает bounding-box-расчёт
//      nested groups → перекрытия (issues elk#776, elkjs#112).
//   2. Порядок children в layer задаётся через `elk.position '(idx,0)'` на КАЖДОМ
//      child + `crossingMinimization.semiInteractive: true` на parent. Initial x
//      не работает для layered алгоритма; `crossingMinimization.strategy: NONE`
//      даёт непредсказуемый layout.

import { useEffect, useMemo, useState } from "react"
import ELK, { type ElkExtendedEdge, type ElkNode } from "elkjs/lib/elk.bundled.js"
import type { Edge, Node } from "@xyflow/react"

const SIZE = {
  prompt: { width: 280, height: 110 },
  // fork height вычисляется динамически из числа branches (см. sizeFor).
  // Базовое значение оставлено для типов без data.branches (не должно быть).
  fork: { width: 300, height: 220 },
  end: { width: 140, height: 60 },
} as const

// Реальная высота fork-карточки: header (~80px) + N × branch-row (~36px при
// одной строке текста, до ~64px при wrap) + bottom padding (12px). ELK должен
// получить достаточный bbox чтобы карточка не залазила в swimlane'ы ниже.
// 44px на ветку — чтобы хватило на 2 строки текста в branch-pill.
function forkHeight(branchCount: number): number {
  return 80 + branchCount * 44 + 12
}

const elk = new ELK()

const ROOT_OPTIONS: Record<string, string> = {
  "elk.algorithm": "layered",
  "elk.direction": "DOWN",
  "elk.layered.spacing.nodeNodeBetweenLayers": "110",
  "elk.spacing.nodeNode": "120",
  "elk.layered.nodePlacement.bk.fixedAlignment": "BALANCED",
  "elk.hierarchyHandling": "INCLUDE_CHILDREN",
  // Уважаем `elk.position` детей как ORDERING-hint (не как координаты).
  "elk.layered.crossingMinimization.semiInteractive": "true",
}

// Padding только для group-узлов: внутри swimlane оставляем место сверху для
// заголовка (если он есть). На root и листьях padding не нужен.
const GROUP_OPTIONS: Record<string, string> = {
  ...ROOT_OPTIONS,
  "elk.padding": "[top=28,left=20,bottom=20,right=20]",
}

interface BuildResult {
  nodes: Node[]
  edges: Edge[]
}

export function useELKLayout(rawNodes: Node[], rawEdges: Edge[]): BuildResult {
  const [layouted, setLayouted] = useState<BuildResult>({ nodes: [], edges: [] })
  const cacheKey = useMemo(
    () =>
      JSON.stringify({
        n: rawNodes.map((n) => [n.id, n.parentId, n.type]),
        e: rawEdges.map((e) => [e.id, e.source, e.target]),
      }),
    [rawNodes, rawEdges],
  )

  useEffect(() => {
    let cancelled = false
    layoutGraph(rawNodes, rawEdges).then((result) => {
      if (!cancelled) setLayouted(result)
    })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cacheKey])

  return layouted
}

async function layoutGraph(rawNodes: Node[], rawEdges: Edge[]): Promise<BuildResult> {
  if (rawNodes.length === 0) {
    return { nodes: [], edges: [] }
  }

  // === parentMap: stepID → parentID (или undefined для root) ===
  const parentMap = new Map<string, string | undefined>()
  for (const n of rawNodes) {
    parentMap.set(n.id, n.parentId ?? undefined)
  }

  // === childrenByParent: для рекурсивной сборки ELK иерархии ===
  const childrenByParent = new Map<string | undefined, Node[]>()
  // Также фиксируем индекс ребёнка внутри parent — это и есть `elk.position`-hint.
  const childIndex = new Map<string, number>()
  for (const n of rawNodes) {
    const arr = childrenByParent.get(n.parentId ?? undefined) ?? []
    childIndex.set(n.id, arr.length)
    arr.push(n)
    childrenByParent.set(n.parentId ?? undefined, arr)
  }

  // === LCA-распределение edges ===
  // ancestorsOf(id) — путь от id до root (включая сам id).
  const ancestorsOf = (id: string): string[] => {
    const out: string[] = []
    let cur: string | undefined = id
    while (cur) {
      out.push(cur)
      cur = parentMap.get(cur)
    }
    return out
  }
  const findLCA = (a: string, b: string): string | undefined => {
    const ancA = new Set(ancestorsOf(a))
    for (const node of ancestorsOf(b)) {
      if (ancA.has(node)) {
        // LCA — это сам узел только если он predок другого. Для нашей tree-
        // структуры edges идут sibling↔sibling или parent↔child, поэтому LCA
        // должен быть PARENT обоих, не сам узел. Если LCA == a или b, берём
        // его parent.
        if (node === a || node === b) return parentMap.get(node)
        return node
      }
    }
    return undefined // оба root → root container
  }
  // edgesByContainer[parentID|undefined] = [...]
  const edgesByContainer = new Map<string | undefined, Edge[]>()
  for (const e of rawEdges) {
    const container = findLCA(e.source, e.target) // undefined → root
    const arr = edgesByContainer.get(container) ?? []
    arr.push(e)
    edgesByContainer.set(container, arr)
  }

  const buildElkNode = (node: Node): ElkNode => {
    const children = (childrenByParent.get(node.id) ?? []).map(buildElkNode)
    const idx = childIndex.get(node.id) ?? 0
    // На каждом узле — elk.position(idx, 0) как ordering-hint в semiInteractive.
    // Для top-level (root children) idx — позиция в массиве rootNodes.
    const baseLayout: Record<string, string> = {
      "elk.position": `(${idx},0)`,
    }
    if (children.length > 0) {
      // У группы — собственные layout options + edges, лежащие внутри неё (LCA).
      const innerEdges = (edgesByContainer.get(node.id) ?? []).map(toElkEdge)
      return {
        id: node.id,
        children,
        edges: innerEdges,
        layoutOptions: { ...GROUP_OPTIONS, ...baseLayout },
      }
    }
    const size = sizeFor(node)
    return {
      id: node.id,
      width: size.width,
      height: size.height,
      layoutOptions: baseLayout,
    }
  }

  const elkGraph: ElkNode = {
    id: "root",
    layoutOptions: ROOT_OPTIONS,
    children: (childrenByParent.get(undefined) ?? []).map(buildElkNode),
    edges: (edgesByContainer.get(undefined) ?? []).map(toElkEdge),
  }

  const result = await elk.layout(elkGraph)

  const positions = new Map<string, { x: number; y: number; width?: number; height?: number }>()
  collectPositions(result, positions)

  const positioned: Node[] = rawNodes.map((n) => {
    const p = positions.get(n.id)
    if (!p) return n
    const out: Node = {
      ...n,
      position: { x: p.x, y: p.y },
    }
    if (n.type === "group" && p.width && p.height) {
      out.width = p.width
      out.height = p.height
      out.style = { ...n.style, width: p.width, height: p.height }
    }
    return out
  })

  return { nodes: positioned, edges: rawEdges }
}

function toElkEdge(e: Edge): ElkExtendedEdge {
  return { id: e.id, sources: [e.source], targets: [e.target] }
}

function collectPositions(
  node: ElkNode,
  out: Map<string, { x: number; y: number; width?: number; height?: number }>,
) {
  if (node.id !== "root") {
    out.set(node.id, {
      x: node.x ?? 0,
      y: node.y ?? 0,
      width: node.width,
      height: node.height,
    })
  }
  for (const child of node.children ?? []) {
    collectPositions(child, out)
  }
}

function sizeFor(node: Node): { width: number; height: number } {
  switch (node.type) {
    case "fork": {
      const branches = (node.data as { branches?: unknown[] } | undefined)?.branches
      const count = Array.isArray(branches) ? branches.length : 2
      return { width: SIZE.fork.width, height: forkHeight(count) }
    }
    case "end":
      return SIZE.end
    default:
      return SIZE.prompt
  }
}
