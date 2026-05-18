import { Link } from "react-router-dom"
import { pluralizeRu } from "@/lib/pluralize"

// InsightPromptRow — единая строка для всех insight-листингов (unused / trending /
// declining / most-edited). Один shape данных (PromptInsightRow в api/prompt-insights.ts) →
// один компонент, чтобы оформление и пагинация были идентичны на всех страницах.
interface InsightPromptRowProps {
  promptID: number
  title: string
  uses: number
  // showUses скрывает подпись "N использований" (например, для unused — там uses всегда 0
  // и подпись бесполезна шумом).
  showUses?: boolean
  // actions — слот справа под per-row кнопки (Удалить / Открыть / Merge-with…).
  actions?: React.ReactNode
}

export function InsightPromptRow({
  promptID,
  title,
  uses,
  showUses = true,
  actions,
}: InsightPromptRowProps) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md border px-3 py-2">
      <div className="min-w-0 flex-1">
        <Link
          to={`/prompts/${promptID}`}
          className="block truncate text-sm font-medium hover:underline"
        >
          {title}
        </Link>
        {showUses && (
          <p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
            {uses}{" "}
            {pluralizeRu(uses, "использование", "использования", "использований")}
          </p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}
