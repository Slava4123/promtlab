import { useMemo } from "react"
import CodeMirror, { type ReactCodeMirrorProps } from "@uiw/react-codemirror"
import { markdown } from "@codemirror/lang-markdown"
import { EditorView } from "@codemirror/view"
import {
  templateVariableHighlight,
  templateVariableTheme,
} from "../../lib/codemirror/variable-highlight"

interface CodeEditorProps extends Omit<ReactCodeMirrorProps, "extensions"> {
  /** Дополнительные CodeMirror extensions сверх дефолтных. */
  extraExtensions?: ReactCodeMirrorProps["extensions"]
}

// Promptсодержание — естественный язык (русский/латиница), не код. Дефолтный
// CodeMirror шрифт (monospace) выглядит как typewriter, что неуместно: затрудняет
// чтение длинных промптов и визуально диссонирует с остальным UI на Geist.
// Также убираем встроенный theme="dark" — он жёстко перекрашивал фон в чёрный
// даже в light-теме. Теперь editor наследует --color-card от обёртки.
const editorTheme = EditorView.theme({
  "&": {
    fontFamily:
      '"Geist Variable", ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
    fontSize: "13px",
    color: "var(--color-foreground)",
    backgroundColor: "transparent",
  },
  ".cm-content": {
    fontFamily: "inherit",
    padding: "12px",
    caretColor: "var(--color-brand)",
  },
  ".cm-line": {
    fontFamily: "inherit",
  },
  "&.cm-focused": {
    outline: "none",
  },
  ".cm-cursor": {
    borderLeftColor: "var(--color-brand)",
  },
  ".cm-selectionBackground, ::selection": {
    backgroundColor: "var(--color-brand-muted)",
  },
  "&.cm-focused .cm-selectionBackground": {
    backgroundColor: "var(--color-brand-muted)",
  },
})

export function CodeEditor({ extraExtensions, ...rest }: CodeEditorProps) {
  const extensions = useMemo(
    () => [
      markdown(),
      templateVariableHighlight,
      templateVariableTheme,
      editorTheme,
      EditorView.lineWrapping,
      ...(extraExtensions ?? []),
    ],
    [extraExtensions],
  )

  return (
    <CodeMirror
      basicSetup={{
        lineNumbers: false,
        foldGutter: false,
        highlightActiveLine: false,
        highlightActiveLineGutter: false,
        autocompletion: false,
        searchKeymap: false,
        indentOnInput: false,
      }}
      extensions={extensions}
      {...rest}
    />
  )
}
