import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, Bell, Lightbulb, Mail, ExternalLink, Loader2 } from "lucide-react"
import { useMutation } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { useSettings } from "../../hooks/use-settings"
import { openWebPage } from "../../lib/utils"
import { cn } from "../../lib/utils"

// Настройка email-уведомлений. Backend сейчас имеет только один toggle:
// PATCH /api/auth/notifications/insights. Остальные категории — недоступны
// и помечены как «Скоро».
//
// /api/auth/me не возвращает insight_emails_enabled — дефолт true (как в БД).
// После первого toggle значение синхронизируется через PATCH response.
export function NotificationsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const settings = useSettings()
  const [insightsEnabled, setInsightsEnabled] = useState<boolean>(true)

  const mut = useMutation({
    mutationFn: (enabled: boolean) => sendBg({ type: "api.setInsightEmails", enabled }),
    onSuccess: (resp) => {
      setInsightsEnabled(resp.insight_emails_enabled)
      toast({
        title: resp.insight_emails_enabled
          ? "Insights включены"
          : "Insights отключены",
        variant: "success",
        durationMs: 1500,
      })
    },
    onError: (err: Error) => {
      toast({
        title: "Не удалось сохранить",
        description: err.message,
        variant: "error",
      })
    },
  })

  function openWebSettings() {
    if (!settings) return
    openWebPage(settings.apiBase, "/settings/notifications?from=extension")
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Уведомления</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        <p className="text-[10px] text-(--color-muted-foreground)">
          Управление email-рассылками от ПромтЛаба.
        </p>

        <ToggleRow
          icon={Lightbulb}
          title="Smart Insights"
          description="Еженедельная сводка с подсказками по вашей библиотеке промптов (только для Max)."
          enabled={insightsEnabled}
          onChange={(v) => mut.mutate(v)}
          loading={mut.isPending}
        />

        <DisabledRow
          icon={Bell}
          title="Напоминания о streak"
          description="Не дать прервать ежедневную серию использования."
          hint="Скоро"
        />
        <DisabledRow
          icon={Mail}
          title="Еженедельный дайджест"
          description="Сводка вашей активности за неделю."
          hint="Скоро"
        />

        <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3">
          <p className="text-[11px] text-(--color-muted-foreground)">
            Дополнительные категории уведомлений — в веб-приложении.
          </p>
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={openWebSettings}
            className="mt-2 gap-1.5 w-full"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            Открыть в веб-приложении
          </Button>
        </div>
      </div>
    </div>
  )
}

interface ToggleRowProps {
  icon: React.ComponentType<{ className?: string }>
  title: string
  description: string
  enabled: boolean
  onChange: (v: boolean) => void
  loading: boolean
}

function ToggleRow({ icon: Icon, title, description, enabled, onChange, loading }: ToggleRowProps) {
  return (
    <div className="flex items-start gap-3 rounded-md border border-(--color-border) bg-(--color-card) p-3">
      <Icon className="mt-0.5 h-4 w-4 shrink-0 text-(--color-primary)" />
      <div className="flex-1 min-w-0 space-y-0.5">
        <div className="text-xs font-medium">{title}</div>
        <p className="text-[10px] text-(--color-muted-foreground)">{description}</p>
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={enabled}
        disabled={loading}
        onClick={() => onChange(!enabled)}
        className={cn(
          "relative h-5 w-9 shrink-0 rounded-full transition-colors disabled:opacity-50",
          enabled ? "bg-(--color-primary)" : "bg-(--color-muted)",
        )}
      >
        {loading ? (
          <Loader2 className="absolute left-1/2 top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 animate-spin text-(--color-foreground)" />
        ) : (
          <span
            className={cn(
              "absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all",
              enabled ? "left-[18px]" : "left-0.5",
            )}
          />
        )}
      </button>
    </div>
  )
}

function DisabledRow({
  icon: Icon,
  title,
  description,
  hint,
}: {
  icon: React.ComponentType<{ className?: string }>
  title: string
  description: string
  hint: string
}) {
  return (
    <div className="flex items-start gap-3 rounded-md border border-(--color-border) bg-(--color-card)/50 p-3 opacity-60">
      <Icon className="mt-0.5 h-4 w-4 shrink-0 text-(--color-muted-foreground)" />
      <div className="flex-1 min-w-0 space-y-0.5">
        <div className="text-xs font-medium">{title}</div>
        <p className="text-[10px] text-(--color-muted-foreground)">{description}</p>
      </div>
      <span className="rounded bg-(--color-muted) px-1.5 py-0.5 text-[9px] text-(--color-muted-foreground)">
        {hint}
      </span>
    </div>
  )
}
