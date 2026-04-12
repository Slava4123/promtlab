import { useQuery } from "@tanstack/react-query"
import { fetchAudit, type AuditFilter } from "@/api/admin/audit"
import { fetchHealth } from "@/api/admin/health"

export function useAdminAudit(filter: AuditFilter) {
  return useQuery({
    queryKey: ["admin", "audit", filter],
    queryFn: () => fetchAudit(filter),
    staleTime: 15_000,
  })
}

export function useAdminHealth() {
  return useQuery({
    queryKey: ["admin", "health"],
    queryFn: () => fetchHealth(),
    refetchInterval: 30_000,
  })
}
