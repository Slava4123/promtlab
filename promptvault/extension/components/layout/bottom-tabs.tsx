import { NavLink } from "react-router-dom"
import {
  FileText,
  GitBranch,
  Users,
  History,
  Menu,
} from "lucide-react"
import { cn } from "../../lib/utils"

interface TabItem {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
  end?: boolean
}

// 4 главные tab'а на дне sidepanel + 5й slot — drawer trigger ("ещё").
// Аналитика убрана — это manage-time информация, в side-panel не вписывается
// (узкий экран, дублирует веб). История — use-time контекст «где недавно
// вставлял этот промпт», полезна прямо во время работы с AI.
const TABS: TabItem[] = [
  { to: "/", label: "Промпты", icon: FileText, end: true },
  { to: "/chains", label: "Цепочки", icon: GitBranch },
  { to: "/teams", label: "Команды", icon: Users },
  { to: "/history", label: "История", icon: History },
]

interface BottomTabsProps {
  onOpenDrawer: () => void
}

export function BottomTabs({ onOpenDrawer }: BottomTabsProps) {
  return (
    <nav
      className="sticky bottom-0 z-10 grid grid-cols-5 border-t border-(--color-border) bg-(--color-background)/95 backdrop-blur"
      aria-label="Основная навигация"
    >
      {TABS.map((tab) => (
        <NavLink
          key={tab.to}
          to={tab.to}
          end={tab.end}
          className={({ isActive }) =>
            cn(
              "flex flex-col items-center justify-center gap-0.5 py-2 text-[10px] font-medium transition-colors",
              isActive
                ? "text-(--color-brand)"
                : "text-(--color-muted-foreground) hover:text-(--color-foreground)",
            )
          }
        >
          {({ isActive }) => (
            <>
              <tab.icon
                className={cn(
                  "h-4 w-4",
                  isActive ? "text-(--color-brand)" : "text-current",
                )}
              />
              <span>{tab.label}</span>
            </>
          )}
        </NavLink>
      ))}
      <button
        type="button"
        onClick={onOpenDrawer}
        className="flex flex-col items-center justify-center gap-0.5 py-2 text-[10px] font-medium text-(--color-muted-foreground) hover:text-(--color-foreground)"
        aria-label="Открыть меню"
      >
        <Menu className="h-4 w-4" />
        <span>Ещё</span>
      </button>
    </nav>
  )
}
