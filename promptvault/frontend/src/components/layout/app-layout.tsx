import { useState } from "react"
import { Outlet } from "react-router-dom"
import { Toaster } from "sonner"
import { Search, Bell, Check, X, Users } from "lucide-react"
import { toast } from "sonner"

import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { TooltipProvider } from "@/components/ui/tooltip"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { CommandPalette } from "@/components/command-palette"
import { useMyInvitations, useAcceptInvitation, useDeclineInvitation } from "@/hooks/use-teams"
import { RoleBadge } from "@/components/teams/role-badge"

// TODO: centralize hardcoded dark theme colors (bg-[#0a0a0c], from-[#101015], etc.) as Tailwind theme tokens
export default function AppLayout() {
  const { data: invitations } = useMyInvitations()
  const acceptInvitation = useAcceptInvitation()
  const declineInvitation = useDeclineInvitation()
  const [bellOpen, setBellOpen] = useState(false)

  const pendingCount = invitations?.length ?? 0

  const handleAccept = (id: number) => {
    acceptInvitation.mutate(id, {
      onSuccess: () => toast.success("Вы присоединились к команде"),
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleDecline = (id: number) => {
    declineInvitation.mutate(id, {
      onSuccess: () => toast.success("Приглашение отклонено"),
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  return (
    <TooltipProvider>
      <SidebarProvider>
        <div className="flex min-h-screen w-full">
          <AppSidebar />
          <div className="flex flex-1 flex-col">
            <header className="flex h-10 items-center justify-between px-4">
              <div className="lg:hidden">
                <SidebarTrigger />
              </div>
              <div className="ml-auto flex items-center gap-2">
                {/* Notifications bell */}
                <div className="relative">
                  <button
                    onClick={() => setBellOpen(!bellOpen)}
                    className="relative flex h-7 w-7 cursor-pointer items-center justify-center rounded-lg border border-white/[0.06] bg-white/[0.02] text-zinc-500 transition-colors hover:bg-white/[0.04] hover:text-zinc-300"
                  >
                    <Bell className="h-3.5 w-3.5" />
                    {pendingCount > 0 && (
                      <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-violet-500 px-1 text-[9px] font-bold text-white">
                        {pendingCount}
                      </span>
                    )}
                  </button>

                  {bellOpen && (
                    <>
                      <div className="fixed inset-0 z-40" onClick={() => setBellOpen(false)} />
                      <div
                        className="absolute right-0 top-9 z-50 w-80 rounded-xl shadow-xl"
                        style={{ border: "1px solid rgba(255,255,255,0.08)", background: "#151518" }}
                      >
                        <div className="flex items-center justify-between px-4 py-3 border-b border-white/[0.06]">
                          <p className="text-[0.8rem] font-medium text-white">Приглашения</p>
                          {pendingCount > 0 && (
                            <span className="text-[0.7rem] text-zinc-500">{pendingCount}</span>
                          )}
                        </div>
                        {pendingCount === 0 ? (
                          <div className="px-4 py-6 text-center">
                            <p className="text-[0.78rem] text-zinc-600">Нет приглашений</p>
                          </div>
                        ) : (
                          <div className="max-h-[300px] overflow-y-auto">
                            {invitations?.map((inv) => (
                              <div key={inv.id} className="border-b border-white/[0.04] px-4 py-3 last:border-0">
                                <div className="flex items-center gap-2 mb-1.5">
                                  <Users className="h-3.5 w-3.5 text-violet-400/60" />
                                  <p className="text-[0.78rem] font-medium text-white truncate">{inv.team_name}</p>
                                </div>
                                <p className="text-[0.7rem] text-zinc-500 mb-2">
                                  {inv.inviter_name} приглашает вас как <RoleBadge role={inv.role} />
                                </p>
                                <div className="flex gap-2">
                                  <button
                                    onClick={() => handleAccept(inv.id)}
                                    disabled={acceptInvitation.isPending}
                                    className="flex h-7 flex-1 items-center justify-center gap-1 rounded-lg text-[0.75rem] font-medium text-white transition-all active:scale-[0.97]"
                                    style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)" }}
                                  >
                                    <Check className="h-3 w-3" />
                                    Принять
                                  </button>
                                  <button
                                    onClick={() => handleDecline(inv.id)}
                                    disabled={declineInvitation.isPending}
                                    className="flex h-7 flex-1 items-center justify-center gap-1 rounded-lg text-[0.75rem] text-zinc-500 transition-all hover:text-zinc-300"
                                    style={{ border: "1px solid rgba(255,255,255,0.06)" }}
                                  >
                                    <X className="h-3 w-3" />
                                    Отклонить
                                  </button>
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </>
                  )}
                </div>

                {/* Search button */}
                <button
                  type="button"
                  onClick={() =>
                    window.dispatchEvent(
                      new KeyboardEvent("keydown", { key: "k", metaKey: true }),
                    )
                  }
                  className="flex h-7 cursor-pointer items-center gap-2 rounded-lg border border-white/[0.06] bg-white/[0.02] px-2.5 text-[0.75rem] text-zinc-500 transition-colors hover:bg-white/[0.04] hover:text-zinc-300"
                >
                  <Search className="h-3 w-3" />
                  <span className="hidden sm:inline">Поиск...</span>
                  <kbd className="hidden rounded border border-white/[0.06] bg-white/[0.03] px-1 py-px text-[9px] sm:inline">
                    ⌘K
                  </kbd>
                </button>
              </div>
            </header>
            <main className="flex-1 px-8 py-7">
              <Outlet />
            </main>
          </div>
        </div>
        <CommandPalette />
        <Toaster richColors position="bottom-center" />
      </SidebarProvider>
    </TooltipProvider>
  )
}
