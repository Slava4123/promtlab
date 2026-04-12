import { api } from "@/api/client"

export interface HealthResponse {
  status: string
  time: string
  total_users: number
  admin_users: number
  active_users: number
  frozen_users: number
}

export function fetchHealth() {
  return api<HealthResponse>("/admin/health")
}
