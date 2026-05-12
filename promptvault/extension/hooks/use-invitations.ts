import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import type { TeamInvitation } from "../lib/types"

const KEY = ["invitations", "my"] as const

export function useMyInvitations() {
  return useQuery<TeamInvitation[]>({
    queryKey: KEY,
    queryFn: () => sendBg({ type: "api.listMyInvitations" }),
    staleTime: 60_000,
    refetchOnWindowFocus: true,
  })
}

export function useAcceptInvitation() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (invitationId) => sendBg({ type: "api.acceptInvitation", invitationId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
      void qc.invalidateQueries({ queryKey: ["teams"] })
    },
  })
}

export function useDeclineInvitation() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (invitationId) => sendBg({ type: "api.declineInvitation", invitationId }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: KEY }),
  })
}
