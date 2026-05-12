import { useQuery } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { StreakResponse } from "../lib/types"

export function useStreakDetail() {
  return useQuery<StreakResponse>({
    queryKey: [...qk.streak, "detail"],
    queryFn: () => sendBg({ type: "api.getStreakDetail" }),
    staleTime: 60_000,
  })
}
