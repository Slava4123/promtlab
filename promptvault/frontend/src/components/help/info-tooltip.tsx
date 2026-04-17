import { Info } from "lucide-react"

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface InfoTooltipProps {
  /** Короткий текст подсказки — 1-2 предложения. */
  children: React.ReactNode
  /** aria-label для trigger-кнопки. Дефолт: «Что это?». */
  ariaLabel?: string
}

/**
 * Inline ⓘ-кнопка рядом с label-ом поля. Click-trigger (работает на тач),
 * Esc/click outside закрывает, фокус-trap внутри popover (base-ui дефолт).
 */
export function InfoTooltip({ children, ariaLabel = "Что это?" }: InfoTooltipProps) {
  return (
    <Popover>
      <PopoverTrigger
        aria-label={ariaLabel}
        className="inline-flex h-4 w-4 items-center justify-center rounded text-muted-foreground/70 transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand/40"
      >
        <Info className="h-3 w-3" />
      </PopoverTrigger>
      <PopoverContent
        align="start"
        side="top"
        className="w-[min(18rem,calc(100vw-1.5rem))] p-3 text-[0.78rem] leading-relaxed text-muted-foreground"
      >
        {children}
      </PopoverContent>
    </Popover>
  )
}
