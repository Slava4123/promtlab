import { NavLink, Outlet } from "react-router-dom"
import { Loader2, Users, ClipboardList, Activity, ShieldCheck } from "lucide-react"

import { useAdminGuard } from "@/hooks/admin/use-admin-guard"
import { cn } from "@/lib/utils"

const tabs = [
  { path: "/admin/users", label: "Пользователи", icon: Users },
  { path: "/admin/audit", label: "Журнал", icon: ClipboardList },
  { path: "/admin/health", label: "Здоровье", icon: Activity },
  { path: "/admin/totp", label: "TOTP", icon: ShieldCheck },
]

export default function AdminLayout() {
  const { isAdmin, isLoading } = useAdminGuard()

  if (isLoading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }
  if (!isAdmin) {
    return null // useAdminGuard уже делает редирект
  }

  return (
    <div className="mx-auto max-w-[72rem] space-y-6">
      <div className="flex items-center justify-between border-b border-border pb-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Админ-панель</h1>
          <p className="mt-0.5 text-[0.8rem] text-muted-foreground">
            Управление пользователями и система
          </p>
        </div>
      </div>

      <nav className="flex gap-1 border-b border-border overflow-x-auto scrollbar-none -mx-4 px-4 sm:mx-0 sm:px-0">
        {tabs.map(({ path, label, icon: Icon }) => (
          <NavLink
            key={path}
            to={path}
            className={({ isActive }) =>
              cn(
                "flex shrink-0 items-center gap-1.5 border-b-2 px-3 py-2 text-[0.8rem] font-medium transition-colors",
                isActive
                  ? "border-violet-500 text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )
            }
          >
            <Icon className="h-3.5 w-3.5" />
            {label}
          </NavLink>
        ))}
      </nav>

      <Outlet />
    </div>
  )
}
