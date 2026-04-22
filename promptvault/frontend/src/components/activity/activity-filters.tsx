import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

const EVENT_OPTIONS: Array<{ value: string; label: string }> = [
  { value: "", label: "Все события" },
  { value: "prompt.created", label: "Создание промпта" },
  { value: "prompt.updated", label: "Обновление промпта" },
  { value: "prompt.deleted", label: "Удаление промпта" },
  { value: "collection.created", label: "Создание коллекции" },
  { value: "collection.updated", label: "Обновление коллекции" },
  { value: "collection.deleted", label: "Удаление коллекции" },
  { value: "share.created", label: "Создание share-ссылки" },
  { value: "share.revoked", label: "Отключение share-ссылки" },
  { value: "member.added", label: "Добавление участника" },
  { value: "member.removed", label: "Удаление участника" },
  { value: "role.changed", label: "Изменение роли" },
]

interface ActivityFiltersProps {
  eventType: string
  onEventTypeChange: (v: string) => void
}

export function ActivityFilters({ eventType, onEventTypeChange }: ActivityFiltersProps) {
  // Пустая строка не поддерживается shadcn Select как value — используем "all".
  const current = eventType || "all"
  return (
    <Select
      value={current}
      onValueChange={(v: string | null) => onEventTypeChange(v && v !== "all" ? v : "")}
    >
      <SelectTrigger className="w-[220px]">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {EVENT_OPTIONS.map((opt) => (
          <SelectItem key={opt.value || "all"} value={opt.value || "all"}>
            {opt.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
