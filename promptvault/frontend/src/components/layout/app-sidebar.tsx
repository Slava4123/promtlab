import { useState, useEffect } from "react"
import { useLocation, useNavigate } from "react-router-dom"
import {
  FileText,
  FolderOpen,
  Users,
  Settings,
  Lock,
  ChevronDown,
  Search,
  User,
  Check,
  CreditCard,
  Trash2,
  Clock,
  Sparkles,
  Trophy,
  Shield,
} from "lucide-react"
import type { LucideIcon } from "lucide-react"
import { useQueryClient } from "@tanstack/react-query"

import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarFooter,
  useSidebar,
} from "@/components/ui/sidebar"
import { UserMenu } from "@/components/layout/user-menu"
import { FeedbackDialog } from "@/components/feedback/feedback-dialog"
import { useAuthStore } from "@/stores/auth-store"
import { PlanBadge } from "@/components/subscription/plan-badge"
import { useCollections } from "@/hooks/use-collections"
import { useTeams } from "@/hooks/use-teams"
import { useTrashCount } from "@/hooks/use-trash"
import { useWorkspaceStore } from "@/stores/workspace-store"

interface NavItem {
  title: string
  icon: LucideIcon
  path: string
}

const mainNav: NavItem[] = [
  { title: "Промпты", icon: FileText, path: "/dashboard" },
  { title: "Коллекции", icon: FolderOpen, path: "/collections" },
  { title: "Команды", icon: Users, path: "/teams" },
  { title: "История", icon: Clock, path: "/history" },
  { title: "Достижения", icon: Trophy, path: "/badges" },
]

const secondaryNav: NavItem[] = [
  { title: "Корзина", icon: Trash2, path: "/trash" },
]

const bottomNav: NavItem[] = [
  { title: "Тарифы", icon: CreditCard, path: "/pricing" },
  { title: "Что нового", icon: Sparkles, path: "/changelog" },
  { title: "Настройки", icon: Settings, path: "/settings" },
]

const adminNavItem: NavItem = { title: "Админ", icon: Shield, path: "/admin/users" }

function NavLink({ item, isActive, onClick }: { item: NavItem; isActive: boolean; onClick: () => void }) {
  const Icon = item.icon
  return (
    <button
      onClick={onClick}
      className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-[7px] text-[0.8rem] transition-colors ${
        isActive
          ? "bg-sidebar-accent font-medium text-sidebar-accent-foreground"
          : "text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
      }`}
    >
      <Icon className={`h-[15px] w-[15px] shrink-0 ${isActive ? "text-violet-400" : ""}`} />
      <span>{item.title}</span>
    </button>
  )
}

export function AppSidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const { isMobile, setOpenMobile } = useSidebar()
  const team = useWorkspaceStore((s) => s.team)
  const setTeam = useWorkspaceStore((s) => s.setTeam)
  const clearTeam = useWorkspaceStore((s) => s.clearTeam)
  const teamSlug = team?.teamSlug ?? null
  const teamId = team?.teamId ?? null
  const teamName = team?.teamName ?? null
  const qc = useQueryClient()
  const { data: collections } = useCollections(teamId)
  const { data: teams } = useTeams()
  const { data: trashCounts } = useTrashCount(teamId)
  const hasUnreadChangelog = useAuthStore((s) => s.user?.has_unread_changelog)
  const isAdmin = useAuthStore((s) => s.user?.role === "admin")
  const planId = useAuthStore((s) => s.user?.plan_id ?? "free")
  const trashTotal = (trashCounts?.prompts ?? 0) + (trashCounts?.collections ?? 0) + (trashCounts?.tags ?? 0)
  const [collectionsOpen, setCollectionsOpen] = useState(true)
  const [collectionSearch, setCollectionSearch] = useState("")
  const [switcherOpen, setSwitcherOpen] = useState(false)

  // Derive live name from teams data (handles rename), fallback to stored name
  const currentTeamName = teams?.find(t => t.id === teamId)?.name || teamName

  // Clear stale team context if user was removed from team or team deleted
  useEffect(() => {
    if (teamId && teams && !teams.some(t => t.id === teamId)) {
      clearTeam()
    }
  }, [teamId, teams, clearTeam])

  const go = (path: string) => {
    navigate(path)
    if (isMobile) setOpenMobile(false)
  }

  const invalidateWorkspaceQueries = () => {
    qc.invalidateQueries({ queryKey: ["prompts"] })
    qc.invalidateQueries({ queryKey: ["collections"] })
    qc.invalidateQueries({ queryKey: ["tags"] })
  }

  const stayOrRedirect = () => {
    const path = location.pathname
    // Detail pages are workspace-bound — redirect to parent list
    if (/^\/collections\/\d+/.test(path)) return go("/collections")
    if (/^\/teams\/[^/]+$/.test(path)) return go("/teams")
    if (/^\/prompts\/\d+/.test(path)) return go("/dashboard")
    // List pages — stay
    if (isMobile) setOpenMobile(false)
  }

  const handleSwitchToPersonal = () => {
    clearTeam()
    invalidateWorkspaceQueries()
    setSwitcherOpen(false)
    stayOrRedirect()
  }

  const handleSwitchToTeam = (slug: string, id: number, name: string) => {
    setTeam(slug, id, name)
    invalidateWorkspaceQueries()
    setSwitcherOpen(false)
    stayOrRedirect()
  }

  return (
    <Sidebar>
      <SidebarHeader className="px-4 py-3">
        <button onClick={() => navigate("/")} className="flex items-center gap-2.5">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500/25 to-violet-600/5 ring-1 ring-violet-500/15">
            <Lock className="h-3.5 w-3.5 text-violet-400" />
          </div>
          <span className="text-[0.85rem] font-semibold tracking-tight">ПромтЛаб</span>
        </button>
      </SidebarHeader>

      <SidebarContent>
        <div className="px-2.5 space-y-0.5">

          {/* Workspace Switcher */}
          <div className="relative px-1 pb-2">
            <button
              onClick={() => setSwitcherOpen(!switcherOpen)}
              className="flex w-full items-center gap-2 rounded-lg border border-sidebar-border px-2 py-1.5 text-[0.78rem] font-medium transition-colors hover:bg-sidebar-accent"
            >
              {teamSlug ? (
                <Users className="h-3.5 w-3.5 text-violet-400" />
              ) : (
                <User className="h-3.5 w-3.5 text-muted-foreground" />
              )}
              <span className="flex-1 truncate text-left text-sidebar-foreground">
                {currentTeamName || "Личное пространство"}
              </span>
              <ChevronDown className={`h-3 w-3 text-muted-foreground transition-transform ${switcherOpen ? "rotate-180" : ""}`} />
            </button>

            {switcherOpen && (
              <>
                <div className="fixed inset-0 z-40" onClick={() => setSwitcherOpen(false)} />
                <div
                  className="absolute left-1 right-1 top-10 z-50 rounded-xl py-1 shadow-xl border border-border bg-popover"
                >
                  <button
                    onClick={handleSwitchToPersonal}
                    className="flex w-full items-center gap-2 px-3 py-2 text-[0.78rem] text-sidebar-foreground/60 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
                  >
                    <User className="h-3.5 w-3.5" />
                    <span className="flex-1 text-left">Личное пространство</span>
                    {!teamSlug && <Check className="h-3 w-3 text-violet-400" />}
                  </button>
                  {teams && teams.length > 0 && (
                    <>
                      <div className="mx-3 my-1 border-t border-border" />
                      {teams.map((t) => (
                        <button
                          key={t.id}
                          onClick={() => handleSwitchToTeam(t.slug, t.id, t.name)}
                          className="flex w-full items-center gap-2 px-3 py-2 text-[0.78rem] text-sidebar-foreground/60 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
                        >
                          <Users className="h-3.5 w-3.5 text-violet-400/50" />
                          <span className="flex-1 text-left">{t.name}</span>
                          {teamSlug === t.slug && <Check className="h-3 w-3 text-violet-400" />}
                        </button>
                      ))}
                    </>
                  )}
                </div>
              </>
            )}
          </div>

          <p className="px-2.5 pb-1.5 pt-2 text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Главное</p>
          {mainNav.map((item) => (
            <NavLink
              key={item.path}
              item={item}
              isActive={location.pathname === item.path}
              onClick={() => go(item.path)}
            />
          ))}

          {/* Коллекции */}
          {collections && collections.length > 0 && (
            <div className="!mt-3 border-t border-border pt-3">
              <button
                onClick={() => setCollectionsOpen(!collectionsOpen)}
                className="flex w-full items-center justify-between px-2.5 pb-1.5 pt-2"
              >
                <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">
                  Коллекции
                </p>
                <ChevronDown className={`h-3 w-3 text-muted-foreground transition-transform ${collectionsOpen ? "" : "-rotate-90"}`} />
              </button>
              {collectionsOpen && (
                <div className="space-y-1">
                  {collections.length > 5 && (
                    <div className="relative px-1 pb-0.5">
                      <Search className="absolute left-3 top-1/2 h-3 w-3 -translate-y-1/2 text-muted-foreground" />
                      <input
                        value={collectionSearch}
                        onChange={(e) => setCollectionSearch(e.target.value)}
                        placeholder="Найти..."
                        className="h-7 w-full rounded-md bg-muted/30 pl-7 pr-2 text-[0.72rem] text-muted-foreground outline-none placeholder:text-muted-foreground/50 focus:bg-muted/50"
                      />
                    </div>
                  )}
                  <div className="max-h-[200px] space-y-0.5 overflow-y-auto scrollbar-none">
                  {collections
                    .filter((c) => !collectionSearch || c.name.toLowerCase().includes(collectionSearch.toLowerCase()))
                    .map((c) => (
                    <button
                      key={c.id}
                      onClick={() => go(`/collections/${c.id}`)}
                      className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-[7px] text-[0.8rem] transition-colors ${
                        location.pathname === `/collections/${c.id}`
                          ? "bg-sidebar-accent font-medium text-sidebar-accent-foreground"
                          : "text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
                      }`}
                    >
                      <span className="h-2 w-2 shrink-0 rounded-full" style={{ backgroundColor: c.color || "#8b5cf6" }} />
                      <span className="truncate">{c.name}</span>
                      <span className="ml-auto text-[0.65rem] tabular-nums text-muted-foreground">{c.prompt_count}</span>
                    </button>
                  ))}
                  </div>
                </div>
              )}
            </div>
          )}

          <div className="!mt-3 border-t border-border pt-3">
            {secondaryNav.map((item) => {
              const isTrash = item.path === "/trash"
              return (
                <button
                  key={item.path}
                  onClick={() => go(item.path)}
                  className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-[7px] text-[0.8rem] transition-colors ${
                    location.pathname === item.path
                      ? "bg-sidebar-accent font-medium text-sidebar-accent-foreground"
                      : "text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
                  }`}
                >
                  <item.icon className={`h-[15px] w-[15px] shrink-0 ${location.pathname === item.path ? "text-violet-400" : ""}`} />
                  <span>{item.title}</span>
                  {isTrash && trashTotal > 0 && (
                    <span className="ml-auto rounded-full bg-muted/50 px-1.5 py-px text-[0.6rem] tabular-nums text-muted-foreground">
                      {trashTotal}
                    </span>
                  )}
                </button>
              )
            })}
          </div>

          <div className="!mt-3 border-t border-border pt-3">
            {bottomNav.map((item) => (
              <div key={item.path} className="relative">
                <NavLink
                  item={item}
                  isActive={location.pathname === item.path}
                  onClick={() => go(item.path)}
                />
                {item.path === "/pricing" && planId !== "free" && (
                  <span className="absolute right-2 top-1/2 -translate-y-1/2">
                    <PlanBadge planId={planId as "free" | "pro" | "max"} />
                  </span>
                )}
                {item.path === "/changelog" && hasUnreadChangelog && (
                  <span className="absolute right-2 top-1/2 -translate-y-1/2 h-2 w-2 rounded-full bg-violet-500" />
                )}
              </div>
            ))}
            {isAdmin && (
              <NavLink
                item={adminNavItem}
                isActive={location.pathname.startsWith("/admin")}
                onClick={() => go(adminNavItem.path)}
              />
            )}
          </div>
        </div>
      </SidebarContent>

      <SidebarFooter className="p-3">
        <FeedbackDialog />
        <UserMenu />
        <div className="flex items-center justify-center gap-2 pt-1 text-[0.6rem] text-muted-foreground/40">
          <button onClick={() => go("/legal/terms")} className="hover:text-muted-foreground transition-colors">Условия</button>
          <span>&middot;</span>
          <button onClick={() => go("/legal/privacy")} className="hover:text-muted-foreground transition-colors">Конфиденциальность</button>
          <span>&middot;</span>
          <button onClick={() => go("/legal/offer")} className="hover:text-muted-foreground transition-colors">Оферта</button>
        </div>
      </SidebarFooter>
    </Sidebar>
  )
}
