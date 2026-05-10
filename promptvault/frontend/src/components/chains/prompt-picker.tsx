// PromptPicker — combobox с поиском промптов через Popover + Command (cmdk).
// Отображает recent промпты при открытии, поиск по title при наборе ≥1 символа.
//
// Используется в editor.tsx (классический редактор) и canvas Sheet sidebar (R3d).

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Check, ChevronsUpDown, FileText, Search } from "lucide-react"

import { api } from "@/api/client"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"
import type { PaginatedResponse, Prompt } from "@/api/types"

interface PromptPickerProps {
  /** Выбранный prompt id; null если ничего не выбрано. */
  value: number | null
  /** Callback при выборе промпта. */
  onChange: (promptID: number, prompt: Prompt) => void
  /** Опционально — id команды; если задан, ищем в командных промптах. */
  teamID?: number | null
  /** Текст плейсхолдера в кнопке-триггере. */
  placeholder?: string
  /** Заданный заголовок выбранного промпта (избегает доп. fetch если уже знаем). */
  selectedTitle?: string
  className?: string
}

export function PromptPicker({
  value,
  onChange,
  teamID,
  placeholder = "Выбрать промпт…",
  selectedTitle,
  className,
}: PromptPickerProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")

  // Поиск через GET /api/prompts?q=...&page_size=20. Бэкенд ищет по title+content.
  const { data, isLoading } = useQuery({
    queryKey: ["prompts-picker", query, teamID],
    queryFn: () => {
      const params = new URLSearchParams()
      if (query.trim()) params.set("q", query.trim())
      params.set("page_size", "20")
      params.set("page", "1")
      if (teamID) params.set("team_id", String(teamID))
      return api<PaginatedResponse<Prompt>>(`/prompts?${params}`)
    },
    staleTime: 30_000,
    enabled: open, // запрашиваем только когда picker открыт
    placeholderData: (prev) => prev,
  })

  const prompts = data?.items ?? []
  const buttonLabel = value && selectedTitle ? selectedTitle : value ? `Промпт #${value}` : placeholder

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn("w-full justify-between font-normal", !value && "text-muted-foreground", className)}
        >
          <span className="line-clamp-1 flex items-center gap-2">
            <FileText className="h-4 w-4 shrink-0" />
            {buttonLabel}
          </span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
        <Command shouldFilter={false}>
          <div className="flex items-center border-b px-3" cmdk-input-wrapper="">
            <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
            <CommandInput
              placeholder="Поиск по названию и содержимому…"
              value={query}
              onValueChange={setQuery}
              className="flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
          <CommandList>
            {isLoading && <div className="py-6 text-center text-sm text-muted-foreground">Загрузка…</div>}
            {!isLoading && prompts.length === 0 && (
              <CommandEmpty>
                {query.trim() ? `Нет промптов по «${query}»` : "У вас пока нет промптов"}
              </CommandEmpty>
            )}
            {prompts.length > 0 && (
              <CommandGroup heading={query.trim() ? "Результаты" : "Недавние"}>
                {prompts.map((p) => (
                  <CommandItem
                    key={p.id}
                    value={`${p.id}-${p.title}`}
                    onSelect={() => {
                      onChange(p.id, p)
                      setOpen(false)
                      setQuery("")
                    }}
                    className="flex items-start gap-2"
                  >
                    <Check
                      className={cn(
                        "mt-0.5 h-4 w-4 shrink-0",
                        value === p.id ? "opacity-100" : "opacity-0",
                      )}
                    />
                    <div className="flex-1 overflow-hidden">
                      <p className="line-clamp-1 text-sm font-medium">{p.title}</p>
                      <p className="line-clamp-1 text-xs text-muted-foreground">
                        {p.content.slice(0, 80) || "—"}
                      </p>
                    </div>
                    {p.tags && p.tags.length > 0 && (
                      <div className="flex shrink-0 gap-1">
                        {p.tags.slice(0, 2).map((t) => (
                          <span
                            key={t.id}
                            className="rounded px-1.5 py-0.5 text-[10px]"
                            style={{ backgroundColor: t.color + "20", color: t.color }}
                          >
                            {t.name}
                          </span>
                        ))}
                      </div>
                    )}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
