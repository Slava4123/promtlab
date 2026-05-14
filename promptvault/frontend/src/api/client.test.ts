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
    setAccessToken("existing") // skip proactive refresh
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: 1, name: "test" }),
    })

    const result = await api<{ id: number; name: string }>("/prompts")

    expect(result).toEqual({ id: 1, name: "test" })
    expect(mockFetch).toHaveBeenCalledTimes(1)
    // MN-64: api() передаёт Headers instance, не plain object — проверяем через .get()
    const call = mockFetch.mock.calls[0]
    expect(call[0]).toBe("/api/prompts")
    expect((call[1].headers as Headers).get("Content-Type")).toBe("application/json")
  })

  it("non-200 throws Error with server message", async () => {
    setAccessToken("existing")
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ error: "bad request" }),
    })

    await expect(api("/prompts")).rejects.toThrow("bad request")
  })

  it("204 returns undefined", async () => {
    setAccessToken("existing")
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

    const call = mockFetch.mock.calls[0]
    expect(call[0]).toBe("/api/prompts")
    expect((call[1].headers as Headers).get("Authorization")).toBe("Bearer my-token-123")
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

    // The retry should use the new token (Headers API после MN-64).
    const retryCall = mockFetch.mock.calls[2]
    expect((retryCall[1].headers as Headers).get("Authorization")).toBe("Bearer new-token")

    // Access token should be updated
    expect(getAccessToken()).toBe("new-token")
  })

  it("non-200 with unparseable body falls back to generic message", async () => {
    setAccessToken("existing")
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: () => Promise.reject(new Error("not json")),
    })

    await expect(api("/prompts")).rejects.toThrow("Ошибка запроса")
  })
})

// ---------- proactive refresh (новое поведение после a4ea7f4) ----------

describe("api() — proactive refresh", () => {
  it("вызывает ensureFreshToken ДО запроса если accessToken отсутствует", async () => {
    // Нет токена → должен сделать refresh-запрос, затем основной запрос с новым токеном.
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          access_token: "proactive-token",
          expires_in: 900,
        }),
    })
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: 42 }),
    })

    const result = await api<{ id: number }>("/prompts")

    expect(result).toEqual({ id: 42 })
    expect(mockFetch).toHaveBeenCalledTimes(2)

    // Первый вызов — refresh.
    expect(mockFetch.mock.calls[0][0]).toBe("/api/auth/refresh")

    // Второй вызов — protected эндпоинт с подмешанным Bearer.
    const apiCall = mockFetch.mock.calls[1]
    expect(apiCall[0]).toBe("/api/prompts")
    expect((apiCall[1].headers as Headers).get("Authorization")).toBe("Bearer proactive-token")
  })

  it("при auth-fail proactive refresh пробрасывает 'Сессия истекла' без дополнительного запроса", async () => {
    // 401 на refresh = истинный auth-fail (нет cookie / expired). Нет смысла
    // слать основной запрос без токена — всё равно 401. Сразу throw.
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "no refresh cookie" }),
    })

    await expect(api("/prompts")).rejects.toThrow("Сессия истекла")

    // Только один fetch — refresh. Основной запрос не делается.
    expect(mockFetch).toHaveBeenCalledTimes(1)
    expect(mockFetch.mock.calls[0][0]).toBe("/api/auth/refresh")
  })

  it("при transient ошибке (network) proactive refresh пробрасывает 'transient:' — auth-store не делает logout", async () => {
    // Refresh упал из-за сети (fetch reject), а не auth-fail. Прокидываем
    // transient-ошибку, чтобы auth-store не редиректил юзера на /sign-in
    // при flaky-соединении.
    mockFetch.mockRejectedValueOnce(new TypeError("Failed to fetch"))

    await expect(api("/prompts")).rejects.toThrow(/transient:/)

    // Основной запрос не делается — нет смысла без токена.
    expect(mockFetch).toHaveBeenCalledTimes(1)
  })

  it.each([
    "/auth/login",
    "/auth/register",
    "/auth/refresh",
    "/auth/verify-totp",
  ])("НЕ вызывает proactive refresh для auth-эндпоинта %s", async (path) => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })

    await api(path, { method: "POST" })

    // Только один вызов — proactive refresh пропущен.
    expect(mockFetch).toHaveBeenCalledTimes(1)
    expect(mockFetch.mock.calls[0][0]).toBe(`/api${path}`)
  })

  it("НЕ дёргает proactive refresh если accessToken уже установлен", async () => {
    setAccessToken("existing-token")

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })

    await api("/prompts")

    // Только один вызов — refresh пропущен.
    expect(mockFetch).toHaveBeenCalledTimes(1)
    const call = mockFetch.mock.calls[0]
    expect((call[1].headers as Headers).get("Authorization")).toBe("Bearer existing-token")
  })
})

// ---------- apiVoid() ----------

describe("apiVoid()", () => {
  it("returns void on 204", async () => {
    setAccessToken("existing")
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

    await expect(ensureFreshToken()).rejects.toThrow("Сессия истекла")
    expect(getAccessToken()).toBeNull()
  })
})
