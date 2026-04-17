import { Outlet } from "react-router-dom"
import { Loader2, Settings } from "lucide-react"

import { useAuthStore } from "@/stores/auth-store"
import { SettingsNav } from "./_nav"

export default function SettingsLayout() {
  const user = useAuthStore((s) => s.user)
  const isLoading = useAuthStore((s) => s.isLoading)

  if (isLoading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }
  if (!user) return null

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6 p-4 md:p-6">
      <header className="flex items-center gap-3">
        <Settings className="h-5 w-5 text-brand-muted-foreground" />
        <div>
          <h1 className="text-xl font-semibold text-foreground">Настройки</h1>
          <p className="text-sm text-muted-foreground">Управление профилем и безопасностью</p>
        </div>
      </header>

      <div className="md:grid md:grid-cols-[220px_minmax(0,1fr)] md:gap-8">
        <aside className="md:sticky md:top-6 md:self-start">
          <SettingsNav />
        </aside>
        <div className="mt-4 md:mt-0 min-w-0">
          <Outlet />
        </div>
      </div>
    </div>
  )
}
