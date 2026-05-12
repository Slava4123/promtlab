import { Activity, Clock, ExternalLink, TrendingUp } from "lucide-react"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import { useSettings } from "../../hooks/use-settings"
import { openWebPage } from "../../lib/utils"
import type { Prompt } from "../../lib/types"

interface PromptStatsProps {
  prompt: Prompt
}

// Компактная карточка с метриками промпта на странице detail.
// Полная аналитика (графики по дням, model segmentation) — в веб-приложении
// через deep-link, потому что requires server-side aggregation.
export function PromptStats({ prompt }: PromptStatsProps) {
  const settings = useSettings()

  function openFullAnalytics() {
    if (!settings) return
    openWebPage(
      settings.apiBase,
      `/prompts/${prompt.id}/analytics?from=extension`,
    )
  }

  const isFresh = prompt.usage_count === 0

  return (
    <section className="rounded-md border border-(--color-border) bg-(--color-card)">
      <header className="flex items-center justify-between border-b border-(--color-border) px-3 py-1.5">
        <div className="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
          <Activity className="h-3 w-3" />
          Аналитика
        </div>
        {settings && (
          <button
            type="button"
            onClick={openFullAnalytics}
            className="flex items-center gap-1 text-[10px] text-(--color-primary) hover:underline"
            title="Открыть полную аналитику в веб-приложении"
          >
            Подробнее
            <ExternalLink className="h-2.5 w-2.5" />
          </button>
        )}
      </header>
      <div className="grid grid-cols-3 divide-x divide-(--color-border) text-center">
        <Metric
          icon={TrendingUp}
          value={prompt.usage_count}
          label="использований"
        />
        <Metric
          icon={Clock}
          value={prompt.last_used_at ? formatRelativeDate(prompt.last_used_at) : "—"}
          label={isFresh ? "ещё не использован" : "последний раз"}
          small
        />
        <Metric
          icon={Clock}
          value={formatRelativeDate(prompt.created_at)}
          label="создан"
          small
        />
      </div>
    </section>
  )
}

interface MetricProps {
  icon: React.ComponentType<{ className?: string }>
  value: string | number
  label: string
  small?: boolean
}

function Metric({ icon: Icon, value, label, small }: MetricProps) {
  return (
    <div className="flex flex-col items-center gap-0.5 py-2">
      <Icon className="h-3 w-3 text-(--color-muted-foreground)" />
      <div className={small ? "text-[11px] font-medium" : "text-sm font-semibold"}>
        {value}
      </div>
      <div className="text-[9px] text-(--color-muted-foreground)">{label}</div>
    </div>
  )
}
