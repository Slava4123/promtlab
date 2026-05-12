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

export function CodeEditor({ extraExtensions, ...rest }: CodeEditorProps) {
  const extensions = useMemo(
    () => [
      markdown(),
      templateVariableHighlight,
      templateVariableTheme,
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
      theme="dark"
      extensions={extensions}
      {...rest}
    />
  )
}
