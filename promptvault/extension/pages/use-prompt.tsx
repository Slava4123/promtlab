import { useEffect } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { usePrompt } from "../hooks/use-prompts"
import { useInsertPrompt } from "../hooks/use-insert-prompt"
import { VariableForm } from "../components/variable-form"
import { extractVariables } from "../lib/template"

// Страница "Использовать промпт" — VariableForm с подгрузкой full content.
// Открывается из dashboard при клике на промпт. Если переменных нет —
// auto-insert + редирект назад.
export function UsePromptPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const promptQuery = usePrompt(id ? Number(id) : null)
  const insert = useInsertPrompt()

  const prompt = promptQuery.data ?? null

  useEffect(() => {
    if (!prompt) return
    const vars = extractVariables(prompt.content)
    if (vars.length === 0 && !insert.submitting) {
      void insert.insert(prompt, prompt.content).then((ok) => {
        if (ok) navigate("/", { replace: true })
      })
    }
  }, [prompt, insert, navigate])

  if (promptQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  if (!prompt) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-(--color-muted-foreground)">
        Промпт не найден
      </div>
    )
  }

  const vars = extractVariables(prompt.content)
  if (vars.length === 0) {
    return (
      <div className="flex h-full items-center justify-center gap-2">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
        <span className="text-xs text-(--color-muted-foreground)">Вставляю…</span>
      </div>
    )
  }

  return (
    <VariableForm
      prompt={prompt}
      onBack={() => navigate("/")}
      onSubmit={async (text) => {
        const ok = await insert.insert(prompt, text)
        if (ok) navigate("/", { replace: true })
      }}
      onInsertAll={async (text) => {
        const ok = await insert.insertAll(prompt, text)
        if (ok) navigate("/", { replace: true })
      }}
      submitting={insert.submitting}
      error={insert.error}
      canInsert={insert.activeTab.supported}
      canInsertReason={
        !insert.activeTab.supported
          ? "Откройте ChatGPT, Claude, Gemini, Perplexity, Yandex GPT, GigaChat, DeepSeek, Mistral или Qwen"
          : undefined
      }
    />
  )
}
