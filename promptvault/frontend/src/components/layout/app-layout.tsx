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
        <div className="flex min-h-screen w-full overflow-x-hidden">
          <a href="#main-content" className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-background focus:px-4 focus:py-2 focus:text-sm focus:text-foreground focus:shadow-lg focus:ring-2 focus:ring-violet-500">
            Перейти к содержимому
          </a>
          <AppSidebar />
          <div className="flex flex-1 flex-col">
            <header role="banner" className="flex h-14 items-center justify-between px-4">
              <div className="lg:hidden">
                <SidebarTrigger />
              </div>
              <div className="ml-auto flex items-center gap-2">
                {/* Notifications bell */}
                <div className="relative">
                  <button
                    onClick={() => setBellOpen(!bellOpen)}
                    aria-label="Уведомления"
                    className="relative flex h-11 w-11 cursor-pointer items-center justify-center rounded-lg border border-border bg-muted/20 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                  >
                    <Bell className="h-4 w-4" />
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
                        className="fixed right-2 top-14 z-50 w-[min(20rem,calc(100vw-1rem))] rounded-xl shadow-xl sm:absolute sm:right-0 sm:top-9 sm:w-80 border border-border bg-popover"
                      >
                        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
                          <p className="text-[0.8rem] font-medium text-foreground">Приглашения</p>
                          {pendingCount > 0 && (
                            <span className="text-[0.7rem] text-muted-foreground">{pendingCount}</span>
                          )}
                        </div>
                        {pendingCount === 0 ? (
                          <div className="px-4 py-6 text-center">
                            <p className="text-[0.78rem] text-muted-foreground">Нет приглашений</p>
                          </div>
                        ) : (
                          <div className="max-h-[300px] overflow-y-auto">
                            {invitations?.map((inv) => (
                              <div key={inv.id} className="border-b border-border px-4 py-3 last:border-0">
                                <div className="flex items-center gap-2 mb-1.5">
                                  <Users className="h-3.5 w-3.5 text-violet-400/60" />
                                  <p className="text-[0.78rem] font-medium text-foreground truncate">{inv.team_name}</p>
                                </div>
                                <p className="text-[0.7rem] text-muted-foreground mb-2">
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
                                    className="flex h-7 flex-1 items-center justify-center gap-1 rounded-lg border border-border text-[0.75rem] text-muted-foreground transition-all hover:text-foreground"
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
                  className="flex h-11 min-w-11 cursor-pointer items-center gap-2 rounded-lg border border-border bg-muted/20 px-3 text-[0.8rem] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                >
                  <Search className="h-4 w-4" />
                  <span className="hidden sm:inline">Поиск...</span>
                  <kbd className="hidden rounded border border-border bg-muted/30 px-1 py-px text-[9px] sm:inline">
                    ⌘K
                  </kbd>
                </button>
              </div>
            </header>
            <main id="main-content" role="main" className="flex-1 overflow-x-hidden px-4 py-5 sm:px-8 sm:py-7">
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
