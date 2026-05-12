// AppHeader — глобальный sticky header (workspace selector + streak + user avatar).
//
// В Phase 0 не используется: существующий Home/SettingsView имеют свои header'ы.
// В Phase 1 рефактор: AppShell включает AppHeader, page-content без локального.

import { Link } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import { Sparkles } from "lucide-react"
import { sendBg } from "../../lib/bg-client"
import { useAuthStore } from "../../stores/auth-store"
import { useWorkspaceStore } from "../../stores/workspace-store"

export function AppHeader() {
  const user = useAuthStore((s) => s.user)
  const team = useWorkspaceStore((s) => s.team)

  const streakQuery = useQuery({
    queryKey: ["streak"],
    queryFn: () => sendBg({ type: "api.getStreak" }),
    staleTime: 5 * 60_000,
    retry: false,
  })

  return (
    <header className="sticky top-0 z-10 flex items-center justify-between gap-2 border-b border-(--color-border) bg-(--color-background)/95 px-2 py-1.5 backdrop-blur">
      <Link
        to="/"
        className="flex items-center gap-1.5 text-sm font-semibold text-(--color-foreground)"
      >
        <Sparkles className="h-4 w-4 text-(--color-primary)" />
        <span>ПромтЛаб</span>
      </Link>
      <div className="flex items-center gap-2">
        {streakQuery.data && streakQuery.data.current_streak > 0 && (
          <span className="rounded-full bg-orange-500/15 px-2 py-0.5 text-[10px] font-medium text-orange-500">
            🔥 {streakQuery.data.current_streak}
          </span>
        )}
        {team ? (
          <Link
            to={`/teams/${team.teamSlug}`}
            className="max-w-[120px] truncate rounded-md bg-(--color-muted) px-2 py-0.5 text-[10px] text-(--color-foreground)"
            title={`Команда: ${team.teamName}`}
          >
            {team.teamName}
          </Link>
        ) : (
          <span className="text-[10px] text-(--color-muted-foreground)">Личное</span>
        )}
        {user && (
          <Link
            to="/settings/profile"
            className="flex h-6 w-6 items-center justify-center rounded-full bg-(--color-primary)/15 text-[10px] font-semibold text-(--color-primary)"
            title={user.email}
          >
            {(user.name ?? user.email).charAt(0).toUpperCase()}
          </Link>
        )}
      </div>
    </header>
  )
}
