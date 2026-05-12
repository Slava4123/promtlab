// Hook для insert-flow промпта в активную вкладку. Вынесена из components/app.tsx
// чтобы быть доступной из любой страницы (router-based).

import { useCallback, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { addLocalRecent } from "../lib/storage"
import { hostLabel } from "../lib/messages"
import { ApiError, type Prompt } from "../lib/types"
import { useToast } from "../components/ui/toaster"
import { useActiveTab } from "./use-active-tab"

export interface InsertOptions {
  silent?: boolean
}

export function useInsertPrompt() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const activeTab = useActiveTab()
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [highlightedId, setHighlightedId] = useState<number | null>(null)

  const insert = useCallback(
    async (prompt: Prompt, text: string, options: InsertOptions = {}) => {
      setSubmitting(true)
      setError(null)
      try {
        await sendBg({ type: "cmd.insertPrompt", text })
        void sendBg({ type: "api.incrementUsage", promptId: prompt.id }).catch(() => undefined)
        void queryClient.invalidateQueries({ queryKey: ["prompts", "recent"] })

        void addLocalRecent({
          promptId: prompt.id,
          title: prompt.title,
          insertedAt: Date.now(),
          targetHost: activeTab.host,
        })

        if (!options.silent) {
          const targetLabel = hostLabel(activeTab.host) ?? "цель"
          setHighlightedId(prompt.id)
          setTimeout(() => setHighlightedId(null), 900)
          toast({
            title: `Вставлено в ${targetLabel}`,
            description: prompt.title,
            variant: "success",
            durationMs: 5000,
            action: {
              label: "Отменить",
              icon: "undo",
              onClick: async () => {
                try {
                  await sendBg({ type: "cmd.undoInsert" })
                  toast({ title: "Отменено", variant: "info", durationMs: 1500 })
                } catch {
                  toast({
                    title: "Не получилось отменить",
                    description: "Возможно, вы уже отредактировали поле",
                    variant: "error",
                  })
                }
              },
            },
          })
        }
        return true
      } catch (err) {
        if (err instanceof ApiError) {
          if (err.code === "no_target") {
            setError("Откройте поддерживаемый AI-сайт (ChatGPT, Claude, Gemini, Perplexity, Yandex GPT, GigaChat, DeepSeek, Mistral или Qwen).")
          } else if (err.code === "unauthorized") {
            setError("Ключ больше не действителен.")
          } else {
            setError("Не удалось вставить промпт. Попробуйте ещё раз.")
          }
        } else {
          setError("Не удалось вставить промпт.")
        }
        return false
      } finally {
        setSubmitting(false)
      }
    },
    [activeTab.host, queryClient, toast],
  )

  const insertAll = useCallback(
    async (prompt: Prompt, text: string) => {
      try {
        const result = await sendBg({ type: "cmd.insertPromptAll", text })
        void sendBg({ type: "api.incrementUsage", promptId: prompt.id }).catch(() => undefined)
        void queryClient.invalidateQueries({ queryKey: ["prompts", "recent"] })
        if (result.successes === 0) {
          toast({
            title: "Нет открытых вкладок",
            description: "Откройте ChatGPT, Claude, Gemini, Perplexity, Yandex GPT, GigaChat, DeepSeek, Mistral или Qwen",
            variant: "error",
          })
          return false
        }
        toast({
          title: `Вставлено в ${result.successes} ${pluralTabs(result.successes)}`,
          description: prompt.title,
          variant: "success",
          durationMs: 4000,
        })
        return true
      } catch {
        toast({ title: "Не удалось вставить во все вкладки", variant: "error" })
        return false
      }
    },
    [queryClient, toast],
  )

  return {
    insert,
    insertAll,
    submitting,
    error,
    highlightedId,
    activeTab,
    clearError: () => setError(null),
  }
}

function pluralTabs(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return "вкладку"
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return "вкладки"
  return "вкладок"
}
