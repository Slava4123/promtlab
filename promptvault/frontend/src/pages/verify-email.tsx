import { useState, useEffect, useCallback } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Loader2, MailCheck } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { AuthLayout } from "@/components/auth/auth-layout"
import { api } from "@/api/client"
import { setTokens } from "@/api/client"
import { useAuthStore } from "@/stores/auth-store"
import type { AuthResponse } from "@/api/types"

export default function VerifyEmail() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const emailFromQuery = searchParams.get("email") || ""
  const fetchMe = useAuthStore((s) => s.fetchMe)

  // MJ-37: единый input вместо 6 раздельных. autoComplete="one-time-code"
  // включает iOS/Android SMS autofill (web.dev/sms-otp-form). 6 раздельных
  // input'ов ломали этот UX — браузер не понимает куда вставлять код.
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [cooldown, setCooldown] = useState(60)

  useEffect(() => {
    if (cooldown <= 0) return
    const timer = setTimeout(() => setCooldown(cooldown - 1), 1000)
    return () => clearTimeout(timer)
  }, [cooldown])

  const handleChange = (value: string) => {
    // Только цифры, до 6 символов; iOS SMS autofill вставит сразу 6.
    const digits = value.replace(/\D/g, "").slice(0, 6)
    setCode(digits)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (code.length !== 6) return

    setError("")
    setIsSubmitting(true)
    try {
      const data = await api<AuthResponse>("/auth/verify-email", {
        method: "POST",
        body: JSON.stringify({ email: emailFromQuery, code }),
      })
      setTokens(data.tokens)
      await fetchMe()
      navigate("/dashboard", { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : "Неверный код")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleResend = useCallback(async () => {
    if (cooldown > 0) return
    try {
      await api("/auth/resend-code", {
        method: "POST",
        body: JSON.stringify({ email: emailFromQuery }),
      })
      setCooldown(60)
      setError("")
    } catch {
      setError("Не удалось отправить код")
    }
  }, [cooldown, emailFromQuery])

  return (
    <AuthLayout>
      <div className="mb-6 flex flex-col items-center gap-3 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-violet-500/10">
          <MailCheck className="h-6 w-6 text-violet-400" />
        </div>
        <h1 className="text-xl font-semibold text-foreground">Подтвердите email</h1>
        <p className="text-sm text-muted-foreground">
          Мы отправили 6-значный код на
          <br />
          <span className="font-medium text-foreground">{emailFromQuery}</span>
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div role="alert" aria-live="assertive" className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-center text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="flex justify-center">
          <Input
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            pattern="[0-9]{6}"
            maxLength={6}
            value={code}
            onChange={(e) => handleChange(e.target.value)}
            placeholder="000000"
            className="h-12 w-48 text-center text-2xl font-mono font-semibold tracking-[0.5em] border-border bg-background focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            autoFocus
          />
        </div>

        <Button
          type="submit"
          className="h-10 w-full bg-violet-600 text-white hover:bg-violet-500 active:bg-violet-700"
          disabled={isSubmitting || code.length !== 6}
        >
          {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {isSubmitting ? "Проверка..." : "Подтвердить"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Не пришёл код?{" "}
        {cooldown > 0 ? (
          <span className="text-muted-foreground">Отправить заново ({cooldown}с)</span>
        ) : (
          <button
            onClick={handleResend}
            className="font-medium text-foreground underline underline-offset-4 transition-colors hover:text-foreground"
          >
            Отправить заново
          </button>
        )}
      </p>
    </AuthLayout>
  )
}
