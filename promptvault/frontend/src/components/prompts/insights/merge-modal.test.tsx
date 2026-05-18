import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent, cleanup } from "@testing-library/react"
import { MergeModal } from "./merge-modal"

const pair = {
  prompt_a: { prompt_id: 1, title: "Refactor v1", uses: 5 },
  prompt_b: { prompt_id: 2, title: "Refactor v2", uses: 10 },
  similarity: 0.91,
}

describe("MergeModal", () => {
  it("renders both prompts side-by-side when open", () => {
    cleanup()
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={() => {}} />)
    expect(screen.getByText("Refactor v1")).toBeInTheDocument()
    expect(screen.getByText("Refactor v2")).toBeInTheDocument()
  })

  it("calls onMerge with correct ids when user picks A", () => {
    cleanup()
    const onMerge = vi.fn()
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={onMerge} />)
    fireEvent.click(screen.getByRole("button", { name: /оставить «refactor v1»/i }))
    expect(onMerge).toHaveBeenCalledWith({ keepID: 1, mergeID: 2 })
  })

  it("calls onMerge with reversed ids when user picks B", () => {
    cleanup()
    const onMerge = vi.fn()
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={onMerge} />)
    fireEvent.click(screen.getByRole("button", { name: /оставить «refactor v2»/i }))
    expect(onMerge).toHaveBeenCalledWith({ keepID: 2, mergeID: 1 })
  })

  it("shows warning about lost metadata", () => {
    cleanup()
    render(<MergeModal pair={pair} open onClose={() => {}} onMerge={() => {}} />)
    expect(screen.getByText(/теги.*коллекции.*не переносятся/i)).toBeInTheDocument()
  })
})
