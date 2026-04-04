import { MoreHorizontal, UserMinus, ArrowUpDown } from "lucide-react"
import { useState } from "react"
import type { TeamMember, TeamRole } from "@/api/types"
import { RoleBadge } from "./role-badge"

interface MemberListProps {
  members: TeamMember[]
  currentUserRole: TeamRole
  currentUserId: number
  onChangeRole: (userId: number, role: TeamRole) => void
  onRemove: (userId: number) => void
}

export function MemberList({ members, currentUserRole, currentUserId, onChangeRole, onRemove }: MemberListProps) {
  const [menuOpen, setMenuOpen] = useState<number | null>(null)
  const isOwner = currentUserRole === "owner"

  return (
    <div className="space-y-1">
      {members.map((m) => (
        <div
          key={m.user_id}
          className="flex items-center gap-3 rounded-lg px-3 py-2.5 transition-colors hover:bg-white/[0.02]"
        >
          {/* Avatar */}
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-violet-500/10 text-[0.75rem] font-medium text-violet-400">
            {m.avatar_url ? (
              <img src={m.avatar_url} alt="" className="h-8 w-8 rounded-full object-cover" />
            ) : (
              m.name?.charAt(0).toUpperCase() || m.email.charAt(0).toUpperCase()
            )}
          </div>

          {/* Info */}
          <div className="min-w-0 flex-1">
            <p className="truncate text-[0.8rem] font-medium text-white">
              {m.name || m.email}
              {m.user_id === currentUserId && <span className="ml-1.5 text-zinc-600">(вы)</span>}
            </p>
            <p className="truncate text-[0.7rem] text-zinc-600">{m.email}</p>
          </div>

          {/* Role badge */}
          <RoleBadge role={m.role} />

          {/* Actions */}
          {isOwner && m.role !== "owner" && (
            <div className="relative">
              <button
                onClick={() => setMenuOpen(menuOpen === m.user_id ? null : m.user_id)}
                className="rounded-md p-1 text-zinc-600 transition-colors hover:bg-white/[0.06] hover:text-zinc-300"
              >
                <MoreHorizontal className="h-4 w-4" />
              </button>

              {menuOpen === m.user_id && (
                <>
                  <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(null)} />
                  <div
                    className="absolute right-0 top-8 z-50 w-44 rounded-xl py-1 shadow-xl"
                    style={{ border: "1px solid rgba(255,255,255,0.08)", background: "#151518" }}
                  >
                    <button
                      onClick={() => {
                        onChangeRole(m.user_id, m.role === "editor" ? "viewer" : "editor")
                        setMenuOpen(null)
                      }}
                      className="flex w-full items-center gap-2 px-3 py-2 text-[0.78rem] text-zinc-400 transition-colors hover:bg-white/[0.04] hover:text-white"
                    >
                      <ArrowUpDown className="h-3.5 w-3.5" />
                      {m.role === "editor" ? "Сделать читателем" : "Сделать редактором"}
                    </button>
                    <button
                      onClick={() => {
                        onRemove(m.user_id)
                        setMenuOpen(null)
                      }}
                      className="flex w-full items-center gap-2 px-3 py-2 text-[0.78rem] text-red-400 transition-colors hover:bg-red-500/10"
                    >
                      <UserMinus className="h-3.5 w-3.5" />
                      Удалить из команды
                    </button>
                  </div>
                </>
              )}
            </div>
          )}

          {/* Self-leave for non-owner */}
          {!isOwner && m.user_id === currentUserId && (
            <button
              onClick={() => onRemove(m.user_id)}
              className="rounded-md p-1 text-zinc-600 transition-colors hover:bg-red-500/10 hover:text-red-400"
              title="Покинуть команду"
            >
              <UserMinus className="h-4 w-4" />
            </button>
          )}
        </div>
      ))}
    </div>
  )
}
