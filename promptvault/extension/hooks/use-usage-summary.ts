import { useQuery } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { UsageSummary } from "../lib/types"

export function useUsageSummary() {
  return useQuery<UsageSummary>({
    queryKey: qk.usage,
    queryFn: () => sendBg({ type: "api.getUsageSummary" }),
    staleTime: 60_000,
  })
}
