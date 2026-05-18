import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactNode } from "react"
import DuplicatesPage from "./duplicates"
import * as hooks from "@/hooks/use-prompt-insights"

vi.mock("@/hooks/use-prompt-insights")

function wrap(node: ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

describe("DuplicatesPage", () => {
  beforeEach(() => {
    cleanup()
    vi.mocked(hooks.useMergePrompts).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof hooks.useMergePrompts>)
  })

  it("renders heading and pair list", () => {
    vi.mocked(hooks.useDuplicates).mockReturnValue({
      data: [{ prompt_a: { prompt_id: 1, title: "A", uses: 0 }, prompt_b: { prompt_id: 2, title: "B", uses: 0 }, similarity: 0.9 }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof hooks.useDuplicates>)
    render(wrap(<DuplicatesPage />))
    expect(screen.getByRole("heading", { name: /дубликат/i })).toBeInTheDocument()
    expect(screen.getByText(/A.*B/)).toBeInTheDocument()
  })

  it("opens merge modal when card clicked", async () => {
    vi.mocked(hooks.useDuplicates).mockReturnValue({
      data: [{ prompt_a: { prompt_id: 1, title: "A", uses: 0 }, prompt_b: { prompt_id: 2, title: "B", uses: 0 }, similarity: 0.9 }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof hooks.useDuplicates>)

    render(wrap(<DuplicatesPage />))
    fireEvent.click(screen.getByRole("button", { name: /объединить/i }))
    await waitFor(() => expect(screen.getByText(/похожесть 90%/i)).toBeInTheDocument())
  })

  it("renders empty state", () => {
    vi.mocked(hooks.useDuplicates).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof hooks.useDuplicates>)
    render(wrap(<DuplicatesPage />))
    expect(screen.getByText(/дубликатов не нашлось|нет дубликатов/i)).toBeInTheDocument()
  })
})
