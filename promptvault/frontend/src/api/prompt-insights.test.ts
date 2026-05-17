import { describe, it, expect, vi, beforeEach } from "vitest"
import {
  fetchUnused,
  fetchDuplicates,
  fetchTrending,
  fetchDeclining,
  fetchMostEdited,
  mergePrompts,
} from "./prompt-insights"
import { setAccessToken, clearTokens } from "./client"

const mockFetch = vi.fn()
beforeEach(() => {
  mockFetch.mockReset()
  globalThis.fetch = mockFetch as unknown as typeof fetch
  clearTokens()
  // existing access token чтобы пропустить proactive refresh — иначе api()
  // сделает доп. вызов /api/auth/refresh и тесты захлебнутся в моках.
  setAccessToken("test-token")
})

describe("prompt-insights api", () => {
  it("fetches unused", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ items: [{ prompt_id: 1, title: "X", uses: 0 }] }),
    })
    const items = await fetchUnused()
    expect(items).toHaveLength(1)
    expect(items[0]).toEqual({ prompt_id: 1, title: "X", uses: 0 })
    const call = mockFetch.mock.calls[0]
    expect(call[0]).toBe("/api/prompts/insights/unused")
    expect((call[1].headers as Headers).get("Authorization")).toBe("Bearer test-token")
  })

  it("fetches duplicates with pairs", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({
        items: [
          {
            prompt_a: { prompt_id: 1, title: "A", uses: 0 },
            prompt_b: { prompt_id: 2, title: "B", uses: 0 },
            similarity: 0.9,
          },
        ],
      }),
    })
    const pairs = await fetchDuplicates()
    expect(pairs).toHaveLength(1)
    expect(pairs[0].similarity).toBe(0.9)
    expect(mockFetch.mock.calls[0][0]).toBe("/api/prompts/insights/duplicates")
  })

  it("fetches trending", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ items: [{ prompt_id: 5, title: "Hot", uses: 20 }] }),
    })
    const items = await fetchTrending()
    expect(items[0].title).toBe("Hot")
    expect(mockFetch.mock.calls[0][0]).toBe("/api/prompts/insights/trending")
  })

  it("fetches declining", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ items: [] }),
    })
    const items = await fetchDeclining()
    expect(items).toHaveLength(0)
    expect(mockFetch.mock.calls[0][0]).toBe("/api/prompts/insights/declining")
  })

  it("fetches most-edited", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ items: [{ prompt_id: 8, title: "Churn", uses: 15 }] }),
    })
    const items = await fetchMostEdited()
    expect(items[0].uses).toBe(15)
    expect(mockFetch.mock.calls[0][0]).toBe("/api/prompts/insights/most-edited")
  })

  it("throws on 402", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 402,
      json: async () => ({ error: "pro_required" }),
    })
    await expect(fetchUnused()).rejects.toThrow("pro_required")
  })

  it("merges prompts", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ kept_id: 1, merged_id: 2 }),
    })
    const res = await mergePrompts(1, 2)
    expect(res).toEqual({ kept_id: 1, merged_id: 2 })
    const call = mockFetch.mock.calls[0]
    expect(call[0]).toBe("/api/prompts/1/merge-with/2")
    expect(call[1].method).toBe("POST")
  })

  it("handles empty items array (no items field)", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({}),
    })
    const items = await fetchUnused()
    expect(items).toEqual([])
  })
})
