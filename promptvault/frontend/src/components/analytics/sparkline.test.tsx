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
})
