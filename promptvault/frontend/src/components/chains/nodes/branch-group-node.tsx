// BranchGroupNode — swimlane контейнер вокруг шагов одной ветки fork-шага.
// Тонированный фон, цветной border-left, плашка с названием ветки в углу.
//
// xyflow `type: 'group'` принимает любой кастомный компонент. Дочерние шаги
// получают `parentId` = id этого узла, ELK раскладывает их внутри.

import { memo } from "react"
import { Handle, Position } from "@xyflow/react"
import { GitBranch } from "lucide-react"

export interface BranchGroupNodeData extends Record<string, unknown> {
  /** Название ветки — отображается в шапке контейнера. */
  label: string
  /** Accent-цвет: используется для border-left и фоновой плашки заголовка.
   *  Цвет — акцент, не несущий сигнал; у каждой карточки шага и в edge-label
   *  есть собственное текстовое имя ветки (для color-blind accessibility). */
  color: string
}

function BranchGroupNodeBase({ data }: { data: BranchGroupNodeData }) {
  // Заголовок-pill раньше был внутри swimlane. Убран: его роль (имя ветки)
  // несёт label на стрелке fork→group + цвет границы swimlane. Дубликат сбивал
  // юзера — два pill'а в одной зоне читались как «название потерялось».
  // Имя ветки видно внизу карточки как subtle subtitle для тех, кто прокрутил
  // далеко вниз и не видит входящей стрелки.
  return (
    <div
      className="relative h-full w-full rounded-lg border border-l-4"
      style={{
        borderLeftColor: data.color,
        borderColor: `${data.color}40`,
        backgroundColor: `${data.color}0d`,
      }}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="!h-1 !w-1 !min-h-0 !min-w-0 !border-0 !bg-transparent"
        isConnectable={false}
      />
      {/* Маленькая подпись в нижнем-правом углу: ненавязчиво, как подпись на
          диаграмме. Не перекрывается с edge label сверху. */}
      <div
        className="pointer-events-none absolute bottom-1 right-2 inline-flex items-center gap-1 text-[10px] font-medium uppercase tracking-wide"
        style={{ color: data.color, opacity: 0.7 }}
        aria-label={`Ветка: ${data.label}`}
      >
        <GitBranch className="h-2.5 w-2.5" aria-hidden />
        <span className="max-w-[140px] truncate">{data.label}</span>
      </div>
    </div>
  )
}

export const BranchGroupNode = memo(BranchGroupNodeBase)
