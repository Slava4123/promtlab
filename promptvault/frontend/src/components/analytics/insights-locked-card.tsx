import { Lock, ArrowRight } from "lucide-react"
import { Link } from "react-router-dom"

interface InsightsLockedCardProps {
  title: string
  description: string
}

// InsightsLockedCard — Pro teaser locked card (для Max-only insight types).
// Визуально соответствует InsightActionCard, но dashed border и lock icon.
// CTA — ссылка на /pricing.
export function InsightsLockedCard({ title, description }: InsightsLockedCardProps) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-foreground/2 p-4">
      <div className="mb-1.5 flex items-center gap-2">
        <Lock className="size-4 text-muted-foreground" aria-hidden="true" />
        <span className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
          {title}
        </span>
      </div>
      <p className="mb-2 text-sm text-foreground/70">{description}</p>
      <Link
        to="/pricing"
        className="inline-flex items-center gap-1 text-xs font-medium text-violet-600 dark:text-violet-400"
      >
        Доступно в Max
        <ArrowRight className="size-3" aria-hidden="true" />
      </Link>
    </div>
  )
}
