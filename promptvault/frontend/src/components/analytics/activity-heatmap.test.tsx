import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ActivityHeatmap } from "./activity-heatmap"

afterEach(() => cleanup())

describe("ActivityHeatmap padding", () => {
  it("renders exactly 28 cells when data is empty", () => {
    const { container } = render(<ActivityHeatmap points={[]} />)
    expect(container.querySelectorAll("[data-cell]")).toHaveLength(28)
  })

  it("renders exactly 28 cells when partial data", () => {
    const points = Array.from({ length: 5 }, (_, i) => ({
      day: `2026-05-${String(12 + i).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    expect(container.querySelectorAll("[data-cell]")).toHaveLength(28)
  })

  it("renders 28 cells for 4 weeks of data", () => {
    const points = Array.from({ length: 28 }, (_, i) => ({
      day: `2026-05-${String(i + 1).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    expect(cells.length).toBe(28)
  })

  it("varies opacity by count", () => {
    // Use yesterday and day-before-yesterday so they fall in 28-day window.
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
    const cells = container.querySelectorAll("[data-cell]")
    const opacityAt = (key: string) => {
      const el = container.querySelector(`[data-cell][data-day="${key}"]`) as HTMLElement | null
      return el ? parseFloat(el.style.opacity || "1") : NaN
    }
    expect(opacityAt(iso(dayBefore))).toBeLessThan(opacityAt(iso(yesterday)))
    // sanity: all 28 cells present
    expect(cells.length).toBe(28)
  })

  it("renders russian month in tooltip aria-label", () => {
    // Use yesterday — guaranteed within 28-day window.
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const iso = yesterday.toISOString().slice(0, 10)
    const monthShort = new Intl.DateTimeFormat("ru-RU", {
      day: "numeric",
      month: "short",
    }).format(yesterday)
    const points = [{ day: iso, count: 3 }]
    render(<ActivityHeatmap points={points} />)
    // Escape any special regex chars from the formatted month (e.g., the dot in "1 дек.").
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
})
