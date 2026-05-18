import { TrendingUp } from "lucide-react"

import { InsightPromptRow } from "@/components/prompts/insights/insight-prompt-row"
import { PageLayout } from "@/components/layout/page-layout"
import { useTrending } from "@/hooks/use-prompt-insights"

// Страница "Растущие промпты" — список промптов, у которых использование
// выросло >=2x за неделю. Read-only листинг (без per-row actions): на этой
// странице делать нечего, кроме как кликнуть в промпт и открыть его.
// Иконка TrendingUp используется в insights-panel/sidebar (F10), здесь
// рендерится визуальный акцент у заголовка.
export default function TrendingPage() {
  const { data, isLoading, isError } = useTrending()

  return (
    <PageLayout
      title="Растущие промпты"
      description="Использование выросло ≥2× за неделю."
      maxWidth="md"
      action={<TrendingUp className="size-5 text-emerald-500" aria-hidden />}
    >
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && (
        <p className="text-sm text-destructive">Не удалось загрузить список.</p>
      )}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Пока нет растущих промптов.
        </p>
      )}

      {data && data.length > 0 && (
        <ul className="space-y-2">
          {data.map((p) => (
            <li key={p.prompt_id}>
              <InsightPromptRow
                promptID={p.prompt_id}
                title={p.title}
                uses={p.uses}
              />
            </li>
          ))}
        </ul>
      )}
    </PageLayout>
  )
}
