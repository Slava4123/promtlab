import { lazy, Suspense, useEffect, useRef, useState } from "react"
import { Navigate } from "react-router-dom"
import { useAuthStore } from "@/stores/auth-store"

import { LandingHeader } from "./landing/components/landing-header"
import { HeroSection } from "./landing/sections/hero"
import { CompatibilitySection } from "./landing/sections/compatibility"
import { LandingFooter } from "./landing/sections/landing-footer"

// P-13: Ниже-fold секции в отдельном чанке, lazy. Initial bundle landing'а
// держит только Hero + Compatibility — что видно до первого скролла.
const BelowFold = lazy(() => import("./landing/below-fold"))

export default function Landing() {
  const { isAuthenticated, isLoading } = useAuthStore()
  const [shouldLoadBelowFold, setShouldLoadBelowFold] = useState(false)
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  // Триггерим импорт BelowFold когда sentinel приближается к viewport.
  // rootMargin 600px — предзагружаем до того как юзер увидит пустое место
  // (обычный thumb-скролл на десктопе перепрыгивает ~1000px за тик).
  useEffect(() => {
    if (shouldLoadBelowFold) return
    const el = sentinelRef.current
    if (!el) return
    const obs = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting)) {
          setShouldLoadBelowFold(true)
          obs.disconnect()
        }
      },
      { rootMargin: "600px" },
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [shouldLoadBelowFold])

  if (isLoading) return null
  if (isAuthenticated) return <Navigate to="/dashboard" replace />

  return (
    <div className="dark min-h-screen bg-background text-foreground overflow-x-hidden">
      <LandingHeader />
      <main>
        <HeroSection />
        <CompatibilitySection />
        <div ref={sentinelRef} aria-hidden="true" />
        {shouldLoadBelowFold && (
          <Suspense fallback={null}>
            <BelowFold />
          </Suspense>
        )}
      </main>
      <LandingFooter />
    </div>
  )
}
