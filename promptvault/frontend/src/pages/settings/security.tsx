import { useState } from "react"
import { useForm, type UseFormRegisterReturn } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Lock, Eye, EyeOff, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { useAuthStore } from "@/stores/auth-store"
import {
  useInitiateSetPassword,
  useConfirmSetPassword,
  useChangePassword,
} from "@/hooks/use-settings"
import { Button } from "@/components/ui/button"
import { SectionHeader } from "./_section-header"

const setPasswordSchema = z
  .object({
    password: z.string().min(8, "Минимум 8 символов").max(128),
    confirm: z.string(),
  })
  .refine((d) => d.password === d.confirm, {
    message: "Пароли не совпадают",
    path: ["confirm"],
  })

const changePasswordSchema = z
  .object({
    old_password: z.string().min(1, "Введите текущий пароль"),
    new_password: z.string().min(8, "Минимум 8 символов").max(128),
    confirm: z.string(),
  })
  .refine((d) => d.new_password === d.confirm, {
    message: "Пароли не совпадают",
    path: ["confirm"],
  })

export default function SettingsSecurityPage() {
  const user = useAuthStore((s) => s.user)
  const fetchMe = useAuthStore((s) => s.fetchMe)
  const [showPasswords, setShowPasswords] = useState(false)

  if (!user) return null

  return (
    <section>
      <SectionHeader
        title="Безопасность"
        description={user.has_password ? "Сменить пароль" : "Установить пароль для входа без OAuth"}
      />
      <div key={user.has_password ? "change" : "set"} className="max-w-md">
        {user.has_password ? (
          <ChangePasswordForm
            showPasswords={showPasswords}
            toggleShow={() => setShowPasswords((v) => !v)}
          />
        ) : (
          <SetPasswordForm
            showPasswords={showPasswords}
            toggleShow={() => setShowPasswords((v) => !v)}
            onUpdate={fetchMe}
          />
        )}
      </div>
    </section>
  )
}

function SetPasswordForm({
  showPasswords,
  toggleShow,
  onUpdate,
}: {
  showPasswords: boolean
  toggleShow: () => void
  onUpdate: () => void
}) {
  const initiateMut = useInitiateSetPassword()
  const confirmMut = useConfirmSetPassword()
  const [codeSent, setCodeSent] = useState(false)
  const [code, setCode] = useState("")
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({ resolver: zodResolver(setPasswordSchema) })

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
        <label className="text-[0.75rem] text-muted-foreground" htmlFor="set-pw-code">Код из email</label>
        <input
          id="set-pw-code"
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
          placeholder="6-значный код"
          className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10 tracking-widest"
          inputMode="numeric"
          maxLength={6}
        />
      </div>
      <PasswordField
        label="Новый пароль"
        placeholder="Минимум 8 символов"
        show={showPasswords}
        toggleShow={toggleShow}
        ToggleIcon={ToggleIcon}
        register={register("password")}
        error={errors.password?.message}
      />
      <PasswordField
        label="Подтвердите пароль"
        placeholder="Повторите пароль"
        show={showPasswords}
        toggleShow={toggleShow}
        ToggleIcon={ToggleIcon}
        register={register("confirm")}
        error={errors.confirm?.message}
      />
      <div className="flex gap-2">
        <SubmitButton isPending={confirmMut.isPending} label="Установить пароль" />
        <button
          type="button"
          onClick={handleSendCode}
          disabled={initiateMut.isPending}
          className="text-xs text-muted-foreground hover:text-foreground disabled:opacity-50"
        >
          Отправить заново
        </button>
      </div>
    </form>
  )
}

function ChangePasswordForm({
  showPasswords,
  toggleShow,
}: {
  showPasswords: boolean
  toggleShow: () => void
}) {
  const changePasswordMut = useChangePassword()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({ resolver: zodResolver(changePasswordSchema) })

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
      <PasswordField
        label="Текущий пароль"
        show={showPasswords}
        toggleShow={toggleShow}
        ToggleIcon={ToggleIcon}
        register={register("old_password")}
        error={errors.old_password?.message}
      />
      <PasswordField
        label="Новый пароль"
        placeholder="Минимум 8 символов"
        show={showPasswords}
        toggleShow={toggleShow}
        ToggleIcon={ToggleIcon}
        register={register("new_password")}
        error={errors.new_password?.message}
      />
      <PasswordField
        label="Подтвердите пароль"
        placeholder="Повторите пароль"
        show={showPasswords}
        toggleShow={toggleShow}
        ToggleIcon={ToggleIcon}
        register={register("confirm")}
        error={errors.confirm?.message}
      />
      <SubmitButton isPending={changePasswordMut.isPending} label="Изменить пароль" />
    </form>
  )
}

function PasswordField({
  label,
  placeholder,
  show,
  toggleShow,
  ToggleIcon,
  register,
  error,
}: {
  label: string
  placeholder?: string
  show: boolean
  toggleShow: () => void
  ToggleIcon: typeof Eye
  register: UseFormRegisterReturn
  error?: string
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
        <button
          type="button"
          onClick={toggleShow}
          className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center h-11 w-11 rounded-md text-muted-foreground hover:text-foreground hover:bg-foreground/[0.04]"
          aria-label={show ? "Скрыть пароль" : "Показать пароль"}
        >
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
