// Palette для известных семейств моделей. Неузнанные модели получают серый.
// Извлечено из model-segmentation-chart.tsx для переиспользования
// в models-donut.tsx (analytics redesign 2026-05-17).
// `sonnet-N / opus-N / haiku-N` ловим явно, потому что backend хранит
// короткое имя модели без `claude-` префикса (например, `sonnet-4`,
// `haiku-4-5`) — старый regex `/^claude/i` их пропускал и обе попадали
// в DEFAULT_COLOR серый, donut выглядел нейтральным «серый + серый».
export const MODEL_COLORS: Array<[RegExp, string]> = [
  [/^claude|^(sonnet|opus|haiku)/i, "#cc7a3e"], // оранж-коричневый как Anthropic brand
  [/^gpt|^o1|^o3/i, "#10a37f"], // зелёный как OpenAI
  [/deepseek/i, "#4a7fff"], // синий
  [/gemini|google/i, "#8ab4f8"],
  [/llama|meta/i, "#0668e1"],
  [/mistral/i, "#ff7000"],
]

export const DEFAULT_COLOR = "#a78bfa" // brand violet для «Без модели» и неопознанных

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
