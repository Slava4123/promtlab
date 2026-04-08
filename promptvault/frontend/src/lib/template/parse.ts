/**
 * Template variables parser for prompt templates.
 *
 * Grammar (BNF, Unicode-aware):
 *   template    ::= (literal | variable)*
 *   variable    ::= "{{" identifier "}}"
 *   identifier  ::= (letter | "_") (letter | digit | "_")*
 *   letter      ::= any Unicode letter (\p{L}) — latin, cyrillic, CJK, etc.
 *   digit       ::= any Unicode decimal digit (\p{N})
 *
 * Examples of valid identifiers: `name`, `имя`, `имя_клиента`, `Name2`,
 *   `компания1`, `_private`, `名前`, `ИмяКомпании`.
 *
 * Anything that doesn't match `variable` is treated as literal text, including
 * non-standard forms like `{{ name }}` (with spaces), `{{my-var}}`, `{{1}}`
 * (leading digit), `{{}}` (empty), `{{has space}}`. This gives an implicit
 * escape mechanism: wrap the braces with a space (`{{ name }}`) to render them
 * literally.
 *
 * Invariants:
 *   - `extractVariables` de-duplicates names, preserving first-occurrence order.
 *   - `renderTemplate` does a single pass: substituted values are NOT re-scanned,
 *     so `{{a}} -> "{{b}}"` produces the literal `{{b}}`, not recursion.
 *   - Missing keys render as empty string, NOT as the literal placeholder.
 *   - Uses a function replacer (not string replacer) so regex metacharacters
 *     in values (`$1`, `$&`, `\`) are never interpreted.
 */

// Global regex used by extractVariables and renderTemplate.
// Each call creates a local iterator (matchAll) or passes to replace() directly,
// so the shared `lastIndex` state of the `g` flag is safe.
// The `u` flag enables Unicode property escapes \p{L} (letter) and \p{N} (digit).
const VARIABLE_REGEX = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu

// Non-global copy for test() — avoids mutating lastIndex between calls.
const HAS_VARIABLE_REGEX = /\{\{[\p{L}_][\p{L}\p{N}_]*\}\}/u

/**
 * Extracts unique variable names from a template, in order of first appearance.
 */
export function extractVariables(content: string): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  for (const match of content.matchAll(VARIABLE_REGEX)) {
    const name = match[1]
    if (name && !seen.has(name)) {
      seen.add(name)
      result.push(name)
    }
  }
  return result
}

/**
 * Returns true if the content contains at least one valid variable placeholder.
 * Cheaper than extractVariables when only existence matters.
 */
export function hasVariables(content: string): boolean {
  return HAS_VARIABLE_REGEX.test(content)
}

/**
 * Renders a template by substituting `{{name}}` placeholders with values[name].
 * Missing or empty values render as empty string.
 */
export function renderTemplate(
  template: string,
  values: Record<string, string>,
): string {
  return template.replace(VARIABLE_REGEX, (_match, name: string) => values[name] ?? "")
}
