import { Star, FileText, Copy } from "lucide-react"
import type { Prompt } from "@/api/types"

interface PromptCardProps {
  prompt: Prompt
  onToggleFavorite: (id: number) => void
  onClick: (id: number) => void
  onUse?: (prompt: Prompt) => void
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

export function PromptCard({ prompt, onToggleFavorite, onClick, onUse, style }: PromptCardProps) {
  return (
    <div
      className={`group cursor-pointer rounded-xl border p-4 transition-all duration-200 hover:-translate-y-0.5 ${
        prompt.favorite
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
      <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
        {prompt.model && (
          <span className="flex items-center gap-1">
            <span className={`h-1.5 w-1.5 rounded-full ${getModelDot(prompt.model)}`} />
            {prompt.model}
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
        <div className="h-8 w-8 animate-pulse rounded-lg bg-white/[0.04]" />
        <div className="h-4 flex-1 animate-pulse rounded-md bg-white/[0.04]" />
      </div>
      <div className="mb-2 h-3 w-full animate-pulse rounded-md bg-white/[0.03]" />
      <div className="mb-3 h-3 w-2/3 animate-pulse rounded-md bg-white/[0.03]" />
      <div className="flex gap-1">
        <div className="h-4 w-16 animate-pulse rounded-full bg-white/[0.03]" />
        <div className="h-4 w-12 animate-pulse rounded-full bg-white/[0.03]" />
      </div>
    </div>
  )
}
