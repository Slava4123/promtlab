import { Lock } from "lucide-react"

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background p-4">
      {/* Лого */}
      <div className="mb-8 flex flex-col items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-b from-violet-500/20 to-violet-500/5 ring-1 ring-violet-500/20">
          <Lock className="h-6 w-6 text-violet-400" />
        </div>
        <span className="text-xl font-semibold tracking-tight text-foreground">ПромтЛаб</span>
      </div>

      {/* Карточка */}
      <div className="w-full max-w-[26rem] rounded-2xl border border-border bg-card p-8 shadow-2xl shadow-black/10 dark:shadow-black/40">
        {children}
      </div>
    </div>
  )
}
