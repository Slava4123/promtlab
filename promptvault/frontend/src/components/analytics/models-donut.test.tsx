import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { ModelsDonut } from "./models-donut"
import type { ModelUsageRow } from "@/api/analytics"

afterEach(() => cleanup())

describe("ModelsDonut", () => {
  it("renders empty state for no data", () => {
    render(<ModelsDonut data={[]} />)
    expect(screen.getByText(/нет данных/i)).toBeInTheDocument()
  })

  it("renders legend with percentages", () => {
    const data: ModelUsageRow[] = [
      { model: "claude-3-opus", uses: 60 },
      { model: "gpt-4", uses: 30 },
      { model: "gemini", uses: 10 },
    ]
    render(<ModelsDonut data={data} />)
    expect(screen.getByText(/60%/)).toBeInTheDocument()
    expect(screen.getByText(/30%/)).toBeInTheDocument()
    expect(screen.getByText(/10%/)).toBeInTheDocument()
  })

  it("collapses tail beyond top-6 into «Другие»", () => {
    const data: ModelUsageRow[] = Array.from({ length: 8 }, (_, i) => ({
      model: `model-${i}`,
      uses: 10,
    }))
    render(<ModelsDonut data={data} />)
    expect(screen.getByText(/Другие/)).toBeInTheDocument()
  })
})
