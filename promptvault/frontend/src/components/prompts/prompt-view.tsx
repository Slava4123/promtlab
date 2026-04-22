import { useEffect, useState } from "react"
import { Eye, Code2 } from "lucide-react"
import { PromptContent } from "./prompt-content"
import { MarkdownEditor } from "./markdown-editor"
import { cn } from "@/lib/utils"

type Mode = "rendered" | "source"

interface PromptViewProps {
  content: string
  className?: string
  /** Persist-ключ для localStorage (уникальный per use-case: share, public, dialog). */
  storageKey?: string
  defaultMode?: Mode
}

interface ToggleButtonProps {
  active: boolean
  onClick: () => void
  icon: React.ComponentType<{ className?: string }>
  label: string
}

function ToggleButton({ active, onClick, icon: Icon, label }: ToggleButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[0.75rem] font-medium transition-colors",
        active
          ? "bg-background text-foreground shadow-sm"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      {label}
    </button>
  )
}

/**
 * PromptView — отображение готового промпта читателю с переключателем
 * «Рендер ↔ Исходник». Используется на публичных страницах (shared, public)
 * и в preview-диалогах.
 *
 * В режиме «Рендер» показывает PromptContent (Markdown + GFM + highlight).
 * В режиме «Исходник» — тот же CodeMirror что в редакторе, но read-only.
 * Это даёт читателю визуальную подсветку markdown-разметки (## H1, таблицы,
 * code-блоки с nested-языками) без возможности редактировать.
 */
export function PromptView({
  content,
  className,
  storageKey,
  defaultMode = "rendered",
}: PromptViewProps) {
  const [mode, setMode] = useState<Mode>(() => {
    if (typeof window === "undefined" || !storageKey) return defaultMode
    const stored = window.localStorage.getItem(storageKey)
    return stored === "source" || stored === "rendered" ? stored : defaultMode
  })

  useEffect(() => {
    if (!storageKey || typeof window === "undefined") return
    window.localStorage.setItem(storageKey, mode)
  }, [mode, storageKey])

  return (
    <div className={cn("space-y-3", className)}>
      <div
        role="tablist"
        aria-label="Режим просмотра промпта"
        className="inline-flex rounded-lg border border-border bg-muted/30 p-0.5"
      >
        <ToggleButton
          active={mode === "rendered"}
          onClick={() => setMode("rendered")}
          icon={Eye}
          label="Рендер"
        />
        <ToggleButton
          active={mode === "source"}
          onClick={() => setMode("source")}
          icon={Code2}
          label="Исходник"
        />
      </div>

      <div role="tabpanel" aria-label={mode === "rendered" ? "Рендер" : "Исходник"}>
        {mode === "rendered" ? (
          <PromptContent content={content} />
        ) : (
          <MarkdownEditor value={content} readOnly minHeight="200px" />
        )}
      </div>
    </div>
  )
}
