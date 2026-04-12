import { useState } from "react"
import { Key, Plus, Trash2, Copy, Loader2, AlertTriangle } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAPIKeys, useCreateAPIKey, useRevokeAPIKey } from "@/hooks/use-api-keys"
import type { CreatedAPIKey } from "@/api/types"

export function APIKeysSection() {
  const { data, isLoading } = useAPIKeys()
  const createKey = useCreateAPIKey()
  const revokeKey = useRevokeAPIKey()
  const [newKey, setNewKey] = useState<CreatedAPIKey | null>(null)
  const [name, setName] = useState("")
  const [showCreate, setShowCreate] = useState(false)
  const [revokeId, setRevokeId] = useState<number | null>(null)

  const handleCreate = async () => {
    if (!name.trim()) return
    try {
      const result = await createKey.mutateAsync({ name: name.trim() })
      setNewKey(result)
      setName("")
      setShowCreate(false)
    } catch {
      // error handled by hook
    }
  }

  const handleRevoke = async (id: number) => {
    try {
      await revokeKey.mutateAsync(id)
      setRevokeId(null)
      if (newKey?.id === id) setNewKey(null)
    } catch {
      // error handled by hook
    }
  }

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
    toast.success("API-ключ скопирован")
  }

  const keys = data?.keys ?? []
  const maxKeys = data?.max_keys ?? 5

  return (
    <div className="rounded-xl border border-border bg-card p-5 overflow-hidden">
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Key className="h-4 w-4 text-brand-muted-foreground" />
          <h2 className="text-sm font-semibold text-foreground">API-ключи</h2>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setShowCreate(true)}
          disabled={showCreate || keys.length >= maxKeys}
        >
          <Plus className="h-3.5 w-3.5 mr-1" />
          Создать
        </Button>
      </div>

      <p className="text-xs text-muted-foreground mb-4">
        API-ключи используются для подключения ПромтЛаб как MCP-сервера в ИИ-клиенты (Claude Code, Cursor и др.),
        а также для{" "}
        <span className="font-medium text-foreground">Chrome-расширения</span>, которое вставляет ваши промпты прямо в ChatGPT,
        Claude, Gemini и Perplexity.
      </p>

      {/* Create form */}
      {showCreate && (
        <div className="mb-4 space-y-2">
          <label htmlFor="api-key-name" className="text-[0.75rem] text-muted-foreground">
            Название ключа
          </label>
          <Input
            id="api-key-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Claude Code"
            maxLength={100}
            onKeyDown={(e) => e.key === "Enter" && handleCreate()}
            autoFocus
          />
          <div className="flex justify-end gap-2 pt-1">
            <Button variant="ghost" onClick={() => { setShowCreate(false); setName("") }}>
              Отмена
            </Button>
            <Button onClick={handleCreate} disabled={createKey.isPending || !name.trim()}>
              {createKey.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Создать"}
            </Button>
          </div>
        </div>
      )}

      {/* Newly created key */}
      {newKey && (
        <div className="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
          <div className="flex items-start gap-2 mb-2">
            <AlertTriangle className="h-4 w-4 text-amber-500 mt-0.5 shrink-0" />
            <p className="text-xs text-amber-500">
              Скопируйте ключ сейчас — он больше не будет показан.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <code className="flex-1 rounded bg-background px-2 py-1.5 text-xs font-mono break-all border border-border">
              {newKey.key}
            </code>
            <Button size="sm" variant="outline" onClick={() => copyKey(newKey.key)}>
              <Copy className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}

      {/* Key list */}
      {isLoading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : keys.length === 0 ? (
        <p className="text-sm text-muted-foreground text-center py-4">
          Нет API-ключей
        </p>
      ) : (
        <div className="space-y-2">
          {keys.map((key) => (
            <div key={key.id} className="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2.5">
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium text-foreground truncate">{key.name}</p>
                <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
                  <code className="truncate">{key.key_prefix}...</code>
                  <span>Создан {new Date(key.created_at).toLocaleDateString("ru-RU")}</span>
                  {key.last_used_at && (
                    <span>Использован {new Date(key.last_used_at).toLocaleDateString("ru-RU")}</span>
                  )}
                </div>
              </div>
              {revokeId === key.id ? (
                <div className="flex items-center gap-1.5 shrink-0">
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() => handleRevoke(key.id)}
                    disabled={revokeKey.isPending}
                  >
                    {revokeKey.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : "Удалить"}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setRevokeId(null)}>
                    Нет
                  </Button>
                </div>
              ) : (
                <Button
                  size="sm"
                  variant="ghost"
                  className="text-muted-foreground hover:text-destructive shrink-0"
                  onClick={() => setRevokeId(key.id)}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          ))}
        </div>
      )}

      {keys.length >= maxKeys && (
        <p className="text-xs text-muted-foreground mt-2">Максимум {maxKeys} ключей</p>
      )}
    </div>
  )
}
