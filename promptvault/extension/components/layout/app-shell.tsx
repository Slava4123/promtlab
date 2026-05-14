import { useEffect, useRef, useState } from "react"
import { Outlet, useLocation } from "react-router-dom"
import { BottomTabs } from "./bottom-tabs"
import { Drawer, useDrawer } from "./drawer"
import { CommandPalette } from "../command-palette"
import { useCommandPalette } from "../../hooks/use-command-palette"
import { QuotaExceededDialog } from "../subscription/quota-exceeded-dialog"
import { QuickSaveDialog } from "../prompts/quick-save-dialog"
import { ChangelogPopup } from "../changelog-popup"

// Layout оболочка для всех authenticated-страниц. Sticky bottom-tabs снизу,
// drawer slide-in для остального меню, глобальная Cmd+K command palette.
export function AppShell() {
  const drawer = useDrawer()
  const palette = useCommandPalette()
  const location = useLocation()
  const mainRef = useRef<HTMLElement | null>(null)
  const [pageAnnouncement, setPageAnnouncement] = useState("")

  // На каждую смену route:
  // 1. Закрываем drawer и palette.
  // 2. H3: переводим focus на <main> — Tab-навигация продолжается с верха
  //    нового контента, screen-reader перечитывает page-context.
  // 3. Объявляем смену route в live-region. Берём pathname как fallback;
  //    отдельные страницы могут override через document.title (TODO).
  // drawer/palette НЕ в deps — иначе закрытие drawer внутри эффекта
  // вызовет бесконечный цикл. setPageAnnouncement в effect правомерен:
  // ARIA-live announcement — это side-effect синхронизации с DOM.
  /* eslint-disable react-hooks/exhaustive-deps, react-hooks/set-state-in-effect */
  useEffect(() => {
    if (drawer.open) drawer.closeDrawer()
    if (palette.open) palette.closePalette()
    mainRef.current?.focus()
    const announcement = document.title || `Открыта ${location.pathname}`
    setPageAnnouncement(announcement)
  }, [location.pathname])
  /* eslint-enable react-hooks/exhaustive-deps, react-hooks/set-state-in-effect */

  return (
    <div className="flex h-full flex-col">
      <main
        ref={mainRef}
        tabIndex={-1}
        className="flex-1 overflow-y-auto outline-none"
      >
        <Outlet />
      </main>
      <BottomTabs onOpenDrawer={drawer.openDrawer} />
      <Drawer open={drawer.open} onClose={drawer.closeDrawer} />
      <CommandPalette open={palette.open} onClose={palette.closePalette} />
      <QuotaExceededDialog />
      <QuickSaveDialog />
      <ChangelogPopup />
      {/* Live-region для screen-reader. visually-hidden, но aria-live="polite"
          объявляет каждую смену route. */}
      <div
        role="status"
        aria-live="polite"
        aria-atomic="true"
        className="sr-only"
      >
        {pageAnnouncement}
      </div>
    </div>
  )
}
