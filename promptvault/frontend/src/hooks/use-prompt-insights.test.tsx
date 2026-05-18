import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactNode } from "react"
import {
  useUnusedPrompts,
  useDuplicates,
  useTrending,
  useDeclining,
  useMostEdited,
  useMergePrompts,
} from "./use-prompt-insights"
import * as api from "@/api/prompt-insights"

vi.mock("@/api/prompt-insights")

function makeWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  }
  return Wrapper
}

describe("use-prompt-insights", () => {
  beforeEach(() => vi.resetAllMocks())

  it("useUnusedPrompts returns data", async () => {
    vi.mocked(api.fetchUnused).mockResolvedValue([{ prompt_id: 1, title: "X", uses: 0 }])
    const { result } = renderHook(() => useUnusedPrompts(), { wrapper: makeWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toHaveLength(1)
  })

  it("useDuplicates returns pairs", async () => {
    vi.mocked(api.fetchDuplicates).mockResolvedValue([
      {
        prompt_a: { prompt_id: 1, title: "A", uses: 0 },
        prompt_b: { prompt_id: 2, title: "B", uses: 0 },
        similarity: 0.9,
      },
    ])
    const { result } = renderHook(() => useDuplicates(), { wrapper: makeWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.[0].similarity).toBe(0.9)
  })

  it("useTrending returns rows", async () => {
    vi.mocked(api.fetchTrending).mockResolvedValue([{ prompt_id: 5, title: "Hot", uses: 20 }])
    const { result } = renderHook(() => useTrending(), { wrapper: makeWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.[0].title).toBe("Hot")
  })

  it("useDeclining returns rows", async () => {
    vi.mocked(api.fetchDeclining).mockResolvedValue([])
    const { result } = renderHook(() => useDeclining(), { wrapper: makeWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toHaveLength(0)
  })

  it("useMostEdited returns rows", async () => {
    vi.mocked(api.fetchMostEdited).mockResolvedValue([{ prompt_id: 8, title: "Churn", uses: 15 }])
    const { result } = renderHook(() => useMostEdited(), { wrapper: makeWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.[0].uses).toBe(15)
  })

  it("useMergePrompts triggers mutation", async () => {
    vi.mocked(api.mergePrompts).mockResolvedValue({ kept_id: 1, merged_id: 2 })
    const { result } = renderHook(() => useMergePrompts(), { wrapper: makeWrapper() })
    await result.current.mutateAsync({ keepID: 1, mergeID: 2 })
    expect(api.mergePrompts).toHaveBeenCalledWith(1, 2)
  })
})
