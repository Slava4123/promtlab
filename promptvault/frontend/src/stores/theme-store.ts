import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

type Theme = "light" | "dark" | "system"

interface ThemeState {
  theme: Theme
  toggle: () => void
  setTheme: (theme: Theme) => void
}

function getSystemTheme(): "light" | "dark" {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light"
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
        theme: "dark",
        toggle: () =>
          set((state) => {
            const next = state.theme === "dark" ? "light" : state.theme === "light" ? "system" : "dark"
            applyTheme(next)
            return { theme: next }
          }),
        setTheme: (theme) => {
          applyTheme(theme)
          set({ theme })
        },
      }),
      { name: "theme-store" },
    ),
    { name: "theme-store" },
  ),
)

// Применить при загрузке
const stored = localStorage.getItem("theme-store")
if (stored) {
  try {
    const { state } = JSON.parse(stored)
    applyTheme(state.theme)
  } catch {
    applyTheme("dark")
  }
} else {
  applyTheme("dark")
}

// Следить за изменением системной темы
window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
  const { theme } = useThemeStore.getState()
  if (theme === "system") {
    applyTheme("system")
  }
})
