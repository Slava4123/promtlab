import { MAX_PROMPT_CONTENT_LENGTH } from "@/lib/constants"
import { FileImportError, type ParseResult } from "./types"
import { normalizeText } from "./parse-text"

// Max страниц которые разрешаем парсить. Защита от PDF с 10000 страниц (которые
// будут грузить вкладку минутами).
const MAX_PDF_PAGES = 500
// Timeout на полный парсинг — если pdfjs завис на malformed PDF.
const PDF_PARSE_TIMEOUT_MS = 30_000

// Парсит .pdf через pdfjs-dist v5 (CVE-2024-4367 fixed в 4.2.67+, мы на 5.6+).
//
// Security:
//  - `isEvalSupported: false` — отключает eval в font loader (defence-in-depth
//    поверх CVE fix).
//  - `disableFontFace: true` — не рендерим шрифты, экономим CPU.
//
// Worker setup: legacy build + ?url suffix для Vite 8 ESM совместимости.
export async function parsePdfFile(file: File): Promise<ParseResult> {
  // Dynamic import + worker URL через Vite ?url — этот chunk lazy.
  const pdfjsLib = await import("pdfjs-dist/legacy/build/pdf.mjs")
  const workerUrl = (
    await import("pdfjs-dist/legacy/build/pdf.worker.min.mjs?url")
  ).default
  pdfjsLib.GlobalWorkerOptions.workerSrc = workerUrl

  const arrayBuffer = await file.arrayBuffer()

  const loadingTask = pdfjsLib.getDocument({
    data: new Uint8Array(arrayBuffer),
    // Security flags.
    isEvalSupported: false,
    disableFontFace: true,
    useSystemFonts: false,
    // Производительность для больших PDF.
    disableAutoFetch: true,
    disableStream: true,
    // Чуть меньше шума в console.
    verbosity: 0,
  })

  const timeoutHandle = new Promise<never>((_, reject) =>
    setTimeout(
      () =>
        reject(
          new FileImportError(
            "PARSE_FAILED",
            `Парсинг PDF занял больше ${PDF_PARSE_TIMEOUT_MS / 1000}с — возможно файл повреждён`,
            "pdf",
          ),
        ),
      PDF_PARSE_TIMEOUT_MS,
    ),
  )

  let pdf: Awaited<ReturnType<typeof pdfjsLib.getDocument>["promise"]>
  try {
    pdf = await Promise.race([loadingTask.promise, timeoutHandle])
  } catch (err) {
    if (err instanceof FileImportError) throw err
    throw new FileImportError(
      "PARSE_FAILED",
      `Не удалось открыть PDF: ${err instanceof Error ? err.message : "unknown"}`,
      "pdf",
    )
  }

  try {
    const pagesCount = pdf.numPages
    if (pagesCount > MAX_PDF_PAGES) {
      throw new FileImportError(
        "SIZE_EXCEEDED",
        `PDF содержит ${pagesCount} страниц — превышен лимит ${MAX_PDF_PAGES}`,
        "pdf",
      )
    }

    const pageTexts: string[] = []
    for (let pageNum = 1; pageNum <= pagesCount; pageNum++) {
      const page = await pdf.getPage(pageNum)
      const textContent = await page.getTextContent({
        includeMarkedContent: false,
        disableNormalization: false,
      })

      let pageText = ""
      let lastY: number | null = null
      for (const item of textContent.items) {
        // getTextContent возвращает TextItem[] или TextMarkedContent[]; у нас
        // includeMarkedContent:false, значит только TextItem (есть .str).
        if (!("str" in item)) continue
        // item.transform[5] — Y координата (в user space). Если Y сильно
        // изменился — была строка-разрыв, добавляем \n.
        const y = (item.transform?.[5] as number | undefined) ?? 0
        if (lastY !== null && Math.abs(y - lastY) > 5) {
          pageText += "\n"
        }
        pageText += item.str
        if (item.hasEOL) {
          pageText += "\n"
        }
        lastY = y
      }

      pageTexts.push(pageText)
      page.cleanup()
    }

    const joined = pageTexts.join("\n\n").trim()
    const normalized = normalizeText(joined)

    if (normalized.length === 0) {
      throw new FileImportError(
        "EMPTY_RESULT",
        "PDF не содержит текстового слоя — это скан или файл без OCR",
        "pdf",
      )
    }

    const { content, truncated } = truncate(normalized)

    return {
      content,
      kind: "pdf",
      filename: file.name,
      originalBytes: file.size,
      truncated,
      pages: pagesCount,
      warnings: [],
    }
  } finally {
    // Освобождаем ресурсы worker'а. Не ждём промиса — fire-and-forget.
    void pdf.destroy()
  }
}

function truncate(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return { content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH), truncated: true }
}
