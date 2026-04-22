import { detectFileKind } from "./detect"
import { MAX_UPLOAD_BYTES } from "./constants"
import { FileImportError, type ParseResult } from "./types"
import { parseTextFile } from "./parse-text"
import { parseJsonFile } from "./parse-json"

// Главная точка входа фичи file-import. Принимает File (из input/drop), отдаёт
// ParseResult либо бросает FileImportError.
//
// Для тяжёлых парсеров (pdf/docx/html/rtf) используется dynamic import() —
// Vite автоматически создаёт отдельные lazy chunks. Main bundle не растёт.
//
// Вызывается из onFileImport handler в prompt-editor.tsx.
export async function parseFile(file: File): Promise<ParseResult> {
  // 1. Pre-read size check — защита от DoS огромным файлом.
  if (file.size > MAX_UPLOAD_BYTES) {
    throw new FileImportError(
      "SIZE_EXCEEDED",
      `Файл слишком большой: ${formatBytes(file.size)} (лимит ${formatBytes(MAX_UPLOAD_BYTES)})`,
    )
  }
  if (file.size === 0) {
    throw new FileImportError("EMPTY_RESULT", "Файл пустой")
  }

  // 2. Determine FileKind (magic-bytes + extension).
  const detection = await detectFileKind(file)

  // 3. Dispatch to parser. Commit 1 поддерживает только P0: text/markdown/json.
  // Остальные форматы — FileImportError("UNSUPPORTED") с подсказкой, что формат
  // появится в следующих коммитах.
  switch (detection.kind) {
    case "text":
    case "markdown":
      return parseTextFile(file, detection.kind)

    case "json":
      return parseJsonFile(file)

    case "ipynb": {
      // Lazy-chunk: notebook-парсер подгружается только при .ipynb.
      const { parseIpynbFile } = await import("./parse-ipynb")
      return parseIpynbFile(file)
    }

    case "html": {
      // Lazy-chunk: turndown подгружается только при .html.
      const { parseHtmlFile } = await import("./parse-html")
      return parseHtmlFile(file)
    }

    case "rtf": {
      // Lazy-chunk: RTF stripper подгружается только при .rtf.
      const { parseRtfFile } = await import("./parse-rtf")
      return parseRtfFile(file)
    }

    case "docx": {
      // Lazy-chunk: mammoth подгружается только при .docx.
      const { parseDocxFile } = await import("./parse-docx")
      return parseDocxFile(file)
    }

    case "pdf": {
      // Lazy-chunk: pdfjs-dist подгружается только при .pdf.
      const { parsePdfFile } = await import("./parse-pdf")
      return parsePdfFile(file)
    }
  }
}

// Экспортируем для прямого использования в тестах / будущих расширений.
export { detectFileKind } from "./detect"
export { MAX_UPLOAD_BYTES, FILE_KIND_LABELS, ACCEPTED_FILE_EXTENSIONS } from "./constants"
export type { FileKind, ParseResult } from "./types"
export { FileImportError } from "./types"

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
  return `${(bytes / 1024 / 1024).toFixed(1)} МБ`
}
