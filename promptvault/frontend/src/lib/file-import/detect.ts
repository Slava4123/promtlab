import type { FileKind } from "./types"

// Minimum bytes we need to read for magic-byte detection. PDF и DOCX
// идентифицируются первыми 4-8 байтами, HTML — до первых ~100 символов.
const MAGIC_BYTES_WINDOW = 512

// Нормализуем расширение файла → FileKind, если расширение есть в whitelist.
function kindFromExtension(filename: string): FileKind | undefined {
  const lower = filename.toLowerCase()
  if (lower.endsWith(".txt")) return "text"
  if (lower.endsWith(".md") || lower.endsWith(".markdown")) return "markdown"
  if (lower.endsWith(".json")) return "json"
  if (lower.endsWith(".ipynb")) return "ipynb"
  if (lower.endsWith(".html") || lower.endsWith(".htm")) return "html"
  if (lower.endsWith(".rtf")) return "rtf"
  if (lower.endsWith(".docx")) return "docx"
  if (lower.endsWith(".pdf")) return "pdf"
  return undefined
}

// Распознаём FileKind по первым байтам. Возвращаем undefined если не можем
// уверенно определить (это не сразу ошибка — для plain text magic-bytes нет).
function kindFromMagicBytes(bytes: Uint8Array): FileKind | undefined {
  // PDF — обязательно начинается с "%PDF-"
  if (
    bytes[0] === 0x25 &&
    bytes[1] === 0x50 &&
    bytes[2] === 0x44 &&
    bytes[3] === 0x46 &&
    bytes[4] === 0x2d
  ) {
    return "pdf"
  }

  // ZIP-based форматы: DOCX, IPYNB(если zipped, редко), XLSX — PK\x03\x04.
  // Дальше смотрим content-type. В v1 считаем что ZIP = DOCX (ipynb обычно JSON).
  if (
    bytes[0] === 0x50 &&
    bytes[1] === 0x4b &&
    bytes[2] === 0x03 &&
    bytes[3] === 0x04
  ) {
    return "docx"
  }

  // RTF — "{\\rtf"
  if (
    bytes[0] === 0x7b &&
    bytes[1] === 0x5c &&
    bytes[2] === 0x72 &&
    bytes[3] === 0x74 &&
    bytes[4] === 0x66
  ) {
    return "rtf"
  }

  // HTML — смотрим в первых 256 char'ах что-то похожее на <!doctype html> / <html / <body / <head
  // Robust heuristic: декодируем первые 256 байт как UTF-8 (ignoring errors) и ищем tags.
  const head = new TextDecoder("utf-8", { fatal: false })
    .decode(bytes.slice(0, 256))
    .toLowerCase()
    .trimStart()
  if (
    head.startsWith("<!doctype html") ||
    head.startsWith("<html") ||
    head.startsWith("<?xml") && head.includes("<html")
  ) {
    return "html"
  }

  // JSON — очень грубо: первый непробельный символ { или [
  const firstNonWs = head.match(/\S/)?.[0]
  if (firstNonWs === "{" || firstNonWs === "[") {
    // Дополнительная валидация (JSON.parse) будет в parse-json.ts — здесь только намёк.
    return "json"
  }

  return undefined
}

// Определяет FileKind по файлу. Magic-bytes приоритетнее расширения (защита
// от MIME-spoofing). Если magic-bytes не дают ответа — падаем на расширение.
// Если и то и другое не распознано — "text" как безопасный fallback (все
// непонятные файлы пытаемся читать как UTF-8 text — хуже не будет).
//
// ВАЖНО: для markdown-файлов magic-bytes нет (это plain UTF-8 text) — такие
// файлы будут распознаны по расширению в kindFromExtension и НЕ перекрыты
// магическим определением (мы не падаем в kindFromMagicBytes для них).
export async function detectFileKind(file: File): Promise<{
  kind: FileKind
  byExtension?: FileKind
  byMagicBytes?: FileKind
  mismatch: boolean
}> {
  const byExtension = kindFromExtension(file.name)

  const buffer = await file.slice(0, MAGIC_BYTES_WINDOW).arrayBuffer()
  const bytes = new Uint8Array(buffer)
  const byMagicBytes = kindFromMagicBytes(bytes)

  // Если magic-bytes уверенно сказал — доверяем им (защита от переименованных файлов).
  if (byMagicBytes !== undefined) {
    // Special case: .ipynb — это JSON по magic-bytes, но семантически другой
    // формат. Расширение приоритетнее чтобы роутиться в parse-ipynb.
    if (byExtension === "ipynb" && byMagicBytes === "json") {
      return { kind: "ipynb", byExtension, byMagicBytes, mismatch: false }
    }
    // Отметить mismatch, если расширение несовместимо. Напр. .txt с PDF-контентом —
    // технически можем обработать (парсим как PDF), но поднимем warning в caller.
    const mismatch =
      byExtension !== undefined &&
      byExtension !== byMagicBytes &&
      // markdown-файлы "совместимы" с text (и наоборот) — не считаем mismatch.
      !(byExtension === "markdown" && byMagicBytes === "text") &&
      !(byExtension === "text" && byMagicBytes === "markdown") &&
      // JSON может валидно читаться и как text — если расширение .txt/.md, разрешаем.
      !(byExtension !== undefined && byMagicBytes === "json")
    return { kind: byMagicBytes, byExtension, byMagicBytes, mismatch }
  }

  // Magic-bytes не дал ответа → используем расширение.
  if (byExtension !== undefined) {
    return { kind: byExtension, byExtension, byMagicBytes, mismatch: false }
  }

  // Последняя соломинка — trying as plain text.
  return { kind: "text", byExtension, byMagicBytes, mismatch: false }
}
