// Russian declension helpers — для UI лейблов типа "5 промптов" / "1 промпт".

export function plural3(n: number, one: string, few: string, many: string): string {
  const mod10 = Math.abs(n) % 10
  const mod100 = Math.abs(n) % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}

export function pluralAfterDo(n: number, one: string, few: string, many: string): string {
  return `${n} ${plural3(n, one, few, many)}`
}
