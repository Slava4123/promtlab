import { api } from "./client"
import type { BadgeListResponse } from "./types"

export function fetchBadges() {
  return api<BadgeListResponse>("/badges")
}
