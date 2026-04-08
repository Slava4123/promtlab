import { useState, useCallback } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { Plus, ArrowLeft, FileText, FolderOpen, PackagePlus, Check, Loader2, Search } from "lucide-react"
import { toast } from "sonner"
import { useQueryClient } from "@tanstack/react-query"

import { PromptCard, PromptCardSkeleton } from "@/components/prompts/prompt-card"
import { UsePromptDialog } from "@/components/prompts/use-prompt-dialog"
import { useCollection } from "@/hooks/use-collections"
import { usePrompts, useToggleFavorite, useUpdatePrompt, useIncrementUsage } from "@/hooks/use-prompts"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { hasVariables } from "@/lib/template/parse"
import type { Prompt } from "@/api/types"

// Reuse ICON_MAP from collections page
import {
  Code, Palette, FileCode, Wrench, Rocket, BarChart3, FlaskConical,
  Shield, Lightbulb, BookOpen, Zap, MessageSquare, Globe, Database,
} from "lucide-react"
import type { LucideIcon } from "lucide-react"

const ICON_MAP: Record<string, LucideIcon> = {
  folder: FolderOpen, code: Code, palette: Palette, "file-code": FileCode,
  wrench: Wrench, rocket: Rocket, chart: BarChart3, flask: FlaskConical,
  shield: Shield, lightbulb: Lightbulb, book: BookOpen, zap: Zap,
  message: MessageSquare, globe: Globe, database: Database,
}

export default function CollectionView() {
  const navigate = useNavigate()
  const { id } = useParams()
  const collectionId = Number(id)

  const qc = useQueryClient()
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const { data: collection, isLoading: loadingCollection } = useCollection(collectionId)
  const { data: promptsData, isLoading: loadingPrompts } = usePrompts({ collection_id: collectionId, team_id: teamId })
  const { data: allPromptsData } = usePrompts({ team_id: teamId })
  const toggleFav = useToggleFavorite()
  const updatePrompt = useUpdatePrompt()
  const incrementUsage = useIncrementUsage()
  const [usePromptModal, setUsePromptModal] = useState<Prompt | null>(null)

  const handleUse = useCallback(
    async (prompt: Prompt) => {
      if (hasVariables(prompt.content)) {
        setUsePromptModal(prompt)
        return
      }
      try {
        await navigator.clipboard.writeText(prompt.content)
        incrementUsage.mutate(prompt.id)
        toast.success("Скопировано")
      } catch {
        toast.error("Не удалось скопировать")
      }
    },
    [incrementUsage],
  )

  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [addSearch, setAddSearch] = useState("")
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [adding, setAdding] = useState(false)

  const prompts = promptsData?.pages.flatMap(p => p.items) ?? []
  const allPrompts = allPromptsData?.pages.flatMap(p => p.items) ?? []

  // Промпты которых нет в этой коллекции
  const availablePrompts = allPrompts.filter(
    (p) => !p.collections?.some(c => c.id === collectionId)
  )

  const toggleSelect = (promptId: number) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(promptId)) next.delete(promptId)
      else next.add(promptId)
      return next
    })
  }

  const handleAddSelected = async () => {
    setAdding(true)
    try {
      for (const promptId of selected) {
        const prompt = allPrompts.find(p => p.id === promptId)
        const existingIds = prompt?.collections?.map(c => c.id) || []
        await updatePrompt.mutateAsync({ id: promptId, collection_ids: [...existingIds, collectionId] })
      }
      await qc.invalidateQueries({ queryKey: ["prompts"] })
      await qc.invalidateQueries({ queryKey: ["collection", collectionId] })
      toast.success(`Добавлено ${selected.size} ${selected.size === 1 ? "промпт" : "промптов"}`)
      setAddDialogOpen(false)
      setAddSearch("")
      setSelected(new Set())
    } catch {
      toast.error("Ошибка при добавлении")
    } finally {
      setAdding(false)
    }
  }

  const IconComponent = (collection?.icon && ICON_MAP[collection.icon]) || FolderOpen
  const color = collection?.color || "#a78bfa"

  if (loadingCollection) {
    return (
      <div className="mx-auto max-w-[64rem]">
        <div className="mb-6 flex items-center gap-3">
          <div className="h-5 w-20 animate-pulse rounded-md bg-muted/40" />
        </div>
        <div className="mb-6 flex items-center gap-3">
          <div className="h-10 w-10 animate-pulse rounded-lg bg-muted/40" />
          <div className="h-6 w-48 animate-pulse rounded-md bg-muted/40" />
        </div>
      </div>
    )
  }

  if (!collection) {
    return (
      <div className="mx-auto max-w-[64rem] py-20 text-center">
        <p className="text-muted-foreground">Коллекция не найдена</p>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-[64rem] space-y-6">
      {/* Хлебные крошки */}
      <div className="flex items-center gap-1.5 text-[0.8rem]">
        <button
          onClick={() => navigate("/collections")}
          className="flex items-center gap-1 text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Коллекции
        </button>
        <span className="text-muted-foreground">/</span>
        <span className="text-foreground">{collection.name}</span>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3.5">
          <div
            className="flex h-11 w-11 items-center justify-center rounded-xl ring-1"
            style={{
              background: `${color}12`,
              boxShadow: `inset 0 0 0 1px ${color}20`,
            }}
          >
            <IconComponent style={{ width: 20, height: 20, color }} />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight text-foreground">{collection.name}</h1>
            {collection.description && (
              <p className="mt-0.5 text-[0.8rem] text-muted-foreground">{collection.description}</p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => { setSelected(new Set()); setAddDialogOpen(true) }}
            className="flex h-8 items-center gap-1.5 rounded-lg px-3 text-[0.8rem] font-medium text-muted-foreground transition-all hover:text-foreground active:scale-[0.97] border border-border bg-card"
          >
            <PackagePlus className="h-3.5 w-3.5" />
            Из списка
          </button>
          <button
            onClick={() => navigate(`/prompts/new?collection_id=${collectionId}`)}
            className="flex h-8 items-center gap-1.5 rounded-lg bg-violet-600 px-3.5 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 hover:shadow-violet-500/20 active:scale-[0.97]"
          >
            <Plus className="h-3.5 w-3.5" />
            Новый промпт
          </button>
        </div>
      </div>

      {/* Промпты */}
      {loadingPrompts ? (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <PromptCardSkeleton key={i} />
          ))}
        </div>
      ) : !promptsData || prompts.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div
            className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl ring-1"
            style={{
              background: `${color}08`,
              boxShadow: `inset 0 0 0 1px ${color}15`,
            }}
          >
            <FileText className="h-7 w-7" style={{ color: `${color}90` }} />
          </div>
          <p className="text-base font-medium text-muted-foreground">Коллекция пока пуста</p>
          <p className="mt-1 text-sm text-muted-foreground">Добавьте первый промпт в эту коллекцию</p>
          <button
            onClick={() => navigate(`/prompts/new?collection_id=${collectionId}`)}
            className="mt-5 flex h-8 items-center gap-1.5 rounded-lg bg-violet-600 px-4 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 active:scale-[0.97]"
          >
            <Plus className="h-3.5 w-3.5" />
            Добавить промпт
          </button>
        </div>
      ) : (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {prompts.map((prompt) => (
            <PromptCard
              key={prompt.id}
              prompt={prompt}
              onToggleFavorite={(id) => toggleFav.mutate(id)}
              onClick={(id) => navigate(`/prompts/${id}`)}
              onUse={handleUse}
            />
          ))}
        </div>
      )}

      {usePromptModal && (
        <UsePromptDialog
          prompt={usePromptModal}
          open
          onOpenChange={(o) => !o && setUsePromptModal(null)}
        />
      )}

      {/* Модалка "Добавить из списка" */}
      {addDialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => { setAddDialogOpen(false); setAddSearch("") }}>
          <div
            className="w-full max-w-lg max-h-[80vh] flex flex-col rounded-2xl border border-border bg-card"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="flex items-center justify-between px-6 pt-5 pb-3">
              <div>
                <h2 className="text-lg font-semibold text-foreground">Добавить в коллекцию</h2>
                <p className="mt-0.5 text-[0.75rem] text-muted-foreground">Выберите промпты для добавления в "{collection.name}"</p>
              </div>
              {selected.size > 0 && (
                <span className="rounded-full bg-violet-500/15 px-2.5 py-0.5 text-xs font-medium text-violet-300">
                  {selected.size} выбрано
                </span>
              )}
            </div>

            {/* Search */}
            <div className="relative px-6 pb-2">
              <Search className="absolute left-8.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
              <input
                value={addSearch}
                onChange={(e) => setAddSearch(e.target.value)}
                placeholder="Поиск по названию..."
                className="h-8 w-full rounded-lg border border-border bg-muted/30 pl-8 pr-3 text-[0.8rem] text-foreground outline-none placeholder:text-muted-foreground focus:border-violet-500/25 focus:ring-1 focus:ring-violet-500/10"
              />
            </div>

            {/* List */}
            <div className="flex-1 overflow-auto px-6 py-2 space-y-1.5">
              {availablePrompts.filter(p => !addSearch || p.title.toLowerCase().includes(addSearch.toLowerCase())).length === 0 ? (
                <div className="py-10 text-center">
                  <p className="text-sm text-muted-foreground">{addSearch ? "Ничего не найдено" : "Все промпты уже в коллекциях"}</p>
                </div>
              ) : (
                availablePrompts.filter(p => !addSearch || p.title.toLowerCase().includes(addSearch.toLowerCase())).map((p) => (
                  <button
                    key={p.id}
                    onClick={() => toggleSelect(p.id)}
                    className={`flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-all ${
                      selected.has(p.id)
                        ? "bg-violet-500/10 ring-1 ring-violet-500/20"
                        : "hover:bg-muted"
                    }`}
                  >
                    <div className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-md transition-all ${
                      selected.has(p.id)
                        ? "bg-violet-500 text-white"
                        : "border border-border bg-muted/30"
                    }`}>
                      {selected.has(p.id) && <Check className="h-3 w-3" />}
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[0.82rem] font-medium text-foreground">{p.title}</p>
                      <p className="mt-0.5 truncate text-[0.72rem] text-muted-foreground">{p.content}</p>
                    </div>
                    {p.model && (
                      <span className="shrink-0 text-[0.65rem] text-muted-foreground">{p.model}</span>
                    )}
                  </button>
                ))
              )}
            </div>

            {/* Footer */}
            <div className="flex items-center justify-end gap-2 border-t border-border px-6 py-4">
              <button
                onClick={() => { setAddDialogOpen(false); setAddSearch("") }}
                className="flex h-9 items-center rounded-lg px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-foreground border border-border bg-card"
              >
                Отмена
              </button>
              <button
                onClick={handleAddSelected}
                disabled={selected.size === 0 || adding}
                className="flex h-9 items-center gap-2 rounded-lg px-5 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97] disabled:opacity-50"
                style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 4px 16px -2px rgba(124,58,237,0.25)" }}
              >
                {adding && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                Добавить{selected.size > 0 ? ` (${selected.size})` : ""}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
