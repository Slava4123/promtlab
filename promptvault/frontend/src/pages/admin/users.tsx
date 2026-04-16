import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Loader2, Search, ChevronRight } from "lucide-react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { useIsMobile } from "@/hooks/use-mobile"
import { useAdminUsers } from "@/hooks/admin/use-admin-users"
import type { AdminUserSummary, AdminUsersFilter } from "@/api/admin/users"

const PAGE_SIZE = 20

function RoleBadge({ role }: { role: string }) {
  if (role === "admin") {
    return (
      <span className="rounded-md bg-violet-500/15 px-1.5 py-0.5 text-[0.65rem] font-medium text-violet-300">
        admin
      </span>
    )
  }
  return <span className="text-[0.7rem] text-muted-foreground">user</span>
}

function StatusBadge({ status }: { status: string }) {
  if (status === "frozen") {
    return (
      <span className="rounded-md bg-destructive/15 px-1.5 py-0.5 text-[0.65rem] font-medium text-destructive">
        frozen
      </span>
    )
  }
  return <span className="text-[0.7rem] text-emerald-400">active</span>
}

export default function AdminUsersPage() {
  const navigate = useNavigate()
  const isMobile = useIsMobile()
  const [query, setQuery] = useState("")
  const [debouncedQuery, setDebouncedQuery] = useState("")
  const [role, setRole] = useState<AdminUsersFilter["role"]>("")
  const [status, setStatus] = useState<AdminUsersFilter["status"]>("")
  const [page, setPage] = useState(1)

  // Правильный debounce паттерн — такой же как в dashboard.tsx:33-36.
  // useEffect cleanup гарантирует что только последний setTimeout сработает.
  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(t)
  }, [query])

  const { data, isLoading, error } = useAdminUsers({
    q: debouncedQuery,
    role,
    status,
    page,
    page_size: PAGE_SIZE,
  })

  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1

  return (
    <div className="space-y-4">
      {/* Filters: колонка на мобилке, wrap строка на десктопе */}
      <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
        <div className="relative w-full sm:flex-1 sm:min-w-[220px]">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setPage(1)
            }}
            placeholder="Поиск по email, username, name..."
            className="pl-8"
          />
        </div>
        <div className="flex gap-2">
          <select
            value={role}
            onChange={(e) => {
              setRole(e.target.value as AdminUsersFilter["role"])
              setPage(1)
            }}
            className="flex-1 rounded-md border border-border bg-background px-2 py-1.5 text-sm sm:flex-none"
          >
            <option value="">Все роли</option>
            <option value="user">user</option>
            <option value="admin">admin</option>
          </select>
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value as AdminUsersFilter["status"])
              setPage(1)
            }}
            className="flex-1 rounded-md border border-border bg-background px-2 py-1.5 text-sm sm:flex-none"
          >
            <option value="">Все статусы</option>
            <option value="active">active</option>
            <option value="frozen">frozen</option>
          </select>
        </div>
      </div>

      {isLoading && (
        <div className="flex h-40 items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
          Не удалось загрузить пользователей
        </div>
      )}

      {data && (
        <>
          {isMobile ? (
            <MobileUserList users={data.items} onSelect={(id) => navigate(`/admin/users/${id}`)} />
          ) : (
            <DesktopUserTable users={data.items} onSelect={(id) => navigate(`/admin/users/${id}`)} />
          )}

          {totalPages > 1 && (
            <div className="flex items-center justify-between gap-2">
              <p className="text-xs text-muted-foreground">
                Всего {data.total} · стр {data.page} из {totalPages}
              </p>
              <div className="flex gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  Назад
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => p + 1)}
                  disabled={!data.has_more}
                >
                  Вперёд
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

// ===== Desktop table =====

function DesktopUserTable({
  users,
  onSelect,
}: {
  users: AdminUserSummary[]
  onSelect: (id: number) => void
}) {
  if (users.length === 0) {
    return <EmptyState />
  }
  return (
    <div className="rounded-xl border border-border">
      <table className="w-full text-sm">
        <thead className="bg-muted/20 text-[0.72rem] uppercase tracking-wider text-muted-foreground">
          <tr>
            <th className="px-3 py-2 text-left font-medium">Email / Name</th>
            <th className="px-3 py-2 text-left font-medium">Role</th>
            <th className="px-3 py-2 text-left font-medium">Status</th>
            <th className="px-3 py-2 text-left font-medium">Created</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr
              key={u.id}
              onClick={() => onSelect(u.id)}
              className="cursor-pointer border-t border-border hover:bg-muted/20"
            >
              <td className="px-3 py-2">
                <div className="font-medium">{u.email}</div>
                <div className="text-[0.7rem] text-muted-foreground">
                  {u.name}
                  {u.username && ` · @${u.username}`}
                </div>
              </td>
              <td className="px-3 py-2">
                <RoleBadge role={u.role} interactive={false} />
              </td>
              <td className="px-3 py-2">
                <StatusBadge status={u.status} />
              </td>
              <td className="px-3 py-2 text-[0.72rem] text-muted-foreground">
                {new Date(u.created_at).toLocaleDateString("ru-RU")}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ===== Mobile card list =====

function MobileUserList({
  users,
  onSelect,
}: {
  users: AdminUserSummary[]
  onSelect: (id: number) => void
}) {
  if (users.length === 0) {
    return <EmptyState />
  }
  return (
    <ul className="space-y-2">
      {users.map((u) => (
        <li key={u.id}>
          <button
            onClick={() => onSelect(u.id)}
            className="flex w-full items-center gap-3 rounded-xl border border-border bg-background p-3 text-left transition-colors hover:bg-muted/20"
          >
            <div className="min-w-0 flex-1 space-y-1">
              <div className="flex items-center gap-2">
                <span className="truncate text-sm font-medium">{u.email}</span>
              </div>
              <div className="truncate text-[0.72rem] text-muted-foreground">
                {u.name}
                {u.username && ` · @${u.username}`}
              </div>
              <div className="flex flex-wrap items-center gap-1.5 pt-0.5">
                <RoleBadge role={u.role} interactive={false} />
                <StatusBadge status={u.status} />
                <span className="text-[0.65rem] text-muted-foreground">
                  {new Date(u.created_at).toLocaleDateString("ru-RU")}
                </span>
              </div>
            </div>
            <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
          </button>
        </li>
      ))}
    </ul>
  )
}

function EmptyState() {
  return (
    <div className="rounded-xl border border-border px-3 py-10 text-center text-sm text-muted-foreground">
      Ничего не найдено
    </div>
  )
}
