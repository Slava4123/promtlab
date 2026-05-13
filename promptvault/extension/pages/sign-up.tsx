import { useState, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, Loader2, Sparkles, UserPlus } from "lucide-react"
import { useMutation } from "@tanstack/react-query"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { useSettings } from "../hooks/use-settings"
import { useAuthStore } from "../stores/auth-store"
import { sendBg } from "../lib/bg-client"
import { setApiKey, setApiBase } from "../lib/storage"
import { useToast } from "../components/ui/toaster"
import { ApiError } from "../lib/types"

export function SignUpPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const settings = useSettings()
  const setUser = useAuthStore((s) => s.setUser)
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [name, setName] = useState("")
  const [referredBy, setReferredBy] = useState("")
  const [error, setError] = useState<string | null>(null)

  const registerMut = useMutation({
    mutationFn: () =>
      sendBg({
        type: "api.registerEmailPassword",
        email: email.trim(),
        password,
        name: name.trim(),
        referredBy: referredBy.trim() || undefined,
      }),
    onSuccess: async ({ apiKey, user }) => {
      if (settings) {
        await setApiBase(settings.apiBase)
      }
      await setApiKey(apiKey)
      setUser(user)
      toast({
        title: "Аккаунт создан",
        description: "Подтвердите email — мы отправили код.",
        variant: "success",
      })
      navigate("/", { replace: true })
    },
    onError: (err: Error) => {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Не удалось зарегистрироваться")
      }
    },
  })

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!email.trim() || !password || !name.trim()) return
    setError(null)
    registerMut.mutate()
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
          <Sparkles className="h-4 w-4 text-(--color-brand)" />
          <h1 className="text-base font-semibold">Создать аккаунт</h1>
        </div>
      </div>

      <form onSubmit={onSubmit} className="flex flex-col gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="reg-name">Имя</Label>
          <Input
            id="reg-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Иван Петров"
            required
            disabled={registerMut.isPending}
            maxLength={100}
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="reg-email">Email</Label>
          <Input
            id="reg-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            required
            disabled={registerMut.isPending}
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="reg-password">Пароль</Label>
          <Input
            id="reg-password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Минимум 8 символов"
            required
            minLength={8}
            disabled={registerMut.isPending}
          />
          <p className="text-[10px] text-(--color-muted-foreground)">
            Минимум 8 символов.
          </p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="reg-ref" className="text-(--color-muted-foreground)">
            Реферальный код (необязательно)
          </Label>
          <Input
            id="reg-ref"
            value={referredBy}
            onChange={(e) => setReferredBy(e.target.value)}
            placeholder="12345678"
            maxLength={8}
            disabled={registerMut.isPending}
            className="font-mono"
          />
        </div>

        {error && (
          <div className="rounded-md border border-(--color-destructive) bg-(--color-destructive)/10 p-2 text-xs text-(--color-destructive)">
            {error}
          </div>
        )}

        <Button
          type="submit"
          variant="brand"
          disabled={
            registerMut.isPending ||
            !email.trim() ||
            !password ||
            !name.trim() ||
            password.length < 8
          }
          className="w-full gap-1.5"
        >
          {registerMut.isPending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <UserPlus className="h-3.5 w-3.5" />
          )}
          Создать аккаунт
        </Button>

        <p className="text-center text-[10px] text-(--color-muted-foreground)">
          Уже есть аккаунт?{" "}
          <a href="#/sign-in" className="text-(--color-brand) hover:underline">
            Войти
          </a>
        </p>
      </form>
    </div>
  )
}
