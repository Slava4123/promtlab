import type { Badge } from "@/api/types"
import { cn } from "@/lib/utils"

interface BadgeCardProps {
  badge: Badge
}

function formatUnlockedDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString("ru-RU", {
      day: "numeric",
      month: "long",
      year: "numeric",
    })
  } catch {
    return ""
  }
}

export function BadgeCard({ badge }: BadgeCardProps) {
  const { unlocked, unlocked_at, progress, target, icon, title, description } = badge
  const percent = target > 0 ? Math.min(100, Math.round((progress / target) * 100)) : 0

  return (
    <div
      className={cn(
        "flex flex-col gap-2 rounded-xl border p-4 transition-colors",
        unlocked
          ? "border-violet-500/30 bg-violet-500/5"
          : "border-border bg-muted/20",
      )}
    >
      <div className="flex items-start gap-3">
        <div
          className={cn(
            "flex h-11 w-11 shrink-0 items-center justify-center rounded-lg text-2xl",
            unlocked ? "bg-violet-500/15 grayscale-0" : "bg-muted/40 grayscale opacity-50",
          )}
          aria-hidden
        >
          {icon}
        </div>
        <div className="min-w-0 flex-1">
          <h3
            className={cn(
              "text-[0.9rem] font-semibold leading-tight",
              unlocked ? "text-foreground" : "text-muted-foreground",
            )}
          >
            {title}
          </h3>
          <p className="mt-0.5 text-[0.75rem] text-muted-foreground line-clamp-2">
            {description}
          </p>
        </div>
      </div>

      <div className="mt-1">
        {unlocked && unlocked_at ? (
          <p className="text-[0.7rem] text-violet-400">
            Разблокировано {formatUnlockedDate(unlocked_at)}
          </p>
        ) : (
          <>
            <div className="flex items-center justify-between text-[0.7rem] text-muted-foreground">
              <span>Прогресс</span>
              <span className="tabular-nums">
                {progress}/{target}
              </span>
            </div>
            <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-muted/40">
              <div
                className="h-full bg-violet-500/50 transition-[width]"
                style={{ width: `${percent}%` }}
              />
            </div>
          </>
        )}
      </div>
    </div>
  )
}
