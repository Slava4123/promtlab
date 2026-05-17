import { Lock } from "lucide-react"
import { Link } from "react-router-dom"

interface InsightsLockedCardProps {
  title: string
  description: string
}

// InsightsLockedCard — teaser-карточка для инсайтов, доступных только на Max.
// Показывается Pro-юзерам рядом с уже доступными типами (unused/duplicates),
// чтобы дать наглядное представление о ценности апгрейда. Стиль — border-dashed
// + muted (как UpgradeGate), чтобы визуально отличался от «живых» карточек.
export function InsightsLockedCard({ title, description }: InsightsLockedCardProps) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-muted/20 p-4">
      <div className="mb-2 flex items-center gap-2">
        <Lock className="h-4 w-4 text-muted-foreground" />
        <h3 className="text-sm font-medium text-muted-foreground">{title}</h3>
      </div>
      <p className="mb-3 text-xs text-muted-foreground">{description}</p>
      <Link
        to="/pricing"
        className="text-xs font-medium text-violet-600 hover:underline dark:text-violet-400"
      >
        Доступно в Max →
      </Link>
    </div>
  )
}
