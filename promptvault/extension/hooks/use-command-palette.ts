import { useEffect, useState } from "react"

// Global Cmd+K listener. Подключается в AppShell, открывает CommandPalette.
export function useCommandPalette() {
  const [open, setOpen] = useState(false)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      const isMod = e.metaKey || e.ctrlKey
      if (isMod && e.key.toLowerCase() === "k" && !e.shiftKey) {
        // Не перехватываем если фокус в input (Cmd+K локально для поиска)
        const t = e.target as HTMLElement | null
        if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) {
          return
        }
        e.preventDefault()
        setOpen(true)
      }
      if (e.key === "Escape" && open) {
        setOpen(false)
      }
    }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [open])

  return {
    open,
    setOpen,
    openPalette: () => setOpen(true),
    closePalette: () => setOpen(false),
  }
}
