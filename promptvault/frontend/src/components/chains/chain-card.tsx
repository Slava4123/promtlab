// Phase 16 UI polish: карточка цепочки в filled-state /chains.
// Заменяет inline grid из chains/index.tsx. Изменения относительно прошлого:
//   - 5 кнопок (Запустить/Дерево/История/Редактор/Удалить) → 2 главные (Запустить, Редактор) + ⋯ menu
//   - mini-graph SVG для визуальной структуры (prompt = квадрат, fork = ромб, +N для длинных)
//   - badges metadata строкой (3 шага · 1 ветвление · 12 запусков) с RU склонением

import { Link, useNavigate } from "react-router-dom"
import { History, MoreHorizontal, Network, Pencil, PlayCircle, Trash2 } from "lucide-react"

import type { Chain } from "@/api/types"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { pluralizeRu } from "@/lib/pluralize"
import { ChainMiniGraph } from "./chain-mini-graph"

interface ChainCardProps {
  chain: Chain
  /** Может ли юзер редактировать (owner/editor в команде) — управляет ⋯ menu Удалить. */
  canWrite: boolean
  /** Viewer ли в команде — меняет надпись «Редактор» на «Просмотр». */
  isViewer: boolean
  /** Удалить (open ConfirmDialog в parent'е). */
  onDelete: (id: number, name: string) => void
}

export function ChainCard({ chain, canWrite, isViewer, onDelete }: ChainCardProps) {
  const navigate = useNavigate()
  const stepCount = chain.step_count ?? 0
  const hasBranching = chain.has_branching ?? false
  const runsCount = chain.saved_runs_count ?? 0
  const stepsPreview = chain.steps_preview ?? []

  const forksCount = stepsPreview.filter((s) => s.step_type === "fork").length

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex-row items-start justify-between gap-2 space-y-0 pb-2">
        <CardTitle className="line-clamp-1 text-base">{chain.name}</CardTitle>
        <DropdownMenu>
          <DropdownMenuTrigger
            className="-mr-2 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Дополнительные действия"
          >
            <MoreHorizontal className="h-4 w-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/chains/${chain.id}/canvas`)}>
              <Network className="mr-2 h-4 w-4" />
              Дерево
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => navigate(`/chains/${chain.id}/runs`)}>
              <History className="mr-2 h-4 w-4" />
              История
            </DropdownMenuItem>
            {canWrite && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => onDelete(chain.id, chain.name)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Удалить
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-3">
        {chain.description && (
          <p className="line-clamp-2 text-[0.78rem] text-muted-foreground">{chain.description}</p>
        )}
        <ChainMiniGraph stepsPreview={stepsPreview} totalSteps={stepCount} />
        <p className="text-[0.7rem] text-muted-foreground">
          {stepCount} {pluralizeRu(stepCount, "шаг", "шага", "шагов")}
          {hasBranching && forksCount > 0 && (
            <>
              {" · "}
              {forksCount} {pluralizeRu(forksCount, "ветвление", "ветвления", "ветвлений")}
            </>
          )}
          {runsCount > 0 && (
            <>
              {" · "}
              {runsCount} {pluralizeRu(runsCount, "запуск", "запуска", "запусков")}
            </>
          )}
        </p>
        <div className="mt-auto flex flex-wrap gap-2 pt-1">
          <Button size="sm" asChild>
            <Link to={`/chains/${chain.id}/run`}>
              <PlayCircle className="mr-2 h-4 w-4" />
              Запустить
            </Link>
          </Button>
          <Button size="sm" variant="outline" asChild>
            <Link to={`/chains/${chain.id}/edit`}>
              <Pencil className="mr-2 h-4 w-4" />
              {isViewer ? "Просмотр" : "Редактор"}
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
