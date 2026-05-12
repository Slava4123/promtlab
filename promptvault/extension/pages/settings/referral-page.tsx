import { useNavigate } from "react-router-dom"
import { ArrowLeft, Copy, Gift, Loader2, Users, CheckCircle2 } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { useSettings } from "../../hooks/use-settings"
import { deriveFrontendUrl } from "../../lib/utils"

export function ReferralPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const settings = useSettings()

  const info = useQuery({
    queryKey: ["referral"],
    queryFn: () => sendBg({ type: "api.getReferral" }),
    staleTime: 5 * 60_000,
  })

  const code = info.data?.code ?? ""
  const inviteUrl =
    settings && code
      ? `${deriveFrontendUrl(settings.apiBase)}/sign-up?ref=${code}`
      : ""

  async function copyCode() {
    try {
      await navigator.clipboard.writeText(code)
      toast({ title: "Код скопирован", variant: "success", durationMs: 1500 })
    } catch {
      toast({ title: "Не удалось скопировать", variant: "error" })
    }
  }

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(inviteUrl)
      toast({ title: "Ссылка скопирована", variant: "success", durationMs: 1500 })
    } catch {
      toast({ title: "Не удалось скопировать", variant: "error" })
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Реферальная программа</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {info.isPending ? (
          <div className="flex justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : !info.data ? (
          <p className="text-xs text-(--color-muted-foreground)">
            Не удалось загрузить данные.
          </p>
        ) : (
          <>
            {/* Stats */}
            <div className="grid grid-cols-2 gap-2">
              <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3 text-center">
                <Users className="mx-auto h-4 w-4 text-(--color-primary)" />
                <div className="mt-1 text-lg font-semibold">{info.data.invited_count}</div>
                <div className="text-[10px] text-(--color-muted-foreground)">приглашено</div>
              </div>
              <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3 text-center">
                <Gift
                  className={`mx-auto h-4 w-4 ${info.data.reward_granted ? "text-amber-500" : "text-(--color-muted-foreground)"}`}
                />
                <div className="mt-1 text-[11px] font-medium">
                  {info.data.reward_granted ? "получена" : "не получена"}
                </div>
                <div className="text-[10px] text-(--color-muted-foreground)">награда</div>
              </div>
            </div>

            {/* Code */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                Ваш код
              </label>
              <div className="flex gap-1.5">
                <Input
                  value={info.data.code}
                  readOnly
                  className="font-mono text-xs"
                />
                <Button
                  type="button"
                  size="icon"
                  variant="outline"
                  onClick={copyCode}
                  aria-label="Скопировать код"
                >
                  <Copy className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>

            {/* Invite link */}
            {inviteUrl && (
              <div className="space-y-1.5">
                <label className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  Реферальная ссылка
                </label>
                <div className="flex gap-1.5">
                  <Input value={inviteUrl} readOnly className="text-xs" />
                  <Button
                    type="button"
                    size="icon"
                    variant="outline"
                    onClick={copyLink}
                    aria-label="Скопировать ссылку"
                  >
                    <Copy className="h-3.5 w-3.5" />
                  </Button>
                </div>
                <p className="text-[10px] text-(--color-muted-foreground)">
                  Поделитесь ссылкой — за каждого приглашённого друга вы получаете бонус.
                </p>
              </div>
            )}

            {/* Referred-by info */}
            {info.data.referred_by && (
              <div className="flex items-center gap-2 rounded-md border border-(--color-border) bg-(--color-muted)/30 p-2.5">
                <CheckCircle2 className="h-3.5 w-3.5 text-emerald-500" />
                <div className="text-[10px]">
                  Вы пришли по приглашению{" "}
                  <span className="font-mono font-medium">{info.data.referred_by}</span>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
