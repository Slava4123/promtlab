// EndNode — терминальный узел дерева, появляется в Run mode когда
// активная ветка достигла конца цепочки.

import { memo } from "react"
import { Handle, Position, type NodeProps } from "@xyflow/react"
import { Flag } from "lucide-react"

import { Card, CardContent } from "@/components/ui/card"

export interface EndNodeData extends Record<string, unknown> {
  label?: string
}

function EndNodeBase({ data }: NodeProps) {
  const d = data as EndNodeData
  return (
    <Card className="w-[160px] border-green-500/50 bg-green-500/5">
      <Handle type="target" position={Position.Top} className="!bg-green-500 !w-2 !h-2" />
      <CardContent className="flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium text-green-700 dark:text-green-400">
        <Flag className="h-4 w-4" />
        {d.label ?? "Конец"}
      </CardContent>
    </Card>
  )
}

export const EndNode = memo(EndNodeBase)
