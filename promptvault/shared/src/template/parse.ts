/**
 * Template variables parser for prompt templates. Shared между frontend и extension.
 *
 * Grammar (BNF, Unicode-aware):
 *   template    ::= (literal | variable)*
 *   variable    ::= "{{" identifier "}}"
 *   identifier  ::= (letter | "_") (letter | digit | "_")*
 *   letter      ::= any Unicode letter (\p{L}) — latin, cyrillic, CJK, etc.
 *   digit       ::= any Unicode decimal digit (\p{N})
 *
 * Invariants:
 *   - `extractVariables` de-duplicates names, preserving first-occurrence order.
 *   - `renderTemplate` does a single pass: substituted values are NOT re-scanned.
 *   - Missing keys render as empty string, NOT as the literal placeholder.
 *   - Uses a function replacer so regex metacharacters in values are never interpreted.
 */

const VARIABLE_REGEX = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu
const HAS_VARIABLE_REGEX = /\{\{[\p{L}_][\p{L}\p{N}_]*\}\}/u

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

export function hasVariables(content: string): boolean {
  return HAS_VARIABLE_REGEX.test(content)
}

export function renderTemplate(
  template: string,
  values: Record<string, string>,
): string {
  return template.replace(VARIABLE_REGEX, (_match, name: string) => values[name] ?? "")
}
