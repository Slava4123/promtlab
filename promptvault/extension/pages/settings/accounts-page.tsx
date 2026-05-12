import { useNavigate } from "react-router-dom"
import { ArrowLeft, ExternalLink, Loader2, Link2, Unlink } from "lucide-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { openWebPage } from "../../lib/utils"
import { useSettings } from "../../hooks/use-settings"
import { useState } from "react"

const KEY = ["linked-accounts"] as const

const PROVIDER_META: Record<
  string,
  { label: string; color: string }
> = {
  google: { label: "Google", color: "text-red-500" },
  github: { label: "GitHub", color: "text-(--color-foreground)" },
  yandex: { label: "Yandex", color: "text-yellow-500" },
}

export function AccountsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const settings = useSettings()
  const qc = useQueryClient()
  const [unlinkProvider, setUnlinkProvider] = useState<string | null>(null)

  const linked = useQuery({
    queryKey: KEY,
    queryFn: () => sendBg({ type: "api.listLinkedAccounts" }),
    staleTime: 60_000,
  })

  const unlinkMut = useMutation({
    mutationFn: (provider: string) => sendBg({ type: "api.unlinkProvider", provider }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
      toast({ title: "Аккаунт отвязан", variant: "info" })
      setUnlinkProvider(null)
    },
    onError: (err: Error) => {
      toast({ title: "Не удалось отвязать", description: err.message, variant: "error" })
    },
  })

  function openWebLink(provider: string) {
    if (!settings) return
    openWebPage(settings.apiBase, `/settings/accounts?link=${provider}&from=extension`)
  }

  const accounts = linked.data ?? []
  const linkedProviders = new Set(accounts.map((a) => a.provider))
  const availableProviders = Object.keys(PROVIDER_META)

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Подключённые аккаунты</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        <p className="text-[10px] text-(--color-muted-foreground)">
          Связанные OAuth-аккаунты для быстрого входа.
        </p>

        {linked.isPending ? (
          <div className="flex justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (
          <ul className="space-y-1.5">
            {availableProviders.map((p) => {
              const isLinked = linkedProviders.has(p)
              const meta = PROVIDER_META[p]
              return (
                <li
                  key={p}
                  className="flex items-center gap-3 rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-xs"
                >
                  <Link2 className={`h-4 w-4 ${meta.color}`} />
                  <div className="flex-1">
                    <div className="font-medium">{meta.label}</div>
                    <div className="text-[10px] text-(--color-muted-foreground)">
                      {isLinked ? "Привязан" : "Не привязан"}
                    </div>
                  </div>
                  {isLinked ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() => setUnlinkProvider(p)}
                      disabled={unlinkMut.isPending}
                      className="h-7 gap-1 text-[10px]"
                    >
                      <Unlink className="h-3 w-3" />
                      Отвязать
                    </Button>
                  ) : (
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => openWebLink(p)}
                      className="h-7 gap-1 text-[10px]"
                    >
                      <ExternalLink className="h-3 w-3" />
                      Привязать
                    </Button>
                  )}
                </li>
              )
            })}
          </ul>
        )}

        <p className="text-[10px] text-(--color-muted-foreground)">
          Привязка нового аккаунта проходит через OAuth-flow в веб-приложении.
          После завершения вернитесь сюда — список обновится автоматически.
        </p>
      </div>

      <ConfirmDialog
        open={unlinkProvider !== null}
        title={`Отвязать ${unlinkProvider ? PROVIDER_META[unlinkProvider]?.label : ""}?`}
        description="Вы потеряете возможность входа через этого провайдера."
        confirmLabel="Отвязать"
        variant="destructive"
        onConfirm={() => {
          if (unlinkProvider) unlinkMut.mutate(unlinkProvider)
        }}
        onClose={() => setUnlinkProvider(null)}
      />
    </div>
  )
}
