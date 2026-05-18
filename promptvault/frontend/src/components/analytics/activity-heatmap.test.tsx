import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ActivityHeatmap } from "./activity-heatmap"

afterEach(() => cleanup())

// 52-week GitHub-style heatmap: 53 колонки × 7 строк, последняя колонка
// частично заполнена в зависимости от weekday сегодня. Точное число cells:
// 52 * 7 + todayWeekday + 1 ∈ [365, 371].
const MIN_CELLS = 365
const MAX_CELLS = 371

describe("ActivityHeatmap", () => {
  it("renders ~365 cells when data is empty (52 weeks coverage)", () => {
    const { container } = render(<ActivityHeatmap points={[]} />)
    const cells = container.querySelectorAll("[data-cell]")
    expect(cells.length).toBeGreaterThanOrEqual(MIN_CELLS)
    expect(cells.length).toBeLessThanOrEqual(MAX_CELLS)
  })

  it("renders ~365 cells with partial data", () => {
    const points = Array.from({ length: 5 }, (_, i) => ({
      day: `2026-05-${String(12 + i).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    expect(cells.length).toBeGreaterThanOrEqual(MIN_CELLS)
    expect(cells.length).toBeLessThanOrEqual(MAX_CELLS)
  })

  it("renders all 365+ cells regardless of data shape", () => {
    const points = Array.from({ length: 30 }, (_, i) => ({
      day: `2026-05-${String(i + 1).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    expect(cells.length).toBeGreaterThanOrEqual(MIN_CELLS)
  })

  it("assigns higher tier to higher count days", () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const dayBefore = new Date()
    dayBefore.setDate(dayBefore.getDate() - 2)
    const iso = (d: Date) => d.toISOString().slice(0, 10)
    const points = [
      { day: iso(dayBefore), count: 0 },
      { day: iso(yesterday), count: 100 },
    ]
    const { container } = render(<ActivityHeatmap points={points} />)
    const tierAt = (key: string) => {
      const el = container.querySelector(`[data-cell][data-day="${key}"]`) as HTMLElement | null
      return el ? parseInt(el.dataset.tier ?? "0", 10) : -1
    }
    expect(tierAt(iso(dayBefore))).toBe(0)
    expect(tierAt(iso(yesterday))).toBe(4) // max → tier 4
  })

  it("renders russian month in tooltip aria-label", () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const iso = yesterday.toISOString().slice(0, 10)
    const monthShort = new Intl.DateTimeFormat("ru-RU", {
      day: "numeric",
      month: "short",
    }).format(yesterday)
    const points = [{ day: iso, count: 3 }]
    render(<ActivityHeatmap points={points} />)
    const escaped = monthShort.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
    const matches = screen.getAllByLabelText(new RegExp(escaped, "i"))
    expect(matches.length).toBeGreaterThan(0)
  })

  it("uses russian plural form for count", () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const iso = yesterday.toISOString().slice(0, 10)
    const points = [{ day: iso, count: 1 }]
    render(<ActivityHeatmap points={points} />)
    expect(screen.getAllByLabelText(/1 использование(?!\w)/i).length).toBeGreaterThan(0)
  })

  it("displays total uses count in header", () => {
    const points = [
      { day: "2026-05-10", count: 3 },
      { day: "2026-05-11", count: 7 },
    ]
    render(<ActivityHeatmap points={points} />)
    expect(screen.getByText(/10 использований за год/i)).toBeInTheDocument()
  })

  it("renders weekday and legend labels in russian", () => {
    render(<ActivityHeatmap points={[]} />)
    expect(screen.getByText("Пн")).toBeInTheDocument()
    expect(screen.getByText("Ср")).toBeInTheDocument()
    expect(screen.getByText("Пт")).toBeInTheDocument()
    expect(screen.getByText("Меньше")).toBeInTheDocument()
    expect(screen.getByText("Больше")).toBeInTheDocument()
  })
})
