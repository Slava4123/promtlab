import { Archive } from "lucide-react"

import { InsightPromptRow } from "@/components/prompts/insights/insight-prompt-row"
import { PageLayout } from "@/components/layout/page-layout"
import { useMostEdited } from "@/hooks/use-prompt-insights"

// Страница "Часто правят" — список промптов с большим числом версий.
// Read-only (тот же паттерн, что и trending.tsx / declining.tsx).
export default function MostEditedPage() {
  const { data, isLoading, isError } = useMostEdited()

  return (
    <PageLayout
      title="Часто правят"
      description="Промпты с большим числом версий"
      maxWidth="md"
      action={<Archive className="size-5 text-blue-500" aria-hidden />}
    >
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && (
        <p className="text-sm text-destructive">Не удалось загрузить список.</p>
      )}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Нет частоправленных промптов.
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
