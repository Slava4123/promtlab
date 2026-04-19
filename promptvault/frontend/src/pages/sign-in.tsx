import { useState } from "react"
import { useNavigate, useSearchParams, Link } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Eye, EyeOff, Loader2, ShieldCheck } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { AuthLayout } from "@/components/auth/auth-layout"
import { useAuthStore } from "@/stores/auth-store"
import { popCheckoutIntent, useCheckout } from "@/hooks/use-subscription"
import { readReferralCookie } from "@/lib/referral"

// withReferral добавляет ?ref=CODE к OAuth redirect, чтобы backend-cookie
// oauth_ref заполнился при Redirect'е (M-7).
function withReferral(path: string): string {
  const ref = readReferralCookie()
  if (!ref) return path
  const sep = path.includes("?") ? "&" : "?"
  return `${path}${sep}ref=${encodeURIComponent(ref)}`
}

const loginSchema = z.object({
  email: z.email("Введите корректный email"),
  password: z.string().min(1, "Введите пароль"),
})

type LoginForm = z.infer<typeof loginSchema>

// Шаг TOTP — показывается после успешной проверки password для admin'ов.
interface TOTPStep {
  preAuthToken: string
  email: string
}

// safeReturnURL защищает от open-redirect: принимает только same-origin пути,
// начинающиеся с "/" (но НЕ "//" — protocol-relative атака).
function safeReturnURL(raw: string | null): string | null {
  if (!raw) return null
  if (!raw.startsWith("/") || raw.startsWith("//")) return null
  return raw
}

export default function SignIn() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const returnURL = safeReturnURL(searchParams.get("return_url"))
  const login = useAuthStore((s) => s.login)
  const verifyTOTP = useAuthStore((s) => s.verifyTOTP)
  const checkout = useCheckout()
  const [error, setError] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [totpStep, setTotpStep] = useState<TOTPStep | null>(null)
  const [totpCode, setTotpCode] = useState("")
  const [totpSubmitting, setTotpSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  // После успешного login — если был saved checkout intent, запускаем checkout сразу (M-14).
  const resumeAfterLogin = () => {
    const pending = popCheckoutIntent()
    if (pending) {
      checkout.mutate(pending)
      return true
    }
    return false
  }

  const onSubmit = async (data: LoginForm) => {
    setError("")
    try {
      const result = await login(data.email, data.password)
      switch (result.kind) {
        case "ok":
          if (resumeAfterLogin()) return
          if (returnURL) {
            // OAuth authorize и другие сценарии с return_url: редиректим
            // на abs-URL, чтобы попасть на backend endpoint (не SPA route).
            window.location.href = returnURL
            return
          }
          navigate("/dashboard")
          return
        case "totp_required":
          setTotpStep({ preAuthToken: result.preAuthToken, email: result.email })
          return
        case "totp_enrollment_required":
          navigate("/admin/totp")
          return
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Ошибка входа"
      if (msg === "Email не подтверждён") {
        navigate(`/verify-email?email=${encodeURIComponent(data.email)}`)
        return
      }
      setError(msg)
    }
  }

  const onTOTPSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!totpStep || !totpCode) return
    setError("")
    setTotpSubmitting(true)
    try {
      const result = await verifyTOTP(totpStep.preAuthToken, totpCode)
      if (result.used_backup_code) {
        // Опционально можно показать toast; для MVP — тихий redirect.
      }
      if (resumeAfterLogin()) return
      if (returnURL) {
        window.location.href = returnURL
        return
      }
      navigate("/admin/users")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Неверный код")
    } finally {
      setTotpSubmitting(false)
    }
  }

  if (totpStep) {
    return (
      <AuthLayout>
        <div className="mb-6 text-center">
          <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-violet-500/15">
            <ShieldCheck className="h-6 w-6 text-violet-400" />
          </div>
          <h1 className="text-xl font-semibold text-foreground">Двухфакторная проверка</h1>
          <p className="mt-1.5 text-sm text-muted-foreground">
            Введите 6-значный код из приложения Authenticator или один из backup-кодов
          </p>
        </div>
        <form onSubmit={onTOTPSubmit} className="space-y-4">
          {error && (
            <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}
          <div className="space-y-1.5">
            <Label htmlFor="totp_code" className="text-foreground">Код</Label>
            <Input
              id="totp_code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              placeholder="000000"
              autoFocus
              value={totpCode}
              onChange={(e) => {
                setTotpCode(e.target.value)
                if (error) setError("")
              }}
              className="border-foreground/[0.08] bg-foreground/[0.06] text-center text-lg tracking-widest focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            />
          </div>
          <Button
            type="submit"
            className="w-full bg-violet-600 text-white hover:bg-violet-500 active:bg-violet-700"
            disabled={totpSubmitting || !totpCode}
          >
            {totpSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {totpSubmitting ? "Проверка..." : "Войти"}
          </Button>
          <button
            type="button"
            className="w-full text-xs text-muted-foreground hover:text-foreground"
            onClick={() => {
              setTotpStep(null)
              setTotpCode("")
              setError("")
            }}
          >
            Вернуться к входу
          </button>
        </form>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <div className="mb-6 text-center">
        <h1 className="text-xl font-semibold text-foreground">Вход в аккаунт</h1>
        <p className="mt-1.5 text-sm text-muted-foreground">Войдите, чтобы продолжить работу</p>
      </div>

      {/* OAuth */}
      <div className="space-y-2.5">
        <Button
          variant="outline"
          size="lg"
          className="w-full gap-2.5 border-border bg-card text-foreground hover:bg-muted"
          onClick={() => window.location.href = withReferral("/api/auth/oauth/github")}
        >
          <svg className="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2z" />
          </svg>
          Войти через GitHub
        </Button>
        <Button
          variant="outline"
          size="lg"
          className="w-full gap-2.5 border-border bg-card text-foreground hover:bg-muted"
          onClick={() => window.location.href = withReferral("/api/auth/oauth/google")}
        >
          <svg className="h-[18px] w-[18px]" viewBox="0 0 24 24">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4" />
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05" />
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
          </svg>
          Войти через Google
        </Button>
        <Button
          variant="outline"
          size="lg"
          className="w-full gap-2.5 border-border bg-card text-foreground hover:bg-muted"
          onClick={() => window.location.href = withReferral("/api/auth/oauth/yandex")}
        >
          <svg className="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none">
            <path d="M2.04 12c0-5.523 4.476-10 10-10 5.522 0 10 4.477 10 10s-4.478 10-10 10c-5.524 0-10-4.477-10-10z" fill="#FC3F1D"/>
            <path d="M13.32 7.666h-.924c-1.694 0-2.585.858-2.585 2.123 0 1.43.616 2.1 1.881 2.959l1.045.704-3.003 4.487H7.49l2.695-4.014c-1.55-1.111-2.42-2.19-2.42-4.015 0-2.288 1.595-3.85 4.62-3.85h3.003v11.868H13.32V7.666z" fill="#fff"/>
          </svg>
          Войти через Яндекс
        </Button>
      </div>

      {/* Разделитель */}
      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-white/[0.06]" />
        </div>
        <div className="relative flex justify-center text-xs uppercase">
          <span className="bg-card px-3 text-muted-foreground">или по email</span>
        </div>
      </div>

      {/* Форма */}
      <form onSubmit={handleSubmit(onSubmit)} noValidate onChange={() => error && setError("")} className="space-y-4">
        {error && (
          <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="space-y-1.5">
          <Label htmlFor="email" className="text-foreground">Email</Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            aria-invalid={!!errors.email}
            aria-describedby={errors.email ? "email-error" : undefined}
            className="border-foreground/[0.08] bg-foreground/[0.06] focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            {...register("email")}
          />
          {errors.email && (
            <p id="email-error" className="text-sm text-destructive">{errors.email.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="password" className="text-foreground">Пароль</Label>
            <Link to="/forgot-password" className="rounded-md px-2 py-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:bg-muted min-h-[44px] flex items-center">
              Забыли пароль?
            </Link>
          </div>
          <div className="relative">
            <Input
              id="password"
              type={showPassword ? "text" : "password"}
              autoComplete="current-password"
              placeholder="••••••••"
              aria-invalid={!!errors.password}
              aria-describedby={errors.password ? "password-error" : undefined}
              className="border-foreground/[0.08] bg-foreground/[0.06] pr-10 focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
              {...register("password")}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center h-11 w-11 rounded-md text-muted-foreground transition-colors hover:text-foreground/70 hover:bg-muted"
              tabIndex={-1}
              aria-label={showPassword ? "Скрыть пароль" : "Показать пароль"}
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          </div>
          {errors.password && (
            <p id="password-error" className="text-sm text-destructive">{errors.password.message}</p>
          )}
        </div>

        <Button
          type="submit"
          className="w-full bg-violet-600 text-white hover:bg-violet-500 active:bg-violet-700"
          disabled={isSubmitting}
        >
          {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {isSubmitting ? "Вход..." : "Войти"}
        </Button>
      </form>

      {/* Регистрация */}
      <p className="mt-6 text-center text-sm text-muted-foreground">
        Нет аккаунта?{" "}
        <Link to="/sign-up" className="inline-flex items-center min-h-[44px] font-medium text-foreground underline underline-offset-4 transition-colors hover:text-foreground">
          Зарегистрироваться
        </Link>
      </p>
    </AuthLayout>
  )
}
