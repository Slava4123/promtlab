import { useEffect, useRef, useState } from "react"
import { useNavigate, useParams, useSearchParams } from "react-router-dom"
import { useForm, useWatch, Controller, type Control } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { ArrowLeft, Loader2, FileText, Sparkles, FolderOpen, Tag, History, Copy, Trash2, Share2, MoreHorizontal } from "lucide-react"
import { toast } from "sonner"

import { usePrompt, useCreatePrompt, useUpdatePrompt, useIncrementUsage, useDeletePrompt } from "@/hooks/use-prompts"
import { Button } from "@/components/ui/button"
import { useCollections } from "@/hooks/use-collections"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { TagInput } from "@/components/tags/tag-input"
import { CollectionsCombobox } from "@/components/prompts/collections-combobox"
import { PromptSplitEditor } from "@/components/prompts/prompt-split-editor"
import { FileImportButton } from "@/components/prompts/file-import-button"
import { FileImportDropZone } from "@/components/prompts/file-import-drop-zone"
import {
  FileImportChoiceDialog,
  type FileImportChoice,
} from "@/components/prompts/file-import-choice-dialog"
import { parseFile, FileImportError, type ParseResult } from "@/lib/file-import/parsers"
import { HelpPopover } from "@/components/help/help-popover"
import { InfoTooltip } from "@/components/help/info-tooltip"
import { UsePromptDialog } from "@/components/prompts/use-prompt-dialog"
import { ShareDialog } from "@/components/prompts/share-dialog"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu"
import { hasVariables } from "@/lib/template/parse"
import {
  MAX_PROMPT_CONTENT_LENGTH,
  MAX_PROMPT_TITLE_LENGTH,
  MAX_CHANGE_NOTE_LENGTH,
  CONTENT_LENGTH_WARNING,
  CONTENT_LENGTH_DANGER,
} from "@/lib/constants"
import type { Prompt } from "@/api/types"

const promptSchema = z.object({
  title: z.string().min(1, "Введите название").max(MAX_PROMPT_TITLE_LENGTH),
  content: z
    .string()
    .min(1, "Введите содержимое промпта")
    .max(MAX_PROMPT_CONTENT_LENGTH, `Максимум ${MAX_PROMPT_CONTENT_LENGTH.toLocaleString("ru-RU")} символов`),
  model: z.string().max(100).optional(),
})

type PromptForm = z.infer<typeof promptSchema>

// CharCounter — изолированная подписка на content через useWatch.
// Без этого watch("content") в родителе ре-рендерил бы весь editor (~400 строк)
// на каждое нажатие клавиши — см. P-8.
function CharCounter({ control }: { control: Control<PromptForm> }) {
  const value = useWatch({ control, name: "content" }) ?? ""
  const len = value.length
  const cls = len > CONTENT_LENGTH_DANGER
    ? "text-red-400"
    : len > CONTENT_LENGTH_WARNING
      ? "text-amber-400"
      : "text-muted-foreground"
  return (
    <span className={`text-[0.7rem] tabular-nums ${cls}`}>
      {len.toLocaleString("ru-RU")}/{MAX_PROMPT_CONTENT_LENGTH.toLocaleString("ru-RU")}
    </span>
  )
}

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
  const deletePrompt = useDeletePrompt()
  const [collectionIds, setCollectionIds] = useState<number[]>(preselectedCollectionId ? [preselectedCollectionId] : [])
  const [tagIds, setTagIds] = useState<number[]>([])
  const [changeNote, setChangeNote] = useState("")
  const [isPublic, setIsPublic] = useState(false)
  const [usePromptModal, setUsePromptModal] = useState<Prompt | null>(null)
  const [shareOpen, setShareOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [noteExpanded, setNoteExpanded] = useState(false)
  // Состояние загрузки файла: null когда idle, File когда парсим. При
  // непустом content подсовываем choiceDialog чтобы спросить стратегию вставки.
  const [isImporting, setIsImporting] = useState(false)
  const [pendingImport, setPendingImport] = useState<
    | { result: ParseResult }
    | null
  >(null)

  const {
    register,
    handleSubmit,
    reset,
    control,
    setValue,
    getValues,
    formState: { errors, isSubmitting },
  } = useForm<PromptForm>({
    resolver: zodResolver(promptSchema),
  })

  // Синхронизируем загруженные с сервера данные в локальное state формы.
  // Это legitimate sync external async data (prompt приходит от TanStack Query).
  useEffect(() => {
    if (existing) {
      reset({
        title: existing.title,
        content: existing.content,
        model: existing.model || "",
      })
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setCollectionIds(existing.collections?.map(c => c.id) || [])
      setTagIds(existing.tags?.map(t => t.id) || [])
      setIsPublic(existing.is_public ?? false)
    }
  }, [existing, reset])

  const onSubmit = async (data: PromptForm) => {
    try {
      if (isEdit) {
        await updatePrompt.mutateAsync({ id: promptId, ...data, change_note: changeNote || undefined, collection_ids: collectionIds, tag_ids: tagIds, is_public: isPublic })
        setChangeNote("")
        toast.success("Промпт обновлён")
      } else {
        const created = await createPrompt.mutateAsync({ ...data, team_id: teamId, collection_ids: collectionIds, tag_ids: tagIds, is_public: isPublic })
        toast.success("Промпт создан")
        navigate(`/prompts/${created.id}`, { replace: true })
        return
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка сохранения")
    }
  }

  // Keep a stable ref to onSubmit so the Ctrl+Enter effect doesn't re-subscribe every render.
  // Ref обновляется в useEffect, чтобы не мутировать .current во время рендера (react-hooks/refs).
  const onSubmitRef = useRef(onSubmit)
  useEffect(() => {
    onSubmitRef.current = onSubmit
  }, [onSubmit])

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

  // Применяет ParseResult к форме согласно выбранной стратегии вставки.
  // shouldValidate:true триггерит Zod-валидацию (проверит max=100000).
  const applyImport = (result: ParseResult, choice: FileImportChoice) => {
    const current = getValues("content") ?? ""
    let next: string
    switch (choice) {
      case "replace":
        next = result.content
        break
      case "prepend":
        next = result.content + "\n\n" + current
        break
      case "append":
        next = current + "\n\n" + result.content
        break
    }
    // Жёсткое обрезание если после concat получилось больше лимита.
    if (next.length > MAX_PROMPT_CONTENT_LENGTH) {
      next = next.slice(0, MAX_PROMPT_CONTENT_LENGTH)
      toast.warning(
        `Общая длина превысила ${MAX_PROMPT_CONTENT_LENGTH.toLocaleString("ru-RU")} символов — текст обрезан.`,
      )
    }
    setValue("content", next, { shouldValidate: true, shouldDirty: true })
    // Если parser вернул metadata (prompt-JSON) — обновляем title/model тоже.
    if (result.metadata?.title) {
      setValue("title", result.metadata.title, { shouldValidate: true, shouldDirty: true })
    }
    if (result.metadata?.model) {
      setValue("model", result.metadata.model, { shouldValidate: true, shouldDirty: true })
    }
    const parsedMsg = `Вставлено ${result.content.length.toLocaleString("ru-RU")} символов из ${result.filename}`
    if (result.truncated) {
      toast.warning(`${parsedMsg} (файл был обрезан до лимита).`)
    } else {
      toast.success(parsedMsg)
    }
    // Выводим warnings парсера отдельным toast.
    for (const w of result.warnings) {
      toast.warning(w)
    }
  }

  // Единая точка обработки выбранного файла (из input или drop).
  const handleFileImport = async (file: File) => {
    if (isImporting) return
    setIsImporting(true)
    try {
      const result = await parseFile(file)
      const currentContent = getValues("content") ?? ""
      if (currentContent.trim().length === 0) {
        // Пустой редактор — просто вставляем без вопросов.
        applyImport(result, "replace")
      } else {
        // Непустой — диалог выбора стратегии.
        setPendingImport({ result })
      }
    } catch (err) {
      if (err instanceof FileImportError) {
        toast.error(err.message)
      } else {
        toast.error(err instanceof Error ? err.message : "Не удалось обработать файл")
      }
    } finally {
      setIsImporting(false)
    }
  }

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
          type="button"
          onClick={() => navigate(-1)}
          aria-label="Назад"
          className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-500/[0.08] ring-1 ring-violet-500/10">
          <FileText className="h-4 w-4 text-violet-400" />
        </div>
        <div className="flex-1">
          <h1 className="text-lg font-bold tracking-tight text-foreground">
            {isEdit ? "Редактировать промпт" : "Новый промпт"}
          </h1>
          <p className="text-[0.75rem] text-muted-foreground">
            {isEdit ? "Измените и сохраните" : "Заполните поля и создайте промпт"}
          </p>
        </div>
        {!isEdit && (
          <HelpPopover
            title="Что такое промпт?"
            learnMoreHref="/help"
            learnMoreLabel="Все вопросы — в Помощи"
            ariaLabel="Подсказка по созданию промпта"
          >
            <p>
              Промпт — это шаблон сообщения для AI-моделей. Хранится в вашей библиотеке,
              версионируется при каждом сохранении.
            </p>
            <p>
              В тексте можно использовать <code className="rounded bg-muted px-1 text-[0.78em]">{"{{переменные}}"}</code> —
              они подставятся при использовании.
            </p>
            <p>
              Поддерживается Markdown: заголовки, таблицы, code-блоки, списки задач. Превью покажет итоговый вид.
            </p>
          </HelpPopover>
        )}
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
              aria-invalid={!!errors.title}
              aria-describedby={errors.title ? "title-error" : undefined}
              className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              {...register("title")}
            />
            {errors.title && (
              <p id="title-error" className="text-[0.75rem] text-red-400">{errors.title.message}</p>
            )}
          </div>

          {/* Содержимое */}
          <div className="space-y-2">
            <div className="flex items-center justify-between gap-2">
              <label htmlFor="content" className="text-[0.8rem] font-medium text-foreground">
                Промпт
              </label>
              <div className="flex items-center gap-3">
                <FileImportButton
                  onFileSelect={handleFileImport}
                  isImporting={isImporting}
                  disabled={isSubmitting}
                />
                <CharCounter control={control} />
              </div>
            </div>
            <Controller
              control={control}
              name="content"
              render={({ field }) => (
                <PromptSplitEditor
                  id="content"
                  value={field.value ?? ""}
                  onChange={field.onChange}
                  maxLength={MAX_PROMPT_CONTENT_LENGTH}
                  placeholder="Введите текст промпта...&#10;&#10;Совет: будьте конкретны и используйте примеры для лучших результатов.&#10;Поддерживается Markdown: заголовки, таблицы, code-блоки, ссылки."
                  aria-invalid={!!errors.content}
                  aria-describedby={errors.content ? "content-error" : undefined}
                />
              )}
            />
            {errors.content && (
              <p id="content-error" className="text-[0.75rem] text-red-400">{errors.content.message}</p>
            )}
          </div>

          {/* Модель + Коллекция */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <label htmlFor="model" className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
                <Sparkles className="h-3 w-3 text-violet-400/60" />
                Модель
                <span className="text-muted-foreground">(необяз.)</span>
                <InfoTooltip ariaLabel="Что значит поле «Модель»?">
                  Под какую AI-модель оптимизирован промпт. Например: <code className="rounded bg-muted px-1">gpt-4o</code>, <code className="rounded bg-muted px-1">claude-sonnet-4</code>. Поле справочное.
                </InfoTooltip>
              </label>
              <input
                id="model"
                placeholder="gpt-4o, claude-sonnet..."
                className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
                {...register("model")}
              />
            </div>

            <div className="space-y-2">
              <label className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
                <FolderOpen className="h-3 w-3 text-violet-400/60" />
                Коллекции
                <span className="text-muted-foreground">(необяз.)</span>
                <InfoTooltip ariaLabel="Что такое коллекции?">
                  Группировка промптов по темам — например «Код-ревью», «Маркетинг». Один промпт может быть в нескольких коллекциях одновременно.
                </InfoTooltip>
                {collectionIds.length > 0 && (
                  <span className="ml-auto text-[0.7rem] text-violet-400">{collectionIds.length} выбрано</span>
                )}
              </label>
              <CollectionsCombobox
                collections={collections}
                value={collectionIds}
                onChange={setCollectionIds}
              />
            </div>
          </div>

          {/* Теги */}
          <div className="space-y-2">
            <label className="flex items-center gap-1.5 text-[0.8rem] font-medium text-foreground">
              <Tag className="h-3 w-3 text-violet-400/60" />
              Теги
              <span className="text-muted-foreground">(необяз.)</span>
              <InfoTooltip ariaLabel="Что такое теги?">
                Метки для фильтрации и поиска по библиотеке. В отличие от коллекций — плоские, без иерархии. Удобно для кросс-категорий: «срочно», «черновик», «en».
              </InfoTooltip>
            </label>
            <TagInput selectedTagIds={tagIds} onChange={setTagIds} />
          </div>

          {/* Публичный доступ — slug генерится backend'ом по id после INSERT,
              на create-форме URL ещё неизвестен → показываем общий placeholder. */}
          <label className="flex items-start gap-3 rounded-lg border border-border bg-muted/20 p-3 text-sm">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              className="mt-0.5 h-4 w-4 cursor-pointer accent-brand"
            />
            <span className="flex-1">
              <span className="font-medium text-foreground">Публичный промпт</span>
              <span className="ml-2 text-muted-foreground">
                {isPublic
                  ? existing?.slug
                    ? `Доступен по ссылке /p/${existing.slug}`
                    : "Будет доступен по публичной ссылке после сохранения"
                  : "Только вы видите этот промпт"}
              </span>
            </span>
          </label>

          {/* Заметка к изменению (collapsible, X-4) — скрываем по умолчанию, большинство
              юзеров не описывают каждое изменение, и поле занимает визуальный вес зря. */}
          {isEdit && !noteExpanded && !changeNote && (
            <button
              type="button"
              onClick={() => setNoteExpanded(true)}
              className="flex items-center gap-1.5 text-[0.75rem] text-muted-foreground underline underline-offset-4 hover:text-foreground"
            >
              <History className="h-3 w-3" />
              Добавить заметку к изменению
            </button>
          )}
          {isEdit && (noteExpanded || changeNote) && (
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
                maxLength={MAX_CHANGE_NOTE_LENGTH}
                placeholder="Что изменилось? Например: улучшил формулировку"
                className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              />
            </div>
          )}
        </div>

        {/* Подсказка */}
        <div className="flex items-center gap-2.5 rounded-xl px-4 py-3 text-[0.82rem] text-muted-foreground" style={{ border: "1px solid rgba(139,92,246,0.08)", background: "rgba(139,92,246,0.04)" }}>
          <span className="text-base">💡</span>
          <span>Промпт сохраняется с версионированием — каждое изменение создаёт новую версию</span>
        </div>

        {/* Кнопки — X-4: primary слева, secondary в DropdownMenu "Ещё". */}
        <div className="flex flex-wrap items-center gap-2.5">
          <Button type="submit" variant="brand" size="sm" disabled={isSubmitting}>
            {isSubmitting && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            {isEdit ? "Сохранить" : "Создать"}
          </Button>
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="flex h-9 items-center rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-colors hover:text-foreground"
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
              className="flex h-9 items-center gap-1.5 rounded-lg border border-violet-500/30 bg-violet-500/10 px-4 text-[0.8rem] font-medium text-violet-600 transition-colors hover:bg-violet-500/20 hover:text-violet-700 dark:text-violet-300 dark:hover:text-violet-200"
            >
              <Copy className="h-3.5 w-3.5" />
              Использовать
            </button>
          )}
          {isEdit && (
            <DropdownMenu>
              <DropdownMenuTrigger className="ml-auto flex h-9 items-center gap-1.5 rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-colors hover:text-foreground">
                Ещё <MoreHorizontal className="h-3.5 w-3.5" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => navigate(`/prompts/${promptId}/versions`)}>
                  <History className="h-3.5 w-3.5" />
                  История версий
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setShareOpen(true)}>
                  <Share2 className="h-3.5 w-3.5" />
                  Поделиться
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
                  <Trash2 className="h-3.5 w-3.5" />
                  Удалить
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </form>

      <FileImportDropZone
        onFileDrop={handleFileImport}
        disabled={isImporting || isSubmitting}
      />

      <FileImportChoiceDialog
        open={pendingImport !== null}
        onOpenChange={(open) => {
          if (!open) setPendingImport(null)
        }}
        filename={pendingImport?.result.filename ?? ""}
        charCount={pendingImport?.result.content.length ?? 0}
        onChoose={(choice) => {
          if (pendingImport) {
            applyImport(pendingImport.result, choice)
            setPendingImport(null)
          }
        }}
      />

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Удалить промпт?"
        description="Промпт переместится в корзину. Восстановить можно в разделе «Корзина» в течение 30 дней."
        confirmLabel="Удалить"
        variant="destructive"
        isPending={deletePrompt.isPending}
        onConfirm={() => {
          deletePrompt.mutate(promptId, {
            onSuccess: () => {
              toast("Промпт перемещён в корзину")
              setDeleteOpen(false)
              navigate("/dashboard")
            },
            onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка удаления"),
          })
        }}
      />

      {usePromptModal && (
        <UsePromptDialog
          prompt={usePromptModal}
          open
          onOpenChange={(o) => !o && setUsePromptModal(null)}
        />
      )}

      {isEdit && (
        <ShareDialog
          promptId={promptId}
          open={shareOpen}
          onOpenChange={setShareOpen}
        />
      )}
    </div>
  )
}
