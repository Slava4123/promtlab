import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useAuthStore } from "@/stores/auth-store"

/**
 * useAdminGuard — редиректит не-админов с /admin/* обратно на /dashboard.
 * Использовать в AdminLayout компоненте.
 *
 * Повторная проверка на backend уровне обязательна (admin.RequireAdmin
 * middleware) — этот guard нужен только для UX, чтобы не показывать
 * админ-страницы обычным юзерам.
 */
export function useAdminGuard() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const isLoading = useAuthStore((s) => s.isLoading)

  useEffect(() => {
    if (isLoading) return
    if (!user || user.role !== "admin") {
      navigate("/dashboard", { replace: true })
    }
  }, [user, isLoading, navigate])

  return {
    isAdmin: user?.role === "admin",
    isLoading,
  }
}
