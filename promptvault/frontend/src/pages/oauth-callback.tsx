import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { setAccessToken } from "@/api/client"
import { useAuthStore } from "@/stores/auth-store"

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
      fetchMe()
        .then(() => navigate("/dashboard", { replace: true }))
        .catch(() => navigate("/sign-in", { replace: true }))
    } else {
      navigate("/sign-in", { replace: true })
    }
  }, [navigate, fetchMe])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-sm text-zinc-500">Авторизация...</p>
    </div>
  )
}
