import { useRef, useState, useEffect, type RefObject } from "react"
import { Link, Navigate } from "react-router-dom"
import {
  FileText, Sparkles, History, FolderOpen, Users, Cpu,
  Zap, Wand2, BarChart3, Shuffle, Check, ArrowRight,
  Globe, Server, GitBranch, Shield, Search, Puzzle,
  Star, ChevronRight, Flame, Lock,
} from "lucide-react"
import { useAuthStore } from "@/stores/auth-store"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

/* ------------------------------------------------------------------ */
/*  useInView — fade-in при появлении в viewport                     */
/* ------------------------------------------------------------------ */
function useInView(opts?: IntersectionObserverInit): [RefObject<HTMLDivElement | null>, boolean] {
  const ref = useRef<HTMLDivElement | null>(null)
  const [visible, setVisible] = useState(false)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const obs = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) { setVisible(true); obs.disconnect() } },
      { threshold: 0.15, ...opts },
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [])
  return [ref, visible]
}

function FadeIn({ children, className, delay = 0 }: { children: React.ReactNode; className?: string; delay?: number }) {
  const [ref, visible] = useInView()
  return (
    <div
      ref={ref}
      className={cn("transition-all duration-700", className)}
      style={{
        transitionTimingFunction: "cubic-bezier(0.22, 1, 0.36, 1)",
        transitionDelay: `${delay}ms`,
        opacity: visible ? 1 : 0,
        transform: visible ? "translateY(0)" : "translateY(24px)",
      }}
    >
      {children}
    </div>
  )
}

/* ------------------------------------------------------------------ */
/*  Data                                                              */
/* ------------------------------------------------------------------ */
const features = [
  { icon: FileText,  title: "Библиотека промптов",  desc: "Храни, ищи и организуй все промпты в одном месте. Избранное, теги, фильтрация — всё под рукой.", span: "lg:col-span-2" },
  { icon: Sparkles,  title: "AI-ассистент",          desc: "Claude Sonnet 4 улучшает, переписывает и анализирует промпты. SSE-стриминг — результат в реальном времени.", span: "" },
  { icon: History,   title: "Версионирование",       desc: "Каждое изменение сохраняется. Сравнивай версии, откатывайся к любой предыдущей.", span: "" },
  { icon: FolderOpen,title: "Коллекции и теги",      desc: "Группируй по проектам, темам, клиентам. Цветные теги, фильтрация, поиск Cmd+K.", span: "" },
  { icon: Users,     title: "Команды с ролями",      desc: "Owner, editor, viewer — управляй доступом. Совместная работа над промптами.", span: "" },
  { icon: Cpu,       title: "MCP для Claude",        desc: "Твои промпты доступны прямо из Claude Desktop и Cursor. 24 инструмента через MCP.", span: "lg:col-span-2" },
]

const aiActions = [
  { icon: Wand2,     title: "Улучшить",     desc: "Сделать промпт точнее и эффективнее" },
  { icon: Shuffle,   title: "Переписать",   desc: "Другой стиль, тон или формат" },
  { icon: BarChart3, title: "Анализ",       desc: "Оценка качества и рекомендации" },
  { icon: Zap,       title: "Вариации",     desc: "4 альтернативных варианта промпта" },
]

const plans = [
  {
    name: "Free", price: "0", period: "",
    desc: "Для знакомства с продуктом",
    highlight: false,
    features: ["50 промптов", "3 коллекции", "5 AI-запросов в день", "1 команда (до 3 чел.)", "Поиск и теги"],
  },
  {
    name: "Pro", price: "599", period: "/мес",
    desc: "Для ежедневной работы с AI",
    highlight: true,
    features: ["500 промптов", "Безлимит коллекций", "100 AI-запросов в день", "5 команд (до 10 чел.)", "Экспорт JSON/MD", "Приоритетная поддержка"],
  },
  {
    name: "Max", price: "1 299", period: "/мес",
    desc: "Для команд и бизнеса",
    highlight: false,
    features: ["Безлимит промптов", "Безлимит коллекций", "Безлимит AI-запросов", "Безлимит команд", "API-доступ", "Экспорт + API", "Приоритетная поддержка"],
  },
]

const advantages = [
  { icon: Globe,     title: "На русском языке",   desc: "Единственный менеджер промптов с полностью русским интерфейсом" },
  { icon: Server,    title: "Self-hosted",         desc: "Данные на твоём сервере. Никакого AWS или Vercel — полный контроль" },
  { icon: GitBranch, title: "Версионирование",     desc: "Полная история изменений каждого промпта. Ни один конкурент этого не даёт" },
  { icon: Shield,    title: "Команды с ролями",    desc: "Owner / editor / viewer. Конкуренты работают только для соло-пользователей" },
]

/* ------------------------------------------------------------------ */
/*  Component                                                          */
/* ------------------------------------------------------------------ */
export default function Landing() {
  const { isAuthenticated, isLoading } = useAuthStore()
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    window.addEventListener("scroll", onScroll, { passive: true })
    return () => window.removeEventListener("scroll", onScroll)
  }, [])

  if (isLoading) return null
  if (isAuthenticated) return <Navigate to="/dashboard" replace />

  return (
    <div className="dark min-h-screen bg-background text-foreground overflow-x-hidden">

      {/* ============ HEADER ============ */}
      <header
        className={cn(
          "fixed inset-x-0 top-0 z-50 transition-all duration-300",
          scrolled
            ? "border-b border-border/50 bg-background/80 backdrop-blur-xl"
            : "bg-transparent",
        )}
      >
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500/25 to-violet-600/5 ring-1 ring-violet-500/15">
              <Lock className="h-4 w-4 text-violet-400" />
            </div>
            <span className="text-[0.95rem] font-semibold tracking-tight">ПромтЛаб</span>
          </Link>

          <nav className="hidden items-center gap-6 text-sm text-muted-foreground sm:flex">
            <a href="#features" className="transition-colors hover:text-foreground">Возможности</a>
            <a href="#pricing" className="transition-colors hover:text-foreground">Тарифы</a>
          </nav>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" asChild>
              <Link to="/sign-in">Войти</Link>
            </Button>
            <Button variant="brand" size="sm" asChild>
              <Link to="/sign-up">Начать бесплатно</Link>
            </Button>
          </div>
        </div>
      </header>

      {/* ============ HERO ============ */}
      <section className="relative flex min-h-[90vh] flex-col items-center justify-center px-6 pt-20 text-center">
        {/* Glow */}
        <div className="pointer-events-none absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 h-[500px] w-[700px] rounded-full bg-violet-500/8 blur-[120px]" />
        <div className="pointer-events-none absolute top-1/3 left-1/3 h-[300px] w-[300px] rounded-full bg-violet-600/5 blur-[80px]" />

        <FadeIn>
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-violet-500/20 bg-violet-500/5 px-4 py-1.5 text-sm text-violet-300">
            <Flame className="h-3.5 w-3.5" />
            Открытый бета-тест
          </div>
        </FadeIn>

        <FadeIn delay={100}>
          <h1 className="mx-auto max-w-3xl text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
            Твои промпты заслуживают
            <span className="block bg-gradient-to-r from-violet-400 to-violet-200 bg-clip-text text-transparent">
              лучше, чем заметки
            </span>
          </h1>
        </FadeIn>

        <FadeIn delay={200}>
          <p className="mx-auto mt-5 max-w-xl text-lg text-muted-foreground sm:text-xl">
            Хватит терять промпты в чатах и закладках.
            Храни, улучшай с AI и используй повторно — один инструмент для всего.
          </p>
        </FadeIn>

        <FadeIn delay={300}>
          <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row">
            <Button variant="brand" size="lg" asChild>
              <Link to="/sign-up" className="gap-2">
                Начать бесплатно <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
            <Button variant="outline" size="lg" asChild>
              <Link to="#features">Подробнее</Link>
            </Button>
          </div>
          <p className="mt-3 text-xs text-muted-foreground/50">
            Бесплатно навсегда. Без карты.
          </p>
        </FadeIn>

        {/* Animated interface mock */}
        <FadeIn delay={500} className="mt-16 w-full max-w-3xl">
          <div className="relative rounded-xl border border-border/50 bg-card/50 p-1 shadow-2xl shadow-violet-500/5 ring-1 ring-white/5 backdrop-blur-sm">
            <div className="flex items-center gap-1.5 px-3 py-2">
              <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
              <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
              <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
              <div className="mx-auto text-xs text-muted-foreground/30">promtlabs.ru</div>
            </div>
            <div className="rounded-lg bg-background/80 p-6">
              <div className="flex items-center gap-3 border-b border-border/30 pb-4">
                <Search className="h-4 w-4 text-muted-foreground/40" />
                <span className="text-sm text-muted-foreground/40">Поиск промптов...</span>
                <span className="ml-auto rounded-md border border-border/30 px-1.5 py-0.5 text-[0.65rem] text-muted-foreground/30">⌘K</span>
              </div>
              <div className="mt-4 space-y-3">
                {[
                  { title: "Код-ревьюер", tags: ["development", "review"], fav: true },
                  { title: "Рерайт статьи", tags: ["content", "writing"], fav: false },
                  { title: "SQL-оптимизатор", tags: ["database", "perf"], fav: true },
                ].map((p, i) => (
                  <div key={i} className="flex items-center gap-3 rounded-lg border border-border/20 bg-card/30 px-4 py-3 transition-colors hover:border-violet-500/20">
                    <FileText className="h-4 w-4 text-violet-400/60" />
                    <span className="text-sm">{p.title}</span>
                    <div className="ml-auto flex items-center gap-2">
                      {p.tags.map(t => (
                        <span key={t} className="rounded-md bg-violet-500/10 px-2 py-0.5 text-[0.6rem] text-violet-300/70">{t}</span>
                      ))}
                      {p.fav && <Star className="h-3.5 w-3.5 text-amber-400/50" />}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </FadeIn>
      </section>

      {/* ============ SOCIAL PROOF ============ */}
      <section className="py-16">
        <FadeIn>
          <div className="mx-auto max-w-4xl px-6 text-center">
            <p className="mb-6 text-sm font-medium uppercase tracking-wider text-muted-foreground/40">
              Для тех, кто работает с
            </p>
            <div className="flex flex-wrap items-center justify-center gap-8 text-muted-foreground/30">
              {["ChatGPT", "Claude", "Gemini", "Perplexity", "Midjourney"].map(name => (
                <span key={name} className="text-lg font-semibold tracking-tight transition-colors hover:text-muted-foreground/60">{name}</span>
              ))}
            </div>
          </div>
        </FadeIn>
      </section>

      {/* ============ FEATURES ============ */}
      <section id="features" className="scroll-mt-20 py-20">
        <div className="mx-auto max-w-6xl px-6">
          <FadeIn>
            <div className="mb-14 text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">Всё для работы с промптами</h2>
              <p className="mt-3 text-muted-foreground">Не очередной блокнот. Полноценное рабочее пространство.</p>
            </div>
          </FadeIn>

          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f, i) => (
              <FadeIn key={f.title} delay={i * 80} className={f.span}>
                <div className="group relative h-full rounded-xl border border-border/50 bg-card/30 p-6 transition-all duration-300 hover:border-violet-500/20 hover:bg-card/50 ring-1 ring-white/[0.02]">
                  <div className="mb-4 inline-flex rounded-lg bg-violet-500/10 p-2.5">
                    <f.icon className="h-5 w-5 text-violet-400" />
                  </div>
                  <h3 className="mb-2 font-semibold">{f.title}</h3>
                  <p className="text-sm leading-relaxed text-muted-foreground">{f.desc}</p>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ============ AI SECTION ============ */}
      <section className="py-20">
        <div className="mx-auto max-w-6xl px-6">
          <div className="grid items-center gap-12 lg:grid-cols-2">
            <FadeIn>
              <div>
                <div className="mb-4 inline-flex items-center gap-2 rounded-full bg-violet-500/10 px-3 py-1 text-sm text-violet-300">
                  <Sparkles className="h-3.5 w-3.5" />
                  AI-ассистент
                </div>
                <h2 className="text-3xl font-bold sm:text-4xl">
                  Claude Sonnet 4
                  <span className="block text-muted-foreground">у тебя под рукой</span>
                </h2>
                <p className="mt-4 text-muted-foreground leading-relaxed">
                  Промпт можно улучшить за секунды. AI анализирует структуру, находит слабые места
                  и предлагает варианты — прямо в редакторе, в реальном времени.
                </p>
              </div>
            </FadeIn>

            <FadeIn delay={150}>
              <div className="grid gap-3 sm:grid-cols-2">
                {aiActions.map((a, i) => (
                  <div
                    key={a.title}
                    className="group relative rounded-xl border border-border/50 bg-card/30 p-5 transition-all duration-300 hover:border-violet-500/20 hover:bg-card/50"
                  >
                    <a.icon className="mb-3 h-5 w-5 text-violet-400" />
                    <h3 className="mb-1 text-sm font-semibold">{a.title}</h3>
                    <p className="text-xs leading-relaxed text-muted-foreground">{a.desc}</p>
                  </div>
                ))}
              </div>
            </FadeIn>
          </div>
        </div>
      </section>

      {/* ============ EXTRA FEATURES ROW ============ */}
      <section className="py-12">
        <div className="mx-auto max-w-6xl px-6">
          <FadeIn>
            <div className="grid gap-4 sm:grid-cols-3">
              {[
                { icon: Puzzle, title: "Браузерное расширение", desc: "Вставляй промпты в ChatGPT, Claude, Gemini одним кликом" },
                { icon: Search, title: "Глобальный поиск", desc: "Cmd+K — найди что угодно за секунду. Промпты, коллекции, теги" },
                { icon: Flame, title: "Стрики и бейджи", desc: "Геймификация помогает выработать привычку работы с промптами" },
              ].map(f => (
                <div key={f.title} className="rounded-xl border border-border/30 bg-card/20 p-5 transition-all duration-300 hover:border-border/60">
                  <f.icon className="mb-3 h-5 w-5 text-muted-foreground/60" />
                  <h3 className="mb-1 text-sm font-semibold">{f.title}</h3>
                  <p className="text-xs text-muted-foreground">{f.desc}</p>
                </div>
              ))}
            </div>
          </FadeIn>
        </div>
      </section>

      {/* ============ PRICING ============ */}
      <section id="pricing" className="scroll-mt-20 py-20">
        <div className="mx-auto max-w-5xl px-6">
          <FadeIn>
            <div className="mb-14 text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">Простые и честные тарифы</h2>
              <p className="mt-3 text-muted-foreground">Начни бесплатно. Перейди на Pro, когда будешь готов.</p>
            </div>
          </FadeIn>

          <div className="grid gap-6 sm:grid-cols-3">
            {plans.map((plan, i) => (
              <FadeIn key={plan.name} delay={i * 100}>
                <div
                  className={cn(
                    "relative flex h-full flex-col rounded-2xl border p-6 transition-all duration-300",
                    plan.highlight
                      ? "border-violet-500/40 bg-violet-500/5 shadow-lg shadow-violet-500/10 ring-1 ring-violet-500/20"
                      : "border-border/50 bg-card/30 hover:border-border",
                  )}
                >
                  {plan.highlight && (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-brand px-3 py-0.5 text-xs font-medium text-white">
                      Популярный
                    </div>
                  )}
                  <div className="mb-4">
                    <h3 className="text-lg font-semibold">{plan.name}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">{plan.desc}</p>
                  </div>
                  <div className="mb-6">
                    <span className="text-4xl font-bold">{plan.price}₽</span>
                    <span className="text-sm text-muted-foreground">{plan.period}</span>
                  </div>
                  <ul className="mb-8 flex-1 space-y-2.5">
                    {plan.features.map(f => (
                      <li key={f} className="flex items-start gap-2 text-sm text-muted-foreground">
                        <Check className="mt-0.5 h-4 w-4 shrink-0 text-violet-400" />
                        {f}
                      </li>
                    ))}
                  </ul>
                  <Button
                    variant={plan.highlight ? "brand" : "outline"}
                    className="w-full"
                    asChild
                  >
                    <Link to="/sign-up">{plan.price === "0" ? "Начать бесплатно" : "Выбрать " + plan.name}</Link>
                  </Button>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ============ ADVANTAGES ============ */}
      <section className="py-20">
        <div className="mx-auto max-w-6xl px-6">
          <FadeIn>
            <div className="mb-14 text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">Почему ПромтЛаб</h2>
              <p className="mt-3 text-muted-foreground">То, чего нет у конкурентов.</p>
            </div>
          </FadeIn>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {advantages.map((a, i) => (
              <FadeIn key={a.title} delay={i * 80}>
                <div className="rounded-xl border border-border/30 bg-card/20 p-6 text-center transition-all duration-300 hover:border-violet-500/15 hover:bg-card/40">
                  <div className="mx-auto mb-4 inline-flex rounded-lg bg-violet-500/10 p-3">
                    <a.icon className="h-5 w-5 text-violet-400" />
                  </div>
                  <h3 className="mb-2 text-sm font-semibold">{a.title}</h3>
                  <p className="text-xs leading-relaxed text-muted-foreground">{a.desc}</p>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ============ CTA BANNER ============ */}
      <section className="py-20">
        <FadeIn>
          <div className="relative mx-auto max-w-4xl px-6">
            {/* Glow */}
            <div className="pointer-events-none absolute inset-0 -z-10 rounded-3xl bg-violet-500/5 blur-[60px]" />

            <div className="relative rounded-2xl border border-violet-500/20 bg-card/30 px-8 py-16 text-center ring-1 ring-white/5 backdrop-blur-sm sm:px-16">
              <h2 className="text-3xl font-bold sm:text-4xl">
                Готов навести порядок
                <span className="block text-violet-400">в своих промптах?</span>
              </h2>
              <p className="mx-auto mt-4 max-w-md text-muted-foreground">
                Регистрация за 30 секунд. Без карты. 50 промптов бесплатно навсегда.
              </p>
              <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
                <Button variant="brand" size="lg" asChild>
                  <Link to="/sign-up" className="gap-2">
                    Начать бесплатно <ArrowRight className="h-4 w-4" />
                  </Link>
                </Button>
              </div>
            </div>
          </div>
        </FadeIn>
      </section>

      {/* ============ FOOTER ============ */}
      <footer className="border-t border-border/30 py-8">
        <div className="mx-auto flex max-w-6xl flex-col items-center gap-3 px-6 sm:flex-row sm:justify-between">
          <div className="flex items-center gap-2 text-sm text-muted-foreground/40">
            <Lock className="h-3.5 w-3.5" />
            ПромтЛаб &copy; {new Date().getFullYear()}
          </div>
          <div className="flex items-center gap-4 text-xs text-muted-foreground/40">
            <Link to="/legal/terms" className="transition-colors hover:text-foreground">Условия использования</Link>
            <Link to="/legal/privacy" className="transition-colors hover:text-foreground">Конфиденциальность</Link>
          </div>
        </div>
      </footer>
    </div>
  )
}
