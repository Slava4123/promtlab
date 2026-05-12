import { useLocation } from "react-router-dom"
import { Construction, ExternalLink } from "lucide-react"
import { Button } from "../components/ui/button"
import { useSettings } from "../hooks/use-settings"
import { openWebPage } from "../lib/utils"

interface PlaceholderProps {
  title: string
  description?: string
  phase?: string
  // Override для случаев, когда extension-роут не совпадает с frontend-роутом
  // (например /notifications в extension → /settings/notifications в вебе).
  webPath?: string
}

// Generic placeholder для тяжёлых страниц, которые остаются только в веб-приложении
// (chain editor + canvas, team branding/analytics/activity, settings security/accounts/...).
// CTA-кнопка ведёт по тому же pathname на frontend через openWebPage().
export function PlaceholderPage({ title, description, phase, webPath }: PlaceholderProps) {
  const location = useLocation()
  const settings = useSettings()
  const target = webPath ?? location.pathname

  function openInWeb() {
    if (!settings) return
    const separator = target.includes("?") ? "&" : "?"
    openWebPage(settings.apiBase, `${target}${separator}from=extension`)
  }

  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 px-6 py-8 text-center">
      <Construction className="h-12 w-12 text-(--color-muted-foreground)" />
      <div>
        <h2 className="text-base font-semibold text-(--color-foreground)">{title}</h2>
        {description && (
          <p className="mt-1 text-sm text-(--color-muted-foreground)">{description}</p>
        )}
      </div>
      {settings && (
        <Button type="button" size="sm" onClick={openInWeb} className="gap-1.5">
          <ExternalLink className="h-3.5 w-3.5" />
          Открыть в веб-приложении
        </Button>
      )}
      {phase && (
        <span className="rounded-full bg-(--color-muted) px-2 py-0.5 text-[10px] uppercase tracking-wide text-(--color-muted-foreground)">
          {phase}
        </span>
      )}
    </div>
  )
}
