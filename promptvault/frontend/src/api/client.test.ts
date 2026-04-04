import {
  api,
  apiVoid,
  setTokens,
  clearTokens,
  getAccessToken,
  setAccessToken,
  ensureFreshToken,
} from "./client"

const mockFetch = vi.fn()
global.fetch = mockFetch

beforeEach(() => {
  mockFetch.mockReset()
  clearTokens()
})

// ---------- api() ----------

describe("api()", () => {
  it("successful JSON response", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: 1, name: "test" }),
    })

    const result = await api<{ id: number; name: string }>("/prompts")

    expect(result).toEqual({ id: 1, name: "test" })
    expect(mockFetch).toHaveBeenCalledTimes(1)
    expect(mockFetch).toHaveBeenCalledWith(
      "/api/prompts",
      expect.objectContaining({
        headers: expect.objectContaining({
          "Content-Type": "application/json",
        }),
      }),
    )
  })

  it("non-200 throws Error with server message", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ error: "bad request" }),
    })

    await expect(api("/prompts")).rejects.toThrow("bad request")
  })

  it("204 returns undefined", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(null),
    })

    const result = await api("/prompts/1", { method: "DELETE" })

    expect(result).toBeUndefined()
  })

  it("attaches Authorization header when token set", async () => {
    setAccessToken("my-token-123")

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ ok: true }),
    })

    await api("/prompts")

    expect(mockFetch).toHaveBeenCalledWith(
      "/api/prompts",
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer my-token-123",
        }),
      }),
    )
  })

  it("auto-refresh on 401", async () => {
    setAccessToken("expired-token")

    // First call returns 401
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "token expired" }),
    })

    // Refresh call succeeds
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          access_token: "new-token",
                    expires_in: 900,
        }),
    })

    // Retry call succeeds
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: 1 }),
    })

    const result = await api<{ id: number }>("/prompts")

    expect(result).toEqual({ id: 1 })
    expect(mockFetch).toHaveBeenCalledTimes(3)

    // The retry should use the new token
    const retryCall = mockFetch.mock.calls[2]
    expect(retryCall[1].headers["Authorization"]).toBe("Bearer new-token")

    // Access token should be updated
    expect(getAccessToken()).toBe("new-token")
  })

  it("non-200 with unparseable body falls back to generic message", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: () => Promise.reject(new Error("not json")),
    })

    await expect(api("/prompts")).rejects.toThrow("request failed")
  })
})

// ---------- apiVoid() ----------

describe("apiVoid()", () => {
  it("returns void on 204", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(null),
    })

    const result = await apiVoid("/prompts/1", { method: "DELETE" })

    expect(result).toBeUndefined()
  })
})

// ---------- token management ----------

describe("setTokens / clearTokens / getAccessToken", () => {
  it("sets and retrieves access token", () => {
    expect(getAccessToken()).toBeNull()

    setTokens({
      access_token: "abc",
      expires_in: 900,
    })

    expect(getAccessToken()).toBe("abc")
  })

  it("clearTokens resets to null", () => {
    setAccessToken("some-token")
    expect(getAccessToken()).toBe("some-token")

    clearTokens()
    expect(getAccessToken()).toBeNull()
  })
})

// ---------- ensureFreshToken ----------

describe("ensureFreshToken", () => {
  it("deduplication — only one refresh request for concurrent calls", async () => {
    // Refresh endpoint returns new tokens
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          access_token: "fresh-token",
          expires_in: 900,
        }),
    })

    // Call ensureFreshToken twice simultaneously
    const [r1, r2] = await Promise.all([
      ensureFreshToken(),
      ensureFreshToken(),
    ])

    // Only one fetch call should have been made (deduplication)
    expect(mockFetch).toHaveBeenCalledTimes(1)
    expect(r1).toBeUndefined()
    expect(r2).toBeUndefined()
    expect(getAccessToken()).toBe("fresh-token")
  })

  it("failed refresh clears tokens and throws", async () => {
    setAccessToken("old-token")

    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "invalid refresh token" }),
    })

    await expect(ensureFreshToken()).rejects.toThrow("refresh failed")
    expect(getAccessToken()).toBeNull()
  })
})
