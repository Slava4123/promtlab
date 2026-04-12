import { useState, useCallback } from "react"
import { cn } from "@/lib/utils"
import { FadeIn } from "../components/fade-in"
import { useAutoAdvance } from "../hooks/use-auto-advance"
import { AppMockupFrame } from "../components/app-mockup-frame"
import { MockupPromptList } from "../components/mockup-prompt-list"
import { MockupAiStream } from "../components/mockup-ai-stream"
import { MockupDiffViewer } from "../components/mockup-diff-viewer"
import { demoTabs } from "../data/landing-content"

function TabPanel({ active, children }: { active: boolean; children: React.ReactNode }) {
  return (
    <div
      role="tabpanel"
      className={cn(
        "transition-all duration-300",
        active ? "opacity-100" : "pointer-events-none absolute inset-0 opacity-0",
      )}
      aria-hidden={!active}
    >
      {children}
    </div>
  )
}

export function ProductDemoSection() {
  const { activeIndex, progress, goTo, pause, resume } = useAutoAdvance({
    count: demoTabs.length,
    interval: 6000,
  })
  const [hasInteracted, setHasInteracted] = useState(false)

  const handleTabClick = useCallback((index: number) => {
    goTo(index)
    setHasInteracted(true)
  }, [goTo])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "ArrowRight") {
        handleTabClick((activeIndex + 1) % demoTabs.length)
      } else if (e.key === "ArrowLeft") {
        handleTabClick((activeIndex - 1 + demoTabs.length) % demoTabs.length)
      }
    },
    [activeIndex, handleTabClick],
  )

  return (
    <section id="demo" className="scroll-mt-20 py-16 sm:py-24">
      <div className="mx-auto max-w-5xl px-6">
        <FadeIn>
          <div className="mb-10 text-center">
            <h2 className="text-3xl font-bold sm:text-4xl">Смотри, как это работает</h2>
            <p className="mt-3 text-muted-foreground">Три главных сценария — в одном окне.</p>
          </div>
        </FadeIn>

        <FadeIn delay={150}>
          <div
            onMouseEnter={pause}
            onMouseLeave={resume}
            onFocus={pause}
            onBlur={resume}
          >
            {/* Tabs */}
            <div
              role="tablist"
              className="mb-6 flex items-center gap-1 rounded-lg border border-border/30 bg-card/20 p-1"
              onKeyDown={handleKeyDown}
            >
              {demoTabs.map((tab, i) => (
                <button
                  key={tab.id}
                  role="tab"
                  aria-selected={activeIndex === i}
                  tabIndex={activeIndex === i ? 0 : -1}
                  onClick={() => handleTabClick(i)}
                  className={cn(
                    "relative flex-1 rounded-md px-4 py-2 text-sm font-medium transition-all duration-200",
                    activeIndex === i
                      ? "bg-violet-500/10 text-foreground"
                      : "text-muted-foreground hover:text-foreground/70",
                  )}
                >
                  {tab.label}
                  {/* Progress bar */}
                  {activeIndex === i && !hasInteracted && (
                    <div className="absolute inset-x-1 bottom-0 h-0.5 overflow-hidden rounded-full bg-violet-500/10">
                      <div
                        className="h-full bg-violet-400/40 rounded-full"
                        style={{
                          transform: `scaleX(${progress})`,
                          transformOrigin: "left",
                          transition: "transform 100ms linear",
                        }}
                      />
                    </div>
                  )}
                </button>
              ))}
            </div>

            {/* Tab content */}
            <AppMockupFrame>
              <div className="relative min-h-[260px] sm:min-h-[280px]">
                <TabPanel active={activeIndex === 0}>
                  <MockupPromptList />
                </TabPanel>
                <TabPanel active={activeIndex === 1}>
                  <MockupAiStream active={activeIndex === 1} />
                </TabPanel>
                <TabPanel active={activeIndex === 2}>
                  <MockupDiffViewer />
                </TabPanel>
              </div>
            </AppMockupFrame>
          </div>
        </FadeIn>
      </div>
    </section>
  )
}
