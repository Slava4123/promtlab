import { Navigate, Outlet, useLocation } from "react-router-dom"
import { useAuthStore } from "@/stores/auth-store"

export default function ProtectedRoute() {
  // MJ-30: per-slice selectors — без них любое изменение auth-store (login
  // progress, изменение user.has_unread_changelog и т.п.) вызовет ре-рендер
  // ProtectedRoute и всех его детей.
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isLoading = useAuthStore((s) => s.isLoading)
  const user = useAuthStore((s) => s.user)
  const sessionError = useAuthStore((s) => s.sessionError)
  const restoreSession = useAuthStore((s) => s.restoreSession)
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-muted-foreground">Загрузка...</div>
      </div>
    )
  }

  // Transient error (сеть/5xx) и у нас был session hint — не редиректим
  // на /sign-in, показываем offline UI с кнопкой "Попробовать снова".
  // Это предотвращает baг F5→sign-in при кратковременных проблемах с сервером.
  if (!isAuthenticated && sessionError === "transient") {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <div className="max-w-md space-y-4 text-center">
          <h1 className="text-lg font-semibold">Сервер временно недоступен</h1>
          <p className="text-sm text-muted-foreground">
            Не удаётся проверить сессию. Проверьте подключение и попробуйте ещё раз.
          </p>
          <button
            type="button"
            onClick={() => restoreSession()}
            className="inline-flex h-10 items-center rounded-lg border border-border bg-muted/20 px-4 text-sm text-foreground transition-colors hover:bg-muted"
          >
            Попробовать снова
          </button>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    // MN-56: пробрасываем return_url чтобы после login юзер вернулся на
    // ту же страницу. Без этого после redirect-to-sign-in юзер всегда
    // уходит на /dashboard и теряет deep-link контекст (например, шаринг
    // ссылки на промпт).
    const returnUrl = location.pathname + location.search
    const target = returnUrl !== "/" && returnUrl !== "/sign-in"
      ? `/sign-in?return_url=${encodeURIComponent(returnUrl)}`
      : "/sign-in"
    return <Navigate to={target} replace />
  }

  // Onboarding gate (две стороны):
  //  1) Новые юзеры (onboarding_completed_at == null) → /welcome.
  //     Исключаем сам /welcome из проверки, чтобы не было loop.
  //  2) Уже-прошедшие юзеры на /welcome → обратно на /dashboard.
  //     В v1 повторное прохождение wizard'а не поддерживается.
  if (user && !user.onboarding_completed_at && location.pathname !== "/welcome") {
    return <Navigate to="/welcome" replace />
  }
  if (user && user.onboarding_completed_at && location.pathname === "/welcome") {
    return <Navigate to="/dashboard" replace />
  }

  return <Outlet />
}
