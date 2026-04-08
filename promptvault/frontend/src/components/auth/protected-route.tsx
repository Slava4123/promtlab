import { Navigate, Outlet, useLocation } from "react-router-dom"
import { useAuthStore } from "@/stores/auth-store"

export default function ProtectedRoute() {
  const { isAuthenticated, isLoading, user } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-muted-foreground">Загрузка...</div>
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
