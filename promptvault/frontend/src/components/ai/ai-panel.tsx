import { useState, useCallback, useEffect } from "react"
import Markdown from "react-markdown"
import { Sparkles, RefreshCw, Search, FileText, ChevronDown, Square, Copy, Check, X, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { useSSE } from "@/hooks/use-sse"
import { ModelSelector } from "./model-selector"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { AIAction } from "@/api/types"
import { DismissibleBanner } from "@/components/hints/dismissible-banner"

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
    if (!isStreaming) return
    const id = setInterval(() => setElapsed((e) => e + 1), 1000)
    return () => clearInterval(id)
  }, [isStreaming])

  const handleAction = useCallback((action: AIAction) => {
    setActiveAction(action)
    setElapsed(0)
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
    <div className="rounded-xl overflow-hidden border border-brand/10 bg-brand/[0.02]">
      {/* Toggle bar */}
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2.5 px-4 py-3 text-left transition-colors hover:bg-brand/[0.04]"
      >
        <Sparkles className="h-3.5 w-3.5 text-brand-muted-foreground" />
        <span className="text-[0.82rem] font-medium text-brand-muted-foreground">AI-ассистент</span>
        <ChevronDown className={`ml-auto h-3.5 w-3.5 text-muted-foreground transition-transform ${expanded ? "rotate-180" : ""}`} />
      </button>

      {expanded && (
        <div className="space-y-4 px-4 pb-4">
          {/* M-13: AI feature hint — показываем только при первом раскрытии панели */}
          <DismissibleBanner
            id="ai_button"
            title="4 режима AI"
            description="Улучшить (конкретика + структура), Переписать (под стиль), Анализ (оценка промпта), Вариации (3 разных версии)."
            tone="violet"
          />

          {/* Auto-select model (hidden, single model) */}
          <div className="hidden">
            <ModelSelector value={selectedModel} onChange={(v) => setSelectedModel(v)} />
          </div>

          {/* Rewrite style selector */}
          {activeAction === "rewrite" && isStreaming && (
            <div className="flex flex-wrap gap-1.5">
              {rewriteStyles.map((s) => (
                <span
                  key={s.value}
                  className="rounded-md px-2 py-0.5 text-[0.72rem] font-medium text-brand-muted-foreground bg-brand/[0.12]"
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
                      className={`flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-[0.78rem] font-medium transition-colors disabled:opacity-30 ${
                        isActive
                          ? "border-brand/30 bg-brand/15 text-brand-muted-foreground"
                          : "border-foreground/[0.08] bg-foreground/[0.03] text-muted-foreground"
                      }`}
                    >
                      <Icon className="h-3.5 w-3.5" />
                      {label}
                    </button>
                    <Select value={rewriteStyle} onValueChange={(v) => { if (v) setRewriteStyle(v) }} modal={false}>
                      <SelectTrigger size="sm" className="h-[30px] text-[0.72rem] text-muted-foreground">
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
                  className={`flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-[0.78rem] font-medium transition-colors disabled:opacity-30 ${
                    isActive
                      ? "border-brand/30 bg-brand/15 text-brand-muted-foreground"
                      : "border-foreground/[0.08] bg-foreground/[0.03] text-muted-foreground"
                  }`}
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
              className="flex items-center gap-2 rounded-lg border border-destructive/15 bg-destructive/[0.06] px-3 py-2.5 text-[0.78rem] text-red-300"
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
              className="relative min-h-[120px] max-h-[400px] overflow-y-auto rounded-lg border border-foreground/[0.05] bg-foreground/[0.02] px-3.5 py-3 text-[0.82rem] leading-relaxed text-foreground"
              aria-live="polite"
            >
              {/* Skeleton + bouncing dots while waiting for first token */}
              {isStreaming && !data && (
                <div className="space-y-3 py-1">
                  <div className="h-3 w-4/5 rounded bg-foreground/[0.06] animate-pulse" />
                  <div className="h-3 w-3/5 rounded bg-foreground/[0.06] animate-pulse" />
                  <div className="h-3 w-full rounded bg-foreground/[0.06] animate-pulse" />
                  <div className="h-3 w-2/3 rounded bg-foreground/[0.06] animate-pulse" />
                  <div className="flex items-center gap-1.5 pt-2">
                    <span className="w-1.5 h-1.5 bg-brand rounded-full animate-bounce" style={{ animationDelay: "0s" }} />
                    <span className="w-1.5 h-1.5 bg-brand rounded-full animate-bounce" style={{ animationDelay: "0.15s" }} />
                    <span className="w-1.5 h-1.5 bg-brand rounded-full animate-bounce" style={{ animationDelay: "0.3s" }} />
                  </div>
                </div>
              )}
              {/* Streaming content */}
              {data && (
                <div className="ai-markdown prose prose-invert prose-sm max-w-none">
                  <Markdown>{data}</Markdown>
                  {isStreaming && <span className="inline-block w-1.5 h-4 ml-0.5 bg-brand animate-pulse rounded-sm" />}
                </div>
              )}
              {/* Spinner + elapsed timer */}
              {isStreaming && (
                <div className="absolute top-2 right-2 flex items-center gap-1.5">
                  <span className="text-[0.7rem] text-muted-foreground tabular-nums">{elapsed} сек</span>
                  <Loader2 className="h-3.5 w-3.5 animate-spin text-brand-muted-foreground/60" />
                </div>
              )}
            </div>
          )}

          {/* Action bar below result */}
          {hasResult && !isStreaming && (
            <div className="flex items-center gap-2">
              {activeAction !== "analyze" && (
                <Button
                  variant="brand"
                  size="sm"
                  onClick={handleApply}
                  className="gap-1.5"
                >
                  <Check className="h-3.5 w-3.5" />
                  Применить
                </Button>
              )}
              <button
                type="button"
                onClick={handleCopy}
                className="flex items-center gap-1.5 rounded-lg border border-foreground/[0.08] bg-foreground/[0.03] px-3 py-1.5 text-[0.78rem] text-muted-foreground transition-colors hover:text-foreground"
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
