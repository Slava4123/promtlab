import { useMutation } from "@tanstack/react-query"
import { api } from "@/api/client"
import type { FeedbackRequest, FeedbackResponse } from "@/api/types"

export function useSubmitFeedback() {
  return useMutation({
    mutationFn: (data: FeedbackRequest) =>
      api<FeedbackResponse>("/feedback", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  })
}
