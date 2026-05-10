import {
  Decoration,
  MatchDecorator,
  ViewPlugin,
  EditorView,
  type DecorationSet,
  type ViewUpdate,
} from "@codemirror/view"

/**
 * Подсветка `{{var}}` в CodeMirror-редакторе промптов.
 *
 * Регулярка ИДЕНТИЧНА grammar'у в `frontend/src/lib/template/parse.ts` и
 * `backend/internal/template/template.go`:
 *   variable := "{{" identifier "}}"
 *   identifier := (letter | "_") (letter | digit | "_")*
 *
 * Это значит:
 *   - `{{name}}`, `{{имя}}`, `{{_x}}` — подсвечиваются
 *   - `{{ name }}` (с пробелами) — НЕ подсвечивается
 *     (это и есть implicit escape: чтобы вставить буквальные `{{`,
 *     поставьте пробел внутри)
 *   - `{{1abc}}`, `{{my-var}}` — НЕ подсвечиваются (невалидный identifier)
 *
 * Если когда-то появится backend-расширение синтаксиса (фильтры, секции),
 * этот файл должен меняться синхронно с парсерами обеих сторон.
 */
const VARIABLE_REGEX = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu

const variableMatcher = new MatchDecorator({
  regexp: VARIABLE_REGEX,
  decoration: Decoration.mark({ class: "cm-template-var" }),
})

const variableHighlighterPlugin = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet
    constructor(view: EditorView) {
      this.decorations = variableMatcher.createDeco(view)
    }
    update(update: ViewUpdate) {
      this.decorations = variableMatcher.updateDeco(update, this.decorations)
    }
  },
  {
    decorations: (v) => v.decorations,
  },
)

const variableTheme = EditorView.baseTheme({
  ".cm-template-var": {
    backgroundColor: "oklch(from var(--primary) l c h / 14%)",
    color: "var(--primary)",
    borderRadius: "3px",
    padding: "0 2px",
    fontWeight: "600",
  },
})

export const templateVariableHighlight = [variableHighlighterPlugin, variableTheme]
