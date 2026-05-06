// BRAND_COLORS — preset для color-palette-picker'а на странице брендинга.
// 12 hex-цветов, выбраны под dual-mode (читаются и на светлом, и на тёмном
// фоне публичной share-страницы), без чисто-чёрного/чисто-белого. Аналог
// COLORS из pages/collections.tsx, но скорректированный под «brand-tone»:
//   - synthwave-violet и indigo для tech-команд
//   - вибрант (red/orange/pink) для маркетинга/контента
//   - emerald/teal/sky для финансов и data
//   - slate как нейтральный fallback
//
// Если будете расширять список — держите ≥3 контрастных пары относительно
// background popover'а в обеих темах (eyeball-test или WCAG AA для текста).
export interface BrandColor {
  value: string
  label: string
}

export const BRAND_COLORS: BrandColor[] = [
  { value: "#0066CC", label: "Корпоративный синий" },
  { value: "#2563EB", label: "Indigo" },
  { value: "#6366F1", label: "Brand Violet" },
  { value: "#8B5CF6", label: "Violet" },
  { value: "#DB2777", label: "Pink" },
  { value: "#EF4444", label: "Red" },
  { value: "#F97316", label: "Orange" },
  { value: "#F59E0B", label: "Amber" },
  { value: "#10B981", label: "Emerald" },
  { value: "#14B8A6", label: "Teal" },
  { value: "#0EA5E9", label: "Sky" },
  { value: "#475569", label: "Slate" },
]

// HEX_REGEX — формат #RRGGBB (без альфы, без короткой #RGB-формы).
// Симметричен серверной валидации в backend/internal/usecases/team/branding.go.
export const HEX_REGEX = /^#[0-9a-fA-F]{6}$/

// isValidHex — true если строка соответствует #RRGGBB.
export function isValidHex(s: string): boolean {
  return HEX_REGEX.test(s)
}
