import TurndownService from "turndown"
import { MAX_PROMPT_CONTENT_LENGTH } from "@/lib/constants"
import { FileImportError, type ParseResult } from "./types"
import { normalizeText } from "./parse-text"

// Набор опасных HTML-тегов которые вырезаем ДО turndown'а. turndown сам
// удалит <script>/<style> при `.remove()`, но мы чистим более агрессивно
// через DOMParser + ручную чистку on*-атрибутов и javascript:-URL.
const DANGEROUS_TAGS = new Set([
  "script",
  "style",
  "iframe",
  "frame",
  "frameset",
  "object",
  "embed",
  "applet",
  "noscript",
  "link",
  "meta",
])

// Атрибуты-обработчики событий (onclick, onmouseover и т.д.) — все on*.
function sanitizeDom(root: Document) {
  // 1. Удалить опасные теги целиком.
  for (const tag of DANGEROUS_TAGS) {
    for (const el of Array.from(root.getElementsByTagName(tag))) {
      el.remove()
    }
  }
  // 2. Обойти все элементы и снять on*-атрибуты и javascript:-URL.
  const all = root.querySelectorAll("*")
  for (const el of Array.from(all)) {
    for (const attr of Array.from(el.attributes)) {
      if (attr.name.startsWith("on")) {
        el.removeAttribute(attr.name)
        continue
      }
      // href="javascript:..." / src="javascript:..." — снести
      if (
        (attr.name === "href" || attr.name === "src" || attr.name === "xlink:href") &&
        /^\s*javascript:/i.test(attr.value)
      ) {
        el.removeAttribute(attr.name)
      }
    }
  }
}

// Парсит HTML-файл в markdown:
//  1. DOMParser → document
//  2. sanitize (удалить script/iframe/style, on* атрибуты, javascript:-URL)
//  3. turndown с headingStyle='atx' + codeBlockStyle='fenced' + strikethrough + table rules
//
// XSS-защита: любой `<script>` или `onerror=` — не попадают даже в сырой HTML
// до turndown'а, значит в итоговом markdown их просто нет. Дополнительно
// финальный markdown всё равно проходит через rehype-sanitize в <PromptContent/>.
export async function parseHtmlFile(file: File): Promise<ParseResult> {
  const text = await file.text()

  let doc: Document
  try {
    doc = new DOMParser().parseFromString(text, "text/html")
  } catch (err) {
    throw new FileImportError(
      "PARSE_FAILED",
      `Не удалось распарсить HTML: ${err instanceof Error ? err.message : "unknown"}`,
      "html",
    )
  }

  sanitizeDom(doc)

  const service = new TurndownService({
    headingStyle: "atx",
    codeBlockStyle: "fenced",
    fence: "```",
    bulletListMarker: "-",
    emDelimiter: "*",
    strongDelimiter: "**",
    linkStyle: "inlined",
  })
  // Strikethrough: <del>, <s>, <strike> ("strike" — HTML4 legacy, не в HTMLElementTagNameMap)
  service.addRule("strikethrough", {
    filter: ["del", "s", "strike"] as (keyof HTMLElementTagNameMap)[],
    replacement: (content) => `~~${content}~~`,
  })
  // GFM tables: turndown по умолчанию рушит <table> в текст. Добавляем простое
  // правило конвертации в markdown-таблицу.
  service.addRule("table", {
    filter: "table",
    replacement: (_content, node) => {
      return htmlTableToMarkdown(node as HTMLTableElement)
    },
  })
  // Если что-то прошло мимо DOMParser sanitize — на всякий случай.
  service.remove(["script", "style", "iframe", "noscript"])

  // Берём body (если нет — целый документ). Turndown принимает HTMLElement.
  const source = doc.body ?? doc.documentElement
  const rawMarkdown = service.turndown(source.innerHTML)
  const normalized = normalizeText(rawMarkdown).trim()

  if (normalized.length === 0) {
    throw new FileImportError(
      "EMPTY_RESULT",
      "HTML не содержит текстового контента после очистки",
      "html",
    )
  }

  const { content, truncated } = truncate(normalized)

  return {
    content,
    kind: "html",
    filename: file.name,
    originalBytes: file.size,
    truncated,
    warnings: [],
  }
}

// Минималистичная конверсия <table> в GFM-markdown. Берём <th> из первой row
// или все <td>/<th> из первой tr как header. Игнорируем colspan/rowspan (редко
// встречается в промптах).
function htmlTableToMarkdown(table: HTMLTableElement): string {
  const rows = Array.from(table.querySelectorAll("tr"))
  if (rows.length === 0) return ""

  const cells = rows.map((tr) =>
    Array.from(tr.querySelectorAll("td, th")).map((cell) =>
      cellText(cell as HTMLElement),
    ),
  )
  if (cells.length === 0 || cells[0].length === 0) return ""

  const colCount = Math.max(...cells.map((r) => r.length))
  const header = padRow(cells[0], colCount)
  const separator = Array(colCount).fill("---")
  const body = cells.slice(1).map((r) => padRow(r, colCount))

  const lines = [
    "| " + header.join(" | ") + " |",
    "| " + separator.join(" | ") + " |",
    ...body.map((r) => "| " + r.join(" | ") + " |"),
  ]
  return "\n\n" + lines.join("\n") + "\n\n"
}

function cellText(el: HTMLElement): string {
  // Без markdown внутри ячеек (turndown-рекурсия на cells теряется). Plain text.
  return (el.textContent ?? "").replace(/\s+/g, " ").trim()
}

function padRow(row: string[], colCount: number): string[] {
  if (row.length >= colCount) return row
  return [...row, ...Array(colCount - row.length).fill("")]
}

function truncate(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return { content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH), truncated: true }
}
