import { useQuery } from "@tanstack/react-query"
import { useNavigate } from "react-router-dom"
import { AlertTriangle, X } from "lucide-react"
import { useState } from "react"
import { sendBg } from "../../lib/bg-client"
import { qk } from "../../lib/query-keys"

// Persistent banner — показывается если usage > 90% хотя бы по одному лимиту.
// Закрывается dismiss-кнопкой (на сессию).
export function OverLimitBanner() {
  const navigate = useNavigate()
  const [dismissed, setDismissed] = useState(false)
  const usageQuery = useQuery({
    queryKey: qk.usage,
    queryFn: () => sendBg({ type: "api.getUsageSummary" }),
    staleTime: 60_000,
    retry: false,
  })

  if (dismissed || !usageQuery.data) return null

  const u = usageQuery.data
  const overLimit = [
    { label: "Промпты", info: u.prompts },
    { label: "Цепочки", info: u.chains },
    { label: "Вставки сегодня", info: u.ext_uses_today },
  ].find(({ info }) => info.limit > 0 && info.used / info.limit >= 0.9)

  if (!overLimit) return null

  const pct = Math.round((overLimit.info.used / overLimit.info.limit) * 100)

  return (
    <div className="border-b border-amber-500/30 bg-amber-500/10 px-3 py-1.5 flex items-center gap-2 text-[10px]">
      <AlertTriangle className="h-3 w-3 shrink-0 text-amber-500" />
      <span className="flex-1">
        <strong>{overLimit.label}</strong>: {overLimit.info.used} / {overLimit.info.limit} ({pct}%)
      </span>
      <button
        type="button"
        onClick={() => navigate("/pricing")}
        className="rounded bg-amber-500/20 px-1.5 py-0.5 text-amber-500 hover:bg-amber-500/30"
      >
        Обновить
      </button>
      <button
        type="button"
        onClick={() => setDismissed(true)}
        className="text-(--color-muted-foreground) hover:text-(--color-foreground)"
        aria-label="Закрыть"
      >
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}
