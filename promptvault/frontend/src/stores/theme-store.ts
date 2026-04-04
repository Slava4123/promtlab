import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

type Theme = "light" | "dark"

interface ThemeState {
  theme: Theme
  toggle: () => void
  setTheme: (theme: Theme) => void
}

export const useThemeStore = create<ThemeState>()(
  devtools(
    persist(
      (set) => ({
        theme: "dark",
        toggle: () =>
          set((state) => {
            const next = state.theme === "dark" ? "light" : "dark"
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

function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle("dark", theme === "dark")
}

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
