import { useState, useCallback, useEffect } from "react"
import Markdown from "react-markdown"
import { Sparkles, RefreshCw, Search, FileText, ChevronDown, Square, Copy, Check, X, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { useSSE } from "@/hooks/use-sse"
import { ModelSelector } from "./model-selector"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { AIAction } from "@/api/types"

interface AIPanelProps {
  content: string
  onApply: (text: string, note: string) => void
}

const actions: { key: AIAction; label: string; icon: typeof Sparkles; description: string }[] = [
  { key: "enhance", label: "Улучшить", icon: Sparkles, description: "Сделать конкретнее и структурированнее" },
  { key: "rewrite", label: "Переписать", icon: RefreshCw, description: "Переписать в другом стиле" },
  { key: "analyze", label: "Анализ", icon: Search, description: "Оценка качества промпта" },
  { key: "variations", label: "Вариации", icon: FileText, description: "3 разных варианта" },
]

const rewriteStyles = [
  { value: "formal", label: "Формальный" },
  { value: "concise", label: "Лаконичный" },
  { value: "creative", label: "Креативный" },
  { value: "detailed", label: "Детальный" },
  { value: "technical", label: "Технический" },
]

const actionLabels: Record<AIAction, string> = {
  enhance: "Улучшено через AI",
  rewrite: "Переписано через AI",
  analyze: "Анализ через AI",
  variations: "Вариация через AI",
}

export function AIPanel({ content, onApply }: AIPanelProps) {
  const [expanded, setExpanded] = useState(false)
  const [selectedModel, setSelectedModel] = useState("")
  const [activeAction, setActiveAction] = useState<AIAction | null>(null)
  const [rewriteStyle, setRewriteStyle] = useState("concise")
  const { data, isStreaming, error, start, abort } = useSSE()
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    if (!isStreaming) { setElapsed(0); return }
    const id = setInterval(() => setElapsed((e) => e + 1), 1000)
    return () => clearInterval(id)
  }, [isStreaming])

  const handleAction = useCallback((action: AIAction) => {
    setActiveAction(action)
    const body: Record<string, unknown> = {
      content,
      model: selectedModel,
    }
    if (action === "rewrite") {
      body.style = rewriteStyle
    }
    if (action === "variations") {
      body.count = 3
    }
    start(`/ai/${action}`, body)
  }, [content, selectedModel, rewriteStyle, start])

  const handleApply = () => {
    if (data && activeAction) {
      onApply(data, actionLabels[activeAction])
      toast.success("Результат применён")
    }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(data)
      toast.success("Скопировано")
    } catch {
      toast.error("Не удалось скопировать")
    }
  }

  const hasContent = content.trim().length > 0
  const hasResult = data.length > 0

  return (
    <div className="rounded-xl overflow-hidden" style={{ border: "1px solid rgba(139,92,246,0.1)", background: "rgba(139,92,246,0.02)" }}>
      {/* Toggle bar */}
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2.5 px-4 py-3 text-left transition-colors hover:bg-violet-500/[0.04]"
      >
        <Sparkles className="h-3.5 w-3.5 text-violet-400" />
        <span className="text-[0.82rem] font-medium text-violet-300">AI-ассистент</span>
        <ChevronDown className={`ml-auto h-3.5 w-3.5 text-zinc-600 transition-transform ${expanded ? "rotate-180" : ""}`} />
      </button>

      {expanded && (
        <div className="space-y-4 px-4 pb-4">
          {/* Auto-select model (hidden, single model) */}
          <div className="hidden">
            <ModelSelector value={selectedModel} onChange={setSelectedModel} />
          </div>

          {/* Rewrite style selector */}
          {activeAction === "rewrite" && isStreaming && (
            <div className="flex flex-wrap gap-1.5">
              {rewriteStyles.map((s) => (
                <span
                  key={s.value}
                  className="rounded-md px-2 py-0.5 text-[0.72rem] font-medium text-violet-300"
                  style={{ background: "rgba(139,92,246,0.12)" }}
                >
                  {s.label}
                </span>
              ))}
            </div>
          )}

          {/* Action buttons */}
          <div className="flex flex-wrap gap-2">
            {actions.map(({ key, label, icon: Icon }) => {
              const isActive = activeAction === key && isStreaming

              if (key === "rewrite" && !isStreaming) {
                // Show rewrite with style dropdown
                return (
                  <div key={key} className="flex items-center gap-1">
                    <button
                      type="button"
                      disabled={!hasContent || !selectedModel || (isStreaming && activeAction !== key)}
                      onClick={() => handleAction("rewrite")}
                      className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-[0.78rem] font-medium transition-all disabled:opacity-30"
                      style={{
                        border: "1px solid rgba(255,255,255,0.08)",
                        background: isActive ? "rgba(139,92,246,0.15)" : "rgba(255,255,255,0.03)",
                        color: isActive ? "#a78bfa" : "#a1a1aa",
                      }}
                    >
                      <Icon className="h-3.5 w-3.5" />
                      {label}
                    </button>
                    <Select value={rewriteStyle} onValueChange={setRewriteStyle} modal={false}>
                      <SelectTrigger size="sm" className="h-[30px] text-[0.72rem] text-zinc-400">
                        <SelectValue>
                          {rewriteStyles.find((s) => s.value === rewriteStyle)?.label}
                        </SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        {rewriteStyles.map((s) => (
                          <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                )
              }

              return (
                <button
                  key={key}
                  type="button"
                  disabled={!hasContent || !selectedModel || (isStreaming && activeAction !== key)}
                  onClick={() => isActive ? abort() : handleAction(key)}
                  className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-[0.78rem] font-medium transition-all disabled:opacity-30"
                  style={{
                    border: `1px solid ${isActive ? "rgba(139,92,246,0.3)" : "rgba(255,255,255,0.08)"}`,
                    background: isActive ? "rgba(139,92,246,0.15)" : "rgba(255,255,255,0.03)",
                    color: isActive ? "#a78bfa" : "#a1a1aa",
                  }}
                >
                  {isActive ? (
                    <>
                      <Square className="h-3 w-3" />
                      Остановить
                    </>
                  ) : (
                    <>
                      <Icon className="h-3.5 w-3.5" />
                      {label}
                    </>
                  )}
                </button>
              )
            })}
          </div>

          {/* Error */}
          {error && (
            <div
              className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-[0.78rem] text-red-300"
              style={{ border: "1px solid rgba(239,68,68,0.15)", background: "rgba(239,68,68,0.06)" }}
            >
              <X className="h-3.5 w-3.5 flex-shrink-0" />
              <span className="flex-1">{error}</span>
              <button
                type="button"
                onClick={() => activeAction && handleAction(activeAction)}
                className="text-red-400 hover:text-red-300 text-[0.75rem] font-medium"
              >
                Повторить
              </button>
            </div>
          )}

          {/* Result area */}
          {(hasResult || isStreaming) && (
            <div
              className="relative min-h-[120px] max-h-[400px] overflow-y-auto rounded-lg px-3.5 py-3 text-[0.82rem] leading-relaxed text-zinc-200"
              style={{ border: "1px solid rgba(255,255,255,0.05)", background: "rgba(255,255,255,0.02)" }}
            >
              {/* Skeleton + bouncing dots while waiting for first token */}
              {isStreaming && !data && (
                <div className="space-y-3 py-1">
                  <div className="h-3 w-4/5 rounded bg-white/[0.04] animate-pulse" />
                  <div className="h-3 w-3/5 rounded bg-white/[0.04] animate-pulse" />
                  <div className="h-3 w-full rounded bg-white/[0.04] animate-pulse" />
                  <div className="h-3 w-2/3 rounded bg-white/[0.04] animate-pulse" />
                  <div className="flex items-center gap-1.5 pt-2">
                    <span className="w-1.5 h-1.5 bg-violet-400 rounded-full animate-bounce" style={{ animationDelay: "0s" }} />
                    <span className="w-1.5 h-1.5 bg-violet-400 rounded-full animate-bounce" style={{ animationDelay: "0.15s" }} />
                    <span className="w-1.5 h-1.5 bg-violet-400 rounded-full animate-bounce" style={{ animationDelay: "0.3s" }} />
                  </div>
                </div>
              )}
              {/* Streaming content */}
              {data && (
                <div className="ai-markdown prose prose-invert prose-sm max-w-none">
                  <Markdown>{data}</Markdown>
                  {isStreaming && <span className="inline-block w-1.5 h-4 ml-0.5 bg-violet-400 animate-pulse rounded-sm" />}
                </div>
              )}
              {/* Spinner + elapsed timer */}
              {isStreaming && (
                <div className="absolute top-2 right-2 flex items-center gap-1.5">
                  <span className="text-[0.7rem] text-zinc-500 tabular-nums">{elapsed} сек</span>
                  <Loader2 className="h-3.5 w-3.5 animate-spin text-violet-400/60" />
                </div>
              )}
            </div>
          )}

          {/* Action bar below result */}
          {hasResult && !isStreaming && (
            <div className="flex items-center gap-2">
              {activeAction !== "analyze" && (
                <button
                  type="button"
                  onClick={handleApply}
                  className="flex items-center gap-1.5 rounded-lg px-4 py-1.5 text-[0.78rem] font-medium text-white transition-all active:scale-[0.97]"
                  style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 2px 8px -1px rgba(124,58,237,0.25)" }}
                >
                  <Check className="h-3.5 w-3.5" />
                  Применить
                </button>
              )}
              <button
                type="button"
                onClick={handleCopy}
                className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-[0.78rem] text-zinc-400 transition-all hover:text-zinc-200"
                style={{ border: "1px solid rgba(255,255,255,0.08)", background: "rgba(255,255,255,0.03)" }}
              >
                <Copy className="h-3.5 w-3.5" />
                Копировать
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
