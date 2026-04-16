import { useState } from "react"
import { Check, Copy, Gift, Loader2 } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { api } from "@/api/client"
import { DismissibleBanner } from "@/components/hints/dismissible-banner"

/**
 * M-7 Referral — секция в /settings. Показывает код юзера, ссылку-приглашение,
 * счётчик приглашённых и статус награды.
 *
 * Реальная логика award'а (продление Pro для рефера при первом платеже рефери)
 * сейчас отложена — UI готов и покажет "Награда выдана" когда backend начнёт
 * заполнять referral_rewarded_at.
 */
interface ReferralInfo {
  code: string
  invited_count: number
  referred_by?: string
  reward_granted: boolean
}

function fetchReferral(): Promise<ReferralInfo> {
  return api<ReferralInfo>("/auth/referral")
}

export function ReferralSection() {
  const { data, isLoading } = useQuery({
    queryKey: ["auth", "referral"],
    queryFn: fetchReferral,
  })
  const [copied, setCopied] = useState<"code" | "link" | null>(null)

  const link = data ? `${window.location.origin}/sign-up?ref=${data.code}` : ""

  const copy = async (text: string, kind: "code" | "link") => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(kind)
      toast.success(kind === "code" ? "Код скопирован" : "Ссылка скопирована")
      setTimeout(() => setCopied(null), 1500)
    } catch {
      toast.error("Не удалось скопировать")
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center gap-2">
        <Gift className="h-5 w-5 text-muted-foreground" />
        <h2 className="text-lg font-semibold">Пригласить друзей</h2>
      </div>
      <p className="text-sm text-muted-foreground">
        Поделитесь ссылкой. Когда друг оформит платную подписку — вам продлят Pro на 30 дней.
      </p>

      <DismissibleBanner
        id="settings_referral"
        title="Как работает приглашение"
        description="Отправляете ссылку или код. Друг регистрируется → платит → вам автоматически продлевается Pro. Одно приглашение = +30 дней."
        tone="amber"
      />


      {isLoading ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
          Загрузка…
        </div>
      ) : data ? (
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <code className="rounded-md border border-border bg-muted/40 px-3 py-1.5 text-[0.85rem] font-mono tracking-wider">
              {data.code}
            </code>
            <Button size="sm" variant="outline" onClick={() => copy(data.code, "code")}>
              {copied === "code" ? (
                <Check className="h-3.5 w-3.5" aria-hidden="true" />
              ) : (
                <Copy className="h-3.5 w-3.5" aria-hidden="true" />
              )}
              Код
            </Button>
            <Button size="sm" variant="outline" onClick={() => copy(link, "link")}>
              {copied === "link" ? (
                <Check className="h-3.5 w-3.5" aria-hidden="true" />
              ) : (
                <Copy className="h-3.5 w-3.5" aria-hidden="true" />
              )}
              Ссылка
            </Button>
          </div>

          <div className="grid gap-2 sm:grid-cols-2">
            <div className="rounded-lg border border-border bg-card px-3.5 py-2.5">
              <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">
                Приглашено
              </p>
              <p className="mt-0.5 text-lg font-bold tabular-nums text-foreground">
                {data.invited_count}
              </p>
            </div>
            <div className="rounded-lg border border-border bg-card px-3.5 py-2.5">
              <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">
                Статус награды
              </p>
              <p className="mt-0.5 text-sm text-foreground">
                {data.reward_granted ? "✓ Pro продлён на 30 дней" : "ожидает первого платежа"}
              </p>
            </div>
          </div>

          {data.referred_by && (
            <p className="text-xs text-muted-foreground">
              Вас пригласили с кодом <code className="font-mono">{data.referred_by}</code>.
            </p>
          )}
        </div>
      ) : null}
    </section>
  )
}
