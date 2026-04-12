import { Sparkles } from "lucide-react"
import { useTypewriter } from "../hooks/use-typewriter"
import { mockAiOriginal, mockAiImproved } from "../data/landing-content"

export function MockupAiStream({ active }: { active: boolean }) {
  const { displayText, cursor } = useTypewriter(mockAiImproved, {
    speed: 18,
    startDelay: 800,
    enabled: active,
  })

  return (
    <div className="space-y-3">
      {/* Original prompt */}
      <div className="rounded-lg border border-border/20 bg-card/20 p-3">
        <div className="mb-1.5 text-[0.6rem] font-medium uppercase tracking-wider text-muted-foreground/40">
          Оригинал
        </div>
        <p className="text-xs leading-relaxed text-muted-foreground/70">{mockAiOriginal}</p>
      </div>

      {/* AI result streaming */}
      <div className="rounded-lg border border-violet-500/20 bg-violet-500/5 p-3">
        <div className="mb-1.5 flex items-center gap-1.5 text-[0.6rem] font-medium uppercase tracking-wider text-violet-400/70">
          <Sparkles className="h-3 w-3" />
          Улучшено ИИ
        </div>
        <div className="text-xs leading-relaxed text-foreground/80 whitespace-pre-wrap">
          {active ? displayText : mockAiImproved}
          {cursor && (
            <span
              className="ml-0.5 inline-block h-3.5 w-[2px] bg-violet-400"
              style={{ animation: "cursor-blink 0.8s step-end infinite" }}
            />
          )}
        </div>
      </div>
    </div>
  )
}
