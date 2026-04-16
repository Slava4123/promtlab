import { FadeIn } from "../components/fade-in"
import { compatibleWith } from "../data/landing-content"

export function CompatibilitySection() {
  return (
    <section className="py-12 sm:py-16">
      <FadeIn>
        <div className="mx-auto max-w-4xl px-6 text-center">
          <p className="mb-6 text-sm font-medium uppercase tracking-wider text-muted-foreground">
            Работает с
          </p>
          <div className="flex flex-wrap items-center justify-center gap-8">
            {compatibleWith.map(name => (
              <span
                key={name}
                className="text-lg font-semibold tracking-tight text-muted-foreground/70 transition-colors hover:text-foreground"
              >
                {name}
              </span>
            ))}
          </div>
        </div>
      </FadeIn>
    </section>
  )
}
