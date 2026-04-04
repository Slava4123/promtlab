import type { TokenPair } from "./types"

const API_BASE = "/api"

let accessToken: string | null = null
let refreshPromise: Promise<void> | null = null

export function setTokens(tokens: TokenPair) {
  accessToken = tokens.access_token
  // refresh_token теперь в HttpOnly cookie — не храним на клиенте
}

export function setAccessToken(token: string) {
  accessToken = token
}

export function clearTokens() {
  accessToken = null
}

export function getAccessToken() {
  return accessToken
}

export async function ensureFreshToken(): Promise<void> {
  if (!refreshPromise) {
    refreshPromise = refreshAccessToken().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

async function refreshAccessToken(): Promise<void> {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    credentials: "include", // отправляет HttpOnly cookie
    headers: { "Content-Type": "application/json" },
  })

  if (!res.ok) {
    clearTokens()
    throw new Error("refresh failed")
  }

  const tokens: TokenPair = await res.json()
  setTokens(tokens)
}

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${API_BASE}${path}`

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  }

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`
  }

  // credentials: include для auth-эндпоинтов (cookie)
  const credentials = path.startsWith("/auth") ? "include" as const : undefined

  let res = await fetch(url, { ...options, headers, credentials })

  // Auto-refresh на 401
  if (res.status === 401) {
    try {
      await ensureFreshToken()
      headers["Authorization"] = `Bearer ${accessToken}`
      res = await fetch(url, { ...options, headers, credentials })
    } catch {
      throw new Error("unauthorized")
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "request failed" }))
    throw new Error(body.error || `HTTP ${res.status}`)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}

export async function apiVoid(
  path: string,
  options: RequestInit = {},
): Promise<void> {
  await api<unknown>(path, options)
}
