import { useState, useRef, useEffect } from "react"
import { X, Plus } from "lucide-react"
import { useTags, useCreateTag } from "@/hooks/use-tags"
import { useWorkspaceStore } from "@/stores/workspace-store"

interface TagInputProps {
  selectedTagIds: number[]
  onChange: (ids: number[]) => void
}

export function TagInput({ selectedTagIds, onChange }: TagInputProps) {
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const { data: tags } = useTags(teamId)
  const createTag = useCreateTag()
  const [input, setInput] = useState("")
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const selectedTags = tags?.filter((t) => selectedTagIds.includes(t.id)) || []
  const filtered = tags?.filter(
    (t) =>
      !selectedTagIds.includes(t.id) &&
      t.name.toLowerCase().includes(input.toLowerCase()),
  ) || []
  const exactMatch = tags?.some((t) => t.name.toLowerCase() === input.trim().toLowerCase())

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClick)
    return () => document.removeEventListener("mousedown", handleClick)
  }, [])

  const removeTag = (id: number) => {
    onChange(selectedTagIds.filter((tid) => tid !== id))
  }

  const addTag = (id: number) => {
    onChange([...selectedTagIds, id])
    setInput("")
    setOpen(false)
  }

  const handleCreate = async () => {
    const name = input.trim()
    if (!name) return
    try {
      const tag = await createTag.mutateAsync({ name, team_id: teamId })
      onChange([...selectedTagIds, tag.id])
      setInput("")
      setOpen(false)
    } catch {
      // tag already exists, ignore
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault()
      if (filtered.length > 0) {
        addTag(filtered[0].id)
      } else if (input.trim() && !exactMatch) {
        handleCreate()
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
        className="flex flex-wrap items-center gap-1.5 rounded-lg border border-border bg-background px-3 py-2 min-h-[40px]"
        onClick={() => setOpen(true)}
      >
        {selectedTags.map((tag) => (
          <span
            key={tag.id}
            className="flex items-center gap-1 rounded-full px-2 py-[2px] text-[0.72rem] font-medium"
            style={{
              backgroundColor: (tag.color || "#6366f1") + "18",
              color: (tag.color || "#6366f1") + "cc",
            }}
          >
            {tag.name}
            <button
              type="button"
              onClick={(e) => { e.stopPropagation(); removeTag(tag.id) }}
              className="ml-0.5 rounded-full p-0.5 transition-colors hover:bg-white/10"
            >
              <X className="h-2.5 w-2.5" />
            </button>
          </span>
        ))}
        <input
          value={input}
          onChange={(e) => { setInput(e.target.value); setOpen(true) }}
          onFocus={() => setOpen(true)}
          onKeyDown={handleKeyDown}
          placeholder={selectedTags.length === 0 ? "Добавить теги..." : ""}
          className="flex-1 min-w-[80px] bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
        />
      </div>

      {open && (filtered.length > 0 || (input.trim() && !exactMatch)) && (
        <div
          className="absolute z-50 mt-1 w-full overflow-y-auto rounded-lg border border-border bg-popover py-1 shadow-xl" style={{ maxHeight: "240px" }}
        >
          {filtered.map((tag) => (
            <button
              key={tag.id}
              type="button"
              onClick={() => addTag(tag.id)}
              className="flex w-full items-center gap-2 px-3 py-1.5 text-[0.8rem] text-foreground transition-colors hover:bg-muted"
            >
              <span
                className="h-2.5 w-2.5 shrink-0 rounded-full"
                style={{ backgroundColor: tag.color || "#6366f1" }}
              />
              {tag.name}
            </button>
          ))}
          {input.trim() && !exactMatch && (
            <button
              type="button"
              onClick={handleCreate}
              disabled={createTag.isPending}
              className="flex w-full items-center gap-2 px-3 py-1.5 text-[0.8rem] text-violet-400 transition-colors hover:bg-muted"
            >
              <Plus className="h-3.5 w-3.5" />
              Создать «{input.trim()}»
            </button>
          )}
        </div>
      )}
    </div>
  )
}
