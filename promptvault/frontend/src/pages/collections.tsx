import { useState } from "react"
import { useNavigate } from "react-router-dom"
import {
  Plus, FolderOpen, Pencil, Trash2, Loader2, FileText, AlertTriangle,
  Code, Palette, FileCode, Wrench, Rocket, BarChart3, FlaskConical,
  Shield, Lightbulb, BookOpen, Zap, MessageSquare, Globe, Database,
  type LucideIcon,
} from "lucide-react"
import { toast } from "sonner"

import { useCollections, useCreateCollection, useUpdateCollection, useDeleteCollection } from "@/hooks/use-collections"
import type { Collection } from "@/api/types"
import { useWorkspaceStore } from "@/stores/workspace-store"

const COLORS = [
  { value: "#a78bfa", label: "Фиолетовый" },
  { value: "#60a5fa", label: "Синий" },
  { value: "#22d3ee", label: "Голубой" },
  { value: "#34d399", label: "Зелёный" },
  { value: "#fbbf24", label: "Жёлтый" },
  { value: "#fb923c", label: "Оранжевый" },
  { value: "#f87171", label: "Красный" },
  { value: "#f472b6", label: "Розовый" },
]

const ICON_OPTIONS: { value: string; Icon: LucideIcon; label: string }[] = [
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
  ICON_OPTIONS.map((i) => [i.value, i.Icon])
)

function CollectionIcon({ icon, color, size = 16 }: { icon?: string; color?: string; size?: number }) {
  const IconComponent = (icon && ICON_MAP[icon]) || FolderOpen
  return <IconComponent style={{ width: size, height: size, color: color || "#8b5cf6" }} />
}

export default function Collections() {
  const navigate = useNavigate()
  const team = useWorkspaceStore((s) => s.team)
  const teamId = team?.teamId ?? null
  const teamName = team?.teamName ?? null
  const { data: collections, isLoading } = useCollections(teamId)
  const createCollection = useCreateCollection()
  const updateCollection = useUpdateCollection()
  const deleteCollection = useDeleteCollection()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [editing, setEditing] = useState<Collection | null>(null)
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [color, setColor] = useState(COLORS[0].value)
  const [icon, setIcon] = useState("")

  const openCreate = () => {
    setEditing(null)
    setName("")
    setDescription("")
    setColor(COLORS[0].value)
    setIcon("")
    setDialogOpen(true)
  }

  const openEdit = (c: Collection) => {
    setEditing(c)
    setName(c.name)
    setDescription(c.description)
    setColor(c.color || COLORS[0].value)
    setIcon(c.icon || "")
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!name.trim()) return
    try {
      if (editing) {
        await updateCollection.mutateAsync({ id: editing.id, name, description, color, icon })
        toast.success("Коллекция обновлена")
      } else {
        await createCollection.mutateAsync({ name, description, color, icon, team_id: teamId })
        toast.success("Коллекция создана")
      }
      setDialogOpen(false)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
    }
  }

  const confirmDelete = (id: number) => {
    setDeletingId(id)
    setDeleteDialogOpen(true)
  }

  const handleDelete = () => {
    if (!deletingId) return
    deleteCollection.mutate(deletingId, {
      onSuccess: () => { toast.success("Коллекция удалена"); setDeleteDialogOpen(false) },
    })
  }

  return (
    <div className="mx-auto max-w-[64rem] space-y-5">
      {/* Header */}
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{teamName ? `Коллекции — ${teamName}` : "Коллекции"}</h1>
          <p className="mt-0.5 text-[0.8rem] text-muted-foreground">Группируйте промпты по темам и проектам</p>
        </div>
        <button
          onClick={openCreate}
          className="flex h-8 items-center gap-1.5 rounded-lg bg-violet-600 px-3.5 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 hover:shadow-violet-500/20 active:scale-[0.97]"
        >
          <Plus className="h-3.5 w-3.5" />
          Новая коллекция
        </button>
      </div>

      {/* List */}
      {isLoading ? (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-border bg-card p-5">
              <div className="mb-3 h-9 w-9 animate-pulse rounded-lg bg-muted/40" />
              <div className="mb-2 h-4 w-2/3 animate-pulse rounded-md bg-muted/40" />
              <div className="h-3 w-1/2 animate-pulse rounded-md bg-muted/30" />
            </div>
          ))}
        </div>
      ) : !collections || collections.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-violet-500/[0.08] ring-1 ring-violet-500/10">
            <FolderOpen className="h-7 w-7 text-violet-400/60" />
          </div>
          <p className="text-base font-medium text-muted-foreground">Пока нет коллекций</p>
          <p className="mt-1 text-sm text-muted-foreground">Создайте первую коллекцию для организации промптов</p>
          <button
            onClick={openCreate}
            className="mt-5 flex h-8 items-center gap-1.5 rounded-lg bg-violet-600 px-4 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 active:scale-[0.97]"
          >
            <Plus className="h-3.5 w-3.5" />
            Создать коллекцию
          </button>
        </div>
      ) : (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {collections.map((c) => (
            <div
              key={c.id}
              className="group cursor-pointer rounded-xl border border-border bg-card p-5 transition-all duration-200 hover:-translate-y-0.5"
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = `${c.color || "#8b5cf6"}30`; (e.currentTarget as HTMLElement).style.boxShadow = `0 8px 32px -8px rgba(0,0,0,0.5), 0 0 0 1px ${c.color || "#8b5cf6"}15` }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = `${c.color || "#8b5cf6"}15`; (e.currentTarget as HTMLElement).style.boxShadow = "none" }}
              onClick={() => navigate(`/collections/${c.id}`)}
            >
              <div className="mb-3 flex items-start justify-between">
                <div
                  className="flex h-9 w-9 items-center justify-center rounded-lg ring-1 text-sm"
                  style={{
                    background: `${c.color || "#8b5cf6"}12`,
                    boxShadow: `inset 0 0 0 1px ${c.color || "#8b5cf6"}20`,
                  }}
                >
                  <CollectionIcon icon={c.icon} color={c.color} />
                </div>
                <div className="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                  <button
                    className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                    onClick={(e) => { e.stopPropagation(); openEdit(c) }}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </button>
                  <button
                    className="rounded-md p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-400"
                    onClick={(e) => { e.stopPropagation(); confirmDelete(c.id) }}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
              <h3 className="mb-1 text-[0.85rem] font-medium text-foreground">{c.name}</h3>
              {c.description && (
                <p className="mb-3 text-[0.75rem] text-muted-foreground line-clamp-2">{c.description}</p>
              )}
              <div className="flex items-center gap-1.5 text-[0.7rem] text-muted-foreground">
                <FileText className="h-3 w-3" />
                <span>{c.prompt_count} {c.prompt_count === 1 ? "промпт" : "промптов"}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Dialog */}
      {dialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setDialogOpen(false)}>
          <div className="w-full max-w-md rounded-2xl border border-border bg-card p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-lg font-semibold text-foreground">{editing ? "Редактировать коллекцию" : "Новая коллекция"}</h2>

            {/* Иконка */}
            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-foreground">Иконка</label>
              <div className="flex flex-wrap gap-1.5">
                {ICON_OPTIONS.map((opt) => (
                  <button
                    key={opt.value}
                    onClick={() => setIcon(opt.value)}
                    title={opt.label}
                    className={`flex h-8 w-8 items-center justify-center rounded-lg transition-all ${icon === opt.value || (!icon && opt.value === "folder") ? "ring-2 ring-violet-500 bg-muted/60" : "bg-muted/30 hover:bg-muted"}`}
                  >
                    <opt.Icon className="h-4 w-4" style={{ color }} />
                  </button>
                ))}
              </div>
            </div>

            {/* Название */}
            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-foreground">Название</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Например: Код-ревью"
                autoFocus
                className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
                onKeyDown={(e) => e.key === "Enter" && handleSave()}
              />
            </div>

            {/* Цвет */}
            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-foreground">Цвет</label>
              <div className="flex gap-2">
                {COLORS.map((c) => (
                  <button
                    key={c.value}
                    onClick={() => setColor(c.value)}
                    className={`h-7 w-7 rounded-full transition-all ${color === c.value ? "ring-2 ring-white ring-offset-2 ring-offset-background scale-110" : "hover:scale-110"}`}
                    style={{ background: c.value }}
                    title={c.label}
                  />
                ))}
              </div>
            </div>

            {/* Описание */}
            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-foreground">Описание <span className="text-muted-foreground">(необязательно)</span></label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Для чего эта коллекция?"
                rows={2}
                className="flex w-full resize-none rounded-lg border border-border bg-background px-3.5 py-2.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              />
            </div>

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setDialogOpen(false)}
                className="flex h-9 items-center rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-foreground"
              >
                Отмена
              </button>
              <button
                onClick={handleSave}
                disabled={!name.trim()}
                className="flex h-9 items-center gap-2 rounded-lg px-5 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97] disabled:opacity-50"
                style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 4px 16px -2px rgba(124,58,237,0.25)" }}
              >
                {(createCollection.isPending || updateCollection.isPending) && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                {editing ? "Сохранить" : "Создать"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteDialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setDeleteDialogOpen(false)}>
          <div className="w-full max-w-sm rounded-2xl border border-red-500/15 bg-card p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/10">
                <AlertTriangle className="h-5 w-5 text-red-400" />
              </div>
              <div>
                <h3 className="text-[0.9rem] font-semibold text-foreground">Удалить коллекцию?</h3>
                <p className="text-[0.75rem] text-muted-foreground">Промпты не удалятся, только открепятся</p>
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setDeleteDialogOpen(false)}
                className="flex h-9 items-center rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-foreground"
              >
                Отмена
              </button>
              <button
                onClick={handleDelete}
                className="flex h-9 items-center gap-2 rounded-lg px-4 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97]"
                style={{ background: "linear-gradient(135deg, #dc2626, #b91c1c)", boxShadow: "0 4px 16px -2px rgba(220,38,38,0.25)" }}
              >
                {deleteCollection.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                Удалить
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
