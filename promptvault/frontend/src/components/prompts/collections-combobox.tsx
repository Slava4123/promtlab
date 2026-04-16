import { useMemo, useState } from "react"
import { Check, Search, X } from "lucide-react"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import type { Collection } from "@/api/types"

interface CollectionsComboboxProps {
  collections: Collection[] | undefined
  value: number[]
  onChange: (ids: number[]) => void
}

/**
 * X-8: Multi-select combobox для коллекций в promt-editor.
 *
 * Заменил кастомный chip-picker в prompt-editor.tsx: тот показывал все
 * коллекции плоским списком с условным поиском и "Ещё N+" складыванием —
 * работало плохо при 20+ коллекциях. Combobox:
 *  - chips выбранных всегда сверху (с × для удаления), видно что выбрано
 *    даже при длинном списке;
 *  - cmdk фильтрация + клавиатурная навигация (↑↓/Enter) из коробки;
 *  - скроллируемый список (CommandList max-h-60) — не разрастается на всю
 *    страницу при сотне коллекций.
 *
 * Popover не используем специально: список живёт inline в форме, что
 * снижает когнитивную нагрузку на мобильных и не конфликтует с Dialog'ами.
 */
export function CollectionsCombobox({ collections, value, onChange }: CollectionsComboboxProps) {
  const [search, setSearch] = useState("")

  const byId = useMemo(() => {
    const m = new Map<number, Collection>()
    for (const c of collections ?? []) m.set(c.id, c)
    return m
  }, [collections])

  const selected = useMemo(
    () => value.map((id) => byId.get(id)).filter((c): c is Collection => !!c),
    [value, byId],
  )

  const toggle = (id: number) => {
    onChange(value.includes(id) ? value.filter((x) => x !== id) : [...value, id])
  }

  const empty = !collections || collections.length === 0

  return (
    <div className="space-y-2">
      {selected.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {selected.map((c) => (
            <button
              key={c.id}
              type="button"
              onClick={() => toggle(c.id)}
              className="group flex items-center gap-1 rounded-md px-2 py-1 text-[0.75rem] font-medium transition-colors"
              style={{ background: `${c.color}18`, boxShadow: `inset 0 0 0 1px ${c.color}30`, color: c.color }}
              aria-label={`Убрать ${c.name}`}
            >
              {c.name}
              <X className="h-3 w-3 opacity-60 transition-opacity group-hover:opacity-100" />
            </button>
          ))}
        </div>
      )}
      <Command className="rounded-lg border border-border bg-background" shouldFilter={!empty}>
        {!empty && (
          <div className="flex items-center gap-2 border-b border-border px-2.5 py-1.5">
            <Search className="h-3 w-3 text-muted-foreground" aria-hidden="true" />
            <CommandInput
              value={search}
              onValueChange={setSearch}
              placeholder="Найти коллекцию..."
              className="h-7 border-0 p-0 text-[0.75rem] outline-none focus:ring-0"
            />
          </div>
        )}
        <CommandList className="max-h-60">
          {empty ? (
            <div className="px-3 py-4 text-[0.8rem] text-muted-foreground">Нет коллекций</div>
          ) : (
            <>
              <CommandEmpty>Ничего не найдено</CommandEmpty>
              <CommandGroup>
                {collections!.map((c) => {
                  const isSelected = value.includes(c.id)
                  return (
                    <CommandItem
                      key={c.id}
                      value={c.name}
                      onSelect={() => toggle(c.id)}
                      className="flex cursor-pointer items-center gap-2 text-[0.8rem]"
                    >
                      <span
                        aria-hidden="true"
                        className="h-2 w-2 shrink-0 rounded-full"
                        style={{ background: c.color }}
                      />
                      <span className="flex-1 truncate">{c.name}</span>
                      {isSelected && <Check className="h-3.5 w-3.5 text-violet-400" aria-hidden="true" />}
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            </>
          )}
        </CommandList>
      </Command>
    </div>
  )
}
