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

// P-12: Shared IntersectionObserver.
// На лендинге ~20 секций используют useFadeInView — каждый создавал свой
// IntersectionObserver. Один разделяемый observer экономит память и CPU
// (browser-side callback батч-обрабатывает все элементы).
//
// threshold запекаем в ключ observer'а — разные пороги требуют отдельного
// observer'а по спецификации IntersectionObserverInit. На практике почти
// всё зовётся с threshold=0.15, так что обычно будет один observer на весь lifetime.
type Callback = (entry: IntersectionObserverEntry) => void
const observers = new Map<number, IntersectionObserver>()
const callbacks = new WeakMap<Element, Callback>()

function getObserver(threshold: number): IntersectionObserver {
  // Дискретизируем threshold до 2 знаков — избегаем миллионов observer'ов
  // при случайных float-дрейфах, оставаясь достаточно точными для анимаций.
  const key = Math.round(threshold * 100) / 100
  let obs = observers.get(key)
  if (obs) return obs
  obs = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const cb = callbacks.get(entry.target)
        if (cb) cb(entry)
      }
    },
    { threshold: key },
  )
  observers.set(key, obs)
  return obs
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
    const obs = getObserver(threshold)
    const cb: Callback = (entry) => {
      if (entry.isIntersecting) {
        setVisible(true)
        callbacks.delete(el)
        obs.unobserve(el)
      }
    }
    callbacks.set(el, cb)
    obs.observe(el)
    return () => {
      callbacks.delete(el)
      obs.unobserve(el)
    }
  }, [threshold])

  const style: CSSProperties = {
    opacity: visible ? 1 : 0,
    transform: visible ? "translate(0, 0)" : getTransform(direction, distance),
    transition: `opacity ${duration}ms cubic-bezier(0.22, 1, 0.36, 1), transform ${duration}ms cubic-bezier(0.22, 1, 0.36, 1)`,
    transitionDelay: `${delay}ms`,
  }

  return [ref, style]
}
