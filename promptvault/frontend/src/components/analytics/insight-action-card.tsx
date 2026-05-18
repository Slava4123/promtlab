import { ArrowRight, type LucideIcon } from "lucide-react"
import { Link } from "react-router-dom"
import { cn } from "@/lib/utils"

export type InsightTone = "warning" | "info" | "success"

interface InsightActionCardProps {
  tone: InsightTone
  icon: LucideIcon
  title: string
  description: string
  href: string
  ctaLabel: string
  count?: number
  // empty=true — для discoverability: рендерим карточку даже когда инсайтов
  // этого типа нет. Приглушённая рамка (dashed), без count-badge, неактивный
  // CTA. Юзер видит, какие категории умной аналитики существуют.
  empty?: boolean
}

// Tailwind expects literal class names — no dynamic interpolation.
// Каждый tone — фиксированный набор классов для border/bg/text.
const TONE_BORDER: Record<InsightTone, string> = {
  warning: "border-amber-500/30 bg-amber-500/8",
  info: "border-violet-500/30 bg-violet-500/8",
  success: "border-emerald-500/30 bg-emerald-500/8",
}

const TONE_ACCENT: Record<InsightTone, string> = {
  warning: "text-amber-500",
  info: "text-violet-500",
  success: "text-emerald-500",
}

// InsightActionCard — actionable card для Smart Insights items.
// Цвет отражает tone: warning=amber, info=violet, success=emerald.
// CTA → React Router Link (deep link на конкретный фильтр/раздел).
export function InsightActionCard({
  tone,
  icon: Icon,
  title,
  description,
  href,
  ctaLabel,
  count,
  empty = false,
}: InsightActionCardProps) {
  return (
    <div
      className={cn(
        "rounded-lg border p-4",
        empty
          ? "border-dashed border-foreground/15 bg-transparent text-muted-foreground"
          : TONE_BORDER[tone],
      )}
    >
      <div className="mb-1.5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon
            className={cn("size-4", empty ? "text-muted-foreground" : TONE_ACCENT[tone])}
            aria-hidden="true"
          />
          <span
            className={cn(
              "text-[11px] font-semibold uppercase tracking-wide",
              empty ? "text-muted-foreground" : TONE_ACCENT[tone],
            )}
          >
            {title}
          </span>
        </div>
        {!empty && count !== undefined && (
          <span className="rounded-full bg-foreground/10 px-2 py-0.5 text-[11px] font-medium tabular-nums">
            {count}
          </span>
        )}
      </div>
      <p className={cn("mb-2 text-sm", empty ? "text-muted-foreground" : "text-foreground/90")}>
        {description}
      </p>
      <Link
        to={href}
        className={cn(
          "inline-flex items-center gap-1 text-xs font-medium transition-colors",
          empty
            ? "text-muted-foreground hover:text-foreground/80"
            : "text-foreground/80 hover:text-foreground",
        )}
      >
        {ctaLabel}
        <ArrowRight className="size-3" aria-hidden="true" />
      </Link>
    </div>
  )
}
