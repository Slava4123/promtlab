import { render, screen } from "@testing-library/react"
import { HighlightMatch } from "./highlight"

describe("HighlightMatch", () => {
  it("highlights a single match", () => {
    render(<HighlightMatch text="Hello review world" query="review" />)
    const mark = screen.getByText("review")
    expect(mark.tagName).toBe("MARK")
  })

  it("is case-insensitive", () => {
    render(<HighlightMatch text="Code Review Guide" query="review" />)
    const mark = screen.getByText("Review")
    expect(mark.tagName).toBe("MARK")
  })

  it("highlights multiple occurrences", () => {
    const { container } = render(
      <HighlightMatch text="review my review please" query="review" />,
    )
    const marks = container.querySelectorAll("mark")
    expect(marks).toHaveLength(2)
    expect(marks[0].textContent).toBe("review")
    expect(marks[1].textContent).toBe("review")
  })

  it("escapes regex special characters in query", () => {
    render(<HighlightMatch text="test (value) here" query="(value)" />)
    const mark = screen.getByText("(value)")
    expect(mark.tagName).toBe("MARK")
  })

  it("returns text as-is when query is empty", () => {
    const { container } = render(<HighlightMatch text="Hello world" query="" />)
    expect(container.textContent).toBe("Hello world")
    expect(container.querySelector("mark")).toBeNull()
  })

  it("returns text as-is when there is no match", () => {
    const { container } = render(
      <HighlightMatch text="Hello world" query="xyz" />,
    )
    expect(container.textContent).toBe("Hello world")
    expect(container.querySelector("mark")).toBeNull()
  })

  it("returns null for empty text", () => {
    const { container } = render(<HighlightMatch text="" query="test" />)
    expect(container.textContent).toBe("")
  })

  it("handles query with only whitespace", () => {
    const { container } = render(<HighlightMatch text="Hello world" query="   " />)
    expect(container.textContent).toBe("Hello world")
    expect(container.querySelector("mark")).toBeNull()
  })

  it("handles cyrillic text and query", () => {
    render(<HighlightMatch text="Промпт для код ревью" query="код" />)
    const mark = screen.getByText("код")
    expect(mark.tagName).toBe("MARK")
  })

  it("applies mark styling class", () => {
    render(<HighlightMatch text="test match here" query="match" />)
    const mark = screen.getByText("match")
    expect(mark.className).toContain("bg-brand-muted")
  })
})
