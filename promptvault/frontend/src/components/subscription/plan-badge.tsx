import { Zap, Sparkles, Crown, type LucideIcon } from "lucide-react"
import type { PlanID } from "@/api/types"

const planConfig: Record<
  PlanID,
  { label: string; bg: string; text: string; Icon: LucideIcon }
> = {
  free: { label: "Free", bg: "bg-muted", text: "text-muted-foreground", Icon: Zap },
  pro: { label: "Pro", bg: "bg-violet-500/15", text: "text-violet-600 dark:text-violet-400", Icon: Sparkles },
  pro_yearly: { label: "Pro · год", bg: "bg-violet-500/15", text: "text-violet-600 dark:text-violet-400", Icon: Sparkles },
  max: { label: "Max", bg: "bg-amber-500/15", text: "text-amber-600 dark:text-amber-400", Icon: Crown },
  max_yearly: { label: "Max · год", bg: "bg-amber-500/15", text: "text-amber-600 dark:text-amber-400", Icon: Crown },
}

interface PlanBadgeProps {
  planId: PlanID
  className?: string
}

export function PlanBadge({ planId, className = "" }: PlanBadgeProps) {
  const cfg = planConfig[planId] ?? planConfig.free
  const { Icon } = cfg

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[0.65rem] font-medium ${cfg.bg} ${cfg.text} ${className}`}
    >
      <Icon className="h-3 w-3" />
      {cfg.label}
    </span>
  )
}
