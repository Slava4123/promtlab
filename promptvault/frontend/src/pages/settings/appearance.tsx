import { useThemeStore } from "@/stores/theme-store"
import { SectionHeader } from "./_section-header"

const themes = [
  { id: "dark", label: "Тёмная" },
  { id: "light", label: "Светлая" },
  { id: "system", label: "Системная" },
] as const

export default function SettingsAppearancePage() {
  const { theme, setTheme } = useThemeStore()

  return (
    <section>
      <SectionHeader title="Оформление" description="Тема интерфейса" />
      <div className="flex flex-wrap gap-2">
        {themes.map((t) => (
          <button
            key={t.id}
            onClick={() => setTheme(t.id)}
            className={`rounded-lg border px-4 h-11 text-sm font-medium transition-colors ${
              theme === t.id
                ? "border-brand/40 bg-brand-muted text-brand-muted-foreground"
                : "border-border bg-background text-muted-foreground hover:text-foreground"
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>
    </section>
  )
}
