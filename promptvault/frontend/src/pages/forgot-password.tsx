import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Loader2, KeyRound } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { AuthLayout } from "@/components/auth/auth-layout"
import { api } from "@/api/client"

const emailSchema = z.object({
  email: z.string().email("Введите корректный email"),
})

const resetSchema = z.object({
  code: z.string().length(6, "Введите 6-значный код"),
  new_password: z.string().min(8, "Минимум 8 символов").max(128),
  confirm: z.string(),
}).refine((d) => d.new_password === d.confirm, {
  message: "Пароли не совпадают",
  path: ["confirm"],
})

export default function ForgotPasswordPage() {
  const navigate = useNavigate()
  const [step, setStep] = useState<"email" | "reset">("email")
  const [email, setEmail] = useState("")
  const [error, setError] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const emailForm = useForm({
    resolver: zodResolver(emailSchema),
  })

  const resetForm = useForm({
    resolver: zodResolver(resetSchema),
  })

  const handleSendCode = emailForm.handleSubmit(async (data) => {
    setError("")
    setIsSubmitting(true)
    try {
      await api("/auth/forgot-password", {
        method: "POST",
        body: JSON.stringify({ email: data.email }),
      })
      setEmail(data.email)
      setStep("reset")
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Ошибка")
    } finally {
      setIsSubmitting(false)
    }
  })

  const handleReset = resetForm.handleSubmit(async (data) => {
    setError("")
    setIsSubmitting(true)
    try {
      await api("/auth/reset-password", {
        method: "POST",
        body: JSON.stringify({
          email,
          code: data.code,
          new_password: data.new_password,
        }),
      })
      navigate("/sign-in", { replace: true })
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Неверный код или ошибка")
    } finally {
      setIsSubmitting(false)
    }
  })

  return (
    <AuthLayout>
      <div className="text-center">
        <KeyRound className="mx-auto mb-3 h-10 w-10 text-violet-400" />
        <h1 className="text-xl font-bold text-foreground">
          {step === "email" ? "Сброс пароля" : "Новый пароль"}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {step === "email"
            ? "Введите email для получения кода"
            : `Код отправлен на ${email}`}
        </p>
      </div>

      {error && (
        <div className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-400">
          {error}
        </div>
      )}

      {step === "email" ? (
        <form onSubmit={handleSendCode} noValidate className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="email" className="text-foreground">Email</Label>
            <Input
              id="email"
              type="email"
              placeholder="you@example.com"
              {...emailForm.register("email")}
              className="h-10 border-border bg-card focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            />
            {emailForm.formState.errors.email && (
              <p className="text-xs text-red-400">{emailForm.formState.errors.email.message}</p>
            )}
          </div>

          <Button type="submit" disabled={isSubmitting}
            className="h-10 w-full bg-violet-600 font-medium text-white hover:bg-violet-500">
            {isSubmitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Отправить код
          </Button>

          <p className="text-center text-sm text-muted-foreground">
            Вспомнили пароль?{" "}
            <a href="/sign-in" className="text-violet-400 hover:text-violet-300">Войти</a>
          </p>
        </form>
      ) : (
        <form onSubmit={handleReset} noValidate className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="code" className="text-foreground">Код из email</Label>
            <Input
              id="code"
              inputMode="numeric"
              maxLength={6}
              placeholder="000000"
              {...resetForm.register("code")}
              className="h-10 border-border bg-card tracking-[0.3em] text-center focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            />
            {resetForm.formState.errors.code && (
              <p className="text-xs text-red-400">{resetForm.formState.errors.code.message}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="new_password" className="text-foreground">Новый пароль</Label>
            <Input
              id="new_password"
              type="password"
              placeholder="Минимум 8 символов"
              {...resetForm.register("new_password")}
              className="h-10 border-border bg-card focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            />
            {resetForm.formState.errors.new_password && (
              <p className="text-xs text-red-400">{resetForm.formState.errors.new_password.message}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="confirm" className="text-foreground">Подтвердите пароль</Label>
            <Input
              id="confirm"
              type="password"
              placeholder="Повторите пароль"
              {...resetForm.register("confirm")}
              className="h-10 border-border bg-card focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            />
            {resetForm.formState.errors.confirm && (
              <p className="text-xs text-red-400">{resetForm.formState.errors.confirm.message}</p>
            )}
          </div>

          <Button type="submit" disabled={isSubmitting}
            className="h-10 w-full bg-violet-600 font-medium text-white hover:bg-violet-500">
            {isSubmitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Сменить пароль
          </Button>

          <button type="button" onClick={() => { setStep("email"); setError("") }}
            className="w-full text-center text-sm text-muted-foreground hover:text-foreground">
            Отправить код заново
          </button>
        </form>
      )}
    </AuthLayout>
  )
}
