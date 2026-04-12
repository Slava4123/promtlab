// Парсер template variables `{{имя}}`, портирован из frontend/src/lib/template/parse.ts.
// Unicode-aware: работает с кириллицей.
//
// Правила (консистентно с основным frontend):
//   {{name}}       — валидная переменная
//   {{имя}}        — валидная (Unicode letters)
//   {{user_name}}  — валидная (underscore)
//   {{Name2}}      — валидная (цифры после первого символа)
//   {{ name }}     — НЕ переменная (пробелы)
//   {{1name}}      — НЕ переменная (начинается с цифры)

const VARIABLE_REGEX = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu;

/**
 * Возвращает массив имён переменных в порядке первого появления (без дубликатов).
 */
export function extractVariables(content: string): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const match of content.matchAll(VARIABLE_REGEX)) {
    const name = match[1];
    if (name && !seen.has(name)) {
      seen.add(name);
      result.push(name);
    }
  }
  return result;
}

/**
 * Подставляет значения переменных. Отсутствующие переменные → пустая строка.
 * Function replacer безопасен от regex metacharacters в значениях.
 */
export function renderTemplate(
  template: string,
  values: Record<string, string>,
): string {
  return template.replace(VARIABLE_REGEX, (_match, name: string) => values[name] ?? '');
}
