import { useEffect, useState } from "react"
import ReactDiffViewer, { DiffMethod } from "react-diff-viewer-continued"
import { Columns2, Rows2 } from "lucide-react"

interface VersionDiffProps {
  oldValue: string
  newValue: string
  oldTitle: string
  newTitle: string
}

function useIsMobile(breakpoint = 768) {
  const [isMobile, setIsMobile] = useState(() => window.innerWidth < breakpoint)
  useEffect(() => {
    const mq = window.matchMedia(`(max-width: ${breakpoint - 1}px)`)
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches)
    mq.addEventListener("change", handler)
    return () => mq.removeEventListener("change", handler)
  }, [breakpoint])
  return isMobile
}

export function VersionDiff({ oldValue, newValue, oldTitle, newTitle }: VersionDiffProps) {
  const isMobile = useIsMobile()
  const [splitView, setSplitView] = useState(!isMobile)

  const effectiveSplit = isMobile ? false : splitView

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-[0.75rem] sm:gap-3 sm:text-[0.8rem]">
          <span className="text-red-400/80">{oldTitle}</span>
          <span className="text-zinc-600">vs</span>
          <span className="text-emerald-400/80">{newTitle}</span>
        </div>
        {!isMobile && (
          <button
            onClick={() => setSplitView(!splitView)}
            className="flex h-7 items-center gap-1.5 rounded-md px-2.5 text-[0.72rem] text-zinc-500 transition-colors hover:text-zinc-300"
            style={{ border: "1px solid rgba(255,255,255,0.06)", background: "rgba(255,255,255,0.02)" }}
          >
            {splitView ? <Rows2 className="h-3 w-3" /> : <Columns2 className="h-3 w-3" />}
            {splitView ? "Unified" : "Split"}
          </button>
        )}
      </div>
      <div className="overflow-auto rounded-lg" style={{ border: "1px solid rgba(255,255,255,0.06)" }}>
        <ReactDiffViewer
          oldValue={oldValue}
          newValue={newValue}
          splitView={effectiveSplit}
          compareMethod={DiffMethod.WORDS}
          useDarkTheme
          styles={{
            variables: {
              dark: {
                diffViewerBackground: "#0d0d10",
                addedBackground: "rgba(34,197,94,0.08)",
                removedBackground: "rgba(239,68,68,0.08)",
                wordAddedBackground: "rgba(34,197,94,0.2)",
                wordRemovedBackground: "rgba(239,68,68,0.2)",
                addedGutterBackground: "rgba(34,197,94,0.12)",
                removedGutterBackground: "rgba(239,68,68,0.12)",
                gutterBackground: "#0d0d10",
                gutterBackgroundDark: "#0a0a0d",
                codeFoldBackground: "#101015",
                codeFoldGutterBackground: "#101015",
                emptyLineBackground: "#0d0d10",
                codeFoldContentColor: "#71717a",
              },
            },
            diffContainer: {
              minWidth: "unset",
            },
            contentText: {
              fontSize: "0.8rem",
              lineHeight: "1.6",
              fontFamily: "var(--font-geist-mono, monospace)",
              whiteSpace: "pre-wrap",
              wordBreak: "break-word",
              overflowWrap: "break-word",
            },
            content: {
              overflow: "hidden",
            },
            gutter: {
              minWidth: "2.5rem",
              fontSize: "0.7rem",
            },
          }}
        />
      </div>
    </div>
  )
}
