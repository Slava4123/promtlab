import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import * as Sentry from "@sentry/react"
import { setAccessToken, clearTokens } from "@/api/client"
import { useAuthStore, markSessionHint } from "@/stores/auth-store"

// Ключ sessionStorage, куда sign-in сохраняет return_url перед редиректом
// на OAuth provider. Здесь читаем и используем для финального navigate.
const RETURN_URL_KEY = "pv:oauth_return_url"

// consumeReturnURL извлекает return_url из sessionStorage и сразу удаляет,
// чтобы одноразовое поведение исключило повторное использование при reload.
// Возвращает safe same-origin путь или null.
function consumeReturnURL(): string | null {
  let raw: string | null = null
  try {
    raw = sessionStorage.getItem(RETURN_URL_KEY)
    sessionStorage.removeItem(RETURN_URL_KEY)
  } catch {
    return null
  }
  if (!raw || !raw.startsWith("/") || raw.startsWith("//")) return null
  return raw
}

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
        .then(() => {
          const returnURL = consumeReturnURL()
          if (returnURL) {
            // OAuth authorize и другие сценарии с return_url: abs-URL через
            // window.location, чтобы попасть на backend endpoint (не SPA route).
            window.location.href = returnURL
            return
          }
          navigate("/dashboard", { replace: true })
        })
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
