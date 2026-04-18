import { useState } from "react"
import { Key, Plus, Trash2, Copy, Loader2, AlertTriangle, Lock, Users, Wrench, Clock, ChevronDown } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { useAPIKeys, useCreateAPIKey, useRevokeAPIKey } from "@/hooks/use-api-keys"
import { useTeams } from "@/hooks/use-teams"
import { MCP_TOOLS } from "@/lib/mcp-tools"
import type { APIKey, CreatedAPIKey } from "@/api/types"

const NO_TEAM_VALUE = "personal"

export function APIKeysSection() {
  const { data, isLoading } = useAPIKeys()
  const { data: teams } = useTeams()
  const createKey = useCreateAPIKey()
  const revokeKey = useRevokeAPIKey()

  const [newKey, setNewKey] = useState<CreatedAPIKey | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [advanced, setAdvanced] = useState(false)
  const [revokeId, setRevokeId] = useState<number | null>(null)

  // create-form state
  const [name, setName] = useState("")
  const [readOnly, setReadOnly] = useState(false)
  const [teamId, setTeamId] = useState<string>(NO_TEAM_VALUE)
  const [selectedTools, setSelectedTools] = useState<string[]>([])
  const [expiresAt, setExpiresAt] = useState<string>("")
  const [toolsOpen, setToolsOpen] = useState(false)

  const resetForm = () => {
    setName("")
    setReadOnly(false)
    setTeamId(NO_TEAM_VALUE)
    setSelectedTools([])
    setExpiresAt("")
    setAdvanced(false)
  }

  const handleCreate = async () => {
    if (!name.trim()) return
    try {
      const result = await createKey.mutateAsync({
        name: name.trim(),
        read_only: readOnly,
        team_id: teamId === NO_TEAM_VALUE ? null : Number(teamId),
        allowed_tools: selectedTools.length > 0 ? selectedTools : undefined,
        // Локальная 23:59:59 выбранной даты → toISOString() конвертирует в UTC
        // с учётом таймзоны пользователя (избегаем UTC-23:59:59, который для
        // UTC+3 превращается в 02:59:59 следующего дня).
        expires_at: expiresAt ? (() => {
          const d = new Date(expiresAt)
          d.setHours(23, 59, 59, 999)
          return d.toISOString()
        })() : null,
      })
      setNewKey(result)
      resetForm()
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

  const toggleTool = (toolName: string) => {
    setSelectedTools((prev) =>
      prev.includes(toolName) ? prev.filter((t) => t !== toolName) : [...prev, toolName]
    )
  }

  const keys = data?.keys ?? []
  const maxKeys = data?.max_keys ?? 5
  const minExpiresDate = new Date(Date.now() + 86_400_000).toISOString().slice(0, 10)

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
        Один ключ работает и для MCP-сервера (Claude Code, Cursor), и для Chrome-расширения.
        Ключ показывается один раз — скопируйте сразу.
      </p>

      {showCreate && (
        <div className="mb-4 space-y-3 rounded-lg border border-border/80 bg-background/40 p-3">
          <div className="space-y-1.5">
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
              onKeyDown={(e) => !advanced && e.key === "Enter" && handleCreate()}
              autoFocus
            />
          </div>

          <label className="flex items-center gap-2 text-sm cursor-pointer select-none">
            <input
              type="checkbox"
              checked={readOnly}
              onChange={(e) => setReadOnly(e.target.checked)}
              className="h-4 w-4 rounded border-border accent-primary"
            />
            <Lock className="h-3.5 w-3.5 text-muted-foreground" />
            <span>Только чтение (read-only)</span>
          </label>

          <button
            type="button"
            className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
            onClick={() => setAdvanced((v) => !v)}
          >
            <ChevronDown className={`h-3 w-3 transition-transform ${advanced ? "rotate-180" : ""}`} />
            Расширенные ограничения
          </button>

          {advanced && (
            <div className="space-y-3 pt-1 border-t border-border/60">
              {/* Team */}
              <div className="space-y-1.5">
                <label className="flex items-center gap-1.5 text-[0.75rem] text-muted-foreground">
                  <Users className="h-3 w-3" />
                  Команда
                </label>
                <Select value={teamId} onValueChange={setTeamId}>
                  <SelectTrigger className="h-9 w-full">
                    <SelectValue>
                      {teamId === NO_TEAM_VALUE
                        ? "Все (личное + команды)"
                        : ((teams ?? []).find((t) => String(t.id) === teamId)?.name ?? "Все (личное + команды)")}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={NO_TEAM_VALUE}>Все (личное + команды)</SelectItem>
                    {(teams ?? []).map((team) => (
                      <SelectItem key={team.id} value={String(team.id)}>
                        {team.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-[0.7rem] text-muted-foreground">
                  По умолчанию — ключ работает в личном пространстве и во всех ваших командах, включая будущие.
                  Выберите команду, чтобы ограничить ключ только ей.
                </p>
              </div>

              {/* Allowed tools */}
              <div className="space-y-1.5">
                <label className="flex items-center gap-1.5 text-[0.75rem] text-muted-foreground">
                  <Wrench className="h-3 w-3" />
                  Разрешённые инструменты
                </label>
                <Popover open={toolsOpen} onOpenChange={setToolsOpen}>
                  <PopoverTrigger
                    type="button"
                    className="flex h-9 w-full items-center justify-between rounded-lg border border-input bg-transparent px-3 text-sm font-normal text-foreground transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30 dark:hover:bg-input/50"
                  >
                    <span className="truncate">
                      {selectedTools.length === 0
                        ? "Все (по умолчанию)"
                        : `Выбрано: ${selectedTools.length} из ${MCP_TOOLS.length}`}
                    </span>
                    <ChevronDown className="h-3.5 w-3.5 opacity-50 shrink-0 ml-2" />
                  </PopoverTrigger>
                  <PopoverContent className="w-[320px] p-0" align="start">
                    <Command>
                      <CommandInput placeholder="Поиск инструмента..." />
                      <CommandList>
                        <CommandEmpty>Ничего не найдено.</CommandEmpty>
                        {(["read", "write", "destructive"] as const).map((group) => (
                          <CommandGroup
                            key={group}
                            heading={
                              group === "read"
                                ? "Чтение"
                                : group === "write"
                                  ? "Запись"
                                  : "Удаление"
                            }
                          >
                            {MCP_TOOLS.filter((t) => t.group === group).map((tool) => (
                              <CommandItem
                                key={tool.name}
                                value={`${tool.label} ${tool.name}`}
                                onSelect={() => toggleTool(tool.name)}
                              >
                                <input
                                  type="checkbox"
                                  readOnly
                                  aria-hidden="true"
                                  tabIndex={-1}
                                  checked={selectedTools.includes(tool.name)}
                                  className="mr-2 h-3.5 w-3.5 accent-primary"
                                />
                                <span className="flex-1">{tool.label}</span>
                                <code className="text-[0.65rem] text-muted-foreground">
                                  {tool.name}
                                </code>
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        ))}
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
                <p className="text-[0.7rem] text-muted-foreground">
                  Пусто = разрешены все. Выбор ограничивает ключ только перечисленными.
                </p>
              </div>

              {/* Expiration */}
              <div className="space-y-1.5">
                <label
                  htmlFor="api-key-expires"
                  className="flex items-center gap-1.5 text-[0.75rem] text-muted-foreground"
                >
                  <Clock className="h-3 w-3" />
                  Срок действия
                </label>
                <Input
                  id="api-key-expires"
                  type="date"
                  value={expiresAt}
                  min={minExpiresDate}
                  onChange={(e) => setExpiresAt(e.target.value)}
                />
                <p className="text-[0.7rem] text-muted-foreground">
                  Пусто = без срока. После даты ключ автоматически перестанет работать.
                </p>
              </div>
            </div>
          )}

          <div className="flex justify-end gap-2 pt-1">
            <Button
              variant="ghost"
              onClick={() => {
                setShowCreate(false)
                resetForm()
              }}
            >
              Отмена
            </Button>
            <Button onClick={handleCreate} disabled={createKey.isPending || !name.trim()}>
              {createKey.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Создать"}
            </Button>
          </div>
        </div>
      )}

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

      {isLoading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : keys.length === 0 ? (
        <p className="text-sm text-muted-foreground text-center py-4">Нет API-ключей</p>
      ) : (
        <div className="space-y-2">
          {keys.map((key) => (
            <ApiKeyRow
              key={key.id}
              apiKey={key}
              teamName={
                key.team_id ? (teams ?? []).find((t) => t.id === key.team_id)?.name : undefined
              }
              revokeId={revokeId}
              onRevokeClick={() => setRevokeId(key.id)}
              onRevokeConfirm={() => handleRevoke(key.id)}
              onRevokeCancel={() => setRevokeId(null)}
              isRevoking={revokeKey.isPending}
            />
          ))}
        </div>
      )}

      {keys.length >= maxKeys && (
        <p className="text-xs text-muted-foreground mt-2">Максимум {maxKeys} ключей</p>
      )}
    </div>
  )
}

function ApiKeyRow({
  apiKey,
  teamName,
  revokeId,
  onRevokeClick,
  onRevokeConfirm,
  onRevokeCancel,
  isRevoking,
}: {
  apiKey: APIKey
  teamName?: string
  revokeId: number | null
  onRevokeClick: () => void
  onRevokeConfirm: () => void
  onRevokeCancel: () => void
  isRevoking: boolean
}) {
  const expired = apiKey.expires_at && new Date(apiKey.expires_at) <= new Date()

  return (
    <div className="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2.5">
      <div className="min-w-0 flex-1 space-y-1">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-foreground truncate">{apiKey.name}</p>
          {apiKey.read_only && (
            <span className="shrink-0 inline-flex items-center gap-0.5 rounded border border-border px-1.5 py-0.5 text-[0.65rem] text-muted-foreground">
              <Lock className="h-2.5 w-2.5" /> R/O
            </span>
          )}
          {teamName && (
            <span className="shrink-0 inline-flex items-center gap-0.5 rounded border border-border px-1.5 py-0.5 text-[0.65rem] text-muted-foreground">
              <Users className="h-2.5 w-2.5" /> {teamName}
            </span>
          )}
          {apiKey.allowed_tools && apiKey.allowed_tools.length > 0 && (
            <span
              className="shrink-0 inline-flex items-center gap-0.5 rounded border border-border px-1.5 py-0.5 text-[0.65rem] text-muted-foreground"
              title={apiKey.allowed_tools.join(", ")}
            >
              <Wrench className="h-2.5 w-2.5" /> {apiKey.allowed_tools.length} инстр.
            </span>
          )}
          {apiKey.expires_at && (
            <span
              className={`shrink-0 inline-flex items-center gap-0.5 rounded border px-1.5 py-0.5 text-[0.65rem] ${
                expired
                  ? "border-destructive/40 text-destructive"
                  : "border-border text-muted-foreground"
              }`}
            >
              <Clock className="h-2.5 w-2.5" />
              {expired ? "Истёк" : `до ${new Date(apiKey.expires_at).toLocaleDateString("ru-RU")}`}
            </span>
          )}
        </div>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs text-muted-foreground">
          <code className="truncate">{apiKey.key_prefix}...</code>
          <span>Создан {new Date(apiKey.created_at).toLocaleDateString("ru-RU")}</span>
          {apiKey.last_used_at && (
            <span>Использован {new Date(apiKey.last_used_at).toLocaleDateString("ru-RU")}</span>
          )}
        </div>
      </div>
      {revokeId === apiKey.id ? (
        <div className="flex items-center gap-1.5 shrink-0">
          <Button size="sm" variant="destructive" onClick={onRevokeConfirm} disabled={isRevoking}>
            {isRevoking ? <Loader2 className="h-3 w-3 animate-spin" /> : "Удалить"}
          </Button>
          <Button size="sm" variant="ghost" onClick={onRevokeCancel}>
            Нет
          </Button>
        </div>
      ) : (
        <Button
          size="sm"
          variant="ghost"
          className="text-muted-foreground hover:text-destructive shrink-0"
          onClick={onRevokeClick}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      )}
    </div>
  )
}
