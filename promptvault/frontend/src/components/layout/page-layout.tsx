import { cn } from "@/lib/utils"

const maxWidthMap = {
  sm: "max-w-[26rem]",
  md: "max-w-[48rem]",
  lg: "max-w-[64rem]",
  xl: "max-w-[72rem]",
} as const

interface PageLayoutProps {
  title: string
  description?: string
  action?: React.ReactNode
  maxWidth?: keyof typeof maxWidthMap
  className?: string
  children: React.ReactNode
}

function PageLayout({
  title,
  description,
  action,
  maxWidth = "lg",
  className,
  children,
}: PageLayoutProps) {
  return (
    <div className={cn("mx-auto space-y-5", maxWidthMap[maxWidth], className)}>
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
          {description && (
            <p className="mt-0.5 text-[0.8rem] text-muted-foreground">{description}</p>
          )}
        </div>
        {action}
      </div>
      {children}
    </div>
  )
}

export { PageLayout, type PageLayoutProps }
