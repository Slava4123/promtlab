import { cn } from "@/lib/utils"
import { FadeIn } from "../components/fade-in"
import { advantages } from "../data/landing-content"

export function AdvantagesBentoSection() {
  return (
    <section className="py-16 sm:py-24">
      <div className="mx-auto max-w-6xl px-6">
        <FadeIn>
          <div className="mb-14 text-center">
            <h2 className="text-3xl font-bold sm:text-4xl">Почему ПромтЛаб</h2>
            <p className="mt-3 text-muted-foreground">То, чего нет у конкурентов.</p>
          </div>
        </FadeIn>

        {/* Bento grid with varied sizes */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {advantages.map((a, i) => (
            <FadeIn key={a.title} delay={i * 100}>
              <div
                className={cn(
                  "group relative rounded-2xl border border-border/30 bg-card/20 p-6 transition-all duration-300 hover:border-violet-500/15 hover:bg-card/40",
                  a.span,
                )}
                style={{
                  transform: "scale(1)",
                  transition: "transform 300ms ease, border-color 300ms ease, background 300ms ease",
                }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.transform = "scale(1.01)" }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.transform = "scale(1)" }}
              >
                <div className="mb-4 inline-flex rounded-lg bg-violet-500/10 p-3">
                  <a.icon
                    className={cn(
                      "h-5 w-5 text-violet-400",
                      a.title.includes("достижений") && "text-amber-400",
                    )}
                    style={a.title.includes("достижений") ? { animation: "streak-flame 2s ease-in-out infinite" } : undefined}
                  />
                </div>
                <h3 className="mb-2 font-semibold">{a.title}</h3>
                <p className="text-sm leading-relaxed text-muted-foreground">{a.desc}</p>

              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  )
}
