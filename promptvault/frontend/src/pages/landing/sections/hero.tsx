import { Link } from "react-router-dom"
import { ArrowRight, Flame, FileText, Star, Search } from "lucide-react"
import { Button } from "@/components/ui/button"
import { FadeIn } from "../components/fade-in"
import { useTypewriter } from "../hooks/use-typewriter"
import { useFadeInView } from "../hooks/use-fade-in-view"
import { AppMockupFrame } from "../components/app-mockup-frame"
import { hero, mockPrompts } from "../data/landing-content"
import { useState, useEffect } from "react"

function HeroMockup() {
  const [ref, style] = useFadeInView({ delay: 500, duration: 900, distance: 40 })
  const [mockupVisible, setMockupVisible] = useState(false)

  useEffect(() => {
    if (style.opacity === 1) {
      const t = setTimeout(() => setMockupVisible(true), 300)
      return () => clearTimeout(t)
    }
  }, [style.opacity])

  const { displayText, cursor } = useTypewriter("Код-ревьюер", {
    speed: 60,
    startDelay: 1200,
    enabled: mockupVisible,
  })

  return (
    <div ref={ref} style={style} className="mt-16 w-full max-w-3xl">
      <AppMockupFrame>
        {/* Search bar */}
        <div className="flex items-center gap-3 border-b border-border/30 pb-4">
          <Search className="h-4 w-4 text-muted-foreground/60" />
          <span className="text-sm text-muted-foreground/60">
            {displayText || "Поиск промптов..."}
            {cursor && (
              <span
                className="ml-0.5 inline-block h-4 w-[2px] bg-violet-400"
                style={{ animation: "cursor-blink 0.8s step-end infinite" }}
              />
            )}
          </span>
          <span className="ml-auto rounded-md border border-border/30 px-1.5 py-0.5 text-[0.65rem] text-muted-foreground/30">
            ⌘K
          </span>
        </div>

        {/* Prompt cards with stagger */}
        <div className="mt-4 space-y-3">
          {mockPrompts.map((p, i) => (
            <div
              key={p.title}
              className="flex items-center gap-3 rounded-lg border border-border/20 bg-card/30 px-4 py-3 transition-all duration-500 hover:border-violet-500/20"
              style={{
                opacity: mockupVisible ? 1 : 0,
                transform: mockupVisible ? "translateY(0)" : "translateY(12px)",
                transition: `opacity 500ms ease, transform 500ms ease`,
                transitionDelay: `${800 + i * 150}ms`,
              }}
            >
              <FileText className="h-4 w-4 text-violet-400/60" />
              <span className="text-sm">{p.title}</span>
              <div className="ml-auto flex items-center gap-2">
                {p.tags.map(t => (
                  <span key={t} className="hidden rounded-md bg-violet-500/10 px-2 py-0.5 text-[0.6rem] text-violet-300/70 sm:inline">
                    {t}
                  </span>
                ))}
                {p.fav && <Star className="h-3.5 w-3.5 text-amber-400/50" />}
              </div>
            </div>
          ))}
        </div>
      </AppMockupFrame>
    </div>
  )
}

export function HeroSection() {
  return (
    <section className="relative flex min-h-[90vh] flex-col items-center justify-center px-6 pt-20 text-center">
      {/* Glow background */}
      <div
        className="pointer-events-none absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 h-[500px] w-[700px] rounded-full bg-violet-500/8 blur-[120px]"
        style={{ animation: "pulse-glow 4s ease-in-out infinite" }}
      />
      <div className="pointer-events-none absolute top-1/3 left-1/3 h-[300px] w-[300px] rounded-full bg-violet-600/5 blur-[80px]" />

      {/* Badge */}
      <FadeIn>
        <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-violet-500/20 bg-violet-500/5 px-4 py-1.5 text-sm text-violet-300">
          <Flame className="h-3.5 w-3.5" />
          {hero.badge}
        </div>
      </FadeIn>

      {/* Headline */}
      <FadeIn delay={100}>
        <h1 className="mx-auto max-w-3xl text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
          {hero.headline}
          <span className="block bg-gradient-to-r from-violet-400 to-violet-200 bg-clip-text text-transparent">
            {hero.headlineGradient}
          </span>
        </h1>
      </FadeIn>

      {/* Sub */}
      <FadeIn delay={200}>
        <p className="mx-auto mt-5 max-w-xl text-base text-muted-foreground sm:text-lg">
          {hero.sub}
        </p>
      </FadeIn>

      {/* CTAs */}
      <FadeIn delay={300}>
        <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row">
          <Button variant="brand" size="lg" nativeButton={false} render={<Link to="/sign-up" />} className="gap-2">
            {hero.cta} <ArrowRight className="h-4 w-4" />
          </Button>
          <a href="#features" className="group flex items-center gap-2 rounded-lg border border-border/30 bg-card/20 px-4 py-2.5 text-sm text-muted-foreground transition-colors hover:border-violet-500/20">
            Подробнее
          </a>
        </div>
        <p className="mt-3 text-xs text-muted-foreground">
          {hero.note}
        </p>
      </FadeIn>

      {/* Animated mockup */}
      <HeroMockup />
    </section>
  )
}
