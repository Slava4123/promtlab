import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  fetchAdminFeedbacks,
  fetchAdminFeedbackDetail,
  updateFeedbackStatus,
  deleteFeedback,
  type AdminFeedbacksFilter,
  type FeedbackStatus,
} from "@/api/admin/feedbacks"

export function useAdminFeedbacks(filter: AdminFeedbacksFilter) {
  return useQuery({
    queryKey: ["admin", "feedbacks", filter],
    queryFn: () => fetchAdminFeedbacks(filter),
    staleTime: 30_000,
  })
}

export function useAdminFeedbackDetail(id: number) {
  return useQuery({
    queryKey: ["admin", "feedback", id],
    queryFn: () => fetchAdminFeedbackDetail(id),
    enabled: id > 0,
  })
}

export function useUpdateFeedbackStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      status,
      totpCode,
    }: {
      id: number
      status: FeedbackStatus
      totpCode: string
    }) => updateFeedbackStatus(id, status, totpCode),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ["admin", "feedbacks"] })
      qc.invalidateQueries({ queryKey: ["admin", "feedback", id] })
    },
  })
}

export function useDeleteFeedback() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, totpCode }: { id: number; totpCode: string }) =>
      deleteFeedback(id, totpCode),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin", "feedbacks"] })
    },
  })
}
