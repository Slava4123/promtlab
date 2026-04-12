import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  fetchAdminUsers,
  fetchAdminUserDetail,
  freezeUser,
  unfreezeUser,
  resetUserPassword,
  grantBadge,
  revokeBadge,
  type AdminUsersFilter,
} from "@/api/admin/users"

export function useAdminUsers(filter: AdminUsersFilter) {
  return useQuery({
    queryKey: ["admin", "users", filter],
    queryFn: () => fetchAdminUsers(filter),
    staleTime: 30_000,
  })
}

export function useAdminUserDetail(id: number) {
  return useQuery({
    queryKey: ["admin", "user", id],
    queryFn: () => fetchAdminUserDetail(id),
    enabled: id > 0,
  })
}

export function useFreezeUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => freezeUser(id),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ["admin", "users"] })
      qc.invalidateQueries({ queryKey: ["admin", "user", id] })
    },
  })
}

export function useUnfreezeUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => unfreezeUser(id),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ["admin", "users"] })
      qc.invalidateQueries({ queryKey: ["admin", "user", id] })
    },
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: ({ id, totpCode }: { id: number; totpCode: string }) =>
      resetUserPassword(id, totpCode),
  })
}

export function useGrantBadge() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, badgeId }: { userId: number; badgeId: string }) =>
      grantBadge(userId, badgeId),
    onSuccess: (_, { userId }) => {
      qc.invalidateQueries({ queryKey: ["admin", "user", userId] })
    },
  })
}

export function useRevokeBadge() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, badgeId, totpCode }: { userId: number; badgeId: string; totpCode: string }) =>
      revokeBadge(userId, badgeId, totpCode),
    onSuccess: (_, { userId }) => {
      qc.invalidateQueries({ queryKey: ["admin", "user", userId] })
    },
  })
}
