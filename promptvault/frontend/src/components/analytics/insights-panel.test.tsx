import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import type React from "react"
import { InsightsPanel } from "./insights-panel"
import type { Insight } from "@/api/analytics"

function wrap(node: React.ReactNode) {
  return render(<MemoryRouter>{node}</MemoryRouter>)
}

describe("insights-panel hrefs", () => {
  it("unused_prompts links to /prompts/insights/unused", () => {
    const insights: Insight[] = [
      { type: "unused_prompts", payload: [{ id: 1 }, { id: 2 }] } as unknown as Insight,
    ]
    wrap(<InsightsPanel insights={insights} />)
    const link = screen.getByRole("link", { name: /посмотреть/i })
    expect(link).toHaveAttribute("href", "/prompts/insights/unused")
  })

  it("possible_duplicates links to /prompts/insights/duplicates", () => {
    const insights: Insight[] = [
      { type: "possible_duplicates", payload: [{}] } as unknown as Insight,
    ]
    wrap(<InsightsPanel insights={insights} />)
    expect(screen.getByRole("link", { name: /объединить/i })).toHaveAttribute(
      "href",
      "/prompts/insights/duplicates",
    )
  })

  it("orphan_tags has russian title and tags?filter=orphan link", () => {
    const insights: Insight[] = [
      { type: "orphan_tags", payload: [{}, {}] } as unknown as Insight,
    ]
    wrap(<InsightsPanel insights={insights} />)
    expect(screen.getByText(/теги без промптов/i)).toBeInTheDocument()
    // "Orphan" не должен появляться нигде в текстовом контексте
    expect(screen.queryByText(/orphan/i)).not.toBeInTheDocument()
    expect(screen.getByRole("link", { name: /очистить/i })).toHaveAttribute(
      "href",
      "/tags?filter=orphan",
    )
  })

  it("empty_collections links to /collections?filter=empty", () => {
    const insights: Insight[] = [
      { type: "empty_collections", payload: [{}] } as unknown as Insight,
    ]
    wrap(<InsightsPanel insights={insights} />)
    expect(screen.getByRole("link", { name: /очистить/i })).toHaveAttribute(
      "href",
      "/collections?filter=empty",
    )
  })

  it("trending/declining/most_edited link to dedicated routes", () => {
    const cases: Array<[Insight["type"], string]> = [
      ["trending", "/prompts/insights/trending"],
      ["declining", "/prompts/insights/declining"],
      ["most_edited", "/prompts/insights/most-edited"],
    ]
    for (const [t, href] of cases) {
      const insights: Insight[] = [{ type: t, payload: [{}] } as unknown as Insight]
      const { container, unmount } = render(
        <MemoryRouter>
          <InsightsPanel insights={insights} />
        </MemoryRouter>,
      )
      expect(container.querySelector(`a[href="${href}"]`)).not.toBeNull()
      unmount()
    }
  })
})
