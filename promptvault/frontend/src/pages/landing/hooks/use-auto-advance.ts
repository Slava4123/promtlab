import { useState, useEffect, useCallback, useRef } from "react"

interface AutoAdvanceOptions {
  count: number
  interval?: number
}

export function useAutoAdvance({ count, interval = 6000 }: AutoAdvanceOptions) {
  const [activeIndex, setActiveIndex] = useState(0)
  const [progress, setProgress] = useState(0)
  const pausedRef = useRef(false)
  const startRef = useRef(0)

  const pause = useCallback(() => { pausedRef.current = true }, [])
  const resume = useCallback(() => {
    pausedRef.current = false
    startRef.current = performance.now()
    setProgress(0)
  }, [])

  const goTo = useCallback((index: number) => {
    setActiveIndex(index)
    startRef.current = performance.now()
    setProgress(0)
  }, [])

  useEffect(() => {
    startRef.current = performance.now()
    let rafId: number

    const tick = (now: number) => {
      if (!pausedRef.current) {
        const elapsed = now - startRef.current
        const p = Math.min(elapsed / interval, 1)
        setProgress(p)

        if (p >= 1) {
          setActiveIndex(prev => (prev + 1) % count)
          startRef.current = now
          setProgress(0)
        }
      }
      rafId = requestAnimationFrame(tick)
    }

    rafId = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(rafId)
  }, [count, interval])

  return { activeIndex, progress, goTo, pause, resume }
}
