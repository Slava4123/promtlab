import { Link } from "react-router-dom"
import { Lock } from "lucide-react"

export function LandingFooter() {
  return (
    <footer className="border-t border-border/30 py-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center gap-3 px-6 sm:flex-row sm:justify-between">
        <div className="flex items-center gap-2 text-sm text-muted-foreground/40">
          <Lock className="h-3.5 w-3.5" />
          ПромтЛаб &copy; {new Date().getFullYear()}
        </div>
        <div className="flex items-center gap-4 text-xs text-muted-foreground/40">
          <Link to="/legal/terms" className="transition-colors hover:text-foreground">
            Условия использования
          </Link>
          <Link to="/legal/privacy" className="transition-colors hover:text-foreground">
            Конфиденциальность
          </Link>
          <Link to="/legal/offer" className="transition-colors hover:text-foreground">
            Публичная оферта
          </Link>
        </div>
      </div>
    </footer>
  )
}
