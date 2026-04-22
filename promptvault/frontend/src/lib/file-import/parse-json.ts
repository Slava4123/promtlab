import { z } from "zod"
import {
  MAX_PROMPT_CONTENT_LENGTH,
  MAX_PROMPT_TITLE_LENGTH,
} from "@/lib/constants"
import { FileImportError, type ParseResult } from "./types"
import { normalizeText } from "./parse-text"

// Zod-схема для prompt-JSON. content — обязательное non-empty string; title
// и model — опциональные короткие строки. Дополнительные поля в JSON
// игнорируются (passthrough — не strict, чтобы toleratable parse expanded
// форматов в будущем).
const promptJsonSchema = z.object({
  content: z.string().min(1, "Поле content обязательно и не может быть пустым"),
  title: z.string().max(MAX_PROMPT_TITLE_LENGTH).optional(),
  model: z.string().max(100).optional(),
})

type PromptJson = z.infer<typeof promptJsonSchema>

// Парсит JSON-файл. Две стратегии:
//
// 1. Prompt-JSON (строго по Zod): { content: string, title?: string, model?: string } —
//    возвращает content + metadata.title/model. Caller (prompt-editor) сразу
//    заполнит title/model поля формы через setValue.
//
// 2. Любой другой валидный JSON (не соответствует схеме) — pretty-print
//    и вставляется как plain content + warning с примером правильного формата.
//
// Invalid JSON → FileImportError("PARSE_FAILED").
export async function parseJsonFile(file: File): Promise<ParseResult> {
  const text = await file.text()
  const normalized = normalizeText(text)

  let parsed: unknown
  try {
    parsed = JSON.parse(normalized)
  } catch (err) {
    throw new FileImportError(
      "PARSE_FAILED",
      `Файл не является валидным JSON: ${err instanceof Error ? err.message : "unknown"}`,
      "json",
    )
  }

  // Case 1: prompt-JSON shape (валидация через Zod).
  const parsedResult = promptJsonSchema.safeParse(parsed)
  if (parsedResult.success) {
    return buildPromptJsonResult(parsedResult.data, file)
  }

  // Case 2: generic JSON → pretty-print как fallback + warning.
  const pretty = JSON.stringify(parsed, null, 2)
  const { content, truncated } = truncate(pretty)
  return {
    content,
    kind: "json",
    filename: file.name,
    originalBytes: file.size,
    truncated,
    warnings: [
      'JSON не соответствует формату промпта — вставлен как отформатированный текст. ' +
        'Ожидаемый формат: {"content": "текст промпта", "title": "название", "model": "модель"}',
    ],
  }
}

function buildPromptJsonResult(data: PromptJson, file: File): ParseResult {
  const { content, truncated } = truncate(data.content)
  const metadata: NonNullable<ParseResult["metadata"]> = {}
  if (data.title && data.title.length > 0) metadata.title = data.title
  if (data.model && data.model.length > 0) metadata.model = data.model
  return {
    content,
    kind: "json",
    filename: file.name,
    originalBytes: file.size,
    truncated,
    warnings: [],
    ...(Object.keys(metadata).length > 0 ? { metadata } : {}),
  }
}

function truncate(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return { content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH), truncated: true }
}

// Re-export для тестов.
export { promptJsonSchema }
