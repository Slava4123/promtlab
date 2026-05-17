import { Sparkles } from "lucide-react"
import type { NarrativeSegments } from "@/lib/analytics-narrative"

interface NarrativeBannerProps {
  segments: NarrativeSegments
}

// NarrativeBanner — top-of-page AI-style summary без LLM-вызова.
// Сегменты собираются в buildNarrative() из existing data.
// Визуально: violet gradient + Sparkles icon. Static informational div (no link).
export function NarrativeBanner({ segments }: NarrativeBannerProps) {
  const hasAction = segments.actionHint !== null

  return (
    <div className="flex items-center gap-3 rounded-lg border border-violet-500/25 bg-gradient-to-r from-violet-500/10 to-violet-500/5 px-4 py-3">
      <Sparkles className="size-5 shrink-0 text-violet-500" aria-hidden="true" />
      <div className="flex-1 text-sm leading-relaxed">
        <span className="font-medium">{segments.summary}</span>
        {segments.topModel && (
          <>
            <span className="mx-1.5 text-muted-foreground">·</span>
            <span>{segments.topModel}</span>
          </>
        )}
        {segments.streak && (
          <>
            <span className="mx-1.5 text-muted-foreground">·</span>
            <span>{segments.streak}</span>
          </>
        )}
        {hasAction && (
          <div className="mt-0.5 text-xs text-muted-foreground">{segments.actionHint}</div>
        )}
      </div>
    </div>
  )
}
