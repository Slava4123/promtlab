import { useState, useEffect, useCallback } from "react"
import { useNavigate } from "react-router-dom"
import { FileText, FolderOpen, Settings, Plus, Tag } from "lucide-react"

import {
  CommandDialog,
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandSeparator,
} from "@/components/ui/command"
import { useSearch } from "@/hooks/use-search"
import { useWorkspaceStore } from "@/stores/workspace-store"

export function CommandPalette() {
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const [debouncedQuery, setDebouncedQuery] = useState("")
  const navigate = useNavigate()

  // Debounce 300ms
  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(t)
  }, [query])

  // Global Cmd+K / Ctrl+K
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault()
        setOpen((prev) => !prev)
      }
    }
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [])

  const { data } = useSearch(debouncedQuery, teamId)

  const runAction = useCallback(
    (cb: () => void) => {
      setOpen(false)
      setQuery("")
      cb()
    },
    [],
  )

  const hasResults =
    data && (data.prompts.length > 0 || data.collections.length > 0 || data.tags.length > 0)

  return (
    <CommandDialog
      open={open}
      onOpenChange={setOpen}
      title="Поиск"
      description="Поиск по промптам, коллекциям и тегам"
    >
      <Command shouldFilter={false}>
        <CommandInput
          placeholder="Поиск..."
          value={query}
          onValueChange={setQuery}
        />
        <CommandList>
          <CommandEmpty>
            {debouncedQuery.length >= 2 ? "Ничего не найдено" : "Начните вводить для поиска..."}
          </CommandEmpty>

          {/* Результаты поиска */}
          {hasResults && (
            <>
              {data.prompts.length > 0 && (
                <CommandGroup heading="Промпты">
                  {data.prompts.map((item) => (
                    <CommandItem
                      key={`p-${item.id}`}
                      onSelect={() => runAction(() => navigate(`/prompts/${item.id}`))}
                    >
                      <FileText className="text-violet-400" />
                      <div className="flex flex-col gap-0.5 overflow-hidden">
                        <span className="truncate">{item.title}</span>
                        {item.description && (
                          <span className="truncate text-xs text-muted-foreground">
                            {item.description}
                          </span>
                        )}
                      </div>
                    </CommandItem>
                  ))}
                </CommandGroup>
              )}

              {data.collections.length > 0 && (
                <CommandGroup heading="Коллекции">
                  {data.collections.map((item) => (
                    <CommandItem
                      key={`c-${item.id}`}
                      onSelect={() => runAction(() => navigate(`/collections/${item.id}`))}
                    >
                      <span
                        className="inline-block h-3 w-3 shrink-0 rounded-full"
                        style={{ backgroundColor: item.color || "#6366f1" }}
                      />
                      <span className="truncate">{item.title}</span>
                    </CommandItem>
                  ))}
                </CommandGroup>
              )}

              {data.tags.length > 0 && (
                <CommandGroup heading="Теги">
                  {data.tags.map((item) => (
                    <CommandItem
                      key={`t-${item.id}`}
                      onSelect={() => runAction(() => navigate(`/?tag_ids=${item.id}`))}
                    >
                      <Tag className="shrink-0" style={{ color: item.color || "#6366f1" }} />
                      <span className="truncate">{item.title}</span>
                    </CommandItem>
                  ))}
                </CommandGroup>
              )}

              <CommandSeparator />
            </>
          )}

          {/* Навигация — всегда видна */}
          <CommandGroup heading="Навигация">
            <CommandItem onSelect={() => runAction(() => navigate("/dashboard"))}>
              <FileText />
              <span>Промпты</span>
            </CommandItem>
            <CommandItem onSelect={() => runAction(() => navigate("/collections"))}>
              <FolderOpen />
              <span>Коллекции</span>
            </CommandItem>
            <CommandItem onSelect={() => runAction(() => navigate("/settings"))}>
              <Settings />
              <span>Настройки</span>
            </CommandItem>
            <CommandItem onSelect={() => runAction(() => navigate("/prompts/new"))}>
              <Plus />
              <span>Новый промпт</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </CommandDialog>
  )
}
