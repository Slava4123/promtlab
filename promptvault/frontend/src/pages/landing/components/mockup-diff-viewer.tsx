import { mockDiffOld, mockDiffNew } from "../data/landing-content"

export function MockupDiffViewer() {
  const oldLines = mockDiffOld.split("\n")
  const newLines = mockDiffNew.split("\n")

  return (
    <div className="rounded-lg border border-border/30 bg-card/20 overflow-hidden">
      {/* Diff header */}
      <div className="flex items-center gap-4 border-b border-border/20 px-3 py-2">
        <span className="text-[0.6rem] font-medium text-muted-foreground/50">v1 → v2</span>
        <span className="ml-auto text-[0.6rem] text-red-400/60">−{oldLines.length}</span>
        <span className="text-[0.6rem] text-emerald-400/60">+{newLines.length}</span>
      </div>

      <div className="p-3 font-mono text-[0.6rem] leading-relaxed sm:text-[0.65rem]">
        {/* Removed lines */}
        {oldLines.map((line, i) => (
          <div key={`old-${i}`} className="rounded-sm bg-red-500/8 px-2 py-0.5 text-red-300/70">
            <span className="mr-2 text-red-400/40">−</span>{line}
          </div>
        ))}
        {/* Added lines */}
        {newLines.map((line, i) => (
          <div key={`new-${i}`} className="rounded-sm bg-emerald-500/8 px-2 py-0.5 text-emerald-300/70">
            <span className="mr-2 text-emerald-400/40">+</span>{line}
          </div>
        ))}
      </div>
    </div>
  )
}
