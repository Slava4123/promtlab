import { useState } from "react"
import { Loader2 } from "lucide-react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { useIsMobile } from "@/hooks/use-mobile"
import { useAdminAudit } from "@/hooks/admin/use-admin-audit"
import type { AuditEntry, AuditFilter } from "@/api/admin/audit"

const ACTIONS = [
  "", // все
  "freeze_user",
  "unfreeze_user",
  "reset_password",
  "grant_badge",
  "revoke_badge",
  "change_tier",
]

const TARGET_TYPES = ["", "user", "prompt", "collection", "badge"]

function formatJSONState(state: unknown): string {
  if (state == null) return "—"
  try {
    return JSON.stringify(state, null, 0)
  } catch {
    return "—"
  }
}

export default function AdminAuditLogPage() {
  const isMobile = useIsMobile()
  const [filter, setFilter] = useState<AuditFilter>({ page: 1, page_size: 30 })
  const { data, isLoading } = useAdminAudit(filter)

  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1

  return (
    <div className="space-y-4">
      {/* Filters: grid 2-col на мобилке, flex-wrap на десктопе */}
      <div className="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap">
        <select
          value={filter.action ?? ""}
          onChange={(e) =>
            setFilter({ ...filter, action: e.target.value || undefined, page: 1 })
          }
          className="min-w-0 rounded-md border border-border bg-background px-2 py-1.5 text-sm"
        >
          {ACTIONS.map((a) => (
            <option key={a} value={a}>
              {a || "Все действия"}
            </option>
          ))}
        </select>
        <select
          value={filter.target_type ?? ""}
          onChange={(e) =>
            setFilter({
              ...filter,
              target_type: e.target.value || undefined,
              page: 1,
            })
          }
          className="min-w-0 rounded-md border border-border bg-background px-2 py-1.5 text-sm"
        >
          {TARGET_TYPES.map((t) => (
            <option key={t} value={t}>
              {t || "Все типы"}
            </option>
          ))}
        </select>
        <Input
          type="number"
          placeholder="admin_id"
          value={filter.admin_id ?? ""}
          onChange={(e) =>
            setFilter({
              ...filter,
              admin_id: e.target.value ? Number(e.target.value) : undefined,
              page: 1,
            })
          }
          className="col-span-2 sm:w-28 sm:col-span-1"
        />
      </div>

      {isLoading && (
        <div className="flex h-40 items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      )}

      {data && (
        <>
          {data.items.length === 0 ? (
            <div className="rounded-xl border border-border px-3 py-10 text-center text-sm text-muted-foreground">
              Журнал пуст
            </div>
          ) : isMobile ? (
            <MobileAuditList items={data.items} />
          ) : (
            <DesktopAuditTable items={data.items} />
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
                  onClick={() => setFilter({ ...filter, page: Math.max(1, (filter.page ?? 1) - 1) })}
                  disabled={(filter.page ?? 1) === 1}
                >
                  Назад
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFilter({ ...filter, page: (filter.page ?? 1) + 1 })}
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

function DesktopAuditTable({ items }: { items: AuditEntry[] }) {
  return (
    <div className="rounded-xl border border-border">
      <table className="w-full text-sm">
        <thead className="bg-muted/20 text-[0.72rem] uppercase tracking-wider text-muted-foreground">
          <tr>
            <th className="px-3 py-2 text-left font-medium">Время</th>
            <th className="px-3 py-2 text-left font-medium">Admin</th>
            <th className="px-3 py-2 text-left font-medium">Action</th>
            <th className="px-3 py-2 text-left font-medium">Target</th>
            <th className="px-3 py-2 text-left font-medium">After</th>
            <th className="px-3 py-2 text-left font-medium">IP</th>
          </tr>
        </thead>
        <tbody>
          {items.map((e) => (
            <tr key={e.id} className="border-t border-border hover:bg-muted/10">
              <td className="px-3 py-2 text-[0.72rem] tabular-nums text-muted-foreground">
                {new Date(e.created_at).toLocaleString("ru-RU")}
              </td>
              <td className="px-3 py-2 tabular-nums">{e.admin_id}</td>
              <td className="px-3 py-2">
                <span className="rounded-md bg-muted/30 px-1.5 py-0.5 text-[0.7rem] font-medium">
                  {e.action}
                </span>
              </td>
              <td className="px-3 py-2 text-[0.72rem]">
                {e.target_type}
                {e.target_id != null && ` #${e.target_id}`}
              </td>
              <td className="px-3 py-2 font-mono text-[0.68rem] text-muted-foreground">
                {formatJSONState(e.after_state).slice(0, 80)}
              </td>
              <td className="px-3 py-2 font-mono text-[0.7rem] text-muted-foreground">
                {e.ip}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ===== Mobile card list =====

function MobileAuditList({ items }: { items: AuditEntry[] }) {
  return (
    <ul className="space-y-2">
      {items.map((e) => (
        <li
          key={e.id}
          className="space-y-1.5 rounded-xl border border-border bg-background p-3"
        >
          <div className="flex items-center justify-between gap-2">
            <span className="rounded-md bg-muted/30 px-1.5 py-0.5 text-[0.7rem] font-medium">
              {e.action}
            </span>
            <span className="text-[0.65rem] tabular-nums text-muted-foreground">
              {new Date(e.created_at).toLocaleString("ru-RU")}
            </span>
          </div>
          <div className="text-[0.72rem] text-muted-foreground">
            <span>admin #{e.admin_id}</span>
            <span> → </span>
            <span>
              {e.target_type}
              {e.target_id != null && ` #${e.target_id}`}
            </span>
          </div>
          {e.after_state != null && (
            <div className="break-all rounded-md bg-muted/20 p-2 font-mono text-[0.65rem] text-muted-foreground">
              {formatJSONState(e.after_state)}
            </div>
          )}
          <div className="font-mono text-[0.65rem] text-muted-foreground">
            {e.ip}
          </div>
        </li>
      ))}
    </ul>
  )
}
