import { useState, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, CheckCircle2, KeyRound, Loader2, Mail } from "lucide-react"
import { useMutation } from "@tanstack/react-query"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { sendBg } from "../lib/bg-client"
import { useToast } from "../components/ui/toaster"
import { ApiError } from "../lib/types"

// 2-step flow: ввести email → "код отправлен" → ввести код + новый пароль.
type Stage = "email" | "reset"

export function ForgotPasswordPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [stage, setStage] = useState<Stage>("email")
  const [email, setEmail] = useState("")
  const [code, setCode] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [error, setError] = useState<string | null>(null)

  const forgotMut = useMutation({
    mutationFn: () => sendBg({ type: "api.forgotPassword", email: email.trim() }),
    onSuccess: () => {
      toast({
        title: "Код отправлен",
        description: "Проверьте почту — введите код ниже.",
        variant: "success",
      })
      setStage("reset")
    },
    onError: (err: Error) => {
      // Backend намеренно возвращает 200 при несуществующем email
      // (security). Реальная ошибка — только сеть/rate-limit.
      setError(err instanceof ApiError ? err.message : "Не удалось отправить")
    },
  })

  const resetMut = useMutation({
    mutationFn: () =>
      sendBg({
        type: "api.resetPassword",
        email: email.trim(),
        code: code.trim(),
        newPassword,
      }),
    onSuccess: () => {
      toast({
        title: "Пароль обновлён",
        description: "Войдите с новым паролем.",
        variant: "success",
      })
      navigate("/sign-in", { replace: true })
    },
    onError: (err: Error) => {
      setError(err instanceof ApiError ? err.message : "Не удалось сбросить пароль")
    },
  })

  function onSubmitEmail(e: FormEvent) {
    e.preventDefault()
    if (!email.trim()) return
    setError(null)
    forgotMut.mutate()
  }

  function onSubmitReset(e: FormEvent) {
    e.preventDefault()
    if (!code.trim() || !newPassword) return
    setError(null)
    resetMut.mutate()
  }

  return (
    <div className="flex h-full flex-col overflow-y-auto p-5">
      <div className="mb-3 flex items-center gap-2">
        <button
          type="button"
          onClick={() => navigate("/sign-in")}
          className="rounded p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
          aria-label="Назад"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <div className="flex items-center gap-1.5">
          <KeyRound className="h-4 w-4 text-(--color-primary)" />
          <h1 className="text-base font-semibold">Сброс пароля</h1>
        </div>
      </div>

      {stage === "email" ? (
        <form onSubmit={onSubmitEmail} className="flex flex-col gap-3">
          <p className="text-xs text-(--color-muted-foreground)">
            Введите email — мы отправим 6-значный код для сброса пароля.
          </p>

          <div className="space-y-1.5">
            <Label htmlFor="fp-email">Email</Label>
            <Input
              id="fp-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              required
              disabled={forgotMut.isPending}
            />
          </div>

          {error && (
            <div className="rounded-md border border-(--color-destructive) bg-(--color-destructive)/10 p-2 text-xs text-(--color-destructive)">
              {error}
            </div>
          )}

          <Button
            type="submit"
            disabled={forgotMut.isPending || !email.trim()}
            className="w-full gap-1.5"
          >
            {forgotMut.isPending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Mail className="h-3.5 w-3.5" />
            )}
            Отправить код
          </Button>
        </form>
      ) : (
        <form onSubmit={onSubmitReset} className="flex flex-col gap-3">
          <div className="flex items-center gap-2 rounded-md border border-emerald-500/30 bg-emerald-500/5 p-2.5 text-[11px]">
            <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-emerald-500" />
            <p>
              Код отправлен на <strong>{email}</strong>.
            </p>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="fp-code">Код из письма</Label>
            <Input
              id="fp-code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="123456"
              maxLength={6}
              required
              className="font-mono text-center"
              autoFocus
              disabled={resetMut.isPending}
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="fp-pwd">Новый пароль</Label>
            <Input
              id="fp-pwd"
              type="password"
              autoComplete="new-password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              minLength={8}
              required
              disabled={resetMut.isPending}
            />
            <p className="text-[10px] text-(--color-muted-foreground)">
              Минимум 8 символов.
            </p>
          </div>

          {error && (
            <div className="rounded-md border border-(--color-destructive) bg-(--color-destructive)/10 p-2 text-xs text-(--color-destructive)">
              {error}
            </div>
          )}

          <Button
            type="submit"
            disabled={
              resetMut.isPending ||
              !code.trim() ||
              !newPassword ||
              newPassword.length < 8
            }
            className="w-full gap-1.5"
          >
            {resetMut.isPending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <KeyRound className="h-3.5 w-3.5" />
            )}
            Сбросить пароль
          </Button>

          <button
            type="button"
            onClick={() => setStage("email")}
            className="text-center text-[10px] text-(--color-muted-foreground) hover:underline"
          >
            ← Отправить код повторно
          </button>
        </form>
      )}
    </div>
  )
}
