import { useEffect, useState } from "react"
import ReactDiffViewer, { DiffMethod } from "react-diff-viewer-continued"
import { Columns2, Rows2 } from "lucide-react"
import { useThemeStore } from "@/stores/theme-store"

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

  const { theme } = useThemeStore()
  const isDark = theme === "dark"
  const effectiveSplit = isMobile ? false : splitView

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-[0.75rem] sm:gap-3 sm:text-[0.8rem]">
          <span className="text-red-400/80">{oldTitle}</span>
          <span className="text-muted-foreground">vs</span>
          <span className="text-emerald-400/80">{newTitle}</span>
        </div>
        {!isMobile && (
          <button
            onClick={() => setSplitView(!splitView)}
            className="flex h-7 items-center gap-1.5 rounded-md border border-border bg-card px-2.5 text-[0.72rem] text-muted-foreground transition-colors hover:text-foreground"
          >
            {splitView ? <Rows2 className="h-3 w-3" /> : <Columns2 className="h-3 w-3" />}
            {splitView ? "Unified" : "Split"}
          </button>
        )}
      </div>
      <div className="overflow-auto rounded-lg border border-border">
        <ReactDiffViewer
          oldValue={oldValue}
          newValue={newValue}
          splitView={effectiveSplit}
          compareMethod={DiffMethod.WORDS}
          useDarkTheme={isDark}
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
              light: {
                diffViewerBackground: "#ffffff",
                addedBackground: "rgba(34,197,94,0.08)",
                removedBackground: "rgba(239,68,68,0.08)",
                wordAddedBackground: "rgba(34,197,94,0.2)",
                wordRemovedBackground: "rgba(239,68,68,0.2)",
                addedGutterBackground: "rgba(34,197,94,0.06)",
                removedGutterBackground: "rgba(239,68,68,0.06)",
                gutterBackground: "#f9fafb",
                gutterBackgroundDark: "#f3f4f6",
                codeFoldBackground: "#f9fafb",
                codeFoldGutterBackground: "#f3f4f6",
                emptyLineBackground: "#ffffff",
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
