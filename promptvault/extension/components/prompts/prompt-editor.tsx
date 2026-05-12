import { useState, useMemo, useEffect } from "react"
import { useForm, Controller } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useQuery } from "@tanstack/react-query"
import { ArrowLeft, Save, Eye, Edit3, X } from "lucide-react"
import { Button } from "../ui/button"
import { Input } from "../ui/input"
import { Label } from "../ui/label"
import { Textarea } from "../ui/textarea"
import { useToast } from "../ui/toaster"
import { CodeEditor } from "../ui/code-editor"
import {
  promptSchema,
  MAX_PROMPT_CONTENT_LENGTH,
  CONTENT_LENGTH_WARNING,
  type PromptFormValues,
} from "../../lib/validation/prompt-schema"
import { sendBg } from "../../lib/bg-client"
import { renderTemplate, extractVariables } from "../../lib/template"
import { qk } from "../../lib/query-keys"
import { useWorkspaceStore } from "../../stores/workspace-store"
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

// Список «известных» моделей с подсказками. Юзер может ввести любую свою
// модель (combobox: datalist suggestions + free-text input).
const MODEL_SUGGESTIONS = [
  "gpt-4o",
  "gpt-4o-mini",
  "gpt-4-turbo",
  "o1-preview",
  "o1-mini",
  "claude-opus-4",
  "claude-sonnet-4",
  "claude-haiku-4",
  "gemini-2.5-pro",
  "gemini-2.5-flash",
  "deepseek-v3",
  "deepseek-r1",
  "yandex-gpt-5",
  "yandex-gpt-5-pro",
  "gigachat-pro",
  "gigachat-max",
  "mistral-large",
  "mistral-medium",
  "qwen-max",
  "qwen-plus",
]

export function PromptEditor({ prompt, onCancel, onSubmit, submitting }: PromptEditorProps) {
  const team = useWorkspaceStore((s) => s.team)
  const { toast } = useToast()
  const [mode, setMode] = useState<"edit" | "preview">("edit")

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isDirty },
    watch,
  } = useForm<PromptFormValues>({
    resolver: zodResolver(promptSchema),
    defaultValues: {
      title: prompt?.title ?? "",
      content: prompt?.content ?? "",
      description: "",
      model: prompt?.model ?? "",
      collection_ids: prompt?.collections.map((c) => c.id) ?? [],
      tag_ids: prompt?.tags.map((t) => t.id) ?? [],
      team_id: team?.teamId ?? null,
      is_public: prompt?.is_public ?? false,
      change_note: "",
    },
  })

  const content = watch("content")
  const charCount = content.length
  const isWarning = charCount >= CONTENT_LENGTH_WARNING

  // Загружаем collections + tags для multi-select.
  const collectionsQuery = useQuery({
    queryKey: qk.collections,
    queryFn: () => sendBg({ type: "api.listCollections", teamId: team?.teamId ?? null }),
    staleTime: 60_000,
  })
  const tagsQuery = useQuery({
    queryKey: qk.tags,
    queryFn: () => sendBg({ type: "api.listTags", teamId: team?.teamId ?? null }),
    staleTime: 60_000,
  })

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
          size="sm"
          disabled={submitting}
          className="gap-1.5"
        >
          <Save className="h-3.5 w-3.5" />
          {submitting ? "Сохраняю…" : prompt ? "Сохранить" : "Создать"}
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
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

        {/* Content tabs: Edit / Preview */}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <Label>Содержимое</Label>
            <div className="flex gap-1">
              <Button
                type="button"
                variant={mode === "edit" ? "default" : "ghost"}
                size="sm"
                onClick={() => setMode("edit")}
                className="h-7 px-2 gap-1"
              >
                <Edit3 className="h-3 w-3" />
                Edit
              </Button>
              <Button
                type="button"
                variant={mode === "preview" ? "default" : "ghost"}
                size="sm"
                onClick={() => setMode("preview")}
                className="h-7 px-2 gap-1"
              >
                <Eye className="h-3 w-3" />
                Preview
              </Button>
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

        {/* Model — combobox: свободный ввод + suggestions */}
        <div className="space-y-1">
          <Label htmlFor="prompt-model">Модель</Label>
          <input
            id="prompt-model"
            list="prompt-model-list"
            {...register("model")}
            className="h-9 w-full rounded-md border border-(--color-border) bg-(--color-card) px-2 text-sm"
            placeholder="Любая или выберите из списка"
            autoComplete="off"
          />
          <datalist id="prompt-model-list">
            {MODEL_SUGGESTIONS.map((m) => (
              <option key={m} value={m} />
            ))}
          </datalist>
          <p className="text-[10px] text-(--color-muted-foreground)">
            Можно ввести любую модель или выбрать подсказку
          </p>
        </div>

        {/* Collections multi-select (chip-based) */}
        <div className="space-y-1">
          <Label>Коллекции</Label>
          <Controller
            control={control}
            name="collection_ids"
            render={({ field }) => (
              <ChipMultiSelect
                options={(collectionsQuery.data ?? []).map((c) => ({
                  id: c.id,
                  label: c.name,
                  color: c.color,
                }))}
                value={field.value ?? []}
                onChange={field.onChange}
                emptyLabel="Без коллекции"
              />
            )}
          />
        </div>

        {/* Tags multi-select */}
        <div className="space-y-1">
          <Label>Теги</Label>
          <Controller
            control={control}
            name="tag_ids"
            render={({ field }) => (
              <ChipMultiSelect
                options={(tagsQuery.data ?? []).map((t) => ({
                  id: t.id,
                  label: t.name,
                  color: t.color,
                }))}
                value={field.value ?? []}
                onChange={field.onChange}
                emptyLabel="Без тегов"
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
            className="h-4 w-4 accent-(--color-primary)"
          />
          <Label htmlFor="prompt-public" className="cursor-pointer text-xs">
            Публичный (доступен по share-ссылке)
          </Label>
        </div>

        {/* Change note (только при редактировании) */}
        {prompt && (
          <div className="space-y-1">
            <Label htmlFor="prompt-changenote">Что изменилось (опционально)</Label>
            <Input
              id="prompt-changenote"
              {...register("change_note")}
              placeholder="Добавил пример…"
            />
            <p className="text-[10px] text-(--color-muted-foreground)">
              Появится в истории версий.
            </p>
          </div>
        )}
      </div>
    </form>
  )

  // Local helper subcomponent — inline для простоты.
  function ChipMultiSelect({
    options,
    value,
    onChange,
    emptyLabel,
  }: {
    options: Array<{ id: number; label: string; color?: string }>
    value: number[]
    onChange: (v: number[]) => void
    emptyLabel: string
  }) {
    const [open, setOpen] = useState(false)
    const selected = options.filter((o) => value.includes(o.id))
    function toggle(id: number) {
      onChange(value.includes(id) ? value.filter((v) => v !== id) : [...value, id])
    }
    return (
      <div className="space-y-1">
        <div className="flex flex-wrap gap-1">
          {selected.length === 0 ? (
            <span className="text-[10px] text-(--color-muted-foreground)">{emptyLabel}</span>
          ) : (
            selected.map((s) => (
              <span
                key={s.id}
                className="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-[10px]"
                style={{
                  backgroundColor: s.color ? `${s.color}22` : "var(--color-muted)",
                  color: s.color ?? "var(--color-foreground)",
                }}
              >
                {s.label}
                <button
                  type="button"
                  onClick={() => toggle(s.id)}
                  className="hover:text-(--color-destructive)"
                  aria-label="Убрать"
                >
                  <X className="h-2.5 w-2.5" />
                </button>
              </span>
            ))
          )}
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-6 text-[10px] px-2"
            onClick={() => setOpen((v) => !v)}
          >
            {open ? "Скрыть" : "Выбрать"}
          </Button>
        </div>
        {open && options.length > 0 && (
          <div className="max-h-32 overflow-y-auto rounded-md border border-(--color-border) p-1">
            {options.map((o) => (
              <button
                key={o.id}
                type="button"
                onClick={() => toggle(o.id)}
                className={cn(
                  "flex w-full items-center gap-2 rounded px-2 py-1 text-xs text-left",
                  value.includes(o.id)
                    ? "bg-(--color-primary)/10 text-(--color-primary)"
                    : "hover:bg-(--color-muted)",
                )}
              >
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ backgroundColor: o.color ?? "currentColor" }}
                />
                {o.label}
                {value.includes(o.id) && <span className="ml-auto">✓</span>}
              </button>
            ))}
          </div>
        )}
        {open && options.length === 0 && (
          <p className="text-[10px] text-(--color-muted-foreground)">
            Создайте {emptyLabel.includes("коллекци") ? "коллекцию" : "тег"} в веб-интерфейсе.
          </p>
        )}
      </div>
    )
  }
}
