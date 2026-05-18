import { Flame } from "lucide-react"
import { Card } from "@/components/ui/card"

interface StreakTrackerProps {
  current: number
  longest: number
  activeToday: boolean
}

// StreakTracker — current streak counter + last-7-days dots.
// Заполненные dots = current streak (capped at 7 для отображения).
export function StreakTracker({ current, longest, activeToday }: StreakTrackerProps) {
  const filledCount = Math.min(current, 7)
  return (
    <Card className="p-4">
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-[11px] uppercase tracking-wide text-muted-foreground">Streak</span>
        <Flame className="size-4 text-amber-500" aria-hidden="true" />
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-2xl font-bold tabular-nums">{current}</span>
        <span className="text-xs text-muted-foreground">{`best ${longest}`}</span>
      </div>
      <div className="mt-2 flex gap-1">
        {Array.from({ length: 7 }, (_, i) => {
          const filled = i < filledCount
          return (
            <span
              key={i}
              data-streak-dot
              data-filled={filled}
              className={
                filled
                  ? "h-3 w-3 rounded-sm bg-amber-500"
                  : "h-3 w-3 rounded-sm bg-foreground/10"
              }
              aria-hidden="true"
            />
          )
        })}
      </div>
      {!activeToday && current > 0 && (
        <p className="mt-1 text-[11px] text-muted-foreground">Сегодня ещё нет активности</p>
      )}
    </Card>
  )
}
