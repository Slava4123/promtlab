import { useFadeInView, type FadeInViewOptions } from "../hooks/use-fade-in-view"

export function FadeIn({
  children,
  className,
  delay = 0,
  direction = "up",
  duration,
  distance,
}: {
  children: React.ReactNode
  className?: string
  delay?: number
  direction?: FadeInViewOptions["direction"]
  duration?: number
  distance?: number
}) {
  const [ref, style] = useFadeInView({ direction, delay, duration, distance })
  return (
    <div ref={ref} className={className} style={style}>
      {children}
    </div>
  )
}
