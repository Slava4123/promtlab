import { Link, Navigate } from "react-router-dom"
import { FileText, Sparkles, History, FolderOpen } from "lucide-react"
import { useAuthStore } from "@/stores/auth-store"

const features = [
  {
    icon: FileText,
    title: "Библиотека промптов",
    desc: "Храните, организуйте и находите промпты в одном месте",
  },
  {
    icon: Sparkles,
    title: "AI-улучшение",
    desc: "Улучшайте, переписывайте и анализируйте промпты с помощью AI",
  },
  {
    icon: History,
    title: "История версий",
    desc: "Отслеживайте изменения, сравнивайте версии, откатывайтесь",
  },
  {
    icon: FolderOpen,
    title: "Коллекции и теги",
    desc: "Группируйте промпты по проектам, темам и категориям",
  },
]

export default function Landing() {
  const { isAuthenticated, isLoading } = useAuthStore()

  if (isLoading) return null
  if (isAuthenticated) return <Navigate to="/dashboard" replace />

  return (
    <div className="flex min-h-screen flex-col bg-zinc-950 text-white">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 sm:px-10">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-violet-600">
            <FileText className="h-4 w-4" />
          </div>
          <span className="text-lg font-semibold">ПромтЛаб</span>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/sign-in"
            className="rounded-lg px-4 py-2 text-sm text-zinc-400 transition-colors hover:text-white"
          >
            Войти
          </Link>
          <Link
            to="/sign-up"
            className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium transition-colors hover:bg-violet-500"
          >
            Регистрация
          </Link>
        </div>
      </header>

      {/* Hero */}
      <main className="flex flex-1 flex-col items-center justify-center px-6 text-center">
        <div className="mx-auto max-w-2xl">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
            Управляйте AI-промптами
            <span className="block text-violet-400">как профессионал</span>
          </h1>
          <p className="mt-4 text-lg text-zinc-400 sm:text-xl">
            Сохраняйте, улучшайте и организуйте промпты.
            Встроенный AI-ассистент, версионирование и коллекции.
          </p>
          <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
            <Link
              to="/sign-up"
              className="w-full rounded-lg bg-violet-600 px-6 py-3 text-sm font-medium transition-colors hover:bg-violet-500 sm:w-auto"
            >
              Начать бесплатно
            </Link>
            <Link
              to="/sign-in"
              className="w-full rounded-lg border border-white/10 px-6 py-3 text-sm text-zinc-400 transition-colors hover:bg-white/5 hover:text-white sm:w-auto"
            >
              У меня есть аккаунт
            </Link>
          </div>
        </div>

        {/* Features */}
        <div className="mx-auto mt-20 grid max-w-4xl gap-6 sm:grid-cols-2">
          {features.map((f) => (
            <div
              key={f.title}
              className="rounded-xl border border-white/5 bg-white/[0.02] p-6 text-left transition-colors hover:border-violet-500/20 hover:bg-white/[0.04]"
            >
              <f.icon className="mb-3 h-6 w-6 text-violet-400" />
              <h3 className="text-sm font-semibold">{f.title}</h3>
              <p className="mt-1 text-sm text-zinc-500">{f.desc}</p>
            </div>
          ))}
        </div>
      </main>

      {/* Footer */}
      <footer className="py-6 text-center text-xs text-zinc-600">
        ПромтЛаб &copy; {new Date().getFullYear()}
      </footer>
    </div>
  )
}
