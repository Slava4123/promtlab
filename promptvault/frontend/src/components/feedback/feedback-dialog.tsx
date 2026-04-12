import { useState } from "react"
import { MessageSquare } from "lucide-react"
import { toast } from "sonner"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useSubmitFeedback } from "@/hooks/use-feedback"
import { useAuthStore } from "@/stores/auth-store"

export function FeedbackDialog() {
  const [open, setOpen] = useState(false)
  const [type, setType] = useState<"bug" | "feature" | "other">("feature")
  const [message, setMessage] = useState("")

  const typeLabels: Record<string, string> = {
    bug: "Баг",
    feature: "Идея / Пожелание",
    other: "Другое",
  }
  const user = useAuthStore((s) => s.user)
  const submit = useSubmitFeedback()

  const handleSubmit = () => {
    if (!message.trim()) return

    submit.mutate(
      {
        type,
        message: message.trim(),
        page_url: window.location.pathname,
      },
      {
        onSuccess: () => {
          toast.success("Спасибо за отзыв!")
          setMessage("")
          setType("feature")
          setOpen(false)
        },
        onError: (err) => {
          toast.error(err instanceof Error ? err.message : "Не удалось отправить отзыв")
        },
      },
    )
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
      >
        <MessageSquare className="h-4 w-4" />
        <span>Отправить отзыв</span>
      </button>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Обратная связь</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-2">
          <div className="space-y-2">
            <Label>Тип</Label>
            <Select value={type} onValueChange={(v) => setType(v as typeof type)}>
              <SelectTrigger>
                <SelectValue>{typeLabels[type]}</SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="bug">Баг</SelectItem>
                <SelectItem value="feature">Идея / Пожелание</SelectItem>
                <SelectItem value="other">Другое</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Сообщение</Label>
            <Textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Расскажите, что можно улучшить..."
              maxLength={2000}
              rows={5}
            />
            <p className="text-right text-xs text-muted-foreground">
              {message.length}/2000
            </p>
          </div>
          {user?.email && (
            <p className="text-xs text-muted-foreground">
              Ответ придёт на {user.email}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Отмена
            </Button>
            <Button
              variant="brand"
              onClick={handleSubmit}
              disabled={!message.trim() || submit.isPending}
            >
              {submit.isPending ? "Отправка..." : "Отправить"}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
