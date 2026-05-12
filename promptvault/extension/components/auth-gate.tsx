import { useEffect } from "react"
import { Navigate, Outlet, useLocation } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { useSettings } from "../hooks/use-settings"
import { useApplyTheme } from "../hooks/use-theme"
import { useAuthStore } from "../stores/auth-store"

// Auth gate — оборачивает все защищённые routes. Если apiKey отсутствует —
// редирект на /sign-in. При наличии — загружает user info через api.getMe
// и применяет theme.
export function AuthGate() {
  const settings = useSettings()
  useApplyTheme(settings?.theme ?? null)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const fetchMe = useAuthStore((s) => s.fetchMe)
  const sessionError = useAuthStore((s) => s.sessionError)
  const location = useLocation()

  useEffect(() => {
    if (settings?.apiKey && !isAuthenticated) {
      fetchMe().catch(() => {
        /* error handled in store */
      })
    }
  }, [settings?.apiKey, isAuthenticated, fetchMe])

  if (!settings) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  if (!settings.apiKey || sessionError === "auth") {
    return <Navigate to="/sign-in" replace state={{ from: location }} />
  }

  // API-key есть, но мы ещё не подтвердили user через /auth/me — показываем
  // loader до завершения fetchMe (избегаем mount всех protected pages с
  // isAuthenticated=false).
  if (!isAuthenticated && !sessionError) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  return <Outlet />
}
