import { useNavigate } from "react-router-dom"
import { ArrowLeft, Sun, Moon, Monitor } from "lucide-react"
import { Button } from "../../components/ui/button"
import { Label } from "../../components/ui/label"
import { useThemeStore, type Theme } from "../../stores/theme-store"
import { setTheme as setStorageTheme } from "../../lib/storage"
import { cn } from "../../lib/utils"

const THEMES: { id: Theme; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { id: "light", label: "Светлая", icon: Sun },
  { id: "dark", label: "Тёмная", icon: Moon },
  { id: "system", label: "Системная", icon: Monitor },
]

export function AppearancePage() {
  const navigate = useNavigate()
  const theme = useThemeStore((s) => s.theme)
  const setStoreTheme = useThemeStore((s) => s.setTheme)

  async function pick(t: Theme) {
    setStoreTheme(t)
    await setStorageTheme(t)
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Внешний вид</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        <section className="space-y-2">
          <Label>Тема</Label>
          <div className="grid grid-cols-3 gap-1.5">
            {THEMES.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => pick(id)}
                className={cn(
                  "flex flex-col items-center justify-center gap-1 rounded-md border p-2 text-xs transition-colors",
                  theme === id
                    ? "border-(--color-primary) bg-(--color-primary)/10 text-(--color-primary)"
                    : "border-(--color-border) bg-(--color-card) hover:bg-(--color-muted)/40",
                )}
              >
                <Icon className="h-4 w-4" />
                <span>{label}</span>
              </button>
            ))}
          </div>
          <p className="text-[10px] text-(--color-muted-foreground)">
            «Системная» следует за настройкой вашей ОС.
          </p>
        </section>
      </div>
    </div>
  )
}
