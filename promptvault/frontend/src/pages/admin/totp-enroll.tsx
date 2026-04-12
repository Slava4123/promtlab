import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Loader2, Copy, Check, AlertTriangle, ShieldCheck } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAdminGuard } from "@/hooks/admin/use-admin-guard"
import { totpEnroll, totpConfirmEnrollment, totpStatus } from "@/api/admin/totp"
import type { TOTPEnrollResponse } from "@/api/types"

type WizardStep = "loading" | "generate" | "display" | "confirm" | "done"

export default function AdminTOTPEnrollPage() {
  const { isAdmin, isLoading: guardLoading } = useAdminGuard()
  const navigate = useNavigate()
  const [step, setStep] = useState<WizardStep>("loading")
  const [enrollData, setEnrollData] = useState<TOTPEnrollResponse | null>(null)
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  // При монтировании проверяем текущий статус TOTP.
  // Если уже confirmed — сразу показываем "done".
  useEffect(() => {
    if (guardLoading || !isAdmin) return
    totpStatus()
      .then((s) => {
        if (s.confirmed) {
          setStep("done")
        } else {
          setStep("generate")
        }
      })
      .catch(() => setStep("generate"))
  }, [guardLoading, isAdmin])

  const handleEnroll = async () => {
    setBusy(true)
    setError("")
    try {
      const data = await totpEnroll()
      setEnrollData(data)
      setStep("display")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка")
    } finally {
      setBusy(false)
    }
  }

  const handleConfirm = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!code) return
    setBusy(true)
    setError("")
    try {
      await totpConfirmEnrollment(code)
      toast.success("TOTP настроен", {
        description: "Теперь при входе будет запрашиваться код",
      })
      setStep("done")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Неверный код")
    } finally {
      setBusy(false)
    }
  }

  const copy = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      toast.success(`${label} скопирован`)
    })
  }

  if (guardLoading || step === "loading") {
    return (
      <div className="flex h-40 items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }
  if (!isAdmin) return null

  if (step === "done") {
    return (
      <div className="mx-auto max-w-lg space-y-6 py-8 text-center">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/15">
          <ShieldCheck className="h-6 w-6 text-emerald-400" />
        </div>
        <div>
          <h2 className="text-xl font-semibold">TOTP уже настроен</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Двухфакторная проверка активна. При следующем login введите код из Authenticator.
          </p>
        </div>
        <Button onClick={() => navigate("/admin/users")}>Перейти к админ-панели</Button>
      </div>
    )
  }

  if (step === "generate") {
    return (
      <div className="mx-auto max-w-lg space-y-6 py-8">
        <div className="text-center">
          <h2 className="text-xl font-semibold">Настройка TOTP 2FA</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Для доступа к админ-панели требуется настроить двухфакторную проверку.
          </p>
        </div>
        {error && (
          <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {error}
          </div>
        )}
        <div className="rounded-xl border border-border bg-muted/20 p-4 text-sm text-muted-foreground">
          <p className="font-medium text-foreground">Что понадобится:</p>
          <ul className="mt-2 list-disc pl-5 space-y-1">
            <li>Приложение Authenticator (Google Authenticator / 1Password / Authy / Bitwarden)</li>
            <li>Надёжное место для сохранения backup-кодов (password manager)</li>
          </ul>
        </div>
        <Button onClick={handleEnroll} disabled={busy} className="w-full">
          {busy && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Начать настройку
        </Button>
      </div>
    )
  }

  if (step === "display" && enrollData) {
    return (
      <div className="mx-auto max-w-lg space-y-6 py-8">
        <div>
          <h2 className="text-xl font-semibold">Шаг 1: Добавьте в Authenticator</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Скопируйте secret или откройте otpauth:// ссылку на мобильном устройстве.
          </p>
        </div>

        <div className="space-y-3">
          <div>
            <Label className="text-xs text-muted-foreground">Secret</Label>
            <div className="mt-1 flex items-center gap-2">
              <code className="flex-1 break-all rounded-md bg-muted/40 px-3 py-2 font-mono text-xs">
                {enrollData.secret}
              </code>
              <Button
                variant="outline"
                size="sm"
                onClick={() => copy(enrollData.secret, "Secret")}
              >
                <Copy className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          <div>
            <Label className="text-xs text-muted-foreground">otpauth:// URL</Label>
            <div className="mt-1 flex items-center gap-2">
              <code className="flex-1 break-all rounded-md bg-muted/40 px-3 py-2 font-mono text-xs">
                {enrollData.qr_url}
              </code>
              <Button
                variant="outline"
                size="sm"
                onClick={() => copy(enrollData.qr_url, "URL")}
              >
                <Copy className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-4">
          <div className="flex items-start gap-2">
            <AlertTriangle className="h-4 w-4 shrink-0 text-destructive mt-0.5" />
            <div className="text-sm">
              <p className="font-medium text-destructive">Шаг 2: Сохраните backup-коды</p>
              <p className="mt-0.5 text-xs text-muted-foreground">
                Коды показываются ОДИН РАЗ. Сохраните их в password manager.
                Если потеряете телефон — сможете войти через backup-код.
              </p>
            </div>
          </div>
          <div className="mt-3 grid grid-cols-2 gap-1.5 font-mono text-xs">
            {enrollData.backup_codes.map((code) => (
              <div key={code} className="rounded-md bg-background px-2 py-1.5">
                {code}
              </div>
            ))}
          </div>
          <Button
            variant="outline"
            size="sm"
            className="mt-3 w-full"
            onClick={() => copy(enrollData.backup_codes.join("\n"), "Backup codes")}
          >
            <Copy className="mr-2 h-3.5 w-3.5" />
            Скопировать все коды
          </Button>
        </div>

        <Button className="w-full" onClick={() => setStep("confirm")}>
          Я сохранил(а) — продолжить
        </Button>
      </div>
    )
  }

  if (step === "confirm") {
    return (
      <div className="mx-auto max-w-lg space-y-6 py-8">
        <div className="text-center">
          <h2 className="text-xl font-semibold">Шаг 3: Подтвердите код</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Введите текущий код из Authenticator — это активирует TOTP.
          </p>
        </div>
        <form onSubmit={handleConfirm} className="space-y-4">
          {error && (
            <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}
          <div>
            <Label htmlFor="confirm_code">Код</Label>
            <Input
              id="confirm_code"
              inputMode="numeric"
              autoComplete="one-time-code"
              placeholder="000000"
              autoFocus
              value={code}
              onChange={(e) => {
                setCode(e.target.value)
                if (error) setError("")
              }}
              className="mt-1 text-center text-lg tracking-widest"
            />
          </div>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setStep("display")}
              disabled={busy}
            >
              Назад
            </Button>
            <Button type="submit" className="flex-1" disabled={busy || !code}>
              {busy && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {busy ? "Проверка..." : "Активировать"}
            </Button>
          </div>
        </form>
      </div>
    )
  }

  // Fallback — should never happen.
  return <div className="p-4 text-muted-foreground">Неожиданное состояние: {step}</div>
}

// Unused import suppression for tooling.
void Check
