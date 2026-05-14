import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { Command } from "cmdk"
import { Search, FileText, FolderOpen, Tag as TagIcon, Plus, X } from "lucide-react"
import { sendBg } from "../lib/bg-client"
import { useDebounced } from "../hooks/use-debounced"
import { useWorkspace } from "../hooks/use-workspace"
import type { SearchResult } from "../lib/types"

interface CommandPaletteProps {
  open: boolean
  onClose: () => void
}

// Cmd+K global command palette: search across prompts/collections/tags + quick actions.
export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState("")
  const debounced = useDebounced(query, 250)
  const navigate = useNavigate()
  const { workspaceId } = useWorkspace()
  const [results, setResults] = useState<SearchResult | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open) {
      // Reset search state при закрытии палитры (внешний event).
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setQuery("")
       
      setResults(null)
      return
    }
  }, [open])

  useEffect(() => {
    if (!open) return
    if (!debounced.trim()) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setResults(null)
      return
    }
    let cancelled = false
    setLoading(true)
    sendBg({
      type: "api.searchPrompts",
      q: debounced,
      filter: { teamId: workspaceId },
    })
      .then((r) => {
        if (!cancelled) setResults(r)
      })
      .catch(() => {
        if (!cancelled) setResults(null)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [debounced, open, workspaceId])

  function go(path: string) {
    onClose()
    navigate(path)
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50">
      {/* eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions -- modal backdrop, focus trap на диалоге */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />
      <div className="absolute left-1/2 top-1/4 w-[90%] max-w-md -translate-x-1/2 rounded-lg border border-(--color-border) bg-(--color-background) shadow-2xl">
        <Command label="Глобальный поиск" shouldFilter={false}>
          <div className="flex items-center gap-2 border-b border-(--color-border) px-3 py-2">
            <Search className="h-4 w-4 text-(--color-muted-foreground)" />
            <Command.Input
              value={query}
              onValueChange={setQuery}
              placeholder="Найти промпт, коллекцию, тег…"
              className="flex-1 bg-transparent text-sm outline-none placeholder:text-(--color-muted-foreground)"
              autoFocus
            />
            <button
              type="button"
              onClick={onClose}
              className="rounded p-0.5 text-(--color-muted-foreground) hover:bg-(--color-muted)"
              aria-label="Закрыть"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
          <Command.List className="max-h-96 overflow-y-auto p-1">
            {!query.trim() ? (
              <>
                <Command.Group heading="Быстрые действия">
                  <Command.Item value="new-prompt" onSelect={() => go("/prompts/new")}>
                    <PaletteRow icon={Plus} label="Создать промпт" />
                  </Command.Item>
                  <Command.Item value="new-collection" onSelect={() => go("/collections")}>
                    <PaletteRow icon={FolderOpen} label="Открыть коллекции" />
                  </Command.Item>
                  <Command.Item value="tags" onSelect={() => go("/tags")}>
                    <PaletteRow icon={TagIcon} label="Управление тегами" />
                  </Command.Item>
                </Command.Group>
              </>
            ) : loading ? (
              <Command.Loading>
                <div className="px-3 py-2 text-xs text-(--color-muted-foreground)">
                  Ищу…
                </div>
              </Command.Loading>
            ) : results ? (
              <>
                {results.prompts.length > 0 && (
                  <Command.Group heading="Промпты">
                    {results.prompts.slice(0, 8).map((r) => (
                      <Command.Item
                        key={`p-${r.id}`}
                        value={`p-${r.id}-${r.title}`}
                        onSelect={() => go(`/prompts/${r.id}`)}
                      >
                        <PaletteRow icon={FileText} label={r.title} description={r.description} />
                      </Command.Item>
                    ))}
                  </Command.Group>
                )}
                {results.collections.length > 0 && (
                  <Command.Group heading="Коллекции">
                    {results.collections.slice(0, 5).map((r) => (
                      <Command.Item
                        key={`c-${r.id}`}
                        value={`c-${r.id}-${r.title}`}
                        onSelect={() => go(`/collections/${r.id}`)}
                      >
                        <PaletteRow icon={FolderOpen} label={r.title} color={r.color} />
                      </Command.Item>
                    ))}
                  </Command.Group>
                )}
                {results.tags.length > 0 && (
                  <Command.Group heading="Теги">
                    {results.tags.slice(0, 5).map((r) => (
                      <Command.Item
                        key={`t-${r.id}`}
                        value={`t-${r.id}-${r.title}`}
                        onSelect={() => go(`/tags/${r.id}`)}
                      >
                        <PaletteRow icon={TagIcon} label={r.title} color={r.color} />
                      </Command.Item>
                    ))}
                  </Command.Group>
                )}
                {results.prompts.length === 0 &&
                  results.collections.length === 0 &&
                  results.tags.length === 0 && (
                    <Command.Empty>
                      <div className="px-3 py-2 text-xs text-(--color-muted-foreground)">
                        Ничего не найдено
                      </div>
                    </Command.Empty>
                  )}
              </>
            ) : null}
          </Command.List>
        </Command>
      </div>
    </div>
  )
}

function PaletteRow({
  icon: Icon,
  label,
  description,
  color,
}: {
  icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>
  label: string
  description?: string
  color?: string
}) {
  return (
    <div className="flex items-center gap-2 rounded px-2 py-1.5 text-sm aria-selected:bg-(--color-muted) cursor-pointer">
      <Icon className="h-3.5 w-3.5 shrink-0" style={color ? { color } : undefined} />
      <div className="flex-1 min-w-0">
        <div className="truncate">{label}</div>
        {description && (
          <div className="truncate text-[10px] text-(--color-muted-foreground)">{description}</div>
        )}
      </div>
    </div>
  )
}
