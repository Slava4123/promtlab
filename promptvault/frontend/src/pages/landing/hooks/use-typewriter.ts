import { useState, useEffect, useRef, useCallback } from "react"

interface TypewriterOptions {
  speed?: number
  startDelay?: number
  enabled?: boolean
}

export function useTypewriter(
  text: string,
  opts: TypewriterOptions = {},
) {
  const { speed = 30, startDelay = 0, enabled = true } = opts
  const [displayText, setDisplayText] = useState("")
  const [isComplete, setIsComplete] = useState(false)
  const indexRef = useRef(0)
  const prevTextRef = useRef(text)
  const prevEnabledRef = useRef(enabled)

  const reset = useCallback(() => {
    indexRef.current = 0
    setDisplayText("")
    setIsComplete(false)
  }, [])

  /* Reset when text or enabled changes */
  if (text !== prevTextRef.current || enabled !== prevEnabledRef.current) {
    prevTextRef.current = text
    prevEnabledRef.current = enabled
    indexRef.current = 0
    // We don't call setState here — the effect below will handle it
  }

  useEffect(() => {
    if (!enabled) {
      reset()
      return
    }

    /* Respect prefers-reduced-motion */
    const prefersReduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (prefersReduced) {
      setDisplayText(text)
      setIsComplete(true)
      return
    }

    indexRef.current = 0
    let timer: ReturnType<typeof setTimeout>
    let rafId: number
    let cancelled = false

    const startTyping = () => {
      let last = performance.now()
      const step = (now: number) => {
        if (cancelled) return
        if (now - last >= speed) {
          last = now
          indexRef.current++
          setDisplayText(text.slice(0, indexRef.current))
          if (indexRef.current >= text.length) {
            setIsComplete(true)
            return
          }
        }
        rafId = requestAnimationFrame(step)
      }
      rafId = requestAnimationFrame(step)
    }

    if (startDelay > 0) {
      timer = setTimeout(startTyping, startDelay)
    } else {
      startTyping()
    }

    return () => {
      cancelled = true
      clearTimeout(timer)
      cancelAnimationFrame(rafId)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [text, speed, startDelay, enabled])

  return { displayText, isComplete, cursor: enabled && !isComplete }
}
