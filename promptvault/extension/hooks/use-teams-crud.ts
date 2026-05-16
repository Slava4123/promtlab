import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import type { Team, TeamDetail, TeamDTO } from "../lib/types"
import type { CreateTeamBody } from "../lib/api"

export function useTeams() {
  return useQuery<TeamDTO[]>({
    queryKey: qk.teams,
    queryFn: () => sendBg({ type: "api.listTeams" }),
    staleTime: 60_000,
  })
}

export function useTeam(slug: string | null) {
  return useQuery<TeamDetail>({
    queryKey: slug ? qk.team(slug) : ["team", "none"],
    queryFn: () => {
      if (!slug) throw new Error("no team slug")
      return sendBg({ type: "api.getTeam", slug })
    },
    enabled: slug !== null,
    staleTime: 30_000,
  })
}

export function useCreateTeam() {
  const qc = useQueryClient()
  return useMutation<Team, Error, CreateTeamBody>({
    mutationFn: (body) => sendBg({ type: "api.createTeam", body }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.teams })
      // Teams count в Подписке растёт.
      void qc.invalidateQueries({ queryKey: qk.usage })
    },
  })
}

export function useUpdateTeam(slug: string) {
  const qc = useQueryClient()
  return useMutation<Team, Error, Partial<CreateTeamBody>>({
    mutationFn: (body) => sendBg({ type: "api.updateTeam", slug, body }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.teams })
      void qc.invalidateQueries({ queryKey: qk.team(slug) })
    },
  })
}

export function useDeleteTeam() {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, string>({
    mutationFn: (slug) => sendBg({ type: "api.deleteTeam", slug }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.teams })
      // Teams count уменьшается.
      void qc.invalidateQueries({ queryKey: qk.usage })
    },
  })
}

export function useInviteTeamMember(slug: string) {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, { email: string; role: "editor" | "viewer" }>({
    mutationFn: ({ email, role }) =>
      sendBg({ type: "api.inviteTeamMember", slug, email, role }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.team(slug) }),
  })
}

export function useRemoveTeamMember(slug: string) {
  const qc = useQueryClient()
  return useMutation<{ ok: true }, Error, number>({
    mutationFn: (memberId) => sendBg({ type: "api.removeTeamMember", slug, memberId }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.team(slug) }),
  })
}

export function useUpdateTeamMemberRole(slug: string) {
  const qc = useQueryClient()
  return useMutation<
    { ok: true },
    Error,
    { memberId: number; role: "editor" | "viewer" }
  >({
    mutationFn: ({ memberId, role }) =>
      sendBg({ type: "api.updateTeamMemberRole", slug, memberId, role }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: qk.team(slug) }),
  })
}
