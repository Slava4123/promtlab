import { Link } from "react-router-dom"
import { ArrowRight, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { FadeIn } from "../components/fade-in"
import { betaBanner } from "../data/landing-content"

export function BetaBannerSection() {
  return (
    <section className="py-16 sm:py-24">
      <FadeIn>
        <div className="relative mx-auto max-w-4xl px-6">
          {/* Glow */}
          <div
            className="pointer-events-none absolute inset-0 -z-10 rounded-3xl bg-violet-500/5 blur-[60px]"
            style={{ animation: "pulse-glow 4s ease-in-out infinite" }}
          />

          <div className="relative rounded-2xl border border-violet-500/20 bg-card/30 px-8 py-14 text-center ring-1 ring-white/5 backdrop-blur-sm sm:px-16 sm:py-16">
            <h2 className="text-3xl font-bold sm:text-4xl">
              {betaBanner.title}
            </h2>
            <p className="mt-3 text-muted-foreground">
              {betaBanner.sub}
            </p>

            {/* Feature checks */}
            <div className="mt-6 flex flex-wrap items-center justify-center gap-x-6 gap-y-2">
              {betaBanner.notes.map(note => (
                <div key={note} className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Check className="h-3.5 w-3.5 text-violet-400" />
                  {note}
                </div>
              ))}
            </div>

            <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
              <Button variant="brand" size="lg" nativeButton={false} render={<Link to="/sign-up" />} className="gap-2">
                {betaBanner.cta} <ArrowRight className="h-4 w-4" />
              </Button>
            </div>

            <p className="mt-4 text-xs text-muted-foreground/40">
              {betaBanner.footer}
            </p>
          </div>
        </div>
      </FadeIn>
    </section>
  )
}
