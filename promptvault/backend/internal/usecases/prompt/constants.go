package prompt

// MaxContentLength — максимальная длина поля content в символах UTF-8.
// Применяется на всех слоях:
//   - HTTP DTO (delivery/http/prompt/request.go, delivery/http/ai/request.go) через validator tag `max=15000`
//   - Frontend Zod schema (frontend/src/lib/constants.ts → MAX_PROMPT_CONTENT_LENGTH)
//   - UI счётчик символов (prompt-editor.tsx)
//
// При изменении — обновить все 3 места И frontend/src/lib/constants.ts.
const MaxContentLength = 15000

// MaxTitleLength — максимальная длина поля title.
const MaxTitleLength = 300

// MaxChangeNoteLength — максимальная длина заметки к версии.
const MaxChangeNoteLength = 300
