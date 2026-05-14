import type { TokenPair } from "./types"
import { captureException } from "@/lib/sentry"

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

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
  let res: Response
  try {
    res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include", // отправляет HttpOnly cookie
      headers: { "Content-Type": "application/json" },
    })
  } catch (err) {
    // Network error (сервер недоступен, потеряна сеть, AbortError) — transient,
    // не сбрасываем сессию. Ошибка помечена специальным сообщением, которое
    // auth-store ловит и не редиректит на /sign-in.
    throw new Error("transient: network unavailable", { cause: err })
  }

  // 401/403 — истинный auth-fail (refresh истёк/невалиден). Только в этом случае
  // чистим токены и заставляем юзера перелогиниться.
  if (res.status === 401 || res.status === 403) {
    clearTokens()
    throw new Error("Сессия истекла")
  }
  // 5xx — сервер временно недоступен. Не чистим токены.
  if (!res.ok) {
    throw new Error(`transient: server ${res.status}`)
  }

  const tokens: TokenPair = await res.json()
  setTokens(tokens)
}

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${API_BASE}${path}`

  // MN-64: типизированно через Headers API. Раньше `as Record<string,string>`
  // молча проглатывал case когда caller передавал `Headers` instance или
  // tuple-array — некоторые headers просто терялись.
  const headers = new Headers(options.headers)
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`)
  }

  try {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
    if (tz) headers.set("X-Timezone", tz)
  } catch {
    // Intl unavailable — streak will use UTC fallback
  }

  // credentials: include для auth-эндпоинтов (cookie)
  const credentials = path.startsWith("/auth") ? "include" as const : undefined

  // Proactive refresh: если accessToken отсутствует (page reload до restoreSession,
  // tab restore в новой сессии), пытаемся получить fresh JWT через refresh cookie
  // перед отправкой запроса. Иначе protected-эндпоинт ответит 401 «missing auth».
  // Auth-эндпоинты пропускаем (refresh сам по себе auth-эндпоинт).
  const isAuthEndpoint =
    path.startsWith("/auth/login") ||
    path.startsWith("/auth/register") ||
    path.startsWith("/auth/refresh") ||
    path.startsWith("/auth/verify-totp")
  if (!accessToken && !isAuthEndpoint) {
    try {
      await ensureFreshToken()
      if (accessToken) {
        headers.set("Authorization", `Bearer ${accessToken}`)
      }
    } catch (err) {
      // Transient ошибки (сеть, 5xx) прокидываем — иначе юзер увидит
      // фальш-«Сессия истекла» при flaky-соединении. true auth-fail
      // ('Сессия истекла') тоже прокидываем — нет смысла слать запрос
      // без токена, всё равно 401.
      if (err instanceof Error && err.message.startsWith("transient:")) {
        throw err
      }
      // Auth-fail (no cookie / expired refresh) — пробрасываем.
      throw err
    }
  }

  let res = await fetch(url, { ...options, headers, credentials })

  // Auto-refresh ТОЛЬКО на 401 от auth middleware (истёкший/невалидный JWT).
  // Бизнес-валидация (неверный TOTP код, не найден enrollment) возвращается
  // как 422 Unprocessable Entity — такие ошибки НЕ триггерят refresh retry.
  //
  // Историческая заметка (BUG #1 из QA): раньше любая 401 на protected
  // endpoint делала retry, что приводило к двойному запросу при неверном
  // TOTP коде. Теперь backend маппит бизнес-валидацию в 422, а client
  // ретритит только истинные auth failures (401 без токена или с expired).
  if (res.status === 401 && !isAuthEndpoint && accessToken) {
    try {
      await ensureFreshToken()
      headers.set("Authorization", `Bearer ${accessToken}`)
      res = await fetch(url, { ...options, headers, credentials })
    } catch {
      throw new Error("Сессия истекла, войдите снова")
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Ошибка запроса" }))

    // Quota exceeded (402) — показываем глобальный upgrade dialog.
    // Backend кладёт в body: quota_type, used, limit, plan, upgrade_url (см. errors.RespondQuotaError).
    if (res.status === 402) {
      const { useQuotaStore } = await import("@/stores/quota-store")
      useQuotaStore.getState().show({
        quotaType: body.quota_type || "unknown",
        message: body.error || "Лимит исчерпан",
        used: typeof body.used === "number" ? body.used : undefined,
        limit: typeof body.limit === "number" ? body.limit : undefined,
        plan: typeof body.plan === "string" ? body.plan : undefined,
      })
    }

    const apiError = new ApiError(body.error || `Ошибка сервера (${res.status})`, res.status)
    // Капчурим только 5xx — это server errors, которые разработчик должен увидеть.
    // 4xx обычно user input errors (validation, auth, not found) — не шлём noise.
    if (res.status >= 500) {
      captureException(apiError, {
        tags: {
          api_status: String(res.status),
          api_path: path,
        },
      })
    }
    throw apiError
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

export async function publicApi<T>(path: string): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, {
      credentials: "omit",
      headers: { "Content-Type": "application/json" },
    })
  } catch {
    throw new ApiError("Ошибка сети, проверьте подключение", 0)
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Ошибка запроса" }))
    const apiError = new ApiError(body.error || `Ошибка (${res.status})`, res.status)
    if (res.status >= 500) {
      captureException(apiError, {
        tags: { api_status: String(res.status), api_path: path },
      })
    }
    throw apiError
  }

  try {
    return await res.json()
  } catch {
    throw new ApiError("Ошибка при обработке ответа сервера", 0)
  }
}
