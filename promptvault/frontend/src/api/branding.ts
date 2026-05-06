import { ApiError, api, ensureFreshToken, getAccessToken } from "./client"
import { useQuotaStore } from "@/stores/quota-store"

// --- Types ---

// LogoSource дискриминирует, как фронт показывает логотип.
//   - "url"  — внешняя ссылка (BrandingInfo.logo_url)
//   - "file" — bytea на бэке, отдаётся через GET /branding/logo
//   - "none" — логотипа нет
export type LogoSource = "url" | "file" | "none"

export interface BrandingInfo {
  logo_url?: string
  logo_source?: LogoSource
  effective_logo_url?: string
  tagline?: string
  website?: string
  primary_color?: string
}

export interface BrandingInput {
  logo_url: string
  logo_source?: LogoSource
  tagline: string
  website: string
  primary_color: string
}

export interface LogoUploadResponse {
  logo_source: "file"
  effective_logo_url: string
  size_bytes: number
  content_type: string
}

// --- API functions ---

export function fetchBranding(slug: string): Promise<BrandingInfo> {
  return api<BrandingInfo>(`/teams/${encodeURIComponent(slug)}/branding`)
}

export function updateBranding(slug: string, input: BrandingInput): Promise<BrandingInfo> {
  return api<BrandingInfo>(`/teams/${encodeURIComponent(slug)}/branding`, {
    method: "PUT",
    body: JSON.stringify(input),
  })
}

// uploadLogo — multipart POST. api() жёстко ставит Content-Type: application/json,
// поэтому идём прямым fetch'ем; переиспользуем 401-refresh + 402-quota обработку.
export async function uploadLogo(slug: string, file: File): Promise<LogoUploadResponse> {
  const path = `/api/teams/${encodeURIComponent(slug)}/branding/logo`
  const form = new FormData()
  form.append("file", file)

  const doFetch = async (): Promise<Response> => {
    const headers: Record<string, string> = {}
    const tok = getAccessToken()
    if (tok) headers["Authorization"] = `Bearer ${tok}`
    return fetch(path, { method: "POST", body: form, headers })
  }

  let res = await doFetch()
  if (res.status === 401 && getAccessToken()) {
    try {
      await ensureFreshToken()
      res = await doFetch()
    } catch {
      throw new ApiError("Сессия истекла, войдите снова", 401)
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: `Ошибка сервера (${res.status})` }))
    if (res.status === 402) {
      useQuotaStore.getState().show({
        quotaType: body.quota_type || "branding",
        message: body.error || "Логотип доступен только на Max",
        used: typeof body.used === "number" ? body.used : undefined,
        limit: typeof body.limit === "number" ? body.limit : undefined,
        plan: typeof body.plan === "string" ? body.plan : undefined,
      })
    }
    throw new ApiError(body.error || `Ошибка сервера (${res.status})`, res.status)
  }

  return res.json() as Promise<LogoUploadResponse>
}

export function deleteLogo(slug: string): Promise<{ logo_source: "none" }> {
  return api<{ logo_source: "none" }>(`/teams/${encodeURIComponent(slug)}/branding/logo`, {
    method: "DELETE",
  })
}
