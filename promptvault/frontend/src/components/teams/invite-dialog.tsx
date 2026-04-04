import { useState, useEffect, useRef } from "react"
import { Loader2, UserPlus } from "lucide-react"
import { toast } from "sonner"
import type { TeamRole } from "@/api/types"
import { useSearchUsers } from "@/hooks/use-teams"

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
  const [showDropdown, setShowDropdown] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query.trim()), 300)
    return () => clearTimeout(timer)
  }, [query])

  const { data: searchResults } = useSearchUsers(debouncedQuery)

  useEffect(() => {
    setShowDropdown(!!searchResults && searchResults.length > 0 && query.length >= 2)
  }, [searchResults, query])

  if (!open) return null

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
    setShowDropdown(false)
    inputRef.current?.focus()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div
        className="w-full max-w-md rounded-2xl p-6 space-y-4"
        style={{ border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(145deg, #101015, #0d0d10)" }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-violet-500/10">
            <UserPlus className="h-5 w-5 text-violet-400" />
          </div>
          <h2 className="text-lg font-semibold text-white">Пригласить участника</h2>
        </div>

        <div className="space-y-2 relative">
          <label className="text-[0.8rem] font-medium text-zinc-300">Email или @username</label>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="user@example.com или @username"
            autoFocus
            className="flex h-10 w-full rounded-lg px-3.5 text-sm text-white outline-none transition-all placeholder:text-zinc-600"
            style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.025)" }}
            onFocus={(e) => { e.target.style.borderColor = "rgba(139,92,246,0.4)"; e.target.style.boxShadow = "0 0 0 3px rgba(139,92,246,0.08)" }}
            onBlur={(e) => {
              e.target.style.borderColor = "rgba(255,255,255,0.07)"
              e.target.style.boxShadow = "none"
              // Delay hiding so click on dropdown registers
              setTimeout(() => setShowDropdown(false), 200)
            }}
            onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
          />

          {/* Search dropdown */}
          {showDropdown && searchResults && searchResults.length > 0 && (
            <div
              className="absolute left-0 right-0 top-full z-10 mt-1 max-h-48 overflow-y-auto rounded-lg py-1"
              style={{ border: "1px solid rgba(255,255,255,0.08)", background: "#18181b" }}
            >
              {searchResults.map((u) => (
                <button
                  key={u.id}
                  type="button"
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => handleSelectUser(u.username)}
                  className="flex w-full items-center gap-3 px-3 py-2 text-left transition-colors hover:bg-white/[0.04]"
                >
                  {u.avatar_url ? (
                    <img src={u.avatar_url} alt={u.name} className="h-7 w-7 rounded-full object-cover" />
                  ) : (
                    <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-violet-500/15 text-[0.65rem] font-semibold text-violet-300">
                      {u.name.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[0.8rem] font-medium text-white">{u.name}</p>
                    <p className="truncate text-[0.7rem] text-zinc-500">
                      {u.username && <span className="text-violet-400">@{u.username}</span>}
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
          <label className="text-[0.8rem] font-medium text-zinc-300">Роль</label>
          <div className="flex gap-2">
            {(["editor", "viewer"] as const).map((r) => (
              <button
                key={r}
                onClick={() => setRole(r)}
                className={`flex-1 rounded-lg px-3 py-2 text-[0.8rem] font-medium transition-all ${
                  role === r
                    ? "bg-violet-600/20 text-violet-300 ring-1 ring-violet-500/30"
                    : "text-zinc-500 hover:text-zinc-300"
                }`}
                style={{ border: role === r ? undefined : "1px solid rgba(255,255,255,0.06)", background: role === r ? undefined : "rgba(255,255,255,0.02)" }}
              >
                {r === "editor" ? "Редактор" : "Читатель"}
              </button>
            ))}
          </div>
          <p className="text-[0.7rem] text-zinc-600">
            {role === "editor" ? "Может управлять промптами и коллекциями команды" : "Может только просматривать"}
          </p>
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <button
            onClick={onClose}
            className="flex h-9 items-center rounded-lg px-4 text-[0.8rem] text-zinc-500 transition-all hover:text-zinc-300"
            style={{ border: "1px solid rgba(255,255,255,0.06)", background: "rgba(255,255,255,0.02)" }}
          >
            Отмена
          </button>
          <button
            onClick={handleSubmit}
            disabled={!query.trim() || isPending}
            className="flex h-9 items-center gap-2 rounded-lg px-5 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97] disabled:opacity-50"
            style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 4px 16px -2px rgba(124,58,237,0.25)" }}
          >
            {isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            Пригласить
          </button>
        </div>
      </div>
    </div>
  )
}
