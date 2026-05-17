import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { StreakTracker } from "./streak-tracker"

afterEach(() => cleanup())

describe("StreakTracker", () => {
  it("renders current and longest streak", () => {
    render(<StreakTracker current={5} longest={12} activeToday={true} />)
    expect(screen.getByText("5")).toBeInTheDocument()
    expect(screen.getByText(/best 12/i)).toBeInTheDocument()
  })

  it("renders 7 dots for last 7 days", () => {
    const { container } = render(<StreakTracker current={3} longest={10} activeToday={false} />)
    const dots = container.querySelectorAll("[data-streak-dot]")
    expect(dots.length).toBe(7)
  })

  it("highlights filled dots up to current streak", () => {
    const { container } = render(<StreakTracker current={4} longest={10} activeToday={true} />)
    const filled = container.querySelectorAll("[data-streak-dot][data-filled='true']")
    expect(filled.length).toBe(4)
  })
})
