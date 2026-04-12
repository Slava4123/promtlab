import { FileText, Star, Search } from "lucide-react"
import { mockPrompts } from "../data/landing-content"

export function MockupPromptList() {
  return (
    <div>
      <div className="flex items-center gap-3 border-b border-border/30 pb-3">
        <Search className="h-4 w-4 text-muted-foreground/60" />
        <span className="text-sm text-muted-foreground/50">Поиск промптов...</span>
        <span className="ml-auto rounded-md border border-border/30 px-1.5 py-0.5 text-[0.65rem] text-muted-foreground/30">
          ⌘K
        </span>
      </div>
      <div className="mt-3 space-y-2">
        {mockPrompts.slice(0, 3).map(p => (
          <div
            key={p.title}
            className="flex items-center gap-3 rounded-lg border border-border/20 bg-card/30 px-3 py-2.5 transition-colors hover:border-violet-500/20"
          >
            <FileText className="h-3.5 w-3.5 text-violet-400/60" />
            <span className="text-xs">{p.title}</span>
            <div className="ml-auto flex items-center gap-1.5">
              {p.tags.slice(0, 1).map(t => (
                <span key={t} className="rounded bg-violet-500/10 px-1.5 py-0.5 text-[0.55rem] text-violet-300/70">
                  {t}
                </span>
              ))}
              {p.fav && <Star className="h-3 w-3 text-amber-400/50" />}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
