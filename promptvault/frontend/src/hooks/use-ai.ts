import { useQuery } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { AIModel } from "@/api/types"

export function useAIModels() {
  return useQuery({
    queryKey: ["ai-models"],
    queryFn: () => api<AIModel[]>("/ai/models"),
    staleTime: 30 * 60 * 1000,
  })
}
