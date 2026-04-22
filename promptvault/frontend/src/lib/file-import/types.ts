// Типы публичного API file-import. Не зависят от конкретных парсеров —
// импортируются как clientами (prompt-editor), так и внутренними парсерами.

export type FileKind =
  | "text"      // .txt и любые text/plain
  | "markdown"  // .md, .markdown
  | "json"      // .json
  | "ipynb"     // .ipynb — Jupyter notebook (v2+)
  | "html"      // .html, .htm (v2+)
  | "rtf"       // .rtf (v2+)
  | "docx"      // .docx (v2+)
  | "pdf"       // .pdf (v2+)

export type FileImportErrorCode =
  | "SIZE_EXCEEDED"    // file.size > MAX_UPLOAD_BYTES до парсинга
  | "UNSUPPORTED"      // не удалось определить FileKind или он не в списке поддерживаемых
  | "MAGIC_MISMATCH"   // magic-byte не совпал с расширением (возможная подмена типа)
  | "PARSE_FAILED"     // парсер бросил исключение (malformed file)
  | "EMPTY_RESULT"     // парсер отработал, но content пустой (напр. scan-only PDF)

export class FileImportError extends Error {
  readonly code: FileImportErrorCode
  readonly kind?: FileKind

  constructor(code: FileImportErrorCode, message: string, kind?: FileKind) {
    super(message)
    this.name = "FileImportError"
    this.code = code
    this.kind = kind
  }
}

export interface ParseResult {
  content: string
  kind: FileKind
  filename: string
  originalBytes: number
  // Обрезали ли текст до MAX_PROMPT_CONTENT_LENGTH (ответственность — caller решает, показывать ли диалог truncation).
  truncated: boolean
  // Ненулевые только для соответствующих парсеров.
  detectedEncoding?: string       // .txt/.md, заполняется Commit 9
  encodingConfidence?: number     // 0..1, заполняется Commit 9
  pages?: number                  // .pdf
  // Короткие предупреждения для toast.warning (напр. "PDF не содержит текстового слоя").
  warnings: string[]
  // Заполняется только при парсинге prompt-JSON (Commit 3) — {title, content, model} → позволяет caller'у обновить
  // ещё title/model, а не только content.
  metadata?: {
    title?: string
    model?: string
  }
}
