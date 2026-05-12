// CodeMirror 6 extension для подсветки `{{переменных}}` в prompt-контенте.
// Грамматика идентична shared/template/parse.ts (single source of truth).

import { Decoration, DecorationSet, ViewPlugin, ViewUpdate, EditorView } from "@codemirror/view"
import { RangeSetBuilder } from "@codemirror/state"

const VARIABLE_REGEX = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu

const variableMark = Decoration.mark({
  class: "cm-template-variable",
  inclusiveStart: false,
  inclusiveEnd: false,
})

function buildDecorations(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  for (const { from, to } of view.visibleRanges) {
    const text = view.state.doc.sliceString(from, to)
    for (const match of text.matchAll(VARIABLE_REGEX)) {
      const start = (match.index ?? 0) + from
      const end = start + match[0].length
      builder.add(start, end, variableMark)
    }
  }
  return builder.finish()
}

export const templateVariableHighlight = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet

    constructor(view: EditorView) {
      this.decorations = buildDecorations(view)
    }

    update(update: ViewUpdate) {
      if (update.docChanged || update.viewportChanged) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  {
    decorations: (v: { decorations: DecorationSet }) => v.decorations,
  },
)

export const templateVariableTheme = EditorView.baseTheme({
  ".cm-template-variable": {
    backgroundColor: "rgb(139 92 246 / 0.15)",
    color: "rgb(167 139 250)",
    fontFamily: "monospace",
    borderRadius: "2px",
    padding: "0 1px",
  },
})
