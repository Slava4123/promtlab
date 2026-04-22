import type { FileKind } from "./types"

// Hard-cap на размер файла до FileReader. Защита вкладки от DoS огромным файлом.
// 10 MB покрывает ~95% реалистичных кейсов (5-10 страниц PDF, типовой .docx,
// крупный .md с картинками-embedded-base64 ещё влезет).
export const MAX_UPLOAD_BYTES = 10 * 1024 * 1024

// Строка для <input type="file" accept="...">. Фильтрует picker, но пользователь
// может выбрать "All files" — всё равно валидируем magic-bytes в detect.ts.
export const ACCEPTED_FILE_EXTENSIONS = [
  ".txt",
  ".md",
  ".markdown",
  ".json",
  ".ipynb",
  ".html",
  ".htm",
  ".rtf",
  ".docx",
  ".pdf",
].join(",")

// Человеко-читаемые описания для сообщений об ошибках и UI.
export const FILE_KIND_LABELS: Record<FileKind, string> = {
  text: "Текстовый файл",
  markdown: "Markdown",
  json: "JSON",
  ipynb: "Jupyter Notebook",
  html: "HTML",
  rtf: "RTF",
  docx: "Word (DOCX)",
  pdf: "PDF",
}
