// Application-wide constants.
// ВАЖНО: значения должны синхронизироваться с backend:
//   - MAX_PROMPT_CONTENT_LENGTH → backend/internal/usecases/prompt/constants.go:MaxContentLength
//   - validator tag в delivery/http/prompt/request.go (`max=100000`)

export const MAX_PROMPT_CONTENT_LENGTH = 100000
export const MAX_PROMPT_TITLE_LENGTH = 300
export const MAX_CHANGE_NOTE_LENGTH = 300

// Пороги цветовой индикации счётчика символов.
export const CONTENT_LENGTH_WARNING = Math.floor(MAX_PROMPT_CONTENT_LENGTH * 0.75) // 75 000
export const CONTENT_LENGTH_DANGER = Math.floor(MAX_PROMPT_CONTENT_LENGTH * 0.9)   // 90 000
