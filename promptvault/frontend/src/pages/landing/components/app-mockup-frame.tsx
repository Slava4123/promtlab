import { cn } from "@/lib/utils"

export function AppMockupFrame({
  children,
  className,
  url = "promtlabs.ru",
}: {
  children: React.ReactNode
  className?: string
  url?: string
}) {
  return (
    <div
      className={cn(
        "relative rounded-xl border border-border/50 bg-card/50 p-1 shadow-2xl shadow-violet-500/5 ring-1 ring-white/5 backdrop-blur-sm",
        className,
      )}
      aria-hidden="true"
    >
      {/* Browser chrome */}
      <div className="flex items-center gap-1.5 px-3 py-2">
        <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
        <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
        <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
        <div className="mx-auto text-xs text-muted-foreground/30">{url}</div>
      </div>
      {/* Content area */}
      <div className="rounded-lg bg-background/80 p-4 sm:p-6">
        {children}
      </div>
    </div>
  )
}
