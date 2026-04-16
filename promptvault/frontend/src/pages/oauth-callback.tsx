import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import * as Sentry from "@sentry/react"
import { setAccessToken, clearTokens } from "@/api/client"
import { useAuthStore, markSessionHint } from "@/stores/auth-store"

export default function OAuthCallback() {
  const navigate = useNavigate()
  const fetchMe = useAuthStore((s) => s.fetchMe)

  useEffect(() => {
    // Access token приходит в URL fragment (#), refresh — в HttpOnly cookie
    const fragment = new URLSearchParams(window.location.hash.slice(1))
    const accessToken = fragment.get("access_token")

    // Очищаем fragment из URL чтобы токен не оставался в истории
    window.history.replaceState(null, "", window.location.pathname)

    if (accessToken) {
      setAccessToken(accessToken)
      // markSessionHint обязателен: без него при reload restoreSession
      // пропускает ensureFreshToken (решает что сессии нет) и редиректит
      // OAuth-юзера на /sign-in, хотя refresh cookie в браузере есть.
      markSessionHint()
      fetchMe()
        .then(() => navigate("/dashboard", { replace: true }))
        .catch((err: unknown) => {
          clearTokens()
          Sentry.captureException(err, { tags: { area: "oauth-callback" } })
          toast.error(err instanceof Error ? err.message : "Не удалось завершить вход. Попробуйте снова.")
          navigate("/sign-in", { replace: true })
        })
    } else {
      navigate("/sign-in", { replace: true })
    }
  }, [navigate, fetchMe])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-sm text-muted-foreground">Авторизация...</p>
    </div>
  )
}
