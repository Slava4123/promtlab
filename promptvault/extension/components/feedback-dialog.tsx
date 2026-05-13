import { useState } from "react"
import { Bug, Lightbulb, MessageSquare, Send, X } from "lucide-react"
import { useMutation } from "@tanstack/react-query"
import { Button } from "./ui/button"
import { Textarea } from "./ui/textarea"
import { useToast } from "./ui/toaster"
import { sendBg } from "../lib/bg-client"
import { cn } from "../lib/utils"
import type { FeedbackRequest, FeedbackResponse } from "../lib/types"

type FeedbackType = "bug" | "feature" | "other"

interface FeedbackDialogProps {
  open: boolean
  onClose: () => void
}

const TYPE_META: Record<FeedbackType, { label: string; icon: React.ComponentType<{ className?: string }>; color: string }> = {
  bug: { label: "Баг", icon: Bug, color: "text-(--color-destructive)" },
  feature: { label: "Идея", icon: Lightbulb, color: "text-amber-500" },
  other: { label: "Другое", icon: MessageSquare, color: "text-(--color-brand)" },
}

const MAX_MESSAGE_LEN = 2000

export function FeedbackDialog({ open, onClose }: FeedbackDialogProps) {
  const { toast } = useToast()
  const [type, setType] = useState<FeedbackType>("other")
  const [message, setMessage] = useState("")

  const submitMut = useMutation<FeedbackResponse, Error, FeedbackRequest>({
    mutationFn: (body) => sendBg({ type: "api.submitFeedback", body }),
  })

  function reset() {
    setType("other")
    setMessage("")
  }

  async function handleSubmit() {
    const trimmed = message.trim()
    if (!trimmed) return
    try {
      await submitMut.mutateAsync({
        type,
        message: trimmed,
        page_url: `chrome-extension://${chrome.runtime.id}`,
      })
      toast({
        title: "Спасибо!",
        description: "Ваш отзыв отправлен. Мы обязательно его прочитаем.",
        variant: "success",
      })
      reset()
      onClose()
    } catch (err) {
      toast({
        title: "Не удалось отправить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden
      />
      <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">Обратная связь</h3>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <p className="mt-1 text-[10px] text-(--color-muted-foreground)">
          Поделитесь идеей, сообщите о баге или просто напишите нам.
        </p>

        {/* Type selector */}
        <div className="mt-3 grid grid-cols-3 gap-1.5">
          {(Object.keys(TYPE_META) as FeedbackType[]).map((t) => {
            const meta = TYPE_META[t]
            const Icon = meta.icon
            const active = type === t
            return (
              <button
                key={t}
                type="button"
                onClick={() => setType(t)}
                className={cn(
                  "flex flex-col items-center gap-1 rounded-md border px-2 py-2 text-[10px] transition-colors",
                  active
                    ? "border-(--color-brand) bg-(--color-brand-muted)"
                    : "border-(--color-border) bg-(--color-card) hover:bg-(--color-muted)/40",
                )}
              >
                <Icon className={cn("h-4 w-4", active ? meta.color : "text-(--color-muted-foreground)")} />
                <span className={cn("font-medium", active && meta.color)}>{meta.label}</span>
              </button>
            )
          })}
        </div>

        {/* Message */}
        <div className="mt-3 space-y-1">
          <Textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder={
              type === "bug"
                ? "Что произошло? Какие шаги воспроизведения?"
                : type === "feature"
                  ? "Какую функцию вы бы хотели?"
                  : "Расскажите подробнее…"
            }
            maxLength={MAX_MESSAGE_LEN}
            rows={5}
            className="text-xs"
            autoFocus
          />
          <p className="text-right text-[9px] text-(--color-muted-foreground)">
            {message.length} / {MAX_MESSAGE_LEN}
          </p>
        </div>

        <div className="mt-3 flex justify-end gap-2">
          <Button type="button" variant="outline" size="sm" onClick={onClose}>
            Отмена
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={handleSubmit}
            disabled={submitMut.isPending || !message.trim()}
            className="gap-1.5"
          >
            <Send className="h-3.5 w-3.5" />
            Отправить
          </Button>
        </div>
      </div>
    </div>
  )
}
