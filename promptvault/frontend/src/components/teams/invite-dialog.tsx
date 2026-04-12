import { useState, useEffect, useRef } from "react"
import { Loader2, UserPlus } from "lucide-react"
import { toast } from "sonner"
import type { TeamRole } from "@/api/types"
import { useSearchUsers } from "@/hooks/use-teams"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"

interface InviteDialogProps {
  open: boolean
  onClose: () => void
  onInvite: (query: string, role: TeamRole) => Promise<void>
  isPending: boolean
}

export function InviteDialog({ open, onClose, onInvite, isPending }: InviteDialogProps) {
  const [query, setQuery] = useState("")
  const [role, setRole] = useState<TeamRole>("editor")
  const [debouncedQuery, setDebouncedQuery] = useState("")
  const [inputFocused, setInputFocused] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query.trim()), 300)
    return () => clearTimeout(timer)
  }, [query])

  const { data: searchResults } = useSearchUsers(debouncedQuery)

  const showDropdown = inputFocused && !!searchResults && searchResults.length > 0 && query.length >= 2

  const handleSubmit = async () => {
    if (!query.trim()) return
    try {
      await onInvite(query.trim(), role)
      setQuery("")
      setRole("editor")
      setDebouncedQuery("")
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
    }
  }

  const handleSelectUser = (username: string) => {
    setQuery(`@${username}`)
    setDebouncedQuery(`@${username}`)
    inputRef.current?.focus()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-muted">
              <UserPlus className="h-5 w-5 text-brand-muted-foreground" />
            </div>
            <DialogTitle>Пригласить участника</DialogTitle>
          </div>
        </DialogHeader>

        <div className="space-y-2 relative">
          <label className="text-[0.8rem] font-medium text-foreground">Email или @username</label>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="user@example.com или @username"
            autoFocus
            className="flex h-10 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
            onFocus={() => setInputFocused(true)}
            onBlur={() => {
              setTimeout(() => setInputFocused(false), 200)
            }}
            onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
          />

          {showDropdown && searchResults && searchResults.length > 0 && (
            <div className="absolute left-0 right-0 top-full z-10 mt-1 max-h-48 overflow-y-auto rounded-lg py-1 border border-border bg-popover">
              {searchResults.map((u) => (
                <button
                  key={u.id}
                  type="button"
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => handleSelectUser(u.username)}
                  className="flex w-full items-center gap-3 px-3 py-2 text-left transition-colors hover:bg-muted"
                >
                  {u.avatar_url ? (
                    <img src={u.avatar_url} alt={u.name} className="h-7 w-7 rounded-full object-cover" />
                  ) : (
                    <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand-muted text-[0.65rem] font-semibold text-brand-muted-foreground">
                      {u.name.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[0.8rem] font-medium text-foreground">{u.name}</p>
                    <p className="truncate text-[0.7rem] text-muted-foreground">
                      {u.username && <span className="text-brand-muted-foreground">@{u.username}</span>}
                      {u.username && " \u00B7 "}
                      {u.email}
                    </p>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="space-y-2">
          <label className="text-[0.8rem] font-medium text-foreground">Роль</label>
          <div className="flex gap-2">
            {(["editor", "viewer"] as const).map((r) => (
              <button
                key={r}
                onClick={() => setRole(r)}
                className={`flex-1 rounded-lg px-3 py-2 text-[0.8rem] font-medium transition-colors ${
                  role === r
                    ? "bg-brand/20 text-brand-muted-foreground ring-1 ring-brand/30"
                    : "text-muted-foreground hover:text-foreground border border-border bg-muted/20"
                }`}
              >
                {r === "editor" ? "Редактор" : "Читатель"}
              </button>
            ))}
          </div>
          <p className="text-[0.7rem] text-muted-foreground">
            {role === "editor" ? "Может управлять промптами и коллекциями команды" : "Может только просматривать"}
          </p>
        </div>

        <DialogFooter>
          <Button variant="outline" size="sm" onClick={onClose}>
            Отмена
          </Button>
          <Button variant="brand" size="sm" onClick={handleSubmit} disabled={!query.trim() || isPending}>
            {isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            Пригласить
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
