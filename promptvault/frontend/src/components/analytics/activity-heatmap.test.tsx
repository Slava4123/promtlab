import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ActivityHeatmap } from "./activity-heatmap"

afterEach(() => cleanup())

describe("ActivityHeatmap", () => {
  it("renders 28 cells for 4 weeks of data", () => {
    const points = Array.from({ length: 28 }, (_, i) => ({
      day: `2026-05-${String(i + 1).padStart(2, "0")}`,
      count: i,
    }))
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    expect(cells.length).toBe(28)
  })

  it("renders empty state for no data", () => {
    render(<ActivityHeatmap points={[]} />)
    expect(screen.getByText(/нет активности/i)).toBeInTheDocument()
  })

  it("varies opacity by count", () => {
    const points = [
      { day: "2026-05-01", count: 0 },
      { day: "2026-05-02", count: 100 },
    ]
    const { container } = render(<ActivityHeatmap points={points} />)
    const cells = container.querySelectorAll("[data-cell]")
    const opacities = Array.from(cells).map((c) => parseFloat((c as HTMLElement).style.opacity || "1"))
    expect(opacities[0]).toBeLessThan(opacities[1])
  })
})
