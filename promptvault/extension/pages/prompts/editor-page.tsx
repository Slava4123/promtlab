import { useNavigate, useParams } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { PromptEditor } from "../../components/prompts/prompt-editor"
import { usePrompt } from "../../hooks/use-prompts"
import { useCreatePrompt, useUpdatePrompt } from "../../hooks/use-prompts-crud"
import { useToast } from "../../components/ui/toaster"
import { useWorkspace } from "../../hooks/use-workspace"
import type { PromptFormValues } from "../../lib/validation/prompt-schema"

// Универсальная страница: /prompts/new (без id) или /prompts/:id/edit.
export function PromptEditorPage() {
  const { id } = useParams<{ id?: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const { workspaceId } = useWorkspace()
  const promptId = id ? Number(id) : null
  const isEdit = promptId !== null

  const promptQuery = usePrompt(promptId)
  const createMut = useCreatePrompt()
  const updateMut = useUpdatePrompt(promptId)

  if (isEdit && promptQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  if (isEdit && !promptQuery.data) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-(--color-muted-foreground)">
        Промпт не найден
      </div>
    )
  }

  async function handleSubmit(values: PromptFormValues) {
    const body = {
      title: values.title,
      content: values.content,
      description: values.description ?? "",
      model: values.model ?? "",
      collection_ids: values.collection_ids ?? [],
      tag_ids: values.tag_ids ?? [],
      team_id: values.team_id ?? workspaceId,
      is_public: values.is_public ?? false,
    }

    if (isEdit && promptId !== null) {
      const saved = await updateMut.mutateAsync({ ...body, change_note: values.change_note })
      toast({ title: "Сохранено", variant: "success" })
      navigate(`/prompts/${saved.id}`)
      return saved
    } else {
      const saved = await createMut.mutateAsync(body)
      toast({ title: "Промпт создан", variant: "success" })
      navigate(`/prompts/${saved.id}`)
      return saved
    }
  }

  const submitting = createMut.isPending || updateMut.isPending

  return (
    <PromptEditor
      prompt={promptQuery.data ?? null}
      onSubmit={handleSubmit}
      onSuccess={(saved) => navigate(`/prompts/${saved.id}`)}
      onCancel={() => navigate(-1)}
      submitting={submitting}
    />
  )
}
