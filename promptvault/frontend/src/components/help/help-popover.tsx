import { HelpCircle, ArrowRight } from "lucide-react"
import { Link } from "react-router-dom"

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface HelpPopoverProps {
  title: string
  /** Контент подсказки. Используйте короткие <p> или <ul> внутри. */
  children: React.ReactNode
  /** Если задано — внизу popover-а показывается ссылка «Подробнее». */
  learnMoreHref?: string
  learnMoreLabel?: string
  /** aria-label для trigger-кнопки. Дефолт: «Подсказка». */
  ariaLabel?: string
}

export function HelpPopover({
  title,
  children,
  learnMoreHref,
  learnMoreLabel = "Подробнее в Помощи",
  ariaLabel = "Подсказка",
}: HelpPopoverProps) {
  return (
    <Popover>
      <PopoverTrigger
        aria-label={ariaLabel}
        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-foreground/[0.04] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand/40"
      >
        <HelpCircle className="h-4 w-4" />
      </PopoverTrigger>
      <PopoverContent
        align="end"
        side="bottom"
        className="w-[min(20rem,calc(100vw-1.5rem))] space-y-2 p-4"
      >
        <div className="text-sm font-semibold text-foreground">{title}</div>
        <div className="space-y-2 text-[0.82rem] leading-relaxed text-muted-foreground">
          {children}
        </div>
        {learnMoreHref && (
          <Link
            to={learnMoreHref}
            className="inline-flex items-center gap-1 pt-1 text-[0.78rem] text-brand-muted-foreground underline-offset-4 hover:underline"
          >
            {learnMoreLabel}
            <ArrowRight className="h-3 w-3" />
          </Link>
        )}
      </PopoverContent>
    </Popover>
  )
}
