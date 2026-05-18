import { describe, it, expect, afterEach } from "vitest"
import { render, cleanup } from "@testing-library/react"
import { Sparkline } from "./sparkline"

afterEach(() => cleanup())

describe("Sparkline", () => {
  it("renders nothing for empty points", () => {
    const { container } = render(<Sparkline points={[]} />)
    expect(container.querySelector("svg")).toBeNull()
  })

  it("renders SVG with polyline for normal points", () => {
    const { container } = render(<Sparkline points={[1, 3, 2, 5, 8]} />)
    const svg = container.querySelector("svg")
    expect(svg).not.toBeNull()
    expect(svg?.querySelector("polyline")).not.toBeNull()
  })

  it("uses emerald color for up trend", () => {
    const { container } = render(<Sparkline points={[1, 5]} trend="up" />)
    const polyline = container.querySelector("polyline")
    expect(polyline?.getAttribute("stroke")).toMatch(/emerald|#10b981/i)
  })

  it("uses rose color for down trend", () => {
    const { container } = render(<Sparkline points={[5, 1]} trend="down" />)
    const polyline = container.querySelector("polyline")
    expect(polyline?.getAttribute("stroke")).toMatch(/rose|#ef4444/i)
  })

  it("renders single dot when all points are equal (constant data)", () => {
    const { container } = render(<Sparkline points={[5, 5, 5, 5]} />)
    // Must render a <circle> dot
    expect(container.querySelector("circle")).not.toBeNull()
    // Must NOT render a stroke polyline / path
    expect(
      container.querySelector("polyline[stroke-width], path[stroke-width]"),
    ).toBeNull()
  })

  it("renders single dot for all-zero array", () => {
    const { container } = render(<Sparkline points={[0, 0, 0, 0]} />)
    expect(container.querySelector("circle")).not.toBeNull()
  })

  it("renders trendline (no circle-only) for non-constant data", () => {
    const { container } = render(<Sparkline points={[1, 3, 2, 5]} />)
    // Has the stroke line (polyline or path)
    expect(
      container.querySelector("polyline[stroke-width], path[stroke-width]"),
    ).not.toBeNull()
  })
})
