"use client"

import type * as React from "react"
import * as HoverCardPrimitive from "@radix-ui/react-hover-card"

import { cn } from "@/lib/utils"

// MN-52: @radix-ui/react-hover-card — единственный @radix-ui dep в проекте
// (остальной UI на Base UI). Сохраняем намеренно: Base UI Popover/Tooltip
// требует click, а hover-card нужен hover-trigger UX (chains/nodes preview
// при наведении в graph editor — UX-эквивалентов в Base UI нет).
// Bundle impact <5KB gzip; миграция была бы UX-breaking.
// MN-53: убран React.forwardRef wrapper — в React 19 ref передаётся как
// обычный prop (https://react.dev/blog/2024/12/05/react-19#ref-as-a-prop).
const HoverCard = HoverCardPrimitive.Root
const HoverCardTrigger = HoverCardPrimitive.Trigger

type HoverCardContentProps = React.ComponentProps<typeof HoverCardPrimitive.Content>

function HoverCardContent({
  className,
  align = "center",
  sideOffset = 4,
  ...props
}: HoverCardContentProps) {
  return (
    <HoverCardPrimitive.Content
      align={align}
      sideOffset={sideOffset}
      className={cn(
        "z-50 w-80 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
        className,
      )}
      {...props}
    />
  )
}

export { HoverCard, HoverCardTrigger, HoverCardContent }
