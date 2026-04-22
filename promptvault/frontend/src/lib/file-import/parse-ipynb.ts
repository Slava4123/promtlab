import { z } from "zod"
import { MAX_PROMPT_CONTENT_LENGTH } from "@/lib/constants"
import { FileImportError, type ParseResult } from "./types"
import { normalizeText } from "./parse-text"

// Zod-schema для nbformat v4. Мы лояльны к незнакомым полям (passthrough),
// но минимальные обязательные — массив cells с cell_type и source.
// Spec: https://nbformat.readthedocs.io/en/latest/format_description.html
const cellSchema = z.object({
  cell_type: z.enum(["markdown", "code", "raw"]),
  // source может быть string или string[] (legacy). Нормализуем в join.
  source: z.union([z.string(), z.array(z.string())]),
  metadata: z
    .object({
      language_info: z.object({ name: z.string() }).optional(),
    })
    .passthrough()
    .optional(),
})

const notebookSchema = z.object({
  cells: z.array(cellSchema),
  metadata: z
    .object({
      // Jupyter ≥ 4.x кладёт язык в metadata.kernelspec.language или metadata.language_info.name.
      kernelspec: z.object({ language: z.string() }).passthrough().optional(),
      language_info: z.object({ name: z.string() }).passthrough().optional(),
    })
    .passthrough()
    .optional(),
  nbformat: z.number().optional(),
})

// Парсит Jupyter Notebook (.ipynb) в markdown:
//  - markdown-ячейки → source как есть
//  - code-ячейки → fenced code block с language (из metadata)
//  - raw-ячейки → skip (обычно служебное)
//  - outputs — skip (часто бинарные / картинки, редко осмысленны для промпта)
//
// Между ячейками — \n\n (стандартный markdown paragraph break).
export async function parseIpynbFile(file: File): Promise<ParseResult> {
  const text = await file.text()

  let parsed: unknown
  try {
    parsed = JSON.parse(text)
  } catch (err) {
    throw new FileImportError(
      "PARSE_FAILED",
      `Файл не является валидным JSON-notebook: ${err instanceof Error ? err.message : "unknown"}`,
      "ipynb",
    )
  }

  const validated = notebookSchema.safeParse(parsed)
  if (!validated.success) {
    throw new FileImportError(
      "PARSE_FAILED",
      `Не похоже на Jupyter Notebook (nbformat v4): ${validated.error.issues[0]?.message ?? "unknown"}`,
      "ipynb",
    )
  }

  const data = validated.data
  const defaultLang =
    data.metadata?.kernelspec?.language ??
    data.metadata?.language_info?.name ??
    ""

  const blocks: string[] = []
  for (const cell of data.cells) {
    const source = Array.isArray(cell.source) ? cell.source.join("") : cell.source
    const normalized = normalizeText(source).trim()
    if (normalized.length === 0) continue

    if (cell.cell_type === "markdown") {
      blocks.push(normalized)
    } else if (cell.cell_type === "code") {
      const cellLang = cell.metadata?.language_info?.name ?? defaultLang ?? ""
      blocks.push("```" + cellLang + "\n" + normalized + "\n```")
    }
    // raw skip
  }

  if (blocks.length === 0) {
    throw new FileImportError(
      "EMPTY_RESULT",
      "Notebook не содержит ячеек с текстом/кодом",
      "ipynb",
    )
  }

  const joined = blocks.join("\n\n")
  const { content, truncated } = truncate(joined)

  return {
    content,
    kind: "ipynb",
    filename: file.name,
    originalBytes: file.size,
    truncated,
    warnings: [],
  }
}

function truncate(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return { content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH), truncated: true }
}
