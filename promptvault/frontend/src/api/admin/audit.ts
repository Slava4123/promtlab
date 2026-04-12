import { api } from "@/api/client"
import type { PaginatedResponse } from "@/api/types"

export interface AuditEntry {
  id: number
  admin_id: number
  action: string
  target_type: string
  target_id?: number
  before_state?: unknown
  after_state?: unknown
  ip: string
  user_agent?: string
  created_at: string
}

export interface AuditFilter {
  action?: string
  target_type?: string
  admin_id?: number
  target_id?: number
  from?: string
  to?: string
  page?: number
  page_size?: number
}

export function fetchAudit(filter: AuditFilter = {}) {
  const params = new URLSearchParams()
  if (filter.action) params.set("action", filter.action)
  if (filter.target_type) params.set("target_type", filter.target_type)
  if (filter.admin_id) params.set("admin_id", String(filter.admin_id))
  if (filter.target_id) params.set("target_id", String(filter.target_id))
  if (filter.from) params.set("from", filter.from)
  if (filter.to) params.set("to", filter.to)
  if (filter.page) params.set("page", String(filter.page))
  if (filter.page_size) params.set("page_size", String(filter.page_size))
  const qs = params.toString()
  return api<PaginatedResponse<AuditEntry>>(`/admin/audit${qs ? `?${qs}` : ""}`)
}
