import { useNavigate } from "react-router-dom"
import { ArrowLeft, Lock, Award, Loader2 } from "lucide-react"
import { Button } from "../components/ui/button"
import { useBadges } from "../hooks/use-badges"
import { useStreakDetail } from "../hooks/use-streak-detail"
import type { Badge, BadgeCategory } from "../lib/types"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import { cn } from "../lib/utils"

const CATEGORY_LABELS: Record<BadgeCategory, string> = {
  personal: "Личные",
  team: "Командные",
  milestone: "Этапы",
  streak: "Серии",
}

export function BadgesPage() {
  const navigate = useNavigate()
  const badgesQuery = useBadges()
  const streakQuery = useStreakDetail()

  if (badgesQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const grouped = (badgesQuery.data?.items ?? []).reduce<Record<BadgeCategory, Badge[]>>(
    (acc, b) => {
      const cat = b.category
      ;(acc[cat] ??= []).push(b)
      return acc
    },
    { personal: [], team: [], milestone: [], streak: [] },
  )

  const streak = streakQuery.data

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Достижения</h2>
        <span className="text-[10px] text-(--color-muted-foreground)">
          {badgesQuery.data?.total_unlocked ?? 0} / {badgesQuery.data?.total_count ?? 0}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {/* Streak summary */}
        {streak && (
          <section className="rounded-md border border-(--color-border) bg-gradient-to-br from-orange-500/10 to-red-500/5 p-3">
            <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
              Серия активности
            </div>
            <div className="mt-1 flex items-baseline gap-3">
              <div>
                <div className="text-2xl font-bold text-orange-500">🔥 {streak.current_streak}</div>
                <div className="text-[10px] text-(--color-muted-foreground)">Текущая</div>
              </div>
              <div>
                <div className="text-sm font-semibold">{streak.longest_streak}</div>
                <div className="text-[10px] text-(--color-muted-foreground)">Лучшая</div>
              </div>
              <div className="ml-auto text-[10px] text-(--color-muted-foreground)">
                {streak.active_today ? "Сегодня активен ✓" : "Сегодня ещё нет активности"}
              </div>
            </div>
          </section>
        )}

        {/* Badge groups */}
        {(Object.keys(CATEGORY_LABELS) as BadgeCategory[]).map((cat) => {
          const items = grouped[cat]
          if (!items || items.length === 0) return null
          const unlocked = items.filter((b) => b.unlocked).length
          return (
            <section key={cat}>
              <div className="mb-2 flex items-center justify-between">
                <h3 className="text-xs font-semibold">{CATEGORY_LABELS[cat]}</h3>
                <span className="text-[10px] text-(--color-muted-foreground)">
                  {unlocked} / {items.length}
                </span>
              </div>
              <ul className="grid grid-cols-2 gap-2">
                {items.map((b) => (
                  <BadgeCard key={b.id} badge={b} />
                ))}
              </ul>
            </section>
          )
        })}

        {badgesQuery.data?.items.length === 0 && (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <Award className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Достижений пока нет</p>
          </div>
        )}
      </div>
    </div>
  )
}

function BadgeCard({ badge }: { badge: Badge }) {
  const progress = badge.target > 0 ? Math.min(100, (badge.progress / badge.target) * 100) : 0
  return (
    <li
      className={cn(
        "rounded-md border p-2.5",
        badge.unlocked
          ? "border-amber-500/50 bg-amber-500/5"
          : "border-(--color-border) bg-(--color-card) opacity-70",
      )}
    >
      <div className="flex items-start gap-2">
        <span className={cn("text-2xl", !badge.unlocked && "grayscale")}>{badge.icon || "🏆"}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1 text-xs font-medium">
            {/* title= даёт tooltip с полным именем при hover —
                карточка в 2-колоночном grid обрезает «Коллекционер», «Командный»
                и подобные длинные названия. */}
            <span className="truncate" title={badge.title}>{badge.title}</span>
            {!badge.unlocked && <Lock className="h-2.5 w-2.5 text-(--color-muted-foreground)" />}
          </div>
          <p className="mt-0.5 line-clamp-2 text-[10px] text-(--color-muted-foreground)">
            {badge.description}
          </p>
        </div>
      </div>
      {!badge.unlocked && badge.target > 0 && (
        <div className="mt-2">
          <div className="h-1 overflow-hidden rounded-full bg-(--color-muted)">
            <div
              className="h-full bg-(--color-brand) transition-all duration-(--duration-normal) ease-(--ease-out)"
              style={{ width: `${progress}%` }}
            />
          </div>
          <div className="mt-0.5 text-right text-[9px] text-(--color-muted-foreground)">
            {badge.progress} / {badge.target}
          </div>
        </div>
      )}
      {badge.unlocked && badge.unlocked_at && (
        <div className="mt-1.5 text-[9px] text-amber-500/70">
          Получено {formatRelativeDate(badge.unlocked_at)}
        </div>
      )}
    </li>
  )
}
