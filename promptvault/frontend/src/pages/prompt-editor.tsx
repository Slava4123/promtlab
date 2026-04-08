import { useEffect, useRef, useState } from "react"
import { useNavigate, useParams, useSearchParams } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { ArrowLeft, Loader2, FileText, Sparkles, FolderOpen, Tag, Search, ChevronDown, History, Copy } from "lucide-react"
import { toast } from "sonner"

import { usePrompt, useCreatePrompt, useUpdatePrompt, useIncrementUsage } from "@/hooks/use-prompts"
import { useCollections } from "@/hooks/use-collections"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { TagInput } from "@/components/tags/tag-input"
import { AIPanel } from "@/components/ai/ai-panel"
import { UsePromptDialog } from "@/components/prompts/use-prompt-dialog"
import { hasVariables } from "@/lib/template/parse"
import type { Prompt } from "@/api/types"

const promptSchema = z.object({
  title: z.string().min(1, "Введите название").max(300),
  content: z.string().min(1, "Введите содержимое промпта").max(10000, "Максимум 10 000 символов"),
  model: z.string().max(100).optional(),
})

type PromptForm = z.infer<typeof promptSchema>

export default function PromptEditor() {
  const navigate = useNavigate()
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const isEdit = !!id && id !== "new"
  const promptId = isEdit ? Number(id) : 0
  const preselectedCollectionId = searchParams.get("collection_id") ? Number(searchParams.get("collection_id")) : undefined

  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const { data: existing, isLoading: loadingExisting } = usePrompt(promptId)
  const { data: collections } = useCollections(teamId)
  const createPrompt = useCreatePrompt()
  const updatePrompt = useUpdatePrompt()
  const incrementUsage = useIncrementUsage()
  const [collectionIds, setCollectionIds] = useState<number[]>(preselectedCollectionId ? [preselectedCollectionId] : [])
  const [tagIds, setTagIds] = useState<number[]>([])
  const [collSearch, setCollSearch] = useState("")
  const [collExpanded, setCollExpanded] = useState(false)
  const [changeNote, setChangeNote] = useState("")
  const [usePromptModal, setUsePromptModal] = useState<Prompt | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<PromptForm>({
    resolver: zodResolver(promptSchema),
  })

  // eslint-disable-next-line react-hooks/incompatible-library
  const contentValue = watch("content") || ""

  useEffect(() => {
    if (existing) {
      reset({
        title: existing.title,
        content: existing.content,
        model: existing.model || "",
      })
      setCollectionIds(existing.collections?.map(c => c.id) || [])
      setTagIds(existing.tags?.map(t => t.id) || [])
    }
  }, [existing, reset])

  const onSubmit = async (data: PromptForm) => {
    try {
      if (isEdit) {
        await updatePrompt.mutateAsync({ id: promptId, ...data, change_note: changeNote || undefined, collection_ids: collectionIds, tag_ids: tagIds })
        setChangeNote("")
        toast.success("Промпт обновлён")
      } else {
        const created = await createPrompt.mutateAsync({ ...data, team_id: teamId, collection_ids: collectionIds, tag_ids: tagIds })
        toast.success("Промпт создан")
        navigate(`/prompts/${created.id}`, { replace: true })
        return
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка сохранения")
    }
  }

  // Keep a stable ref to onSubmit so the Ctrl+Enter effect doesn't re-subscribe every render
  const onSubmitRef = useRef(onSubmit)
  onSubmitRef.current = onSubmit

  // Ctrl+Enter to submit
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
        e.preventDefault()
        handleSubmit(onSubmitRef.current)()
      }
    }
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [handleSubmit])

  if (isEdit && loadingExisting) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-[48rem]">
      {/* Header */}
      <div className="mb-8 flex items-center gap-3">
        <button
          onClick={() => navigate(-1)}
          className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-500/[0.08] ring-1 ring-violet-500/10">
          <FileText className="h-4 w-4 text-violet-400" />
        </div>
        <div>
          <h1 className="text-lg font-bold tracking-tight text-foreground">
            {isEdit ? "Редактировать промпт" : "Новый промпт"}
          </h1>
          <p className="text-[0.75rem] text-muted-foreground">
            {isEdit ? "Измените и сохраните" : "Заполните поля и создайте промпт"}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        {/* Карточка с формой */}
        <div className="rounded-xl border border-border bg-card p-6 space-y-5">

          {/* Название */}
          <div className="space-y-2">
            <label htmlFor="title" className="text-[0.8rem] font-medium text-foreground">
              Название
            </label>
            <input
              id="title"
              placeholder="Например: Генерация README для проекта"
              className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              {...register("title")}
            />
            {errors.title && (
              <p className="text-[0.75rem] text-red-400">{errors.title.message}</p>
            )}
          </div>

          {/* Содержимое */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label htmlFor="content" className="text-[0.8rem] font-medium text-foreground">
                Промпт
              </label>
              <span className={`text-[0.7rem] tabular-nums ${contentValue.length > 9000 ? "text-red-400" : contentValue.length > 7500 ? "text-amber-400" : "text-muted-foreground"}`}>
                {contentValue.length.toLocaleString("ru-RU")}/10 000
              </span>
            </div>
            <textarea
              id="content"
              rows={16}
              maxLength={10000}
              placeholder="Введите текст промпта...&#10;&#10;Совет: будьте конкретны и используйте примеры для лучших результатов"
              className="flex w-full min-h-[280px] resize-y rounded-lg border border-border bg-background px-3.5 py-3 text-sm leading-relaxed text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              {...register("content")}
            />
            {errors.content && (
              <p className="text-[0.75rem] text-red-400">{errors.content.message}</p>
            )}
          </div>

          {/* AI-панель */}
          <AIPanel
            content={contentValue}
            onApply={(text, note) => {
              setValue("content", text)
              setChangeNote(note)
            }}
          />

          {/* Модель + Коллекция */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="model" className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
                <Sparkles className="h-3 w-3 text-violet-400/60" />
                Модель
                <span className="text-muted-foreground">(необяз.)</span>
              </label>
              <input
                id="model"
                placeholder="gpt-4o, claude-sonnet..."
                className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
                {...register("model")}
              />
            </div>

            <div className="space-y-2">
              <label className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
                <FolderOpen className="h-3 w-3 text-violet-400/60" />
                Коллекции
                <span className="text-muted-foreground">(необяз.)</span>
                {collectionIds.length > 0 && (
                  <span className="ml-auto text-[0.7rem] text-violet-400">{collectionIds.length} выбрано</span>
                )}
              </label>
              <div className="rounded-lg border border-border bg-background">
                {collections && collections.length > 5 && (
                  <div className="relative px-2.5 pt-2.5 pb-1">
                    <Search className="absolute left-5 top-1/2 h-3 w-3 -translate-y-1/2 text-muted-foreground" />
                    <input
                      value={collSearch}
                      onChange={(e) => { setCollSearch(e.target.value); setCollExpanded(true) }}
                      placeholder="Найти коллекцию..."
                      className="h-7 w-full rounded-md bg-muted pl-7 pr-2 text-[0.72rem] text-foreground outline-none placeholder:text-muted-foreground focus:bg-muted/80"
                    />
                  </div>
                )}
                <div className={`relative flex flex-wrap gap-1.5 px-3 py-2.5 overflow-hidden transition-all ${collExpanded || collSearch ? "" : "max-h-[72px]"}`}>
                  {(!collections || collections.length === 0) ? (
                    <span className="text-[0.8rem] text-muted-foreground">Нет коллекций</span>
                  ) : collections
                    .filter((c) => !collSearch || c.name.toLowerCase().includes(collSearch.toLowerCase()))
                    .map((c) => {
                      const isSelected = collectionIds.includes(c.id)
                      return (
                        <button
                          key={c.id}
                          type="button"
                          onClick={() => setCollectionIds(prev =>
                            isSelected ? prev.filter(id => id !== c.id) : [...prev, c.id]
                          )}
                          className={`flex items-center gap-1 rounded-md px-2 py-1 text-[0.75rem] font-medium transition-all ${
                            isSelected
                              ? "text-white ring-1"
                              : "text-muted-foreground hover:text-foreground hover:bg-muted"
                          }`}
                          style={isSelected ? { background: `${c.color}18`, boxShadow: `inset 0 0 0 1px ${c.color}30`, color: c.color } : undefined}
                        >
                          {c.name}
                        </button>
                      )
                    })}
                  {!collExpanded && !collSearch && collections && collections.length > 10 && (
                    <div className="pointer-events-none absolute inset-x-0 bottom-0 h-5 bg-gradient-to-t from-background to-transparent" />
                  )}
                </div>
                {collections && collections.length > 10 && !collSearch && (
                  <div className="px-3 pb-2">
                    <button
                      type="button"
                      onClick={() => setCollExpanded(!collExpanded)}
                      className="flex items-center gap-1 text-[0.7rem] text-muted-foreground transition-colors hover:text-muted-foreground"
                    >
                      <ChevronDown className={`h-3 w-3 transition-transform ${collExpanded ? "rotate-180" : ""}`} />
                      {collExpanded ? "Свернуть" : `Ещё ${collections.length - 10}+`}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Теги */}
          <div className="space-y-2">
            <label className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
              <Tag className="h-3 w-3 text-violet-400/60" />
              Теги
              <span className="text-muted-foreground">(необяз.)</span>
            </label>
            <TagInput selectedTagIds={tagIds} onChange={setTagIds} />
          </div>

          {/* Заметка к изменению (только в режиме редактирования) */}
          {isEdit && (
            <div className="space-y-2">
              <label htmlFor="change_note" className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
                <History className="h-3 w-3 text-violet-400/60" />
                Заметка к изменению
                <span className="text-muted-foreground">(необяз.)</span>
              </label>
              <input
                id="change_note"
                value={changeNote}
                onChange={(e) => setChangeNote(e.target.value)}
                maxLength={300}
                placeholder="Что изменилось? Например: улучшил формулировку"
                className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              />
            </div>
          )}
        </div>

        {/* Подсказка */}
        <div className="flex items-center gap-2.5 rounded-xl px-4 py-3 text-[0.82rem] text-muted-foreground" style={{ border: "1px solid rgba(139,92,246,0.08)", background: "rgba(139,92,246,0.04)" }}>
          <span className="text-base">💡</span>
          <span>Промпт сохраняется с версионированием — каждое изменение создаёт новую версию</span>
        </div>

        {/* Кнопки */}
        <div className="flex items-center gap-2.5">
          <button
            type="submit"
            disabled={isSubmitting}
            className="flex h-9 items-center gap-2 rounded-lg px-5 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97] disabled:opacity-50"
            style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 4px 16px -2px rgba(124,58,237,0.25)" }}
            onMouseEnter={(e) => { (e.target as HTMLElement).style.boxShadow = "0 6px 24px -2px rgba(124,58,237,0.35)" }}
            onMouseLeave={(e) => { (e.target as HTMLElement).style.boxShadow = "0 4px 16px -2px rgba(124,58,237,0.25)" }}
          >
            {isSubmitting && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            {isEdit ? "Сохранить изменения" : "Создать промпт"}
          </button>
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="flex h-9 items-center rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-foreground"
          >
            Отмена
          </button>
          {isEdit && existing && (
            <button
              type="button"
              onClick={async () => {
                if (hasVariables(existing.content)) {
                  setUsePromptModal(existing)
                  return
                }
                try {
                  await navigator.clipboard.writeText(existing.content)
                  incrementUsage.mutate(existing.id)
                  toast.success("Скопировано")
                } catch {
                  toast.error("Не удалось скопировать")
                }
              }}
              className="ml-auto flex h-9 items-center gap-1.5 rounded-lg border border-violet-500/25 bg-violet-500/10 px-4 text-[0.8rem] font-medium text-violet-300 transition-all hover:bg-violet-500/15 hover:text-violet-200"
            >
              <Copy className="h-3.5 w-3.5" />
              Использовать
            </button>
          )}
          {isEdit && (
            <button
              type="button"
              onClick={() => navigate(`/prompts/${promptId}/versions`)}
              className={`flex h-9 items-center gap-1.5 rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-violet-400 ${existing ? "" : "ml-auto"}`}
            >
              <History className="h-3.5 w-3.5" />
              История версий
            </button>
          )}
        </div>
      </form>

      {usePromptModal && (
        <UsePromptDialog
          prompt={usePromptModal}
          open
          onOpenChange={(o) => !o && setUsePromptModal(null)}
        />
      )}
    </div>
  )
}
