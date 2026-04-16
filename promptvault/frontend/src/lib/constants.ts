// Application-wide constants.
// ВАЖНО: значения должны синхронизироваться с backend:
//   - MAX_PROMPT_CONTENT_LENGTH → backend/internal/usecases/prompt/constants.go:MaxContentLength
//   - validator tags в delivery/http/ai/request.go и prompt/request.go (`max=15000`)

export const MAX_PROMPT_CONTENT_LENGTH = 15000
export const MAX_PROMPT_TITLE_LENGTH = 300
export const MAX_CHANGE_NOTE_LENGTH = 300

// Пороги цветовой индикации счётчика символов.
export const CONTENT_LENGTH_WARNING = Math.floor(MAX_PROMPT_CONTENT_LENGTH * 0.75) // 11 250
export const CONTENT_LENGTH_DANGER = Math.floor(MAX_PROMPT_CONTENT_LENGTH * 0.9)   // 13 500
