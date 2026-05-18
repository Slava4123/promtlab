import { TrendingDown } from "lucide-react"

import { InsightPromptRow } from "@/components/prompts/insights/insight-prompt-row"
import { PageLayout } from "@/components/layout/page-layout"
import { useDeclining } from "@/hooks/use-prompt-insights"

// Страница "Падающие промпты" — список промптов, у которых использование
// снизилось >=2x за неделю. Read-only (тот же паттерн, что и trending.tsx).
export default function DecliningPage() {
  const { data, isLoading, isError } = useDeclining()

  return (
    <PageLayout
      title="Падающие промпты"
      description="Использование снизилось ≥2× за неделю."
      maxWidth="md"
      action={<TrendingDown className="size-5 text-amber-500" aria-hidden />}
    >
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && (
        <p className="text-sm text-destructive">Не удалось загрузить список.</p>
      )}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Пока нет падающих промптов.
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
