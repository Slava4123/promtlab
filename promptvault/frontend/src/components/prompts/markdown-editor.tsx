import { useMemo } from "react"
import CodeMirror, { EditorView, keymap } from "@uiw/react-codemirror"
import { markdown, markdownLanguage } from "@codemirror/lang-markdown"
import { languages } from "@codemirror/language-data"
import { oneDark } from "@codemirror/theme-one-dark"
import { indentWithTab } from "@codemirror/commands"
import { useThemeStore } from "@/stores/theme-store"
import { cn } from "@/lib/utils"

interface MarkdownEditorProps {
  value: string
  onChange?: (value: string) => void
  placeholder?: string
  maxLength?: number
  className?: string
  minHeight?: string
  readOnly?: boolean
  "aria-invalid"?: boolean
  "aria-describedby"?: string
  id?: string
}

// Стилистика редактора: моноширинный fallback, увеличенный line-height для читаемости
// длинных промптов, headers — крупнее, code-span — подсвеченный фон.
// Не используем theme-stealing у highlight.js (только rehype-highlight в preview),
// а даём CodeMirror свои правила подсветки — иначе получим visual clash.
const baseTheme = EditorView.theme({
  "&": {
    fontSize: "14px",
    fontFamily:
      "'Geist Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
    height: "auto",
  },
  ".cm-content": {
    padding: "12px 14px",
    lineHeight: "1.65",
    caretColor: "var(--foreground)",
  },
  ".cm-gutters": {
    display: "none", // промпты — не код, номера строк не нужны
  },
  ".cm-line": {
    padding: "0 0 0 0",
  },
  "&.cm-focused": {
    outline: "none",
  },
  ".cm-scroller": {
    fontFamily:
      "'Geist Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
    minHeight: "inherit",
  },
  ".cm-placeholder": {
    color: "oklch(from var(--muted-foreground) l c h / 70%)",
    fontStyle: "italic",
  },
})

// Визуальная подсветка markdown-разметки: headers крупнее/жирные, bold/italic,
// code-span с фоном, blockquote приглушённый.
// Используем только CSS — ни одного React-render'а на токен, чтобы не тормозить
// на 100K символов.
const markdownStylingDark = EditorView.theme({
  ".cm-line .tok-heading1, .cm-line .tok-heading": {
    fontSize: "1.15em",
    fontWeight: "700",
    color: "oklch(0.985 0 0)",
  },
  ".cm-line .tok-strong": { fontWeight: "700", color: "oklch(0.985 0 0)" },
  ".cm-line .tok-emphasis": { fontStyle: "italic" },
  ".cm-line .tok-monospace": {
    backgroundColor: "oklch(from var(--foreground) l c h / 8%)",
    padding: "0 4px",
    borderRadius: "3px",
  },
  ".cm-line .tok-link": { color: "oklch(0.811 0.111 293)" },
  ".cm-line .tok-url": { color: "oklch(0.708 0.111 254)" },
  ".cm-line .tok-quote": {
    color: "oklch(0.708 0 0)",
    fontStyle: "italic",
  },
})

export function MarkdownEditor({
  value,
  onChange,
  placeholder,
  maxLength,
  className,
  minHeight = "280px",
  readOnly = false,
  id,
  ...aria
}: MarkdownEditorProps) {
  const theme = useThemeStore((s) => s.theme)
  const isDark =
    theme === "dark" ||
    (theme === "system" &&
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)

  // Обрезаем ввод по maxLength (если указан). Обработка — на уровне onChange,
  // чтобы не трогать внутреннее состояние CodeMirror.
  const handleChange = (val: string) => {
    if (!onChange) return
    if (maxLength !== undefined && val.length > maxLength) {
      onChange(val.slice(0, maxLength))
      return
    }
    onChange(val)
  }

  const extensions = useMemo(
    () => [
      markdown({ base: markdownLanguage, codeLanguages: languages }),
      EditorView.lineWrapping,
      keymap.of([indentWithTab]),
      baseTheme,
      ...(isDark ? [markdownStylingDark] : []),
    ],
    [isDark],
  )

  return (
    <div
      id={id}
      className={cn(
        "overflow-hidden rounded-lg border border-border bg-background transition-colors focus-within:border-violet-500/40 focus-within:ring-3 focus-within:ring-violet-500/10",
        className,
      )}
      {...aria}
    >
      <CodeMirror
        value={value}
        onChange={handleChange}
        placeholder={placeholder}
        extensions={extensions}
        theme={isDark ? oneDark : "light"}
        minHeight={minHeight}
        maxHeight="640px"
        editable={!readOnly}
        readOnly={readOnly}
        basicSetup={{
          lineNumbers: false,
          foldGutter: false,
          highlightActiveLine: !readOnly,
          highlightActiveLineGutter: false,
          autocompletion: false,
          searchKeymap: true,
          dropCursor: !readOnly,
          indentOnInput: !readOnly,
          bracketMatching: false,
        }}
      />
    </div>
  )
}
