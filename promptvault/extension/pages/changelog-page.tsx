import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, BookOpen, Loader2 } from "lucide-react"
import { Button } from "../components/ui/button"
import { useChangelog, useMarkChangelogRead } from "../hooks/use-changelog"
import { formatDate } from "@pv/shared/utils/format-date"

const CATEGORY_COLORS: Record<string, string> = {
  feature: "text-emerald-500 bg-emerald-500/10",
  fix: "text-amber-500 bg-amber-500/10",
  improvement: "text-blue-500 bg-blue-500/10",
  security: "text-red-500 bg-red-500/10",
}

const CATEGORY_LABELS: Record<string, string> = {
  feature: "Новое",
  fix: "Исправление",
  improvement: "Улучшение",
  security: "Безопасность",
}

export function ChangelogPage() {
  const navigate = useNavigate()
  const changelogQuery = useChangelog()
  const markRead = useMarkChangelogRead()

  // Помечаем как прочитанное при открытии страницы.
  useEffect(() => {
    if (changelogQuery.data?.has_unread) {
      void markRead.mutateAsync()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [changelogQuery.data?.has_unread])

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Что нового</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {changelogQuery.isPending ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (changelogQuery.data?.entries ?? []).length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <BookOpen className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Пока нет записей</p>
          </div>
        ) : (
          <ul className="space-y-3">
            {(changelogQuery.data?.entries ?? []).map((entry, i) => (
              <li
                key={`${entry.version}-${i}`}
                className="rounded-md border border-(--color-border) bg-(--color-card) p-3"
              >
                <div className="mb-1.5 flex items-center gap-2">
                  <span className="rounded bg-(--color-muted) px-1.5 py-0.5 font-mono text-[10px]">
                    v{entry.version}
                  </span>
                  <span
                    className={
                      "rounded px-1.5 py-0.5 text-[10px] " +
                      (CATEGORY_COLORS[entry.category] ?? "bg-(--color-muted)")
                    }
                  >
                    {CATEGORY_LABELS[entry.category] ?? entry.category}
                  </span>
                  <span className="ml-auto text-[10px] text-(--color-muted-foreground)">
                    {formatDate(entry.date)}
                  </span>
                </div>
                <h3 className="text-sm font-medium">{entry.title}</h3>
                <p className="mt-1 text-xs text-(--color-muted-foreground)">{entry.description}</p>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
