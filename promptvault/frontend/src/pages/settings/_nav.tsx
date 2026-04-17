import { NavLink } from "react-router-dom"

import { cn } from "@/lib/utils"
import { NAV_ITEMS } from "./_nav-config"

export function SettingsNav() {
  return (
    <nav aria-label="Разделы настроек">
      {/* Desktop: sticky vertical list */}
      <div className="hidden md:flex flex-col gap-0.5">
        {NAV_ITEMS.map((item) => (
          <NavLink
            key={item.id}
            to={item.to}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[0.82rem] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand/40",
                isActive
                  ? "bg-brand-muted font-medium text-brand-muted-foreground"
                  : "text-muted-foreground hover:bg-foreground/[0.04] hover:text-foreground",
              )
            }
            end
          >
            {({ isActive }) => (
              <>
                <item.icon className={cn("h-4 w-4 shrink-0", isActive && "text-brand")} />
                <span className="truncate">{item.title}</span>
              </>
            )}
          </NavLink>
        ))}
      </div>

      {/* Mobile: horizontal scrolling pill-tabs */}
      <div className="md:hidden -mx-4 flex gap-1.5 overflow-x-auto px-4 pb-1 [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
        {NAV_ITEMS.map((item) => (
          <NavLink
            key={item.id}
            to={item.to}
            className={({ isActive }) =>
              cn(
                "flex shrink-0 items-center gap-1.5 rounded-full border px-3 text-[0.78rem] transition-colors min-h-[36px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand/40",
                isActive
                  ? "border-brand/40 bg-brand-muted text-brand-muted-foreground font-medium"
                  : "border-border bg-background text-muted-foreground hover:text-foreground",
              )
            }
            end
          >
            <item.icon className="h-3.5 w-3.5 shrink-0" />
            <span>{item.title}</span>
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
