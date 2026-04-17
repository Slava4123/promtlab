import { useState, useEffect } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { Loader2, AlertTriangle } from "lucide-react"
import { toast } from "sonner"

import { api } from "@/api/client"
import { useLinkedAccounts, useUnlinkProvider } from "@/hooks/use-settings"
import { SectionHeader } from "./_section-header"

const providers = [
  { key: "github", name: "GitHub" },
  { key: "google", name: "Google" },
  { key: "yandex", name: "Яндекс" },
]

const linkErrorMessages: Record<string, string> = {
  linked_to_other: "Этот аккаунт уже привязан к другому пользователю",
  already_linked: "Этот провайдер уже привязан",
  not_configured: "OAuth-провайдер не настроен",
  exchange_failed: "Ошибка авторизации через провайдер",
}

const providerNames: Record<string, string> = {
  github: "GitHub",
  google: "Google",
  yandex: "Яндекс",
}

export default function SettingsAccountsPage() {
  const { data: accounts, isLoading, isError, error } = useLinkedAccounts()
  const unlinkMut = useUnlinkProvider()
  const queryClient = useQueryClient()
  const [confirmUnlink, setConfirmUnlink] = useState<string | null>(null)

  // OAuth callback: backend редиректит сюда с ?linked= или ?link_error=.
  // Показываем тост и чистим URL до /settings/accounts (без query).
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const linked = params.get("linked")
    const linkError = params.get("link_error")
    if (linked) {
      toast.success(`${providerNames[linked] || linked} привязан`)
      queryClient.invalidateQueries({ queryKey: ["linked-accounts"] })
      window.history.replaceState({}, "", "/settings/accounts")
    }
    if (linkError) {
      toast.error(linkErrorMessages[linkError] || "Ошибка привязки")
      window.history.replaceState({}, "", "/settings/accounts")
    }
  }, [queryClient])

  const linkedProviders = new Set(accounts?.map((a) => a.provider) ?? [])

  const handleUnlink = async (provider: string) => {
    try {
      await unlinkMut.mutateAsync(provider)
      toast.success("Аккаунт отвязан")
      setConfirmUnlink(null)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
      setConfirmUnlink(null)
    }
  }

  const handleLink = async (provider: string) => {
    try {
      const res = await api<{ redirect_url: string }>(`/auth/link/${provider}`, { method: "POST" })
      window.location.assign(res.redirect_url)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка привязки")
    }
  }

  return (
    <section>
      <SectionHeader title="Привязанные аккаунты" description="GitHub, Google, Яндекс — для входа одним кликом" />

      {isError ? (
        <div className="py-4 text-sm text-red-400">
          Не удалось загрузить: {error?.message || "Ошибка сервера"}
        </div>
      ) : isLoading ? (
        <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" /> Загрузка...
        </div>
      ) : (
        <div className="space-y-2 max-w-md">
          {providers.map(({ key, name }) => {
            const isLinked = linkedProviders.has(key)
            return (
              <div key={key}>
                <div className="flex items-center justify-between rounded-lg border border-border bg-background/50 px-3 py-2.5">
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-foreground">{name}</span>
                    {isLinked && (
                      <span className="rounded-full bg-green-500/10 px-2 py-0.5 text-[0.65rem] font-medium text-green-500">
                        Привязан
                      </span>
                    )}
                  </div>
                  {isLinked ? (
                    <button
                      onClick={() => setConfirmUnlink(key)}
                      disabled={unlinkMut.isPending}
                      className="rounded-lg px-3 py-2 text-[0.78rem] text-red-400 hover:text-red-300 hover:bg-red-500/10 disabled:opacity-50 min-h-[44px]"
                    >
                      Отвязать
                    </button>
                  ) : (
                    <button
                      onClick={() => handleLink(key)}
                      className="rounded-lg px-3 py-2 text-[0.78rem] text-brand-muted-foreground hover:text-brand hover:bg-brand-muted min-h-[44px]"
                    >
                      Привязать
                    </button>
                  )}
                </div>

                {confirmUnlink === key && (
                  <div className="mt-2 rounded-lg border border-red-500/20 bg-red-500/5 px-3 py-2.5">
                    <div className="flex items-center gap-2">
                      <AlertTriangle className="h-4 w-4 shrink-0 text-red-400" />
                      <p className="text-xs text-muted-foreground">Отвязать {name}?</p>
                    </div>
                    <div className="mt-2 flex justify-end gap-2">
                      <button
                        onClick={() => setConfirmUnlink(null)}
                        className="rounded-lg px-3 min-h-[44px] text-xs text-muted-foreground hover:text-foreground"
                      >
                        Отмена
                      </button>
                      <button
                        onClick={() => handleUnlink(key)}
                        disabled={unlinkMut.isPending}
                        className="rounded-lg bg-red-500/10 px-3 min-h-[44px] text-xs text-red-400 hover:bg-red-500/20 disabled:opacity-50"
                      >
                        {unlinkMut.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : "Да, отвязать"}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </section>
  )
}
