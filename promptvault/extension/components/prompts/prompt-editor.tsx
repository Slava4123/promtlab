import { useState, useMemo, useEffect } from "react"
import { useForm, useWatch, Controller } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowLeft, Save, Eye, Edit3 } from "lucide-react"
import { Button } from "../ui/button"
import { Input } from "../ui/input"
import { Label } from "../ui/label"
import { Textarea } from "../ui/textarea"
import { useToast } from "../ui/toaster"
import { CodeEditor } from "../ui/code-editor"
import { TagInput } from "../tags/tag-input"
import { CollectionInput } from "../collections/collection-input"
import {
  promptSchema,
  MAX_PROMPT_CONTENT_LENGTH,
  CONTENT_LENGTH_WARNING,
  type PromptFormValues,
} from "../../lib/validation/prompt-schema"
import { renderTemplate, extractVariables } from "@pv/shared/template"
import { useWorkspace } from "../../hooks/use-workspace"
import { ApiError, type Prompt } from "../../lib/types"
import { cn } from "../../lib/utils"

interface PromptEditorProps {
  // null/undefined = create mode; Prompt = edit
  prompt?: Prompt | null
  onSuccess: (saved: Prompt) => void
  onCancel: () => void
  onSubmit: (values: PromptFormValues) => Promise<Prompt>
  submitting?: boolean
}

export function PromptEditor({ prompt, onCancel, onSubmit, submitting }: PromptEditorProps) {
  const { workspaceId } = useWorkspace()
  const { toast } = useToast()
  const [mode, setMode] = useState<"edit" | "preview">("edit")

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isDirty },
  } = useForm<PromptFormValues>({
    resolver: zodResolver(promptSchema),
    defaultValues: {
      title: prompt?.title ?? "",
      content: prompt?.content ?? "",
      description: "",
      model: prompt?.model ?? "",
      collection_ids: prompt?.collections.map((c) => c.id) ?? [],
      tag_ids: prompt?.tags.map((t) => t.id) ?? [],
      team_id: workspaceId,
      is_public: prompt?.is_public ?? false,
      change_note: "",
    },
  })

  // useWatch вместо watch() — memoization-safe; см. react-hooks/incompatible-library.
  const content = useWatch({ control, name: "content" }) ?? ""
  const charCount = content.length
  const isWarning = charCount >= CONTENT_LENGTH_WARNING

  // Collections / Tags теперь fetch'ятся внутри CollectionInput / TagInput
  // через свои хуки. Раньше parent держал collectionsQuery и передавал options
  // в ChipMultiSelect — теперь не нужно.

  // Unsaved-changes guard через beforeunload.
  useEffect(() => {
    if (!isDirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      e.returnValue = ""
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [isDirty])

  async function submit(values: PromptFormValues) {
    try {
      await onSubmit(values)
    } catch (err) {
      if (err instanceof ApiError && err.code === "validation") {
        toast({ title: "Ошибка валидации", description: err.message, variant: "error" })
      } else if (err instanceof ApiError && err.code === "quota_exceeded") {
        toast({ title: "Лимит исчерпан", description: err.message, variant: "error" })
      } else {
        toast({
          title: "Не удалось сохранить",
          description: err instanceof Error ? err.message : undefined,
          variant: "error",
        })
      }
    }
  }

  const variables = useMemo(() => extractVariables(content), [content])

  return (
    <form onSubmit={handleSubmit(submit)} className="flex h-full flex-col">
      {/* Header */}
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={onCancel}
          aria-label="Назад"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">
          {prompt ? "Редактировать" : "Новый промпт"}
        </h2>
        <Button
          type="submit"
          variant="brand"
          size="sm"
          disabled={submitting}
          className="gap-1.5"
        >
          <Save className="h-3.5 w-3.5" />
          {submitting ? "Сохраняю…" : prompt ? "Сохранить" : "Создать"}
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {/* Title */}
        <div className="space-y-1">
          <Label htmlFor="prompt-title">Название</Label>
          <Input id="prompt-title" {...register("title")} placeholder="Например: Резюме статьи" />
          {errors.title && (
            <p className="text-xs text-(--color-destructive)">{errors.title.message}</p>
          )}
        </div>

        {/* Description (optional) */}
        <div className="space-y-1">
          <Label htmlFor="prompt-description">Описание (опционально)</Label>
          <Textarea
            id="prompt-description"
            {...register("description")}
            rows={2}
            placeholder="Краткое описание для команды"
          />
        </div>

        {/* Content tabs: Редактор / Просмотр */}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <Label>Содержимое</Label>
            {/* Segmented Edit/Preview — active state брендирован.
                Кастомные кнопки вместо Button, чтобы не наследовать
                primary-цвет (default variant); paint только active. */}
            <div className="inline-flex rounded-md border border-(--color-border) p-0.5">
              <button
                type="button"
                onClick={() => setMode("edit")}
                aria-pressed={mode === "edit"}
                className={cn(
                  "inline-flex items-center gap-1 rounded-sm px-2 py-1 text-xs font-medium transition-colors",
                  mode === "edit"
                    ? "bg-(--color-brand-muted) text-(--color-brand)"
                    : "text-(--color-muted-foreground) hover:text-(--color-foreground)",
                )}
              >
                <Edit3 className="h-3 w-3" aria-hidden />
                Редактор
              </button>
              <button
                type="button"
                onClick={() => setMode("preview")}
                aria-pressed={mode === "preview"}
                className={cn(
                  "inline-flex items-center gap-1 rounded-sm px-2 py-1 text-xs font-medium transition-colors",
                  mode === "preview"
                    ? "bg-(--color-brand-muted) text-(--color-brand)"
                    : "text-(--color-muted-foreground) hover:text-(--color-foreground)",
                )}
              >
                <Eye className="h-3 w-3" aria-hidden />
                Просмотр
              </button>
            </div>
          </div>
          {mode === "edit" ? (
            <Controller
              control={control}
              name="content"
              render={({ field }) => (
                <div className="rounded-md border border-(--color-border) bg-(--color-card) min-h-[240px] max-h-[360px] overflow-auto">
                  <CodeEditor
                    value={field.value}
                    onChange={field.onChange}
                    minHeight="240px"
                    placeholder="Введите промпт. Используйте {{переменные}} для подстановки."
                  />
                </div>
              )}
            />
          ) : (
            <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3 text-xs whitespace-pre-wrap min-h-[240px] max-h-[360px] overflow-auto">
              {renderTemplate(content, {}) || (
                <span className="text-(--color-muted-foreground)">Пустой промпт</span>
              )}
            </div>
          )}
          {errors.content && (
            <p className="text-xs text-(--color-destructive)">{errors.content.message}</p>
          )}
          <div className="flex items-center justify-between text-[10px] text-(--color-muted-foreground)">
            <span>
              {variables.length > 0
                ? `Переменных: ${variables.length}`
                : "Переменных нет"}
            </span>
            <span className={cn(isWarning && "text-amber-500")}>
              {charCount.toLocaleString("ru-RU")} / {MAX_PROMPT_CONTENT_LENGTH.toLocaleString("ru-RU")}
            </span>
          </div>
        </div>

        {/* Model — свободный ввод (как frontend), без datalist */}
        <div className="space-y-1">
          <Label htmlFor="prompt-model">Модель (опционально)</Label>
          <Input
            id="prompt-model"
            {...register("model")}
            placeholder="Например: gpt-4o, claude-opus-4"
            autoComplete="off"
          />
        </div>

        {/* Collections — combobox с inline-созданием (mirror TagInput).
            Раньше был ChipMultiSelect — простой dropdown без поиска и без
            создания. Теперь юзер вводит название и либо выбирает из списка,
            либо создаёт новую (с дефолтной иконкой/цветом, дораб. потом). */}
        <div className="space-y-1">
          <Label>Коллекции</Label>
          <Controller
            control={control}
            name="collection_ids"
            render={({ field }) => (
              <CollectionInput
                selectedCollectionIds={field.value ?? []}
                onChange={field.onChange}
              />
            )}
          />
        </div>

        {/* Tags — combobox с inline-созданием */}
        <div className="space-y-1">
          <Label>Теги</Label>
          <Controller
            control={control}
            name="tag_ids"
            render={({ field }) => (
              <TagInput
                selectedTagIds={field.value ?? []}
                onChange={field.onChange}
              />
            )}
          />
        </div>

        {/* Public toggle */}
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="prompt-public"
            {...register("is_public")}
            className="h-4 w-4 accent-(--color-brand)"
          />
          <Label htmlFor="prompt-public" className="cursor-pointer text-xs">
            Публичный (доступен по ссылке)
          </Label>
        </div>

        {/* Spacer перед bottom-tabs — даёт визуальный «воздух» при scroll'е
            до конца. Explicit div вместо pb-N, чтобы CSS override'ы не съели. */}
        <div className="h-12" aria-hidden />
      </div>
    </form>
  )
}
