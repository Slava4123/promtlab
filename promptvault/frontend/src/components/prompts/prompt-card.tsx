import { Star, FileText, Copy, Trash2, Pin } from "lucide-react"
import type { Prompt } from "@/api/types"

interface PromptCardProps {
  prompt: Prompt
  onToggleFavorite: (id: number) => void
  onTogglePin?: (id: number, teamWide: boolean) => void
  onClick: (id: number) => void
  onUse?: (prompt: Prompt) => void
  onDelete?: (id: number) => void
  style?: React.CSSProperties
}

const modelDot: Record<string, string> = {
  "gpt": "bg-emerald-500/70",
  "claude": "bg-orange-500/70",
  "default": "bg-zinc-500/70",
}

function getModelDot(model?: string) {
  if (!model) return modelDot.default
  const m = model.toLowerCase()
  if (m.includes("gpt") || m.includes("openai")) return modelDot.gpt
  if (m.includes("claude") || m.includes("anthropic")) return modelDot.claude
  return modelDot.default
}

export function PromptCard({ prompt, onToggleFavorite, onTogglePin, onClick, onUse, onDelete, style }: PromptCardProps) {
  const isPinned = prompt.pinned_personal || prompt.pinned_team
  return (
    <div
      className={`group cursor-pointer overflow-hidden rounded-xl border p-4 transition-[transform,box-shadow] duration-200 hover:-translate-y-0.5 ${
        isPinned
          ? "border-violet-500/20 bg-violet-500/[0.02] hover:border-violet-500/30 hover:shadow-lg"
          : prompt.favorite
            ? "border-yellow-500/15 bg-card hover:border-yellow-500/25 hover:shadow-lg"
            : "border-border bg-card hover:border-violet-500/15 hover:shadow-lg"
      }`}
      onClick={() => onClick(prompt.id)}
      style={style}
    >
      {/* Header: icon + title + star */}
      <div className="mb-3 flex items-center gap-2.5">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-500/[0.08] ring-1 ring-violet-500/10">
          <FileText className="h-3.5 w-3.5 text-violet-400" />
        </div>
        <h3 className="min-w-0 flex-1 truncate text-[0.82rem] font-medium text-foreground">
          {prompt.title}
        </h3>
        {onUse && (
          <button
            type="button"
            title="Использовать промпт"
            aria-label="Использовать промпт"
            className="shrink-0 text-muted-foreground transition-opacity hover:text-violet-400 sm:opacity-0 sm:group-hover:opacity-100"
            onClick={(e) => { e.stopPropagation(); onUse(prompt) }}
          >
            <Copy className="h-3.5 w-3.5" />
          </button>
        )}
        {onDelete && (
          <button
            type="button"
            aria-label="Удалить промпт"
            className="shrink-0 text-muted-foreground transition-opacity hover:text-red-400 sm:opacity-0 sm:group-hover:opacity-100"
            onClick={(e) => { e.stopPropagation(); onDelete(prompt.id) }}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        )}
        {onTogglePin && (
          <button
            type="button"
            aria-label={isPinned ? "Открепить" : "Закрепить"}
            className={`shrink-0 transition-opacity ${
              isPinned ? "text-violet-400" : "text-muted-foreground sm:opacity-0 sm:group-hover:opacity-100 hover:text-violet-400"
            }`}
            onClick={(e) => { e.stopPropagation(); onTogglePin(prompt.id, false) }}
          >
            <Pin className={`h-3.5 w-3.5 ${isPinned ? "fill-violet-400" : ""}`} />
          </button>
        )}
        <button
          type="button"
          aria-label={prompt.favorite ? "Убрать из избранного" : "Добавить в избранное"}
          className={`shrink-0 transition-opacity ${
            prompt.favorite ? "text-yellow-500" : "text-muted-foreground sm:opacity-0 sm:group-hover:opacity-100 hover:text-yellow-400"
          }`}
          onClick={(e) => { e.stopPropagation(); onToggleFavorite(prompt.id) }}
        >
          <Star className={`h-3.5 w-3.5 ${prompt.favorite ? "fill-yellow-500" : ""}`} />
        </button>
      </div>

      {/* Content preview */}
      <p className="mb-3 line-clamp-2 text-[0.75rem] leading-relaxed text-muted-foreground">
        {prompt.content}
      </p>

      {/* Tags */}
      {prompt.tags.length > 0 && (
        <div className="mb-3 flex flex-wrap gap-1">
          {prompt.tags.map((tag) => (
            <span
              key={tag.id}
              className="rounded-full px-2 py-[2px] text-[10px] font-medium"
              style={{
                backgroundColor: (tag.color || "#8b5cf6") + "12",
                color: (tag.color || "#8b5cf6") + "cc",
              }}
            >
              {tag.name}
            </span>
          ))}
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center gap-2 text-[10px] text-muted-foreground min-w-0">
        {prompt.model && (
          <span className="flex min-w-0 items-center gap-1">
            <span className={`h-1.5 w-1.5 shrink-0 rounded-full ${getModelDot(prompt.model)}`} />
            <span className="truncate">{prompt.model}</span>
          </span>
        )}
        {prompt.usage_count > 0 && <span>· {prompt.usage_count}x</span>}
        <span className="ml-auto">
          {new Date(prompt.updated_at).toLocaleDateString("ru-RU", { day: "numeric", month: "short" })}
        </span>
      </div>
    </div>
  )
}

export function PromptCardSkeleton() {
  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="mb-3 flex items-center gap-2.5">
        <div className="h-8 w-8 animate-pulse rounded-lg bg-foreground/[0.06]" />
        <div className="h-4 flex-1 animate-pulse rounded-md bg-foreground/[0.06]" />
      </div>
      <div className="mb-2 h-3 w-full animate-pulse rounded-md bg-foreground/[0.04]" />
      <div className="mb-3 h-3 w-2/3 animate-pulse rounded-md bg-foreground/[0.04]" />
      <div className="flex gap-1">
        <div className="h-4 w-16 animate-pulse rounded-full bg-foreground/[0.04]" />
        <div className="h-4 w-12 animate-pulse rounded-full bg-foreground/[0.04]" />
      </div>
    </div>
  )
}
