import type { LucideIcon } from "lucide-react"

interface Props {
  title: string
  description?: string
  icon?: LucideIcon
}

export function SectionHeader({ title, description, icon: Icon }: Props) {
  return (
    <div className="mb-6">
      <h2 className="flex items-center gap-2 text-lg font-semibold text-foreground">
        {Icon && <Icon className="size-5" aria-hidden />}
        {title}
      </h2>
      {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
    </div>
  )
}
