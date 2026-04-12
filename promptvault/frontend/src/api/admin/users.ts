import { api } from "@/api/client"
import type { PaginatedResponse, UserRole, UserStatus } from "@/api/types"

export interface AdminUserSummary {
  id: number
  email: string
  name: string
  username?: string
  role: UserRole
  status: UserStatus
  email_verified: boolean
  created_at: string
}

export interface AdminUserDetail {
  id: number
  email: string
  name: string
  username?: string
  avatar_url?: string
  role: UserRole
  status: UserStatus
  email_verified: boolean
  default_model: string
  created_at: string
  updated_at: string
  prompt_count: number
  collection_count: number
  badge_count: number
  total_usage: number
  linked_providers: string[]
  unlocked_badge_ids: string[]
  tier: string
}

export interface AdminUsersFilter {
  q?: string
  role?: UserRole | ""
  status?: UserStatus | ""
  page?: number
  page_size?: number
}

export function fetchAdminUsers(filter: AdminUsersFilter = {}) {
  const params = new URLSearchParams()
  if (filter.q) params.set("q", filter.q)
  if (filter.role) params.set("role", filter.role)
  if (filter.status) params.set("status", filter.status)
  if (filter.page) params.set("page", String(filter.page))
  if (filter.page_size) params.set("page_size", String(filter.page_size))
  const qs = params.toString()
  return api<PaginatedResponse<AdminUserSummary>>(
    `/admin/users${qs ? `?${qs}` : ""}`,
  )
}

export function fetchAdminUserDetail(id: number) {
  return api<AdminUserDetail>(`/admin/users/${id}`)
}

export function freezeUser(id: number) {
  return api<{ ok: boolean; action: string }>(`/admin/users/${id}/freeze`, {
    method: "POST",
  })
}

export function unfreezeUser(id: number) {
  return api<{ ok: boolean; action: string }>(`/admin/users/${id}/unfreeze`, {
    method: "POST",
  })
}

export function resetUserPassword(id: number, totpCode: string) {
  return api<{ ok: boolean; action: string }>(`/admin/users/${id}/reset-password`, {
    method: "POST",
    body: JSON.stringify({ totp_code: totpCode }),
  })
}

export function grantBadge(userId: number, badgeId: string) {
  return api<{ ok: boolean; badge: { id: string; title: string; icon: string } }>(
    `/admin/users/${userId}/badges/${badgeId}/grant`,
    { method: "POST" },
  )
}

export function revokeBadge(userId: number, badgeId: string, totpCode: string) {
  return api<{ ok: boolean; action: string }>(
    `/admin/users/${userId}/badges/${badgeId}`,
    {
      method: "DELETE",
      body: JSON.stringify({ totp_code: totpCode }),
    },
  )
}
