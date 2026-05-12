import { useEffect } from "react"
import { Outlet, useLocation } from "react-router-dom"
import { BottomTabs } from "./bottom-tabs"
import { Drawer, useDrawer } from "./drawer"
import { CommandPalette } from "../command-palette"
import { useCommandPalette } from "../../hooks/use-command-palette"
import { QuotaExceededDialog } from "../subscription/quota-exceeded-dialog"
import { OverLimitBanner } from "../subscription/over-limit-banner"
import { QuickSaveDialog } from "../prompts/quick-save-dialog"
import { ChangelogPopup } from "../changelog-popup"

// Layout оболочка для всех authenticated-страниц. Sticky bottom-tabs снизу,
// drawer slide-in для остального меню, глобальная Cmd+K command palette.
export function AppShell() {
  const drawer = useDrawer()
  const palette = useCommandPalette()
  const location = useLocation()

  // Закрываем drawer и palette при навигации.
  useEffect(() => {
    if (drawer.open) drawer.closeDrawer()
    if (palette.open) palette.closePalette()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.pathname])

  return (
    <div className="flex h-full flex-col">
      <OverLimitBanner />
      <main className="flex-1 overflow-y-auto">
        <Outlet />
      </main>
      <BottomTabs onOpenDrawer={drawer.openDrawer} />
      <Drawer open={drawer.open} onClose={drawer.closeDrawer} />
      <CommandPalette open={palette.open} onClose={palette.closePalette} />
      <QuotaExceededDialog />
      <QuickSaveDialog />
      <ChangelogPopup />
    </div>
  )
}
