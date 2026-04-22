import { useQuery } from "@tanstack/react-query"
import { fetchPromptHistory } from "@/api/activity"

export function usePromptHistory(promptId: number) {
  return useQuery({
    queryKey: ["activity", "prompt", promptId],
    queryFn: () => fetchPromptHistory(promptId),
    enabled: promptId > 0,
  })
}
