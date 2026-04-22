import { MAX_PROMPT_CONTENT_LENGTH } from "@/lib/constants"
import { FileImportError, type ParseResult } from "./types"
import { normalizeText } from "./parse-text"

// Hard cap на конечную длину vs MAX_PROMPT_CONTENT_LENGTH — защита от
// zip-bomb. Если mammoth выдал сильно больше лимита, обрезаем (вместо abort'а
// чтобы не отказать в импорте если юзер готов к truncation).
const MAX_OUTPUT_CHARS = MAX_PROMPT_CONTENT_LENGTH * 5

// Парсит .docx через mammoth v1.11+ (CVE-2025-11849 fix).
// Использует convertToMarkdown напрямую — не идём через HTML.
// Изображения → alt-текст без src (не встраиваем base64, для промптов не нужно).
export async function parseDocxFile(file: File): Promise<ParseResult> {
  const mammoth = (await import("mammoth/mammoth.browser.js")).default

  const arrayBuffer = await file.arrayBuffer()

  const imageHandler = mammoth.images.imgElement(async (image) => {
    // Возвращаем "плейсхолдер" — src="" + alt (если был). mammoth вставит <img alt="..." src="">,
    // что на стадии конверсии в markdown даст `![alt]()`. Caller юзер дополнит сам.
    return {
      src: "",
      alt: image.altText ?? "",
    }
  })

  let result: { value: string; messages: Array<{ type: string; message: string }> }
  try {
    result = await mammoth.convertToMarkdown(
      { arrayBuffer },
      { convertImage: imageHandler },
    )
  } catch (err) {
    throw new FileImportError(
      "PARSE_FAILED",
      `Не удалось обработать DOCX: ${err instanceof Error ? err.message : "unknown"}`,
      "docx",
    )
  }

  // Zip-bomb defense: если mammoth извлёк экстремально много — abort.
  if (result.value.length > MAX_OUTPUT_CHARS) {
    throw new FileImportError(
      "SIZE_EXCEEDED",
      `DOCX выдал ${result.value.length.toLocaleString("ru-RU")} символов — подозрительно для файла размером ${formatBytes(file.size)}`,
      "docx",
    )
  }

  const normalized = normalizeText(result.value).trim()
  if (normalized.length === 0) {
    throw new FileImportError("EMPTY_RESULT", "DOCX не содержит текста", "docx")
  }

  const warnings: string[] = []
  // mammoth возвращает messages с type=warning для неподдерживаемых фичей
  // (custom styles, etc). Собираем первые 3 — больше раздражает.
  const mammothWarnings = result.messages
    .filter((m) => m.type === "warning")
    .slice(0, 3)
  if (mammothWarnings.length > 0) {
    warnings.push(
      `DOCX частично: ${mammothWarnings.map((m) => m.message).join("; ")}`,
    )
  }

  const { content, truncated } = truncate(normalized)

  return {
    content,
    kind: "docx",
    filename: file.name,
    originalBytes: file.size,
    truncated,
    warnings,
  }
}

function truncate(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return { content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH), truncated: true }
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
  return `${(bytes / 1024 / 1024).toFixed(1)} МБ`
}
