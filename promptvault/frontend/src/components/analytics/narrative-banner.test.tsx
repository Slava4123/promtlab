import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { NarrativeBanner } from "./narrative-banner"
import type { NarrativeSegments } from "@/lib/analytics-narrative"

afterEach(() => cleanup())

describe("NarrativeBanner", () => {
  it("renders summary text", () => {
    const segments: NarrativeSegments = {
      summary: "За 7 дней: 234 использований ↑23%",
      topModel: null,
      streak: null,
      actionHint: null,
    }
    render(<NarrativeBanner segments={segments} />)
    expect(screen.getByText(/234 использований/)).toBeInTheDocument()
  })

  it("renders all 4 segments when provided", () => {
    const segments: NarrativeSegments = {
      summary: "За 7 дней: 234 использований ↑23%",
      topModel: "топ-модель Claude (62%)",
      streak: "streak 5 дней",
      actionHint: "5 забытых ждут уборки",
    }
    render(<NarrativeBanner segments={segments} />)
    expect(screen.getByText(/Claude/)).toBeInTheDocument()
    expect(screen.getByText(/streak 5 дней/)).toBeInTheDocument()
    expect(screen.getByText(/5 забытых/)).toBeInTheDocument()
  })

  it("omits null segments gracefully", () => {
    const segments: NarrativeSegments = {
      summary: "За 7 дней пока тихо",
      topModel: null,
      streak: null,
      actionHint: null,
    }
    render(<NarrativeBanner segments={segments} />)
    expect(screen.getByText(/тихо/)).toBeInTheDocument()
    expect(screen.queryByText(/streak/)).toBeNull()
  })
})
