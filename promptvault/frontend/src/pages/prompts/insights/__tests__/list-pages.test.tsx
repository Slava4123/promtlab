import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactNode } from "react"
import TrendingPage from "../trending"
import DecliningPage from "../declining"
import MostEditedPage from "../most-edited"
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

const stub = (rows: Array<{ prompt_id: number; title: string; uses: number }>) => ({
  data: rows, isLoading: false, isError: false,
} as ReturnType<typeof hooks.useTrending>)

describe("list-style insight pages", () => {
  it("trending renders heading + items", () => {
    vi.mocked(hooks.useTrending).mockReturnValue(stub([{ prompt_id: 1, title: "Hot", uses: 20 }]))
    render(wrap(<TrendingPage />))
    expect(screen.getByRole("heading", { name: /растущие/i })).toBeInTheDocument()
    expect(screen.getByText("Hot")).toBeInTheDocument()
  })

  it("declining renders heading", () => {
    vi.mocked(hooks.useDeclining).mockReturnValue(stub([]))
    render(wrap(<DecliningPage />))
    expect(screen.getByRole("heading", { name: /падающие/i })).toBeInTheDocument()
  })

  it("most-edited renders heading", () => {
    vi.mocked(hooks.useMostEdited).mockReturnValue(stub([]))
    render(wrap(<MostEditedPage />))
    expect(screen.getByRole("heading", { name: /часто правят/i })).toBeInTheDocument()
  })
})
