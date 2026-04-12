import { cn } from "@/lib/utils"
import { FadeIn } from "../components/fade-in"
import { featurePillars } from "../data/landing-content"

export function FeatureDeepDivesSection() {
  const pillarPrompts = featurePillars[0]
  const pillarAi = featurePillars[1]
  const pillarVersions = featurePillars[2]

  const PromptsIcon = pillarPrompts.icon
  const AiIcon = pillarAi.icon
  const VersionsIcon = pillarVersions.icon

  return (
    <section id="features" className="scroll-mt-20 py-16 sm:py-24">
      <div className="mx-auto max-w-6xl px-6">
        <FadeIn>
          <div className="mb-14 text-center">
            <h2 className="text-3xl font-bold sm:text-4xl">Всё для работы с промптами</h2>
            <p className="mt-3 text-muted-foreground">Не очередной блокнот. Полноценное рабочее пространство.</p>
          </div>
        </FadeIn>

        {/* Asymmetric grid: first pillar tall left, second & third stacked right */}
        <div className="grid gap-4 lg:grid-cols-5">
          {/* Left: Prompts + Collections — tall card */}
          <FadeIn delay={0} direction="left" className="lg:col-span-3">
            <div className="group h-full rounded-2xl border border-border/50 bg-card/30 p-6 sm:p-8 transition-all duration-300 hover:border-violet-500/15 ring-1 ring-white/[0.02]">
              <div className="mb-4 inline-flex rounded-lg bg-violet-500/10 p-2.5">
                <PromptsIcon className="h-5 w-5 text-violet-400" />
              </div>
              <h3 className="mb-2 text-lg font-semibold">{pillarPrompts.title}</h3>
              <p className="mb-5 text-sm leading-relaxed text-muted-foreground">{pillarPrompts.desc}</p>

              {/* Mini mockup: sidebar with collections */}
              <div className="rounded-lg border border-border/20 bg-background/50 p-3" aria-hidden="true">
                <div className="mb-2 text-[0.6rem] font-medium uppercase tracking-wider text-muted-foreground/40">Коллекции</div>
                {[
                  { name: "Разработка", color: "bg-violet-500", count: 12 },
                  { name: "Контент", color: "bg-blue-500", count: 8 },
                  { name: "Аналитика", color: "bg-emerald-500", count: 5 },
                ].map(c => (
                  <div key={c.name} className="flex items-center gap-2 rounded-md px-2 py-1.5 transition-colors hover:bg-card/50">
                    <div className={cn("h-2 w-2 rounded-sm", c.color)} />
                    <span className="text-xs">{c.name}</span>
                    <span className="ml-auto text-[0.6rem] text-muted-foreground/40">{c.count}</span>
                  </div>
                ))}
              </div>

              <ul className="mt-5 grid gap-2 sm:grid-cols-2">
                {pillarPrompts.bullets.map(b => (
                  <li key={b} className="flex items-start gap-2 text-xs text-muted-foreground">
                    <div className="mt-1 h-1 w-1 shrink-0 rounded-full bg-violet-400/60" />
                    {b}
                  </li>
                ))}
              </ul>
            </div>
          </FadeIn>

          {/* Right column: AI + Versions stacked */}
          <div className="flex flex-col gap-4 lg:col-span-2">
            {/* AI Assistant */}
            <FadeIn delay={150} direction="right">
              <div className="group rounded-2xl border border-border/50 bg-card/30 p-6 transition-all duration-300 hover:border-violet-500/15 ring-1 ring-white/[0.02]">
                <div className="mb-4 inline-flex rounded-lg bg-violet-500/10 p-2.5">
                  <AiIcon className="h-5 w-5 text-violet-400" />
                </div>
                <h3 className="mb-2 text-lg font-semibold">{pillarAi.title}</h3>
                <p className="mb-4 text-sm leading-relaxed text-muted-foreground">{pillarAi.desc}</p>

                {/* AI actions grid */}
                <div className="grid grid-cols-2 gap-2" aria-hidden="true">
                  {pillarAi.bullets.map(b => {
                    const label = b.split(" — ")[0]
                    return (
                      <div key={b} className="rounded-md border border-border/20 bg-background/30 px-2.5 py-2 text-center">
                        <span className="text-[0.65rem] text-muted-foreground/70">{label}</span>
                      </div>
                    )
                  })}
                </div>
              </div>
            </FadeIn>

            {/* Versions */}
            <FadeIn delay={300} direction="right">
              <div className="group rounded-2xl border border-border/50 bg-card/30 p-6 transition-all duration-300 hover:border-violet-500/15 ring-1 ring-white/[0.02]">
                <div className="mb-4 inline-flex rounded-lg bg-violet-500/10 p-2.5">
                  <VersionsIcon className="h-5 w-5 text-violet-400" />
                </div>
                <h3 className="mb-2 text-lg font-semibold">{pillarVersions.title}</h3>
                <p className="mb-4 text-sm leading-relaxed text-muted-foreground">{pillarVersions.desc}</p>

                {/* Mini diff mockup */}
                <div className="rounded-md border border-border/20 bg-background/30 p-2 font-mono text-[0.55rem] leading-relaxed" aria-hidden="true">
                  <div className="rounded-sm bg-red-500/8 px-2 py-0.5 text-red-300/60">
                    <span className="text-red-400/30">&minus;</span> Проверяй код и пиши комментарии.
                  </div>
                  <div className="rounded-sm bg-emerald-500/8 px-2 py-0.5 text-emerald-300/60">
                    <span className="text-emerald-400/30">+</span> Проверяй: типизацию, обработку ошибок, граничные случаи.
                  </div>
                </div>
              </div>
            </FadeIn>
          </div>
        </div>
      </div>
    </section>
  )
}
