import { z } from "zod"

// Source mirror: backend/internal/usecases/prompt/types.go.
export const MAX_PROMPT_TITLE_LENGTH = 100
export const MAX_PROMPT_CONTENT_LENGTH = 50_000
export const CONTENT_LENGTH_WARNING = Math.floor(MAX_PROMPT_CONTENT_LENGTH * 0.7)

export const promptSchema = z.object({
  title: z
    .string()
    .min(1, "Заполните название")
    .max(MAX_PROMPT_TITLE_LENGTH, `Максимум ${MAX_PROMPT_TITLE_LENGTH} символов`),
  content: z
    .string()
    .min(1, "Заполните содержимое")
    .max(MAX_PROMPT_CONTENT_LENGTH, `Максимум ${MAX_PROMPT_CONTENT_LENGTH} символов`),
  // Backend max=2000 (CreatePromptRequest validation после миграции 000071).
  description: z.string().max(2000, "Максимум 2000 символов").optional(),
  model: z.string().optional(),
  collection_ids: z.array(z.number()).optional(),
  tag_ids: z.array(z.number()).optional(),
  team_id: z.number().nullable().optional(),
  is_public: z.boolean().optional(),
  change_note: z.string().max(500).optional(),
})

export type PromptFormValues = z.infer<typeof promptSchema>
