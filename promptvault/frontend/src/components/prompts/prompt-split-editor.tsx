import { useEffect, useState } from "react"
import { Eye, Pencil, Columns } from "lucide-react"
import { MarkdownEditor } from "./markdown-editor"
import { PromptContent } from "./prompt-content"
import { cn } from "@/lib/utils"

type Mode = "editor" | "both" | "preview"

interface PromptSplitEditorProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  maxLength?: number
  className?: string
  id?: string
  "aria-invalid"?: boolean
  "aria-describedby"?: string
}

const STORAGE_KEY = "prompt-editor-mode"

// Breakpoint для переключения split → tab. В split-режиме на десктопе
// показываем editor + preview рядом; на меньших экранах split физически не
// помещается — переключаемся на tab (editor ИЛИ preview).
const MIN_SPLIT_WIDTH = 1024

function getInitialMode(): Mode {
  if (typeof window === "undefined") return "both"
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored === "editor" || stored === "both" || stored === "preview") {
    // На узком экране даже если в localStorage "both" — форсим editor,
    // потому что split не влезет.
    if (stored === "both" && window.innerWidth < MIN_SPLIT_WIDTH) return "editor"
    return stored
  }
  return window.innerWidth >= MIN_SPLIT_WIDTH ? "both" : "editor"
}

interface ToggleButtonProps {
  active: boolean
  onClick: () => void
  icon: React.ComponentType<{ className?: string }>
  label: string
  title?: string
}

function ToggleButton({ active, onClick, icon: Icon, label, title }: ToggleButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      title={title}
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

export function PromptSplitEditor({
  value,
  onChange,
  placeholder,
  maxLength,
  className,
  id,
  ...aria
}: PromptSplitEditorProps) {
  const [mode, setMode] = useState<Mode>(getInitialMode)
  const [canSplit, setCanSplit] = useState(
    typeof window !== "undefined" ? window.innerWidth >= MIN_SPLIT_WIDTH : true,
  )

  // Отслеживаем ширину окна, чтобы "both" автоматически деградировал в "editor"
  // при resize до mobile.
  // MN-43: debounce 100ms — без него каждый pixel resize дёргал re-render,
  // что замечалось при медленном dragging бордера окна на больших экранах.
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | null = null
    function update() {
      if (timer) clearTimeout(timer)
      timer = setTimeout(() => {
        const next = window.innerWidth >= MIN_SPLIT_WIDTH
        setCanSplit(next)
        if (!next && mode === "both") setMode("editor")
      }, 100)
    }
    window.addEventListener("resize", update)
    return () => {
      if (timer) clearTimeout(timer)
      window.removeEventListener("resize", update)
    }
  }, [mode])

  // Персистим выбор в localStorage, чтобы при возврате в редактор он
  // восстановился.
  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, mode)
  }, [mode])

  // ⌘/Ctrl + / — циклический переключатель режима.
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === "/") {
        e.preventDefault()
        setMode((m) => {
          if (!canSplit) return m === "editor" ? "preview" : "editor"
          return m === "editor" ? "both" : m === "both" ? "preview" : "editor"
        })
      }
    }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [canSplit])

  const showEditor = mode === "editor" || mode === "both"
  const showPreview = mode === "preview" || mode === "both"

  return (
    <div className={cn("space-y-2", className)}>
      <div
        role="tablist"
        aria-label="Режим редактора промпта"
        className="inline-flex rounded-lg border border-border bg-muted/30 p-0.5"
      >
        <ToggleButton
          active={mode === "editor"}
          onClick={() => setMode("editor")}
          icon={Pencil}
          label="Редактор"
          title="Только редактор (⌘/)"
        />
        {canSplit && (
          <ToggleButton
            active={mode === "both"}
            onClick={() => setMode("both")}
            icon={Columns}
            label="Оба"
            title="Редактор + превью (⌘/)"
          />
        )}
        <ToggleButton
          active={mode === "preview"}
          onClick={() => setMode("preview")}
          icon={Eye}
          label="Превью"
          title="Только превью (⌘/)"
        />
      </div>

      <div
        className={cn(
          "grid gap-3",
          mode === "both" ? "grid-cols-1 lg:grid-cols-2" : "grid-cols-1",
        )}
      >
        {showEditor && (
          <MarkdownEditor
            value={value}
            onChange={onChange}
            placeholder={placeholder}
            maxLength={maxLength}
            id={id}
            minHeight="320px"
            // 60vh держит редактор в пределах одного экрана даже на коротких
            // ноутбуках (≤768px) — action-buttons остаются видимыми без
            // скролла страницы. Для большого монитора кепится 640px.
            maxHeight="min(640px, 60vh)"
            aria-invalid={aria["aria-invalid"]}
            aria-describedby={aria["aria-describedby"]}
          />
        )}
        {showPreview && (
          <div
            // max-h применяется ВСЕГДА (раньше было только в both — single-tab
            // превью разрастался на любую высоту, заставляя пользователя
            // скроллить страницу до кнопок Сохранить/Отмена).
            className="min-h-[320px] max-h-[min(640px,60vh)] overflow-y-auto rounded-lg border border-border bg-muted/20 px-4 py-3"
            role="tabpanel"
            aria-label="Превью промпта"
          >
            {value.trim() ? (
              <PromptContent content={value} />
            ) : (
              <p className="text-[0.8rem] italic text-muted-foreground">
                Превью появится здесь, когда вы начнёте вводить промпт.
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
