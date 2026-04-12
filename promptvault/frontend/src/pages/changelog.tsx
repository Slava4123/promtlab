import { useEffect } from "react"
import { Sparkles } from "lucide-react"

import { useChangelog, useMarkChangelogSeen } from "@/hooks/use-changelog"

const categoryColors: Record<string, string> = {
  feature: "bg-violet-500/10 text-violet-400 border-violet-500/20",
  fix: "bg-green-500/10 text-green-400 border-green-500/20",
  improvement: "bg-blue-500/10 text-blue-400 border-blue-500/20",
}

export default function Changelog() {
  const { data, isLoading } = useChangelog()
  const markSeen = useMarkChangelogSeen()

  useEffect(() => {
    if (data?.has_unread) {
      markSeen.mutate(undefined, {
        onError: (err) => {
          console.warn("Failed to mark changelog as seen:", err)
        },
      })
    }
  }, [data?.has_unread]) // eslint-disable-line react-hooks/exhaustive-deps

  if (isLoading) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-8">
        <div className="space-y-6">
          {[1, 2, 3].map((i) => (
            <div key={i} className="animate-pulse space-y-2">
              <div className="h-4 w-24 rounded bg-muted" />
              <div className="h-6 w-64 rounded bg-muted" />
              <div className="h-16 w-full rounded bg-muted" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <div className="mb-8 flex items-center gap-2">
        <Sparkles className="h-5 w-5 text-violet-400" />
        <h1 className="text-xl font-bold">Что нового</h1>
      </div>

      <div className="space-y-8">
        {data?.entries.map((entry, i) => (
          <article key={i} className="relative border-l-2 border-border pl-6">
            <span className="absolute -left-[5px] top-1.5 h-2 w-2 rounded-full bg-violet-500" />
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <time>{entry.date}</time>
              <span>v{entry.version}</span>
              <span
                className={`rounded-full border px-2 py-0.5 text-[0.65rem] font-medium ${
                  categoryColors[entry.category] ?? "bg-muted text-muted-foreground border-border"
                }`}
              >
                {entry.category === "feature" ? "новое" : entry.category === "fix" ? "исправление" : "улучшение"}
              </span>
            </div>
            <h2 className="mt-1 text-base font-semibold">{entry.title}</h2>
            <p className="mt-1.5 text-sm text-muted-foreground leading-relaxed">{entry.description}</p>
          </article>
        ))}
      </div>

      {(!data?.entries || data.entries.length === 0) && (
        <p className="text-center text-muted-foreground">Пока нет обновлений</p>
      )}
    </div>
  )
}
