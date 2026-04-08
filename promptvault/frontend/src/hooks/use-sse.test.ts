import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, act } from "@testing-library/react"
import { useSSE } from "./use-sse"

// Mock api/client
vi.mock("@/api/client", () => ({
  getAccessToken: vi.fn(() => "test-token"),
  ensureFreshToken: vi.fn(),
}))

function mockFetchSSE(chunks: string[], status = 200) {
  const encoder = new TextEncoder()
  let chunkIndex = 0

  const mockResponse = {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve({ error: "Server error" }),
    body: {
      getReader: () => ({
        read: () => {
          if (chunkIndex < chunks.length) {
            const data = encoder.encode(chunks[chunkIndex++])
            return Promise.resolve({ done: false, value: data })
          }
          return Promise.resolve({ done: true, value: undefined })
        },
      }),
    },
  }

  return vi.fn(() => Promise.resolve(mockResponse))
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe("useSSE", () => {
  it("initial state is idle", () => {
    const { result } = renderHook(() => useSSE())
    expect(result.current.data).toBe("")
    expect(result.current.isStreaming).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it("streams data from SSE chunks", async () => {
    const fetchMock = mockFetchSSE([
      "data: Hello\n\n",
      "data:  world\n\n",
      "data: [DONE]\n\n",
    ])
    vi.stubGlobal("fetch", fetchMock)

    const { result } = renderHook(() => useSSE())

    await act(async () => {
      await result.current.start("/ai/enhance", { content: "test" })
    })

    expect(result.current.data).toBe("Hello world")
    expect(result.current.isStreaming).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it("handles HTTP error response", async () => {
    const fetchMock = mockFetchSSE([], 500)
    vi.stubGlobal("fetch", fetchMock)

    const { result } = renderHook(() => useSSE())

    await act(async () => {
      await result.current.start("/ai/enhance", { content: "test" })
    })

    expect(result.current.error).toBe("Server error")
    expect(result.current.isStreaming).toBe(false)
  })

  it("handles SSE error event", async () => {
    const fetchMock = mockFetchSSE([
      "event: error\ndata: Rate limit exceeded\n\n",
    ])
    vi.stubGlobal("fetch", fetchMock)

    const { result } = renderHook(() => useSSE())

    await act(async () => {
      await result.current.start("/ai/enhance", { content: "test" })
    })

    expect(result.current.error).toBe("Rate limit exceeded")
    expect(result.current.isStreaming).toBe(false)
  })

  it("sends auth header with token", async () => {
    const fetchMock = mockFetchSSE(["data: [DONE]\n\n"])
    vi.stubGlobal("fetch", fetchMock)

    const { result } = renderHook(() => useSSE())

    await act(async () => {
      await result.current.start("/ai/enhance", { content: "test" })
    })

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/ai/enhance",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer test-token",
        }),
      })
    )
  })
})
