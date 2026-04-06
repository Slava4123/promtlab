import { useEffect } from "react"
import { useAIModels } from "@/hooks/use-ai"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

const STORAGE_KEY = "ai-last-model"

interface ModelSelectorProps {
  value: string
  onChange: (modelId: string) => void
}

export function ModelSelector({ value, onChange }: ModelSelectorProps) {
  const { data: models, isLoading, error } = useAIModels()

  // Restore last used model on first load
  useEffect(() => {
    if (!models || models.length === 0 || value) return

    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored && models.some((m) => m.id === stored)) {
      onChange(stored)
    } else {
      onChange(models[0].id)
    }
  }, [models, value, onChange])

  const handleChange = (modelId: string | null) => {
    if (!modelId) return
    onChange(modelId)
    localStorage.setItem(STORAGE_KEY, modelId)
  }

  if (isLoading) {
    return (
      <div className="flex h-8 items-center px-2.5 text-[0.78rem] text-muted-foreground">
        Загрузка моделей...
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-8 items-center px-2.5 text-[0.78rem] text-red-400">
        Не удалось загрузить модели
      </div>
    )
  }

  const selectedModel = models?.find((m) => m.id === value)

  return (
    <Select value={value} onValueChange={handleChange} modal={false}>
      <SelectTrigger size="sm" className="w-full text-[0.78rem]">
        <SelectValue placeholder="Выберите модель">
          {selectedModel?.name}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {models?.map((m) => (
          <SelectItem key={m.id} value={m.id}>
            {m.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
