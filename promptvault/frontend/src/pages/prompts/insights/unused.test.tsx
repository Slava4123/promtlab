import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactNode } from "react"
import UnusedInsightsPage from "./unused"
import * as insightsHooks from "@/hooks/use-prompt-insights"
import * as promptsHooks from "@/hooks/use-prompts"

vi.mock("@/hooks/use-prompt-insights")
vi.mock("@/hooks/use-prompts")

function wrap(node: ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <MemoryRouter>
      <QueryClientProvider client={qc}>{node}</QueryClientProvider>
    </MemoryRouter>
  )
}

const stubDelete = () => ({
  mutate: vi.fn(),
  isPending: false,
} as unknown as ReturnType<typeof promptsHooks.useDeletePrompt>)

describe("UnusedInsightsPage", () => {
  beforeEach(() => {
    vi.mocked(promptsHooks.useDeletePrompt).mockReturnValue(stubDelete())
  })

  it("renders heading and items", () => {
    vi.mocked(insightsHooks.useUnusedPrompts).mockReturnValue({
      data: [{ prompt_id: 1, title: "Old", uses: 0 }],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof insightsHooks.useUnusedPrompts>)
    render(wrap(<UnusedInsightsPage />))
    expect(screen.getByRole("heading", { name: /забытые промпты/i })).toBeInTheDocument()
    expect(screen.getByText("Old")).toBeInTheDocument()
  })

  it("renders empty state", () => {
    vi.mocked(insightsHooks.useUnusedPrompts).mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof insightsHooks.useUnusedPrompts>)
    render(wrap(<UnusedInsightsPage />))
    expect(screen.getByText(/нет забытых/i)).toBeInTheDocument()
  })

  it("renders loading state", () => {
    vi.mocked(insightsHooks.useUnusedPrompts).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as ReturnType<typeof insightsHooks.useUnusedPrompts>)
    render(wrap(<UnusedInsightsPage />))
    expect(screen.getByText(/загруж/i)).toBeInTheDocument()
  })
})
