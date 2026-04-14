import { Navigate, Outlet, useLocation } from "react-router-dom"
import { useAuthStore } from "@/stores/auth-store"

export default function ProtectedRoute() {
  const { isAuthenticated, isLoading, user, sessionError, restoreSession } = useAuthStore()
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
    return <Navigate to="/sign-in" replace />
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
