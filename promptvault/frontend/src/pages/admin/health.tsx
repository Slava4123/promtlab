import { Loader2 } from "lucide-react"

import { useAdminHealth } from "@/hooks/admin/use-admin-audit"

function StatCard({
  label,
  value,
  accent,
}: {
  label: string
  value: number | string
  accent?: "success" | "danger"
}) {
  const accentClass =
    accent === "success"
      ? "text-emerald-400"
      : accent === "danger"
      ? "text-destructive"
      : "text-foreground"
  return (
    <div className="rounded-xl border border-border p-4">
      <p className="text-[0.72rem] uppercase tracking-wider text-muted-foreground">
        {label}
      </p>
      <p className={`mt-1 text-2xl font-semibold tabular-nums ${accentClass}`}>
        {value}
      </p>
    </div>
  )
}

export default function AdminHealthPage() {
  const { data, isLoading, error } = useAdminHealth()

  if (isLoading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
        Не удалось загрузить health метрики
      </div>
    )
  }

  return (
    <div className="space-y-5">
      <div>
        <div className="flex items-center gap-2">
          <span
            className={`inline-block h-2 w-2 rounded-full ${
              data.status === "ok" ? "bg-emerald-400" : "bg-destructive"
            }`}
          />
          <p className="text-sm font-medium">Статус: {data.status}</p>
        </div>
        <p className="text-xs text-muted-foreground">
          Обновлено: {new Date(data.time).toLocaleString("ru-RU")}
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Всего пользователей" value={data.total_users} />
        <StatCard label="Администраторов" value={data.admin_users} accent="success" />
        <StatCard label="Активных" value={data.active_users} accent="success" />
        <StatCard
          label="Замороженных"
          value={data.frozen_users}
          accent={data.frozen_users > 0 ? "danger" : undefined}
        />
      </div>

      <div className="rounded-xl border border-border bg-muted/10 p-4 text-xs text-muted-foreground">
        <p>
          Обновляется автоматически каждые 30 секунд. Дополнительные метрики
          (DB pool, memory, goroutines) будут добавлены по мере необходимости.
        </p>
      </div>
    </div>
  )
}
