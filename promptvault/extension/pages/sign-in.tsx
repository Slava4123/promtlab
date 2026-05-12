import { useEffect, useState, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import {
  ExternalLink,
  KeyRound,
  Loader2,
  Mail,
  Sparkles,
  UserPlus,
} from "lucide-react"
import { useMutation } from "@tanstack/react-query"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { useSettings } from "../hooks/use-settings"
import { ApiKeySetup } from "../components/api-key-setup"
import { useAuthStore } from "../stores/auth-store"
import { sendBg } from "../lib/bg-client"
import { setApiKey, setApiBase } from "../lib/storage"
import { useToast } from "../components/ui/toaster"
import { ApiError } from "../lib/types"
import { openWebPage, deriveFrontendUrl } from "../lib/utils"
import { cn } from "../lib/utils"

type AuthMode = "email" | "apikey"

// Phase 6: sign-in поддерживает два mode'а:
//   - email/password — login через /api/auth/login → auto-create API key
//     "Chrome Extension" → сохранить как pvlt_*
//   - apikey — старый flow (для пользователей без пароля или для подключения
//     к self-hosted instance без email confirm)
export function SignInPage() {
  const navigate = useNavigate()
  const settings = useSettings()
  const [mode, setMode] = useState<AuthMode>("email")

  useEffect(() => {
    if (settings?.apiKey) navigate("/", { replace: true })
  }, [settings?.apiKey, navigate])

  if (!settings) return null

  return (
    <div className="flex h-full flex-col overflow-y-auto p-5 gap-4">
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-(--color-primary)" />
          <h1 className="text-lg font-semibold">ПромтЛаб</h1>
        </div>
        <p className="text-xs text-(--color-muted-foreground)">
          Войдите, чтобы синхронизировать промпты на 9 AI-сайтах.
        </p>
      </div>

      {/* Mode tabs */}
      <div className="grid grid-cols-2 gap-1 rounded-md border border-(--color-border) p-0.5 text-xs">
        <button
          type="button"
          onClick={() => setMode("email")}
          className={cn(
            "rounded px-2 py-1.5 font-medium transition-colors",
            mode === "email"
              ? "bg-(--color-primary) text-(--color-primary-foreground)"
              : "text-(--color-muted-foreground) hover:bg-(--color-muted)",
          )}
        >
          Email
        </button>
        <button
          type="button"
          onClick={() => setMode("apikey")}
          className={cn(
            "rounded px-2 py-1.5 font-medium transition-colors",
            mode === "apikey"
              ? "bg-(--color-primary) text-(--color-primary-foreground)"
              : "text-(--color-muted-foreground) hover:bg-(--color-muted)",
          )}
        >
          API-ключ
        </button>
      </div>

      {mode === "email" ? (
        <EmailPasswordForm apiBase={settings.apiBase} />
      ) : (
        <div className="flex-1">
          <ApiKeySetup initialBase={settings.apiBase} />
        </div>
      )}
    </div>
  )
}

function EmailPasswordForm({ apiBase }: { apiBase: string }) {
  const { toast } = useToast()
  const setUser = useAuthStore((s) => s.setUser)
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState<string | null>(null)

  const loginMut = useMutation({
    mutationFn: () =>
      sendBg({ type: "api.loginEmailPassword", email: email.trim(), password }),
    onSuccess: async ({ apiKey, user }) => {
      await setApiBase(apiBase)
      await setApiKey(apiKey)
      setUser(user)
      toast({ title: `Привет, ${user.name || user.email}`, variant: "success" })
    },
    onError: (err: Error) => {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Не удалось войти")
      }
    },
  })

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!email.trim() || !password) return
    setError(null)
    loginMut.mutate()
  }

  function openOAuth(provider: "google" | "github" | "yandex") {
    openWebPage(apiBase, `/sign-in?provider=${provider}&from=extension`)
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-3">
      {/* OAuth */}
      <div className="grid grid-cols-3 gap-1.5">
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => openOAuth("google")}
          className="h-8 text-[10px] gap-1"
        >
          <ExternalLink className="h-3 w-3" />
          Google
        </Button>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => openOAuth("github")}
          className="h-8 text-[10px] gap-1"
        >
          <ExternalLink className="h-3 w-3" />
          GitHub
        </Button>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => openOAuth("yandex")}
          className="h-8 text-[10px] gap-1"
        >
          <ExternalLink className="h-3 w-3" />
          Yandex
        </Button>
      </div>

      <p className="text-center text-[10px] text-(--color-muted-foreground)">
        OAuth откроется в новой вкладке. Вернитесь сюда после авторизации.
      </p>

      <div className="relative">
        <div className="absolute inset-y-1/2 left-0 right-0 h-px bg-(--color-border)" />
        <span className="relative mx-auto block w-fit bg-(--color-background) px-2 text-[10px] uppercase tracking-wide text-(--color-muted-foreground)">
          или
        </span>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          type="email"
          autoComplete="username"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@example.com"
          required
          disabled={loginMut.isPending}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="password">Пароль</Label>
        <Input
          id="password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          disabled={loginMut.isPending}
        />
      </div>

      {error && (
        <div className="rounded-md border border-(--color-destructive) bg-(--color-destructive)/10 p-2 text-xs text-(--color-destructive)">
          {error}
        </div>
      )}

      <Button
        type="submit"
        disabled={loginMut.isPending || !email.trim() || !password}
        className="w-full gap-1.5"
      >
        {loginMut.isPending ? (
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
        ) : (
          <Mail className="h-3.5 w-3.5" />
        )}
        Войти
      </Button>

      <div className="flex items-center justify-between text-[10px]">
        <a
          href="#/sign-up"
          className="flex items-center gap-1 text-(--color-primary) hover:underline"
        >
          <UserPlus className="h-3 w-3" />
          Создать аккаунт
        </a>
        <a
          href="#/forgot-password"
          className="flex items-center gap-1 text-(--color-muted-foreground) hover:underline"
        >
          <KeyRound className="h-3 w-3" />
          Забыли пароль?
        </a>
      </div>

      <p className="text-center text-[9px] text-(--color-muted-foreground)">
        Без аккаунта?{" "}
        <button
          type="button"
          onClick={() => openWebPage(apiBase, "/sign-up?from=extension")}
          className="text-(--color-primary) hover:underline"
        >
          Зарегистрироваться на {deriveFrontendUrl(apiBase).replace(/^https?:\/\//, "")}
        </button>
      </p>
    </form>
  )
}
