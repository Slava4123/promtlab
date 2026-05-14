// Lucide-иконки для коллекций. Синхронизировано с frontend/src/pages/collections.tsx
// — один и тот же список 15 иконок с теми же `value` ключами.

import {
  FolderOpen,
  Code,
  Palette,
  FileCode,
  Wrench,
  Rocket,
  BarChart3,
  FlaskConical,
  Shield,
  Lightbulb,
  BookOpen,
  Zap,
  MessageSquare,
  Globe,
  Database,
  type LucideIcon,
} from "lucide-react"

export interface CollectionIconOption {
  value: string
  Icon: LucideIcon
  label: string
}

export const COLLECTION_ICON_OPTIONS: CollectionIconOption[] = [
  { value: "folder", Icon: FolderOpen, label: "Общее" },
  { value: "code", Icon: Code, label: "Разработка" },
  { value: "palette", Icon: Palette, label: "Дизайн" },
  { value: "file-code", Icon: FileCode, label: "Скрипты" },
  { value: "wrench", Icon: Wrench, label: "Инструменты" },
  { value: "rocket", Icon: Rocket, label: "Продакшен" },
  { value: "chart", Icon: BarChart3, label: "Аналитика" },
  { value: "flask", Icon: FlaskConical, label: "Тестирование" },
  { value: "shield", Icon: Shield, label: "Безопасность" },
  { value: "lightbulb", Icon: Lightbulb, label: "Идеи" },
  { value: "book", Icon: BookOpen, label: "Документация" },
  { value: "zap", Icon: Zap, label: "Автоматизация" },
  { value: "message", Icon: MessageSquare, label: "Коммуникация" },
  { value: "globe", Icon: Globe, label: "Веб" },
  { value: "database", Icon: Database, label: "Базы данных" },
]

const ICON_MAP: Record<string, LucideIcon> = Object.fromEntries(
  COLLECTION_ICON_OPTIONS.map((i) => [i.value, i.Icon]),
)

export function getCollectionIcon(value?: string): LucideIcon {
  return (value && ICON_MAP[value]) || FolderOpen
}

interface CollectionIconProps {
  icon?: string
  color?: string
  size?: number
  className?: string
}

export function CollectionIcon({ icon, color, size = 16, className }: CollectionIconProps) {
  /* eslint-disable react-hooks/static-components -- Icon — lookup готового
     lucide-компонента из COLLECTION_ICONS map, не создание нового. */
  const Icon = getCollectionIcon(icon)
  return (
    <Icon
      width={size}
      height={size}
      className={className}
      style={color ? { color } : undefined}
    />
  )
  /* eslint-enable react-hooks/static-components */
}
