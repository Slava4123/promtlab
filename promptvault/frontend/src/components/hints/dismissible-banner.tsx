import { X } from "lucide-react"
import { useHintDismissed, type HintId } from "@/lib/hints"

interface DismissibleBannerProps {
  id: HintId
  title: string
  description: string
  cta?: { label: string; onClick: () => void }
  tone?: "violet" | "emerald" | "amber"
}

const tones: Record<NonNullable<DismissibleBannerProps["tone"]>, string> = {
  violet: "border-violet-500/30 bg-violet-500/[0.04] text-violet-300",
  emerald: "border-emerald-500/30 bg-emerald-500/[0.04] text-emerald-300",
  amber: "border-amber-500/30 bg-amber-500/[0.04] text-amber-300",
}

/**
 * DismissibleBanner — маленький info-блок с кнопкой × (M-13).
 * Состояние dismissed хранится в localStorage по id — после закрытия
 * не показывается повторно на том же устройстве.
 */
export function DismissibleBanner({
  id,
  title,
  description,
  cta,
  tone = "violet",
}: DismissibleBannerProps) {
  const [dismissed, dismiss] = useHintDismissed(id)
  if (dismissed) return null

  return (
    <div className={`flex items-start gap-3 rounded-xl border px-4 py-3 ${tones[tone]}`}>
      <div className="flex-1">
        <p className="text-sm font-medium text-foreground">{title}</p>
        <p className="mt-0.5 text-[0.8rem] text-muted-foreground">{description}</p>
        {cta && (
          <button
            type="button"
            onClick={cta.onClick}
            className="mt-2 text-[0.75rem] font-medium text-foreground underline underline-offset-4 hover:no-underline"
          >
            {cta.label}
          </button>
        )}
      </div>
      <button
        type="button"
        onClick={dismiss}
        aria-label="Скрыть подсказку"
        className="mt-0.5 rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted/40 hover:text-foreground"
      >
        <X className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
    </div>
  )
}
