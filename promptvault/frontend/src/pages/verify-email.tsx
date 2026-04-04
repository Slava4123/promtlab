import { useState, useRef, useEffect, useCallback } from "react"
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

  const [code, setCode] = useState(["", "", "", "", "", ""])
  const [error, setError] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [cooldown, setCooldown] = useState(60)
  const inputRefs = useRef<(HTMLInputElement | null)[]>([])

  useEffect(() => {
    if (cooldown <= 0) return
    const timer = setTimeout(() => setCooldown(cooldown - 1), 1000)
    return () => clearTimeout(timer)
  }, [cooldown])

  const handleChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return

    const newCode = [...code]
    newCode[index] = value.slice(-1)
    setCode(newCode)

    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus()
    }
  }

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === "Backspace" && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus()
    }
  }

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault()
    const pasted = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, 6)
    if (pasted.length === 6) {
      setCode(pasted.split(""))
      inputRefs.current[5]?.focus()
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const fullCode = code.join("")
    if (fullCode.length !== 6) return

    setError("")
    setIsSubmitting(true)
    try {
      const data = await api<AuthResponse>("/auth/verify-email", {
        method: "POST",
        body: JSON.stringify({ email: emailFromQuery, code: fullCode }),
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
        <h1 className="text-xl font-semibold text-white">Подтвердите email</h1>
        <p className="text-sm text-zinc-500">
          Мы отправили 6-значный код на
          <br />
          <span className="font-medium text-zinc-300">{emailFromQuery}</span>
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-center text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="flex justify-center gap-2" onPaste={handlePaste}>
          {code.map((digit, i) => (
            <Input
              key={i}
              ref={(el) => { inputRefs.current[i] = el }}
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={digit}
              onChange={(e) => handleChange(i, e.target.value)}
              onKeyDown={(e) => handleKeyDown(i, e)}
              className="h-12 w-11 text-center text-lg font-semibold border-white/[0.08] bg-white/[0.04] focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            />
          ))}
        </div>

        <Button
          type="submit"
          className="h-10 w-full bg-violet-600 text-white hover:bg-violet-500 active:bg-violet-700"
          disabled={isSubmitting || code.join("").length !== 6}
        >
          {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {isSubmitting ? "Проверка..." : "Подтвердить"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-zinc-500">
        Не пришёл код?{" "}
        {cooldown > 0 ? (
          <span className="text-zinc-600">Отправить заново ({cooldown}с)</span>
        ) : (
          <button
            onClick={handleResend}
            className="font-medium text-zinc-200 underline underline-offset-4 transition-colors hover:text-white"
          >
            Отправить заново
          </button>
        )}
      </p>
    </AuthLayout>
  )
}
