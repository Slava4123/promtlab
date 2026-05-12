import { useQuery } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { BadgeListResponse } from "../lib/types"

export function useBadges() {
  return useQuery<BadgeListResponse>({
    queryKey: qk.badges,
    queryFn: () => sendBg({ type: "api.listBadges" }),
    staleTime: 5 * 60_000,
  })
}
