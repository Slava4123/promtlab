import { api } from "@/api/client"
import type { PaginatedResponse } from "@/api/types"

export type FeedbackType = "bug" | "feature" | "other"
export type FeedbackStatus = "new" | "read" | "archived"

export interface AdminFeedbackItem {
  id: number
  user_id: number
  user_email: string
  user_name?: string
  type: FeedbackType
  status: FeedbackStatus
  message: string
  page_url?: string
  created_at: string
}

export interface AdminFeedbacksFilter {
  type?: FeedbackType | ""
  status?: FeedbackStatus | ""
  q?: string
  page?: number
  page_size?: number
}

export function fetchAdminFeedbacks(filter: AdminFeedbacksFilter = {}) {
  const params = new URLSearchParams()
  if (filter.type) params.set("type", filter.type)
  if (filter.status) params.set("status", filter.status)
  if (filter.q) params.set("q", filter.q)
  if (filter.page) params.set("page", String(filter.page))
  if (filter.page_size) params.set("page_size", String(filter.page_size))
  const qs = params.toString()
  return api<PaginatedResponse<AdminFeedbackItem>>(
    `/admin/feedbacks${qs ? `?${qs}` : ""}`,
  )
}

export function fetchAdminFeedbackDetail(id: number) {
  return api<AdminFeedbackItem>(`/admin/feedbacks/${id}`)
}

// PATCH /admin/feedbacks/{id}/status — sudo (TOTP).
// status: 'new' | 'read' | 'archived'.
export function updateFeedbackStatus(
  id: number,
  status: FeedbackStatus,
  totpCode: string,
) {
  return api<{ ok: boolean; action: string }>(
    `/admin/feedbacks/${id}/status`,
    {
      method: "PATCH",
      body: JSON.stringify({ status, totp_code: totpCode }),
    },
  )
}

// DELETE /admin/feedbacks/{id} — sudo (TOTP). Hard delete без восстановления.
export function deleteFeedback(id: number, totpCode: string) {
  return api<{ ok: boolean; action: string }>(`/admin/feedbacks/${id}`, {
    method: "DELETE",
    body: JSON.stringify({ totp_code: totpCode }),
  })
}
