import { MAX_PROMPT_CONTENT_LENGTH } from "@/lib/constants"
import type { FileKind, ParseResult } from "./types"

// Читает файл как текст в UTF-8 (без encoding detect — это Commit 9).
// Handle BOM и нормализует переводы строк. Возвращает ParseResult.
//
// Используется для FileKind = "text" | "markdown".
export async function parseTextFile(
  file: File,
  kind: Extract<FileKind, "text" | "markdown">,
): Promise<ParseResult> {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)

  // Определяем encoding по BOM. UTF-8 по умолчанию.
  let encoding: string
  let dataOffset = 0
  if (bytes.length >= 3 && bytes[0] === 0xef && bytes[1] === 0xbb && bytes[2] === 0xbf) {
    encoding = "utf-8"
    dataOffset = 3
  } else if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe) {
    encoding = "utf-16le"
    dataOffset = 2
  } else if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    encoding = "utf-16be"
    dataOffset = 2
  } else {
    encoding = "utf-8"
  }

  // fatal:false — битые последовательности заменятся U+FFFD, не выбросит ошибку.
  // В Commit 9 добавим fallback через jschardet при подозрительном decode.
  const decoder = new TextDecoder(encoding, { fatal: false })
  const raw = decoder.decode(bytes.slice(dataOffset))

  let finalText = raw
  let detectedEncoding = encoding
  let encodingConfidence: number | undefined
  const warnings: string[] = []

  // Heuristic для "кракозябр": если >10% символов — U+FFFD, пробуем пере-
  // декодировать через jschardet (cp1251/KOI8-R/IBM866).
  const replacementCharCount = (raw.match(/\uFFFD/g) ?? []).length
  if (replacementCharCount > raw.length * 0.1) {
    // Lazy-chunk: jschardet подгружается только при подозрении на non-UTF-8.
    const { detectAndDecode } = await import("./encoding")
    const result = detectAndDecode(bytes.slice(dataOffset), raw)
    if (result.recovered) {
      finalText = result.content
      detectedEncoding = result.encoding
      encodingConfidence = result.confidence
      warnings.push(
        `Определена кодировка ${result.encoding} (уверенность ${Math.round(result.confidence * 100)}%) — текст переконвертирован в UTF-8`,
      )
    } else {
      warnings.push(
        "Файл содержит некорректные символы — не удалось определить кодировку. " +
          "Откройте файл в редакторе и сохраните его в UTF-8",
      )
    }
  }

  const normalized = normalizeText(finalText)
  const { content, truncated } = truncateToLimit(normalized)

  return {
    content,
    kind,
    filename: file.name,
    originalBytes: file.size,
    truncated,
    detectedEncoding,
    ...(encodingConfidence !== undefined ? { encodingConfidence } : {}),
    warnings,
  }
}

// Нормализация переводов строк и spec-символов.
// \r\n (Windows) и одиночные \r (classic Mac) → \n; NULL-байты вырезаем
// (иногда появляются в .txt из legacy Windows-экспортов).
export function normalizeText(input: string): string {
  // eslint-disable-next-line no-control-regex
  return input.replace(/\r\n/g, "\n").replace(/\r/g, "\n").replace(/\u0000/g, "")
}

// Обрезаем до MAX_PROMPT_CONTENT_LENGTH (100_000 символов). Решение "показать
// ли диалог truncation" принимает caller — мы лишь возвращаем truncated=true.
function truncateToLimit(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return {
    content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH),
    truncated: true,
  }
}
