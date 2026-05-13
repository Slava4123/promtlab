import { useEffect, useRef, useState } from "react"
import { X, Plus } from "lucide-react"
import { useTags, useCreateTag } from "../../hooks/use-tags-crud"
import { useWorkspace } from "../../hooks/use-workspace"

// Палитра дефолтных цветов для новых тегов — циклически.
const DEFAULT_COLORS = [
  "#7c3aed", "#dc2626", "#ea580c", "#d97706",
  "#65a30d", "#059669", "#0891b2", "#2563eb",
  "#4f46e5", "#9333ea", "#c026d3", "#db2777",
]

function pickColor(seed: number): string {
  return DEFAULT_COLORS[seed % DEFAULT_COLORS.length]
}

interface TagInputProps {
  selectedTagIds: number[]
  onChange: (ids: number[]) => void
}

// Combobox с автокомплитом + inline-создание. По образу frontend tag-input.
// Enter — выбрать первое предложение или создать (если такого нет).
// Backspace в пустом input — удалить последний chip.
export function TagInput({ selectedTagIds, onChange }: TagInputProps) {
  const workspace = useWorkspace()
  const teamId = workspace.workspaceId
  const { data: tags } = useTags()
  const createTag = useCreateTag()
  const [input, setInput] = useState("")
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const selectedTags = (tags ?? []).filter((t) => selectedTagIds.includes(t.id))
  const filtered = (tags ?? []).filter(
    (t) =>
      !selectedTagIds.includes(t.id) &&
      t.name.toLowerCase().includes(input.toLowerCase()),
  )
  const exactMatch = (tags ?? []).some(
    (t) => t.name.toLowerCase() === input.trim().toLowerCase(),
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

  function removeTag(id: number) {
    onChange(selectedTagIds.filter((tid) => tid !== id))
  }

  function addTag(id: number) {
    onChange([...selectedTagIds, id])
    setInput("")
    setOpen(false)
  }

  async function handleCreate() {
    const name = input.trim()
    if (!name) return
    try {
      const color = pickColor((tags?.length ?? 0) + selectedTagIds.length)
      const tag = await createTag.mutateAsync({
        name,
        color,
        team_id: teamId,
      })
      onChange([...selectedTagIds, tag.id])
      setInput("")
      setOpen(false)
    } catch {
      // backend может вернуть 409 если тег уже создан в параллели — refetch
      // поднимет его в `tags`, юзер выберет из autocomplete.
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") {
      e.preventDefault()
      if (filtered.length > 0) {
        addTag(filtered[0].id)
      } else if (input.trim() && !exactMatch) {
        void handleCreate()
      }
    }
    if (e.key === "Backspace" && !input && selectedTagIds.length > 0) {
      removeTag(selectedTagIds[selectedTagIds.length - 1])
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
        className="flex flex-wrap items-center gap-1 rounded-md border border-(--color-border) bg-(--color-card) px-2 py-1.5 min-h-[36px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-(--color-primary)/40"
        onClick={() => setOpen(true)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault()
            setOpen(true)
          }
        }}
      >
        {selectedTags.map((tag) => (
          <span
            key={tag.id}
            className="flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium"
            style={{
              backgroundColor: `${tag.color || "#7c3aed"}22`,
              color: tag.color || "#7c3aed",
            }}
          >
            {tag.name}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                removeTag(tag.id)
              }}
              className="rounded p-0.5 hover:bg-(--color-foreground)/10"
              aria-label="Убрать тег"
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
          placeholder={selectedTags.length === 0 ? "Добавить теги…" : ""}
          className="flex-1 min-w-[80px] bg-transparent text-xs text-(--color-foreground) outline-none placeholder:text-(--color-muted-foreground)"
        />
      </div>

      {open && (filtered.length > 0 || (input.trim() && !exactMatch)) && (
        <div
          className="absolute z-50 mt-1 w-full overflow-y-auto rounded-md border border-(--color-border) bg-(--color-background) py-1 shadow-xl"
          style={{ maxHeight: "200px" }}
        >
          {filtered.map((tag) => (
            <button
              key={tag.id}
              type="button"
              onClick={() => addTag(tag.id)}
              className="flex w-full items-center gap-2 px-2 py-1.5 text-xs text-(--color-foreground) transition-colors hover:bg-(--color-muted)"
            >
              <span
                className="h-2 w-2 shrink-0 rounded-full"
                style={{ backgroundColor: tag.color || "#7c3aed" }}
              />
              {tag.name}
            </button>
          ))}
          {input.trim() && !exactMatch && (
            <button
              type="button"
              onClick={handleCreate}
              disabled={createTag.isPending}
              className="flex w-full items-center gap-1.5 px-2 py-1.5 text-xs text-(--color-brand) transition-colors hover:bg-(--color-muted) disabled:opacity-50"
            >
              <Plus className="h-3 w-3" />
              Создать «{input.trim()}»
            </button>
          )}
        </div>
      )}
    </div>
  )
}
