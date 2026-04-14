import { useState, useEffect } from "react"
import { useForm, type UseFormRegisterReturn } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Settings, User, Lock, Link2, Palette, Eye, EyeOff, Loader2, Check, AlertTriangle } from "lucide-react"
import { toast } from "sonner"

import { useAuthStore } from "@/stores/auth-store"
import { useThemeStore } from "@/stores/theme-store"
import { useQueryClient } from "@tanstack/react-query"
import { api } from "@/api/client"
import { Button } from "@/components/ui/button"
import {
  useLinkedAccounts,
  useUpdateProfile,
  useInitiateSetPassword,
  useConfirmSetPassword,
  useChangePassword,
  useUnlinkProvider,
} from "@/hooks/use-settings"
import { APIKeysSection } from "@/components/settings/api-keys-section"
import { ExtensionPromoSection } from "@/components/settings/extension-promo-section"
import { SubscriptionSection } from "@/components/subscription/subscription-section"

// --- Schemas ---

const profileSchema = z.object({
  name: z.string().min(1, "Введите имя").max(100),
  username: z.string().max(30).regex(/^[a-zA-Z0-9_]*$/, "Только латинские буквы, цифры и _").optional().or(z.literal("")),
})

const setPasswordSchema = z.object({
  password: z.string().min(8, "Минимум 8 символов").max(128),
  confirm: z.string(),
}).refine((d) => d.password === d.confirm, {
  message: "Пароли не совпадают",
  path: ["confirm"],
})

const changePasswordSchema = z.object({
  old_password: z.string().min(1, "Введите текущий пароль"),
  new_password: z.string().min(8, "Минимум 8 символов").max(128),
  confirm: z.string(),
}).refine((d) => d.new_password === d.confirm, {
  message: "Пароли не совпадают",
  path: ["confirm"],
})

// --- Page ---

export default function SettingsPage() {
  const user = useAuthStore((s) => s.user)
  const isLoading = useAuthStore((s) => s.isLoading)
  const fetchMe = useAuthStore((s) => s.fetchMe)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!user) return null

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <div className="flex items-center gap-3">
        <Settings className="h-5 w-5 text-brand-muted-foreground" />
        <div>
          <h1 className="text-xl font-semibold text-foreground">Настройки</h1>
          <p className="text-sm text-muted-foreground">Управление профилем и безопасностью</p>
        </div>
      </div>

      <ProfileSection user={user} onUpdate={fetchMe} />
      <PasswordSection key={user.has_password ? "change" : "set"} hasPassword={user.has_password} onUpdate={fetchMe} />
      <LinkedAccountsSection />
      <SubscriptionSection />
      <ExtensionPromoSection />
      <APIKeysSection />
      <ThemeSection />
    </div>
  )
}

// --- Profile ---

function ProfileSection({ user, onUpdate }: { user: { name: string; email: string; avatar_url?: string; username?: string }; onUpdate: () => void }) {
  const updateProfile = useUpdateProfile()
  const { register, handleSubmit, formState: { errors } } = useForm({
    resolver: zodResolver(profileSchema),
    defaultValues: { name: user.name, username: user.username || "" },
  })

  const onSubmit = async (data: { name: string; username?: string }) => {
    try {
      await updateProfile.mutateAsync({ name: data.name, username: data.username || undefined })
      onUpdate()
      toast.success("Профиль обновлён")
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка обновления")
    }
  }

  const initials = user.name
    .split(" ")
    .filter(Boolean)
    .map((w) => w[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) || "?"

  return (
    <Section icon={User} title="Профиль">
      <div className="flex items-center gap-4 mb-4">
        {user.avatar_url ? (
          <img src={user.avatar_url} alt={user.name} className="h-14 w-14 rounded-full object-cover" />
        ) : (
          <div className="flex h-14 w-14 items-center justify-center rounded-full text-lg font-semibold text-brand-foreground [background:var(--brand-gradient)]">
            {initials}
          </div>
        )}
        <div>
          <p className="text-sm font-medium text-foreground">{user.name}</p>
          <p className="text-xs text-muted-foreground">{user.email}</p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
        <div>
          <label className="text-[0.75rem] text-muted-foreground">Имя</label>
          <input
            {...register("name")}
            className="mt-1 h-11 w-full rounded-lg border border-border bg-background px-3 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
          />
          {errors.name && <p className="mt-1 text-xs text-red-400">{errors.name.message}</p>}
        </div>

        <div>
          <label className="text-[0.75rem] text-muted-foreground">Никнейм</label>
          <div className="relative mt-1">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">@</span>
            <input
              {...register("username")}
              placeholder="username"
              className="h-11 w-full rounded-lg border border-border bg-background pl-7 pr-3 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
            />
          </div>
          {errors.username && <p className="mt-1 text-xs text-red-400">{errors.username.message}</p>}
        </div>

        <div>
          <label className="text-[0.75rem] text-muted-foreground">Email</label>
          <input
            value={user.email}
            disabled
            className="mt-1 h-11 w-full rounded-lg border border-border bg-muted px-3 text-sm text-muted-foreground outline-none cursor-not-allowed"
          />
        </div>

        <Button type="submit" variant="brand" disabled={updateProfile.isPending}>
          {updateProfile.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
          Сохранить
        </Button>
      </form>
    </Section>
  )
}

// --- Password ---

function PasswordSection({ hasPassword, onUpdate }: { hasPassword: boolean; onUpdate: () => void }) {
  const [showPasswords, setShowPasswords] = useState(false)

  return (
    <Section icon={Lock} title="Безопасность">
      {hasPassword ? (
        <ChangePasswordForm showPasswords={showPasswords} toggleShow={() => setShowPasswords(!showPasswords)} />
      ) : (
        <SetPasswordForm showPasswords={showPasswords} toggleShow={() => setShowPasswords(!showPasswords)} onUpdate={onUpdate} />
      )}
    </Section>
  )
}

function SetPasswordForm({ showPasswords, toggleShow, onUpdate }: { showPasswords: boolean; toggleShow: () => void; onUpdate: () => void }) {
  const initiateMut = useInitiateSetPassword()
  const confirmMut = useConfirmSetPassword()
  const [codeSent, setCodeSent] = useState(false)
  const [code, setCode] = useState("")
  const { register, handleSubmit, reset, formState: { errors } } = useForm({
    resolver: zodResolver(setPasswordSchema),
  })

  const handleSendCode = async () => {
    try {
      await initiateMut.mutateAsync()
      setCodeSent(true)
      toast.success("Код отправлен на email")
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка отправки кода")
    }
  }

  const onSubmit = handleSubmit(async (data) => {
    try {
      await confirmMut.mutateAsync({ code, password: data.password })
      toast.success("Пароль установлен")
      reset()
      setCode("")
      setCodeSent(false)
      onUpdate()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
    }
  })

  const ToggleIcon = showPasswords ? EyeOff : Eye

  if (!codeSent) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-muted-foreground">
          Для установки пароля нужно подтвердить email. Мы отправим код на вашу почту.
        </p>
        <Button type="button" variant="brand" onClick={handleSendCode} disabled={initiateMut.isPending}>
          {initiateMut.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Lock className="h-4 w-4" />}
          Отправить код
        </Button>
      </div>
    )
  }

  return (
    <form onSubmit={onSubmit} className="space-y-3">
      <div>
        <label className="text-[0.75rem] text-muted-foreground">Код из email</label>
        <input
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
          placeholder="6-значный код"
          className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10 tracking-widest"
          inputMode="numeric"
          maxLength={6}
        />
      </div>
      <PasswordField label="Новый пароль" placeholder="Минимум 8 символов" show={showPasswords}
        toggleShow={toggleShow} ToggleIcon={ToggleIcon} register={register("password")}
        error={errors.password?.message} />
      <PasswordField label="Подтвердите пароль" placeholder="Повторите пароль" show={showPasswords}
        toggleShow={toggleShow} ToggleIcon={ToggleIcon} register={register("confirm")}
        error={errors.confirm?.message} />
      <div className="flex gap-2">
        <SubmitButton isPending={confirmMut.isPending} label="Установить пароль" />
        <button type="button" onClick={handleSendCode} disabled={initiateMut.isPending}
          className="text-xs text-muted-foreground hover:text-foreground disabled:opacity-50">
          Отправить заново
        </button>
      </div>
    </form>
  )
}

function ChangePasswordForm({ showPasswords, toggleShow }: { showPasswords: boolean; toggleShow: () => void }) {
  const changePasswordMut = useChangePassword()
  const { register, handleSubmit, reset, formState: { errors } } = useForm({
    resolver: zodResolver(changePasswordSchema),
  })

  const onSubmit = handleSubmit(async (data) => {
    try {
      await changePasswordMut.mutateAsync({
        old_password: data.old_password,
        new_password: data.new_password,
      })
      toast.success("Пароль изменён")
      reset()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
    }
  })

  const ToggleIcon = showPasswords ? EyeOff : Eye

  return (
    <form onSubmit={onSubmit} className="space-y-3">
      <PasswordField label="Текущий пароль" show={showPasswords}
        toggleShow={toggleShow} ToggleIcon={ToggleIcon} register={register("old_password")}
        error={errors.old_password?.message} />
      <PasswordField label="Новый пароль" placeholder="Минимум 8 символов" show={showPasswords}
        toggleShow={toggleShow} ToggleIcon={ToggleIcon} register={register("new_password")}
        error={errors.new_password?.message} />
      <PasswordField label="Подтвердите пароль" placeholder="Повторите пароль" show={showPasswords}
        toggleShow={toggleShow} ToggleIcon={ToggleIcon} register={register("confirm")}
        error={errors.confirm?.message} />
      <SubmitButton isPending={changePasswordMut.isPending} label="Изменить пароль" />
    </form>
  )
}

function PasswordField({ label, placeholder, show, toggleShow, ToggleIcon, register, error }: {
  label: string; placeholder?: string; show: boolean; toggleShow: () => void
  ToggleIcon: typeof Eye; register: UseFormRegisterReturn; error?: string
}) {
  return (
    <div>
      <label className="text-[0.75rem] text-muted-foreground">{label}</label>
      <div className="relative mt-1">
        <input
          type={show ? "text" : "password"}
          placeholder={placeholder}
          {...register}
          className="h-11 w-full rounded-lg border border-border bg-background px-3 pr-12 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
        />
        <button type="button" onClick={toggleShow}
          className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center h-11 w-11 rounded-md text-muted-foreground hover:text-foreground hover:bg-foreground/[0.04]"
          aria-label={show ? "Скрыть пароль" : "Показать пароль"}>
          <ToggleIcon className="h-4 w-4" />
        </button>
      </div>
      {error && <p className="mt-1 text-xs text-red-400">{error}</p>}
    </div>
  )
}

function SubmitButton({ isPending, label }: { isPending: boolean; label: string }) {
  return (
    <Button type="submit" variant="brand" disabled={isPending}>
      {isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Lock className="h-4 w-4" />}
      {label}
    </Button>
  )
}

// --- Linked Accounts ---

const providers = [
  { key: "github", name: "GitHub" },
  { key: "google", name: "Google" },
  { key: "yandex", name: "Яндекс" },
]

const linkErrorMessages: Record<string, string> = {
  linked_to_other: "Этот аккаунт уже привязан к другому пользователю",
  already_linked: "Этот провайдер уже привязан",
  not_configured: "OAuth-провайдер не настроен",
  exchange_failed: "Ошибка авторизации через провайдер",
}

const providerNames: Record<string, string> = {
  github: "GitHub",
  google: "Google",
  yandex: "Яндекс",
}

function LinkedAccountsSection() {
  const { data: accounts, isLoading, isError, error } = useLinkedAccounts()
  const unlinkMut = useUnlinkProvider()
  const queryClient = useQueryClient()
  const [confirmUnlink, setConfirmUnlink] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const linked = params.get("linked")
    const linkError = params.get("link_error")
    if (linked) {
      toast.success(`${providerNames[linked] || linked} привязан`)
      queryClient.invalidateQueries({ queryKey: ["linked-accounts"] })
      window.history.replaceState({}, "", "/settings")
    }
    if (linkError) {
      toast.error(linkErrorMessages[linkError] || "Ошибка привязки")
      window.history.replaceState({}, "", "/settings")
    }
  }, [queryClient])

  const linkedProviders = new Set(accounts?.map((a) => a.provider) ?? [])

  const handleUnlink = async (provider: string) => {
    try {
      await unlinkMut.mutateAsync(provider)
      toast.success("Аккаунт отвязан")
      setConfirmUnlink(null)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
      setConfirmUnlink(null)
    }
  }

  const handleLink = async (provider: string) => {
    try {
      const res = await api<{ redirect_url: string }>(`/auth/link/${provider}`, { method: "POST" })
      window.location.assign(res.redirect_url)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка привязки")
    }
  }

  return (
    <Section icon={Link2} title="Привязанные аккаунты">
      {isError ? (
        <div className="py-4 text-sm text-red-400">
          Не удалось загрузить: {error?.message || "Ошибка сервера"}
        </div>
      ) : isLoading ? (
        <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" /> Загрузка...
        </div>
      ) : (
        <div className="space-y-2">
          {providers.map(({ key, name }) => {
            const isLinked = linkedProviders.has(key)
            return (
              <div key={key}>
                <div className="flex items-center justify-between rounded-lg border border-border bg-background/50 px-3 py-2.5">
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-foreground">{name}</span>
                    {isLinked && (
                      <span className="rounded-full bg-green-500/10 px-2 py-0.5 text-[0.65rem] font-medium text-green-500">
                        Привязан
                      </span>
                    )}
                  </div>
                  {isLinked ? (
                    <button
                      onClick={() => setConfirmUnlink(key)}
                      disabled={unlinkMut.isPending}
                      className="rounded-lg px-3 py-2 text-[0.78rem] text-red-400 hover:text-red-300 hover:bg-red-500/10 disabled:opacity-50 min-h-[44px]"
                    >
                      Отвязать
                    </button>
                  ) : (
                    <button
                      onClick={() => handleLink(key)}
                      className="rounded-lg px-3 py-2 text-[0.78rem] text-brand-muted-foreground hover:text-brand hover:bg-brand-muted min-h-[44px]"
                    >
                      Привязать
                    </button>
                  )}
                </div>

                {/* Confirm dialog */}
                {confirmUnlink === key && (
                  <div className="mt-2 rounded-lg border border-red-500/20 bg-red-500/5 px-3 py-2.5">
                    <div className="flex items-center gap-2">
                      <AlertTriangle className="h-4 w-4 shrink-0 text-red-400" />
                      <p className="text-xs text-muted-foreground">Отвязать {name}?</p>
                    </div>
                    <div className="mt-2 flex justify-end gap-2">
                      <button
                        onClick={() => setConfirmUnlink(null)}
                        className="rounded-lg px-3 min-h-[44px] text-xs text-muted-foreground hover:text-foreground"
                      >
                        Отмена
                      </button>
                      <button
                        onClick={() => handleUnlink(key)}
                        disabled={unlinkMut.isPending}
                        className="rounded-lg bg-red-500/10 px-3 min-h-[44px] text-xs text-red-400 hover:bg-red-500/20 disabled:opacity-50"
                      >
                        {unlinkMut.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : "Да, отвязать"}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </Section>
  )
}

// --- Theme ---

function ThemeSection() {
  const { theme, setTheme } = useThemeStore()

  return (
    <Section icon={Palette} title="Оформление">
      <div className="flex flex-wrap gap-2">
        {(["dark", "light", "system"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTheme(t)}
            className={`rounded-lg border px-4 h-11 text-sm font-medium transition-colors ${
              theme === t
                ? "border-brand/40 bg-brand-muted text-brand-muted-foreground"
                : "border-border bg-background text-muted-foreground hover:text-foreground"
            }`}
          >
            {t === "dark" ? "Тёмная" : t === "light" ? "Светлая" : "Системная"}
          </button>
        ))}
      </div>
    </Section>
  )
}

// --- Shared ---

function Section({ icon: Icon, title, children }: { icon: typeof Settings; title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-border bg-card p-5 overflow-hidden">
      <div className="mb-4 flex items-center gap-2">
        <Icon className="h-4 w-4 text-brand-muted-foreground" />
        <h2 className="text-sm font-semibold text-foreground">{title}</h2>
      </div>
      {children}
    </div>
  )
}
