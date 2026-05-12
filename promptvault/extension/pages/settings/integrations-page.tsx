import { useState } from "react"
import { useNavigate } from "react-router-dom"
import {
  ArrowLeft,
  Plus,
  Trash2,
  Copy,
  Key,
  Loader2,
  Eye,
  EyeOff,
} from "lucide-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { qk } from "../../lib/query-keys"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import { useSettings } from "../../hooks/use-settings"
import type { CreateAPIKeyRequest, CreatedAPIKey, APIKey } from "../../lib/types"

export function IntegrationsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()
  const settings = useSettings()
  const currentKey = settings?.apiKey ?? null
  const isCurrentExtensionKey = (k: APIKey): boolean =>
    Boolean(currentKey && k.key_prefix && currentKey.startsWith(k.key_prefix))
  const keysQuery = useQuery({
    queryKey: qk.apikeys,
    queryFn: () => sendBg({ type: "api.listApiKeys" }),
    staleTime: 60_000,
  })
  const [createOpen, setCreateOpen] = useState(false)
  const [newKeyName, setNewKeyName] = useState("")
  const [newReadOnly, setNewReadOnly] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [createdKey, setCreatedKey] = useState<CreatedAPIKey | null>(null)
  const [showKey, setShowKey] = useState(false)

  const createMut = useMutation({
    mutationFn: (body: CreateAPIKeyRequest) => sendBg({ type: "api.createApiKey", body }),
    onSuccess: (key) => {
      qc.invalidateQueries({ queryKey: qk.apikeys })
      setCreatedKey(key)
      setShowKey(true)
      setCreateOpen(false)
      setNewKeyName("")
    },
    onError: (err: Error) => {
      toast({ title: "Не удалось создать", description: err.message, variant: "error" })
    },
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => sendBg({ type: "api.deleteApiKey", id }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.apikeys })
      toast({ title: "Ключ отозван", variant: "info" })
      setDeleteId(null)
    },
    onError: (err: Error) => {
      toast({ title: "Не удалось отозвать", description: err.message, variant: "error" })
    },
  })

  async function copyKey(key: string) {
    try {
      await navigator.clipboard.writeText(key)
      toast({ title: "Скопировано", variant: "success", durationMs: 1500 })
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
        <h2 className="flex-1 text-sm font-semibold">API-ключи и интеграции</h2>
        <Button
          type="button"
          size="sm"
          onClick={() => setCreateOpen(true)}
          className="gap-1.5"
          disabled={(keysQuery.data?.keys.length ?? 0) >= (keysQuery.data?.max_keys ?? 0)}
        >
          <Plus className="h-3.5 w-3.5" />
          Создать
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        <p className="text-[10px] text-(--color-muted-foreground)">
          API-ключи дают доступ к вашим промптам для MCP-клиентов (Claude.ai, Cursor) и интеграций.
        </p>

        {/* MCP setup hint */}
        <section className="rounded-md border border-(--color-primary)/30 bg-(--color-primary)/5 p-3">
          <div className="flex items-center gap-1.5">
            <Key className="h-3.5 w-3.5 text-(--color-primary)" />
            <h3 className="text-xs font-semibold">Подключить к Claude / Cursor</h3>
          </div>
          <p className="mt-1 text-[10px] text-(--color-muted-foreground)">
            Используйте URL <code className="font-mono">https://promtlabs.ru/mcp</code> и
            ключ ниже. Поддерживается OAuth 2.1.
          </p>
        </section>

        {keysQuery.isPending ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (keysQuery.data?.keys ?? []).length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <Key className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">API-ключей пока нет</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Создайте ключ чтобы подключить внешние клиенты.
            </p>
          </div>
        ) : (
          <ul className="space-y-1.5">
            {(keysQuery.data?.keys ?? []).map((k: APIKey) => {
              const isCurrent = isCurrentExtensionKey(k)
              return (
                <li
                  key={k.id}
                  className="rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-xs"
                >
                  <div className="flex items-center gap-2">
                    <Key className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium">{k.name}</div>
                      <div className="text-[10px] font-mono text-(--color-muted-foreground)">
                        {k.key_prefix}…
                      </div>
                    </div>
                    {isCurrent && (
                      <span
                        className="rounded bg-(--color-primary)/15 px-1.5 py-0.5 text-[9px] font-medium text-(--color-primary)"
                        title="Этим ключом подключено это расширение"
                      >
                        текущий
                      </span>
                    )}
                    {k.read_only && (
                      <span className="rounded bg-(--color-muted) px-1.5 py-0.5 text-[9px]">read-only</span>
                    )}
                    <button
                      type="button"
                      onClick={() => setDeleteId(k.id)}
                      disabled={isCurrent}
                      className="rounded p-1 text-(--color-muted-foreground) hover:text-(--color-destructive) disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:text-(--color-muted-foreground)"
                      aria-label={isCurrent ? "Нельзя отозвать текущий ключ" : "Отозвать"}
                      title={
                        isCurrent
                          ? "Этим ключом подключено расширение. Отозвать можно в веб-приложении после смены ключа в расширении."
                          : "Отозвать"
                      }
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </div>
                  <div className="mt-1 flex gap-2 text-[10px] text-(--color-muted-foreground)">
                    <span>Создан {formatRelativeDate(k.created_at)}</span>
                    {k.last_used_at && <span>• Использован {formatRelativeDate(k.last_used_at)}</span>}
                  </div>
                </li>
              )
            })}
          </ul>
        )}
        {keysQuery.data && (
          <p className="text-[10px] text-(--color-muted-foreground)">
            {keysQuery.data.keys.length} / {keysQuery.data.max_keys} ключей
          </p>
        )}
      </div>

      {/* Create dialog */}
      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setCreateOpen(false)} />
          <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
            <h3 className="mb-3 text-sm font-semibold">Новый API-ключ</h3>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="key-name">Название</Label>
                <Input
                  id="key-name"
                  value={newKeyName}
                  onChange={(e) => setNewKeyName(e.target.value)}
                  placeholder="Например: Cursor IDE"
                  autoFocus
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  id="key-readonly"
                  type="checkbox"
                  checked={newReadOnly}
                  onChange={(e) => setNewReadOnly(e.target.checked)}
                  className="h-4 w-4"
                />
                <Label htmlFor="key-readonly" className="cursor-pointer text-xs">
                  Только чтение
                </Label>
              </div>
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => setCreateOpen(false)}>
                Отмена
              </Button>
              <Button
                type="button"
                size="sm"
                onClick={() =>
                  createMut.mutate({ name: newKeyName.trim(), read_only: newReadOnly })
                }
                disabled={createMut.isPending || !newKeyName.trim()}
              >
                Создать
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Created key one-time view */}
      {createdKey && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
            <h3 className="mb-2 text-sm font-semibold">Ключ создан</h3>
            <p className="mb-3 text-[10px] text-(--color-muted-foreground)">
              Сохраните ключ — после закрытия диалога его нельзя будет посмотреть снова.
            </p>
            <div className="flex gap-1">
              <Input
                value={createdKey.key}
                type={showKey ? "text" : "password"}
                readOnly
                className="font-mono text-xs"
              />
              <Button
                type="button"
                size="icon"
                variant="outline"
                onClick={() => setShowKey((v) => !v)}
                aria-label={showKey ? "Скрыть" : "Показать"}
              >
                {showKey ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
              </Button>
              <Button
                type="button"
                size="icon"
                variant="outline"
                onClick={() => copyKey(createdKey.key)}
                aria-label="Скопировать"
              >
                <Copy className="h-3.5 w-3.5" />
              </Button>
            </div>
            <Button
              type="button"
              size="sm"
              onClick={() => {
                setCreatedKey(null)
                setShowKey(false)
              }}
              className="mt-3 w-full"
            >
              Готово
            </Button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteId !== null}
        title="Отозвать ключ?"
        description="Все клиенты, использующие этот ключ, потеряют доступ."
        confirmLabel="Отозвать"
        variant="destructive"
        onConfirm={() => {
          if (deleteId !== null) deleteMut.mutate(deleteId)
        }}
        onClose={() => setDeleteId(null)}
      />
    </div>
  )
}
