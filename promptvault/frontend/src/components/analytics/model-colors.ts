// Palette для известных семейств моделей. Неузнанные модели получают серый.
// Извлечено из model-segmentation-chart.tsx для переиспользования
// в models-donut.tsx (analytics redesign 2026-05-17).
export const MODEL_COLORS: Array<[RegExp, string]> = [
  [/^claude/i, "#cc7a3e"], // оранж-коричневый как Anthropic brand
  [/^gpt/i, "#10a37f"], // зелёный как OpenAI
  [/deepseek/i, "#4a7fff"], // синий
  [/gemini|google/i, "#8ab4f8"],
  [/llama|meta/i, "#0668e1"],
  [/mistral/i, "#ff7000"],
]

export const DEFAULT_COLOR = "#94a3b8" // серый для «Без модели» и неопознанных

export const UNKNOWN_MODEL_HINT =
  "Промпты, в которых при создании не указана target-модель в редакторе"

export function colorFor(model: string): string {
  for (const [re, color] of MODEL_COLORS) {
    if (re.test(model)) return color
  }
  return DEFAULT_COLOR
}

// Backend агрегирует строки с пустой `prompts.model` под пустой строкой.
// Для пользователя показываем расшифровку — это не legacy и не баг.
export function labelFor(model: string): string {
  return model === "" ? "Модель не указана" : model
}
