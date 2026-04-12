import { FadeIn } from "../components/fade-in"
import { integrations } from "../data/landing-content"

export function IntegrationsRailSection() {
  return (
    <section id="integrations" className="scroll-mt-20 py-16 sm:py-24">
      <div className="mx-auto max-w-6xl px-6">
        <FadeIn>
          <div className="mb-10 text-center">
            <h2 className="text-3xl font-bold sm:text-4xl">Интеграции</h2>
            <p className="mt-3 text-muted-foreground">Промпты — там, где ты работаешь.</p>
          </div>
        </FadeIn>

        {/* Horizontal rail: flex on desktop, snap-x scroll on mobile */}
        <div className="-mx-6 px-6 sm:mx-0 sm:px-0">
          <div className="flex gap-4 overflow-x-auto pb-4 snap-x snap-mandatory sm:snap-none sm:overflow-visible sm:grid sm:grid-cols-4 sm:pb-0">
            {integrations.map((item, i) => (
              <FadeIn key={item.title} delay={i * 100} direction="right">
                <div className="min-w-[240px] flex-shrink-0 snap-start rounded-xl border border-border/30 bg-card/20 p-5 transition-all duration-300 hover:border-violet-500/15 hover:bg-card/40 sm:min-w-0">
                  <div className="mb-3 flex items-center gap-3">
                    <div className="inline-flex rounded-lg bg-violet-500/10 p-2">
                      <item.icon className="h-4 w-4 text-violet-400" />
                    </div>
                    {item.extra && (
                      <code className="ml-auto rounded-md border border-border/20 bg-background/40 px-2 py-0.5 font-mono text-[0.6rem] text-muted-foreground/50">
                        {item.extra}
                      </code>
                    )}
                  </div>
                  <h3 className="mb-1 text-sm font-semibold">{item.title}</h3>
                  <p className="text-xs leading-relaxed text-muted-foreground">{item.desc}</p>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
