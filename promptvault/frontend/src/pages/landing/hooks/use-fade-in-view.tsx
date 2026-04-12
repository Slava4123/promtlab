import { useRef, useState, useEffect, type CSSProperties, type RefObject } from "react"

type Direction = "up" | "down" | "left" | "right"

export interface FadeInViewOptions {
  direction?: Direction
  delay?: number
  duration?: number
  threshold?: number
  distance?: number
}

const getTransform = (dir: Direction, distance: number) => {
  switch (dir) {
    case "up": return `translateY(${distance}px)`
    case "down": return `translateY(-${distance}px)`
    case "left": return `translateX(${distance}px)`
    case "right": return `translateX(-${distance}px)`
  }
}

export function useFadeInView(
  opts: FadeInViewOptions = {},
): [RefObject<HTMLDivElement | null>, CSSProperties] {
  const { direction = "up", delay = 0, duration = 700, threshold = 0.15, distance = 24 } = opts
  const ref = useRef<HTMLDivElement | null>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true)
          obs.disconnect()
        }
      },
      { threshold },
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [threshold])

  const style: CSSProperties = {
    opacity: visible ? 1 : 0,
    transform: visible ? "translate(0, 0)" : getTransform(direction, distance),
    transition: `opacity ${duration}ms cubic-bezier(0.22, 1, 0.36, 1), transform ${duration}ms cubic-bezier(0.22, 1, 0.36, 1)`,
    transitionDelay: `${delay}ms`,
  }

  return [ref, style]
}
