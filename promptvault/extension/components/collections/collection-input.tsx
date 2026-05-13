import { useEffect, useRef, useState } from "react"
import { X, Plus } from "lucide-react"
import { useCollections, useCreateCollection } from "../../hooks/use-collections-crud"
import { useWorkspace } from "../../hooks/use-workspace"
import { CollectionIcon, COLLECTION_ICON_OPTIONS } from "../../lib/collection-icons"

// Палитра дефолтных цветов — mirror collections-page.tsx COLORS.
// Новые коллекции получают цвет по seed (циклически).
const DEFAULT_COLORS = [
  "#a78bfa", "#60a5fa", "#22d3ee", "#34d399",
  "#fbbf24", "#fb923c", "#f87171", "#f472b6",
]

function pickColor(seed: number): string {
  return DEFAULT_COLORS[seed % DEFAULT_COLORS.length]
}

interface CollectionInputProps {
  selectedCollectionIds: number[]
  onChange: (ids: number[]) => void
}

// Combobox с автокомплитом + inline-создание. Mirror TagInput — тот же UX
// (Enter = выбрать первое или создать, Backspace в пустом — удалить чип).
// В отличие от тегов, коллекция требует icon — берём дефолтный «folder»
// для inline-created. Сменить icon/color можно потом на странице коллекций.
export function CollectionInput({ selectedCollectionIds, onChange }: CollectionInputProps) {
  const workspace = useWorkspace()
  const teamId = workspace.workspaceId
  const { data: collections } = useCollections()
  const createCollection = useCreateCollection()
  const [input, setInput] = useState("")
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const selected = (collections ?? []).filter((c) => selectedCollectionIds.includes(c.id))
  const filtered = (collections ?? []).filter(
    (c) =>
      !selectedCollectionIds.includes(c.id) &&
      c.name.toLowerCase().includes(input.toLowerCase()),
  )
  const exactMatch = (collections ?? []).some(
    (c) => c.name.toLowerCase() === input.trim().toLowerCase(),
  )

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", onClickOutside)
    return () => document.removeEventListener("mousedown", onClickOutside)
  }, [])

  function removeCollection(id: number) {
    onChange(selectedCollectionIds.filter((cid) => cid !== id))
  }

  function addCollection(id: number) {
    onChange([...selectedCollectionIds, id])
    setInput("")
    setOpen(false)
  }

  async function handleCreate() {
    const name = input.trim()
    if (!name) return
    try {
      const color = pickColor((collections?.length ?? 0) + selectedCollectionIds.length)
      const created = await createCollection.mutateAsync({
        name,
        color,
        icon: COLLECTION_ICON_OPTIONS[0].value,
        team_id: teamId,
      })
      onChange([...selectedCollectionIds, created.id])
      setInput("")
      setOpen(false)
    } catch {
      // 409 на race-create — refetch поднимет collection, юзер выберет из autocomplete.
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") {
      e.preventDefault()
      if (filtered.length > 0) {
        addCollection(filtered[0].id)
      } else if (input.trim() && !exactMatch) {
        void handleCreate()
      }
    }
    if (e.key === "Backspace" && !input && selectedCollectionIds.length > 0) {
      removeCollection(selectedCollectionIds[selectedCollectionIds.length - 1])
    }
    if (e.key === "Escape") {
      setOpen(false)
    }
  }

  return (
    <div ref={containerRef} className="relative">
      <div
        role="button"
        tabIndex={0}
        className="flex flex-wrap items-center gap-1 rounded-md border border-(--color-border) bg-(--color-card) px-2 py-1.5 min-h-[36px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-(--color-brand)/40"
        onClick={() => setOpen(true)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault()
            setOpen(true)
          }
        }}
      >
        {selected.map((c) => (
          <span
            key={c.id}
            className="flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium"
            style={{
              backgroundColor: `${c.color || "#a78bfa"}22`,
              color: c.color || "#a78bfa",
            }}
          >
            <CollectionIcon icon={c.icon} size={11} color={c.color || "#a78bfa"} />
            {c.name}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                removeCollection(c.id)
              }}
              className="rounded p-0.5 hover:bg-(--color-foreground)/10"
              aria-label="Убрать коллекцию"
            >
              <X className="h-2.5 w-2.5" />
            </button>
          </span>
        ))}
        <input
          value={input}
          onChange={(e) => {
            setInput(e.target.value)
            setOpen(true)
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={handleKeyDown}
          placeholder={selected.length === 0 ? "Добавить коллекции…" : ""}
          className="flex-1 min-w-[80px] bg-transparent text-xs text-(--color-foreground) outline-none placeholder:text-(--color-muted-foreground)"
        />
      </div>

      {open && (filtered.length > 0 || (input.trim() && !exactMatch)) && (
        <div
          className="absolute z-50 mt-1 w-full overflow-y-auto rounded-md border border-(--color-border) bg-(--color-background) py-1 shadow-xl"
          style={{ maxHeight: "200px" }}
        >
          {filtered.map((c) => (
            <button
              key={c.id}
              type="button"
              onClick={() => addCollection(c.id)}
              className="flex w-full items-center gap-2 px-2 py-1.5 text-xs text-(--color-foreground) transition-colors hover:bg-(--color-muted)"
            >
              <CollectionIcon icon={c.icon} size={14} color={c.color || "#a78bfa"} />
              {c.name}
            </button>
          ))}
          {input.trim() && !exactMatch && (
            <button
              type="button"
              onClick={handleCreate}
              disabled={createCollection.isPending}
              className="flex w-full items-center gap-1.5 px-2 py-1.5 text-xs text-(--color-brand) transition-colors hover:bg-(--color-muted) disabled:opacity-50"
            >
              <Plus className="h-3 w-3" />
              Создать коллекцию «{input.trim()}»
            </button>
          )}
        </div>
      )}
    </div>
  )
}
