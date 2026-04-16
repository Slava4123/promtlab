import { Suspense, lazy, useEffect, useState, createElement, type ComponentType } from "react"
import { Columns2, Rows2, Loader2 } from "lucide-react"
import { useThemeStore } from "@/stores/theme-store"

// react-diff-viewer-continued ≈ 100-150kB gzipped. Ленивая загрузка
// вытаскивает его из main bundle — нужен только на /prompts/:id/versions (P-6).
// Передаём compareMethod как опциональный параметр в виде строкового enum
// из загруженного модуля — так избегаем статического импорта типов.
type DiffViewerProps = {
  oldValue: string
  newValue: string
  splitView?: boolean
  useDarkTheme?: boolean
  compareMethod?: unknown
  styles?: Record<string, unknown>
}

const LazyReactDiffViewer = lazy(async () => {
  const mod = await import("react-diff-viewer-continued")
  const Original = mod.default as unknown as ComponentType<Record<string, unknown>>
  // Инжектим compareMethod=WORDS из enum'а библиотеки, убирая зависимость
  // main bundle от типа DiffMethod.
  const method = mod.DiffMethod.WORDS
  const Wrapped: ComponentType<DiffViewerProps> = (props) =>
    createElement(Original, {
      ...(props as Record<string, unknown>),
      compareMethod: props.compareMethod ?? method,
    })
  return { default: Wrapped }
})

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

const diffViewerStyles = {
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
  diffContainer: { minWidth: "unset" },
  contentText: {
    fontSize: "0.8rem",
    lineHeight: "1.6",
    fontFamily: "var(--font-geist-mono, monospace)",
    whiteSpace: "pre-wrap" as const,
    wordBreak: "break-word" as const,
    overflowWrap: "break-word" as const,
  },
  content: { overflow: "hidden" as const },
  gutter: { minWidth: "2.5rem", fontSize: "0.7rem" },
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
            type="button"
            onClick={() => setSplitView(!splitView)}
            className="flex h-7 items-center gap-1.5 rounded-md border border-border bg-card px-2.5 text-[0.72rem] text-muted-foreground transition-colors hover:text-foreground"
          >
            {splitView ? <Rows2 className="h-3 w-3" /> : <Columns2 className="h-3 w-3" />}
            {splitView ? "Unified" : "Split"}
          </button>
        )}
      </div>
      <div className="overflow-auto rounded-lg border border-border">
        <Suspense
          fallback={
            <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
              <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Загрузка diff-viewer…
            </div>
          }
        >
          <LazyReactDiffViewer
            oldValue={oldValue}
            newValue={newValue}
            splitView={effectiveSplit}
            useDarkTheme={isDark}
            styles={diffViewerStyles}
          />
        </Suspense>
      </div>
    </div>
  )
}
