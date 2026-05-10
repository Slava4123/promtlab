// Phase 16 UI polish: карточка шаблона в empty-state галерее /chains.
// Клик создаёт промпты+цепочку через applyTemplate и редиректит в
// /chains/{id}/edit. На время создания disabled + spinner.

import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"

import { applyTemplate, type ChainTemplate } from "@/lib/chain-templates"
import { Card, CardContent } from "@/components/ui/card"
import { ChainMiniGraph } from "./chain-mini-graph"

interface ChainTemplateCardProps {
  template: ChainTemplate
  /** team_id активного workspace; null = личное пространство. */
  teamId: number | null
}

export function ChainTemplateCard({ template, teamId }: ChainTemplateCardProps) {
  const navigate = useNavigate()
  const [pending, setPending] = useState(false)

  const stepsPreview = template.steps.map((_, i) => ({
    position: i + 1,
    step_type: "prompt" as const,
  }))

  const handleClick = async () => {
    if (pending) return
    setPending(true)
    try {
      const chain = await applyTemplate(template, teamId)
      toast.success("Цепочка создана из шаблона", {
        description: `${template.steps.length} шагов готовы к редактированию`,
      })
      navigate(`/chains/${chain.id}/edit`)
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Не удалось создать цепочку из шаблона",
      )
      setPending(false)
    }
  }

  return (
    <Card
      role="button"
      tabIndex={0}
      onClick={handleClick}
      onKeyDown={(e) => {
        if ((e.key === "Enter" || e.key === " ") && !pending) {
          e.preventDefault()
          handleClick()
        }
      }}
      aria-busy={pending}
      aria-label={`Создать цепочку из шаблона: ${template.title}`}
      className={`group cursor-pointer transition-[transform,box-shadow,border-color] duration-200 hover:-translate-y-0.5 hover:border-violet-500/30 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-500/40 ${
        pending ? "pointer-events-none opacity-70" : ""
      }`}
    >
      <CardContent className="flex flex-col gap-3 p-4">
        <div className="flex items-start gap-2">
          <span className="text-2xl leading-none" aria-hidden="true">
            {template.emoji}
          </span>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-semibold text-foreground">{template.title}</p>
            <p className="mt-0.5 text-[0.72rem] text-muted-foreground line-clamp-2">
              {template.description}
            </p>
          </div>
        </div>
        <ChainMiniGraph stepsPreview={stepsPreview} totalSteps={template.steps.length} />
        <div className="flex items-center justify-between text-[0.7rem] text-muted-foreground">
          <span>{template.steps.length} шага</span>
          <span className="flex items-center gap-1 font-medium text-violet-400 group-hover:text-violet-300">
            {pending ? (
              <>
                <Loader2 className="h-3 w-3 animate-spin" />
                Создаём…
              </>
            ) : (
              "Использовать →"
            )}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
