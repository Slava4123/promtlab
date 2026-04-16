import { useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Eye, EyeOff, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { AuthLayout } from "@/components/auth/auth-layout"
import { api } from "@/api/client"

const registerSchema = z.object({
  name: z.string().min(1, "Введите имя").max(100),
  email: z.email("Введите корректный email"),
  password: z.string().min(8, "Минимум 8 символов").max(128),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Пароли не совпадают",
  path: ["confirmPassword"],
})

type RegisterForm = z.infer<typeof registerSchema>

export default function SignUp() {
  const navigate = useNavigate()
  const [error, setError] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data: RegisterForm) => {
    setError("")
    try {
      await api("/auth/register", {
        method: "POST",
        body: JSON.stringify({ email: data.email, password: data.password, name: data.name }),
      })
      navigate(`/verify-email?email=${encodeURIComponent(data.email)}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка регистрации")
    }
  }

  return (
    <AuthLayout>
      <div className="mb-6 text-center">
        <h1 className="text-xl font-semibold text-foreground">Создать аккаунт</h1>
        <p className="mt-1.5 text-sm text-muted-foreground">Заполните данные для регистрации</p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} noValidate onChange={() => error && setError("")} className="space-y-4">
        {error && (
          <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="space-y-1.5">
          <Label htmlFor="name" className="text-foreground">Имя</Label>
          <Input
            id="name"
            autoComplete="name"
            placeholder="Ваше имя"
            aria-invalid={!!errors.name}
            aria-describedby={errors.name ? "name-error" : undefined}
            className="border-border bg-card focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            {...register("name")}
          />
          {errors.name && (
            <p id="name-error" className="text-sm text-destructive">{errors.name.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="email" className="text-foreground">Email</Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            aria-invalid={!!errors.email}
            aria-describedby={errors.email ? "email-error" : undefined}
            className="border-border bg-card focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
            {...register("email")}
          />
          {errors.email && (
            <p id="email-error" className="text-sm text-destructive">{errors.email.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="password" className="text-foreground">Пароль</Label>
          <div className="relative">
            <Input
              id="password"
              type={showPassword ? "text" : "password"}
              autoComplete="new-password"
              placeholder="Минимум 8 символов"
              aria-invalid={!!errors.password}
              aria-describedby={errors.password ? "password-error" : undefined}
              className="border-border bg-card pr-10 focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
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

        <div className="space-y-1.5">
          <Label htmlFor="confirmPassword" className="text-foreground">Подтвердите пароль</Label>
          <div className="relative">
            <Input
              id="confirmPassword"
              type={showConfirm ? "text" : "password"}
              autoComplete="new-password"
              placeholder="Повторите пароль"
              aria-invalid={!!errors.confirmPassword}
              aria-describedby={errors.confirmPassword ? "confirmPassword-error" : undefined}
              className="border-border bg-card pr-10 focus-visible:border-violet-500/50 focus-visible:ring-violet-500/20"
              {...register("confirmPassword")}
            />
            <button
              type="button"
              onClick={() => setShowConfirm(!showConfirm)}
              className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center h-11 w-11 rounded-md text-muted-foreground transition-colors hover:text-foreground/70 hover:bg-muted"
              tabIndex={-1}
              aria-label={showConfirm ? "Скрыть пароль" : "Показать пароль"}
            >
              {showConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          </div>
          {errors.confirmPassword && (
            <p id="confirmPassword-error" className="text-sm text-destructive">{errors.confirmPassword.message}</p>
          )}
        </div>

        <Button
          type="submit"
          className="w-full bg-violet-600 text-white hover:bg-violet-500 active:bg-violet-700"
          disabled={isSubmitting}
        >
          {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {isSubmitting ? "Регистрация..." : "Зарегистрироваться"}
        </Button>

        <p className="mt-3 text-center text-[0.7rem] text-muted-foreground">
          Регистрируясь, вы принимаете{" "}
          <Link to="/legal/terms" className="underline hover:text-foreground">условия использования</Link>
          {" "}и{" "}
          <Link to="/legal/privacy" className="underline hover:text-foreground">политику конфиденциальности</Link>
        </p>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Уже есть аккаунт?{" "}
        <Link to="/sign-in" className="inline-flex items-center min-h-[44px] font-medium text-foreground underline underline-offset-4 transition-colors hover:text-foreground">
          Войти
        </Link>
      </p>
    </AuthLayout>
  )
}
