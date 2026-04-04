import { Lock } from "lucide-react"

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="flex min-h-screen flex-col items-center justify-center p-4"
      style={{
        backgroundImage: "radial-gradient(circle, rgba(255,255,255,0.03) 1px, transparent 1px)",
        backgroundSize: "24px 24px",
      }}
    >
      {/* Лого */}
      <div className="mb-8 flex flex-col items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-b from-violet-500/20 to-violet-500/5 ring-1 ring-violet-500/20">
          <Lock className="h-6 w-6 text-violet-400" />
        </div>
        <span className="text-xl font-semibold tracking-tight text-white">ПромтЛаб</span>
      </div>

      {/* Карточка */}
      <div className="w-full max-w-[26rem] rounded-2xl border border-white/[0.06] bg-zinc-900/60 p-8 shadow-2xl shadow-black/40 backdrop-blur-sm">
        {children}
      </div>
    </div>
  )
}
