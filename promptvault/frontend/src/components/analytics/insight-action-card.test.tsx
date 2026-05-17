import { describe, it, expect, afterEach } from "vitest"
import { render, screen, cleanup } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { AlertCircle, Copy, TrendingUp } from "lucide-react"
import { InsightActionCard } from "./insight-action-card"

afterEach(() => cleanup())

function renderWithRouter(ui: React.ReactNode) {
  return render(<MemoryRouter>{ui}</MemoryRouter>)
}

describe("InsightActionCard", () => {
  it("renders title, description, count, and CTA", () => {
    renderWithRouter(
      <InsightActionCard
        tone="warning"
        icon={AlertCircle}
        title="Забытые"
        description="5 промптов не использовались 30+ дней"
        href="/prompts?filter=unused"
        count={5}
        ctaLabel="Посмотреть"
      />,
    )
    expect(screen.getByText("Забытые")).toBeInTheDocument()
    expect(screen.getByText(/30\+ дней/)).toBeInTheDocument()
    expect(screen.getByText("5")).toBeInTheDocument()
    const link = screen.getByRole("link", { name: /Посмотреть/ })
    expect(link).toHaveAttribute("href", "/prompts?filter=unused")
  })

  it("applies warning tone (amber color)", () => {
    const { container } = renderWithRouter(
      <InsightActionCard tone="warning" icon={AlertCircle} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    expect(container.innerHTML).toMatch(/amber/)
  })

  it("applies info tone (violet color)", () => {
    const { container } = renderWithRouter(
      <InsightActionCard tone="info" icon={Copy} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    expect(container.innerHTML).toMatch(/violet/)
  })

  it("applies success tone (emerald color)", () => {
    const { container } = renderWithRouter(
      <InsightActionCard tone="success" icon={TrendingUp} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    expect(container.innerHTML).toMatch(/emerald/)
  })

  it("omits count badge when not provided", () => {
    const { container } = renderWithRouter(
      <InsightActionCard tone="warning" icon={AlertCircle} title="X" description="Y" href="#" ctaLabel="→" />,
    )
    // Count badge has tabular-nums class — check none rendered
    expect(container.querySelectorAll(".tabular-nums").length).toBe(0)
  })
})
