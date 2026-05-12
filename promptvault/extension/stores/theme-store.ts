import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

export type Theme = "light" | "dark" | "system"

interface ThemeState {
  theme: Theme
  toggle: () => void
  setTheme: (theme: Theme) => void
}

function getSystemTheme(): "light" | "dark" {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
}

function resolveTheme(theme: Theme): "light" | "dark" {
  return theme === "system" ? getSystemTheme() : theme
}

function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle("dark", resolveTheme(theme) === "dark")
}

export const useThemeStore = create<ThemeState>()(
  devtools(
    persist(
      (set) => ({
        theme: "system",
        toggle: () =>
          set((state) => {
            const next: Theme =
              state.theme === "dark" ? "light" : state.theme === "light" ? "system" : "dark"
            applyTheme(next)
            return { theme: next }
          }),
        setTheme: (theme) => {
          applyTheme(theme)
          set({ theme })
        },
      }),
      { name: "pv-theme-store" },
    ),
    { name: "theme-store" },
  ),
)

const stored = typeof localStorage !== "undefined" ? localStorage.getItem("pv-theme-store") : null
if (stored) {
  try {
    const { state } = JSON.parse(stored)
    applyTheme(state.theme ?? "system")
  } catch {
    applyTheme("system")
  }
} else {
  applyTheme("system")
}

declare global {
  interface Window {
    __pvThemeMediaListenerInstalled?: boolean
  }
}
if (typeof window !== "undefined" && !window.__pvThemeMediaListenerInstalled) {
  window.__pvThemeMediaListenerInstalled = true
  window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
    const { theme } = useThemeStore.getState()
    if (theme === "system") {
      applyTheme("system")
    }
  })
}
