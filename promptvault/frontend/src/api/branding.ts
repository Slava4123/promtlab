import { api } from "./client"

// --- Types ---

export interface BrandingInfo {
  logo_url?: string
  tagline?: string
  website?: string
  primary_color?: string
}

export interface BrandingInput {
  logo_url: string
  tagline: string
  website: string
  primary_color: string
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
