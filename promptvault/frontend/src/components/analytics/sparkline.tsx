interface SparklineProps {
  points: number[]
  trend?: "up" | "down" | "neutral"
  width?: number
  height?: number
}

// Sparkline — мини-график для KPI-карточки. Чистый SVG, без Recharts overhead.
// trend влияет на цвет: up → emerald, down → rose, neutral → slate.
// fill — лёгкий gradient под линией (соответствует цвету stroke на 15%).
export function Sparkline({
  points,
  trend = "neutral",
  width = 120,
  height = 22,
}: SparklineProps) {
  if (points.length === 0) return null

  const stroke =
    trend === "up" ? "#10b981" : trend === "down" ? "#ef4444" : "#94a3b8"
  const fill =
    trend === "up"
      ? "rgba(16,185,129,0.15)"
      : trend === "down"
        ? "rgba(239,68,68,0.15)"
        : "rgba(148,163,184,0.15)"

  // Если все точки равны — плоская линия выглядит как бессмысленное тире.
  // Рисуем одиночную точку на правом краю как индикатор «нет тренда».
  const rawMax = Math.max(...points)
  const rawMin = Math.min(...points)
  if (rawMax === rawMin) {
    return (
      <svg
        width={width}
        height={height}
        viewBox={`0 0 ${width} ${height}`}
        aria-label="нет тренда"
      >
        <circle cx={width - 4} cy={height / 2} r={2.5} fill={stroke} />
      </svg>
    )
  }

  const max = Math.max(...points, 1)
  const min = Math.min(...points, 0)
  const range = max - min || 1
  const stepX = points.length > 1 ? width / (points.length - 1) : width

  const coords = points.map((p, i) => {
    const x = i * stepX
    const y = height - ((p - min) / range) * (height - 2) - 1
    return `${x},${y}`
  })

  const polyPoints = coords.join(" ")
  // Area-path под линией: M start → L каждой точки → V height → H 0 → Z.
  // Используем <path>, чтобы единственный <polyline> в SVG был линией —
  // тесты `container.querySelector("polyline")` находят строго stroke-элемент.
  const areaPath =
    `M 0,${height} ` +
    coords.map((c) => `L ${c}`).join(" ") +
    ` L ${width},${height} Z`

  return (
    <svg width={width} height={height} aria-hidden="true">
      <path d={areaPath} fill={fill} stroke="none" />
      <polyline points={polyPoints} fill="none" stroke={stroke} strokeWidth="1.5" />
    </svg>
  )
}
