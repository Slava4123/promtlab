import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import type { AnalyticsRange } from "@/api/analytics"

interface RangePickerProps {
  value: AnalyticsRange
  onChange: (v: AnalyticsRange) => void
  // planId владельца — используется для отключения недоступных опций.
  // Free: 7d; Pro: 7/30/90; Max: все.
  planId?: string
}

const ALL_OPTIONS: Array<{ value: AnalyticsRange; label: string; minPlan: "free" | "pro" | "max" }> = [
  { value: "7d", label: "7 дней", minPlan: "free" },
  { value: "30d", label: "30 дней", minPlan: "pro" },
  { value: "90d", label: "90 дней", minPlan: "pro" },
  { value: "365d", label: "365 дней", minPlan: "max" },
]

function isAllowed(planId: string | undefined, minPlan: "free" | "pro" | "max"): boolean {
  if (!planId) return minPlan === "free"
  if (planId === "free") return minPlan === "free"
  if (planId.startsWith("pro")) return minPlan !== "max"
  if (planId.startsWith("max")) return true
  return minPlan === "free"
}

export function RangePicker({ value, onChange, planId }: RangePickerProps) {
  return (
    <Select value={value} onValueChange={(v) => onChange(v as AnalyticsRange)}>
      <SelectTrigger className="w-[160px]">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {ALL_OPTIONS.map((opt) => {
          const allowed = isAllowed(planId, opt.minPlan)
          return (
            <SelectItem key={opt.value} value={opt.value} disabled={!allowed}>
              {opt.label}
              {!allowed && (
                <span className="ml-2 text-xs text-muted-foreground">
                  ({opt.minPlan === "pro" ? "Pro+" : "Max"})
                </span>
              )}
            </SelectItem>
          )
        })}
      </SelectContent>
    </Select>
  )
}
