import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api, apiVoid } from "@/api/client"
import type { Team, TeamDetail, TeamInvitation, PendingInvitation, TeamRole, UserSearchResult } from "@/api/types"

export function useTeams() {
  return useQuery({
    queryKey: ["teams"],
    queryFn: () => api<Team[]>("/teams"),
  })
}

export function useTeam(slug: string) {
  return useQuery({
    queryKey: ["team", slug],
    queryFn: () => api<TeamDetail>(`/teams/${slug}`),
    enabled: !!slug,
    refetchInterval: 15_000,
  })
}

export function useCreateTeam() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; description?: string }) =>
      api<Team>("/teams", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["teams"] }),
  })
}

export function useUpdateTeam() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ slug, ...data }: { slug: string; name?: string; description?: string }) =>
      api<Team>(`/teams/${slug}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["teams"] })
      qc.invalidateQueries({ queryKey: ["team", variables.slug] })
    },
  })
}

export function useDeleteTeam() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (slug: string) => apiVoid(`/teams/${slug}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["teams"] }),
  })
}

export function useInviteMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ slug, query, role }: { slug: string; query: string; role: TeamRole }) =>
      api<TeamInvitation>(`/teams/${slug}/invitations`, {
        method: "POST",
        body: JSON.stringify({ query, role }),
      }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["team", variables.slug] })
      qc.invalidateQueries({ queryKey: ["team-invitations", variables.slug] })
    },
  })
}

export function useTeamInvitations(slug: string) {
  return useQuery({
    queryKey: ["team-invitations", slug],
    queryFn: () => api<PendingInvitation[]>(`/teams/${slug}/invitations`),
    enabled: !!slug,
    refetchInterval: 15_000,
  })
}

export function useCancelInvitation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ slug, invitationId }: { slug: string; invitationId: number }) =>
      apiVoid(`/teams/${slug}/invitations/${invitationId}`, { method: "DELETE" }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["team-invitations", variables.slug] })
    },
  })
}

export function useMyInvitations() {
  return useQuery({
    queryKey: ["my-invitations"],
    queryFn: () => api<TeamInvitation[]>("/invitations"),
    refetchInterval: 30_000,
  })
}

export function useAcceptInvitation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (invitationId: number) =>
      apiVoid(`/invitations/${invitationId}/accept`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["my-invitations"] })
      qc.invalidateQueries({ queryKey: ["teams"] })
      qc.invalidateQueries({ queryKey: ["team-invitations"] })
      qc.invalidateQueries({ queryKey: ["team"] })
    },
  })
}

export function useDeclineInvitation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (invitationId: number) =>
      apiVoid(`/invitations/${invitationId}/decline`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["my-invitations"] })
      qc.invalidateQueries({ queryKey: ["team-invitations"] })
    },
  })
}

export function useUpdateMemberRole() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ slug, userId, role }: { slug: string; userId: number; role: TeamRole }) =>
      apiVoid(`/teams/${slug}/members/${userId}`, {
        method: "PUT",
        body: JSON.stringify({ role }),
      }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["team", variables.slug] })
    },
  })
}

export function useRemoveMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ slug, userId }: { slug: string; userId: number }) =>
      apiVoid(`/teams/${slug}/members/${userId}`, { method: "DELETE" }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["team", variables.slug] })
      qc.invalidateQueries({ queryKey: ["teams"] })
    },
  })
}

export function useSearchUsers(query: string) {
  return useQuery({
    queryKey: ["users-search", query],
    queryFn: () => api<UserSearchResult[]>(`/users/search?q=${encodeURIComponent(query)}`),
    enabled: query.length >= 2,
    staleTime: 30_000,
  })
}
