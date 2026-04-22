import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

// value="all" — sentinel для "без фильтра" (shadcn Select не принимает пустую строку).
// На onChange конвертируется в "" перед отправкой на backend.
const EVENT_OPTIONS: Array<{ value: string; label: string }> = [
  { value: "all", label: "Все события" },
  { value: "prompt.created", label: "Создание промпта" },
  { value: "prompt.updated", label: "Обновление промпта" },
  { value: "prompt.deleted", label: "Удаление промпта" },
  { value: "collection.created", label: "Создание коллекции" },
  { value: "collection.updated", label: "Обновление коллекции" },
  { value: "collection.deleted", label: "Удаление коллекции" },
  { value: "share.created", label: "Создание публичной ссылки" },
  { value: "share.revoked", label: "Отключение публичной ссылки" },
  { value: "member.added", label: "Добавление участника" },
  { value: "member.removed", label: "Удаление участника" },
  { value: "role.changed", label: "Изменение роли" },
]

interface ActivityFiltersProps {
  eventType: string
  onEventTypeChange: (v: string) => void
}

export function ActivityFilters({ eventType, onEventTypeChange }: ActivityFiltersProps) {
  const current = eventType || "all"
  // base-ui SelectValue без children показывает сырой value. Резолвим label
  // через render-function по value → label из EVENT_OPTIONS.
  return (
    <Select
      value={current}
      onValueChange={(v: string | null) => onEventTypeChange(v && v !== "all" ? v : "")}
    >
      <SelectTrigger className="w-[220px]">
        <SelectValue>
          {(value: string) => EVENT_OPTIONS.find((o) => o.value === value)?.label ?? "Все события"}
        </SelectValue>
      </SelectTrigger>
      {/*
        alignItemWithTrigger=false — иначе base-ui совмещает выбранный item
        с уровнем trigger, и popup уезжает вверх, если выбран не первый
        элемент. Для длинного списка фильтров это ломает UX.
      */}
      <SelectContent alignItemWithTrigger={false}>
        {EVENT_OPTIONS.map((opt) => (
          <SelectItem key={opt.value} value={opt.value}>
            {opt.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
