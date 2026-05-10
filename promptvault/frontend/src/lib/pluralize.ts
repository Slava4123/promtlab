/**
 * Russian plural form selector.
 *
 * Russian uses three plural forms based on the last 1-2 digits:
 *   - one:  для 1, 21, 31, ..., 101, 121 (но НЕ 11) — «1 использование»
 *   - few:  для 2-4, 22-24, 32-34, ..., 102-104 (но НЕ 12-14) — «2 использования»
 *   - many: для 0, 5-20, 25-30, 11-14, 100, 111-114, ... — «5 использований»
 *
 * Logic (Unicode CLDR plural rules for Russian):
 *   mod10 = n % 10
 *   mod100 = n % 100
 *   if mod10 == 1 && mod100 != 11           → one
 *   if mod10 in 2..4 && mod100 not in 12..14 → few
 *   else                                     → many
 *
 * Negative numbers are normalized via Math.abs — `-1` returns the `one` form.
 * Non-integer numbers are floored — `1.5` returns the `one` form.
 *
 * Source: CLDR plural rules for `ru` —
 *   https://cldr.unicode.org/index/cldr-spec/plural-rules
 *
 * @example
 *   pluralizeRu(1,  "использование", "использования", "использований") // "использование"
 *   pluralizeRu(2,  ...) // "использования"
 *   pluralizeRu(5,  ...) // "использований"
 *   pluralizeRu(11, ...) // "использований"  (особый случай: 11-14 → many)
 *   pluralizeRu(21, ...) // "использование"  (заканчивается на 1, но не 11)
 */
export function pluralizeRu(
  n: number,
  one: string,
  few: string,
  many: string,
): string {
  const abs = Math.abs(Math.trunc(n))
  const mod10 = abs % 10
  const mod100 = abs % 100

  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}
