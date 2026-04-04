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
} from "lucide-react"
import type { LucideIcon } from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarFooter,
} from "@/components/ui/sidebar"
import { UserMenu } from "@/components/layout/user-menu"
import { useCollections } from "@/hooks/use-collections"
import { useTeams } from "@/hooks/use-teams"
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
]

const bottomNav: NavItem[] = [
  { title: "Настройки", icon: Settings, path: "/settings" },
]

function NavLink({ item, isActive, onClick }: { item: NavItem; isActive: boolean; onClick: () => void }) {
  const Icon = item.icon
  return (
    <button
      onClick={onClick}
      className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-[7px] text-[0.8rem] transition-colors ${
        isActive
          ? "bg-sidebar-accent font-medium text-white"
          : "text-zinc-500 hover:bg-white/[0.03] hover:text-zinc-300"
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
  const team = useWorkspaceStore((s) => s.team)
  const setTeam = useWorkspaceStore((s) => s.setTeam)
  const clearTeam = useWorkspaceStore((s) => s.clearTeam)
  const teamSlug = team?.teamSlug ?? null
  const teamId = team?.teamId ?? null
  const teamName = team?.teamName ?? null
  const { data: collections } = useCollections(teamId)
  const { data: teams } = useTeams()
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

  const handleSwitchToPersonal = () => {
    clearTeam()
    setSwitcherOpen(false)
    navigate("/dashboard")
  }

  const handleSwitchToTeam = (slug: string, id: number, name: string) => {
    setTeam(slug, id, name)
    setSwitcherOpen(false)
    navigate("/dashboard")
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
              className="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-[0.78rem] font-medium transition-colors hover:bg-white/[0.04]"
              style={{ border: "1px solid rgba(255,255,255,0.06)" }}
            >
              {teamSlug ? (
                <Users className="h-3.5 w-3.5 text-violet-400" />
              ) : (
                <User className="h-3.5 w-3.5 text-zinc-400" />
              )}
              <span className="flex-1 truncate text-left text-white">
                {currentTeamName || "Личное пространство"}
              </span>
              <ChevronDown className={`h-3 w-3 text-zinc-600 transition-transform ${switcherOpen ? "rotate-180" : ""}`} />
            </button>

            {switcherOpen && (
              <>
                <div className="fixed inset-0 z-40" onClick={() => setSwitcherOpen(false)} />
                <div
                  className="absolute left-1 right-1 top-10 z-50 rounded-xl py-1 shadow-xl"
                  style={{ border: "1px solid rgba(255,255,255,0.08)", background: "#151518" }}
                >
                  <button
                    onClick={handleSwitchToPersonal}
                    className="flex w-full items-center gap-2 px-3 py-2 text-[0.78rem] text-zinc-400 transition-colors hover:bg-white/[0.04] hover:text-white"
                  >
                    <User className="h-3.5 w-3.5" />
                    <span className="flex-1 text-left">Личное пространство</span>
                    {!teamSlug && <Check className="h-3 w-3 text-violet-400" />}
                  </button>
                  {teams && teams.length > 0 && (
                    <>
                      <div className="mx-3 my-1 border-t border-white/[0.06]" />
                      {teams.map((t) => (
                        <button
                          key={t.id}
                          onClick={() => handleSwitchToTeam(t.slug, t.id, t.name)}
                          className="flex w-full items-center gap-2 px-3 py-2 text-[0.78rem] text-zinc-400 transition-colors hover:bg-white/[0.04] hover:text-white"
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

          <p className="px-2.5 pb-1.5 pt-2 text-[0.65rem] font-medium uppercase tracking-wider text-zinc-600">Главное</p>
          {mainNav.map((item) => (
            <NavLink
              key={item.path}
              item={item}
              isActive={location.pathname === item.path}
              onClick={() => navigate(item.path)}
            />
          ))}

          {/* Коллекции */}
          {collections && collections.length > 0 && (
            <div className="!mt-3 border-t border-white/[0.04] pt-3">
              <button
                onClick={() => setCollectionsOpen(!collectionsOpen)}
                className="flex w-full items-center justify-between px-2.5 pb-1.5 pt-2"
              >
                <p className="text-[0.65rem] font-medium uppercase tracking-wider text-zinc-600">
                  Коллекции
                </p>
                <ChevronDown className={`h-3 w-3 text-zinc-600 transition-transform ${collectionsOpen ? "" : "-rotate-90"}`} />
              </button>
              {collectionsOpen && (
                <div className="space-y-1">
                  {collections.length > 5 && (
                    <div className="relative px-1 pb-0.5">
                      <Search className="absolute left-3 top-1/2 h-3 w-3 -translate-y-1/2 text-zinc-600" />
                      <input
                        value={collectionSearch}
                        onChange={(e) => setCollectionSearch(e.target.value)}
                        placeholder="Найти..."
                        className="h-7 w-full rounded-md bg-white/[0.03] pl-7 pr-2 text-[0.72rem] text-zinc-400 outline-none placeholder:text-zinc-700 focus:bg-white/[0.05]"
                      />
                    </div>
                  )}
                  <div className="max-h-[200px] space-y-0.5 overflow-y-auto scrollbar-none">
                  {collections
                    .filter((c) => !collectionSearch || c.name.toLowerCase().includes(collectionSearch.toLowerCase()))
                    .map((c) => (
                    <button
                      key={c.id}
                      onClick={() => navigate(`/collections/${c.id}`)}
                      className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-[7px] text-[0.8rem] transition-colors ${
                        location.pathname === `/collections/${c.id}`
                          ? "bg-sidebar-accent font-medium text-white"
                          : "text-zinc-500 hover:bg-white/[0.03] hover:text-zinc-300"
                      }`}
                    >
                      <span className="h-2 w-2 shrink-0 rounded-full" style={{ backgroundColor: c.color || "#8b5cf6" }} />
                      <span className="truncate">{c.name}</span>
                      <span className="ml-auto text-[0.65rem] tabular-nums text-zinc-600">{c.prompt_count}</span>
                    </button>
                  ))}
                  </div>
                </div>
              )}
            </div>
          )}

          <div className="!mt-3 border-t border-white/[0.04] pt-3">
            {bottomNav.map((item) => (
              <NavLink
                key={item.path}
                item={item}
                isActive={location.pathname === item.path}
                onClick={() => navigate(item.path)}
              />
            ))}
          </div>
        </div>
      </SidebarContent>

      <SidebarFooter className="p-3">
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  )
}
