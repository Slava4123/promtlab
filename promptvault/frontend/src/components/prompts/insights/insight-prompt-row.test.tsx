import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { InsightPromptRow } from "./insight-prompt-row"

function r(node: React.ReactNode) {
  return render(<MemoryRouter>{node}</MemoryRouter>)
}

describe("InsightPromptRow", () => {
  it("renders title and uses label", () => {
    r(<InsightPromptRow promptID={1} title="Refactor X" uses={12} />)
    expect(screen.getByText("Refactor X")).toBeInTheDocument()
    expect(screen.getByText(/12/)).toBeInTheDocument()
  })

  it("renders link to prompt editor", () => {
    r(<InsightPromptRow promptID={42} title="X" uses={0} />)
    const link = screen.getByRole("link", { name: /X/ })
    expect(link).toHaveAttribute("href", "/prompts/42")
  })

  it("renders action slot", () => {
    r(<InsightPromptRow promptID={1} title="X" uses={0} actions={<button>Удалить</button>} />)
    expect(screen.getByRole("button", { name: "Удалить" })).toBeInTheDocument()
  })

  it("hides uses when showUses=false", () => {
    r(<InsightPromptRow promptID={1} title="X" uses={5} showUses={false} />)
    expect(screen.queryByText(/использований|использование/i)).not.toBeInTheDocument()
  })

  it("uses correct pluralization (1 → использование, 2-4 → использования, 5+ → использований)", () => {
    const { rerender } = r(<InsightPromptRow promptID={1} title="T" uses={1} />)
    expect(screen.getByText("1 использование")).toBeInTheDocument()
    rerender(<MemoryRouter><InsightPromptRow promptID={1} title="T" uses={3} /></MemoryRouter>)
    expect(screen.getByText("3 использования")).toBeInTheDocument()
    rerender(<MemoryRouter><InsightPromptRow promptID={1} title="T" uses={5} /></MemoryRouter>)
    expect(screen.getByText("5 использований")).toBeInTheDocument()
  })
})
