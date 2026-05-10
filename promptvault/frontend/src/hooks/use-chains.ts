// Phase 16: TanStack Query hooks для Prompt Chains.
// Паттерны invalidate скопированы с use-collections.ts.

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"

import { api, apiVoid } from "@/api/client"
import type {
  Chain,
  ChainConditions,
  ChainDetail,
  ChainExecution,
  ChainExecutionListResponse,
  ChainListResponse,
  ChainStep,
  ChainStepType,
  VariableMapping,
} from "@/api/types"

const chainsKey = (filters: { teamId?: number | null; limit: number; offset: number }) =>
  ["chains", filters] as const
const chainKey = (id: number) => ["chain", id] as const
const executionKey = (id: number) => ["chain-execution", id] as const

export function useChains(opts: { teamId?: number | null; limit?: number; offset?: number } = {}) {
  const limit = opts.limit ?? 20
  const offset = opts.offset ?? 0
  return useQuery({
    queryKey: chainsKey({ teamId: opts.teamId ?? null, limit, offset }),
    queryFn: () => {
      const params = new URLSearchParams()
      if (opts.teamId) params.set("team_id", String(opts.teamId))
      params.set("limit", String(limit))
      params.set("offset", String(offset))
      return api<ChainListResponse>(`/chains?${params.toString()}`)
    },
  })
}

export function useChain(id: number) {
  return useQuery({
    queryKey: chainKey(id),
    queryFn: () => api<ChainDetail>(`/chains/${id}`),
    enabled: id > 0,
  })
}

export function useCreateChain() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; description?: string; team_id?: number | null }) =>
      api<Chain>("/chains", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["chains"] })
    },
  })
}

export function useUpdateChain() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number; name?: string; description?: string }) =>
      api<Chain>(`/chains/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["chains"] })
      qc.invalidateQueries({ queryKey: chainKey(vars.id) })
    },
  })
}

export function useDeleteChain() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiVoid(`/chains/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["chains"] })
    },
  })
}

export function useAddStep(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: {
      /** Обязателен для prompt-шагов; для fork-шагов не передаётся. */
      prompt_id?: number
      name?: string
      variable_mapping?: VariableMapping
      manual_checkpoint?: boolean
      step_type?: ChainStepType
      conditions?: ChainConditions
      /** Куда вставить (Phase 16 v3, взаимоисключающие):
       *  - after_step_id: после указанного prompt-шага
       *  - parent_fork_id + branch_index: как первый шаг ветки fork-шага
       *  - ничего: tail-mode — в конец главной линии */
      after_step_id?: number
      parent_fork_id?: number
      branch_index?: number
    }) => {
      // conditions включаем в body только когда они есть, чтобы backend не сохранил
      // JSONB-литерал "null" для обычных prompt-шагов.
      const body: Record<string, unknown> = {
        name: data.name ?? "",
        variable_mapping: data.variable_mapping ?? {},
        manual_checkpoint: data.manual_checkpoint ?? false,
        step_type: data.step_type ?? "prompt",
      }
      if (data.prompt_id !== undefined) body.prompt_id = data.prompt_id
      if (data.conditions) body.conditions = data.conditions
      if (data.after_step_id !== undefined) body.after_step_id = data.after_step_id
      if (data.parent_fork_id !== undefined) body.parent_fork_id = data.parent_fork_id
      if (data.branch_index !== undefined) body.branch_index = data.branch_index
      return api<ChainStep>(`/chains/${chainId}/steps`, {
        method: "POST",
        body: JSON.stringify(body),
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: chainKey(chainId) })
      qc.invalidateQueries({ queryKey: ["chains"] })
    },
  })
}

export function useUpdateStep(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      stepId,
      ...data
    }: {
      stepId: number
      name?: string
      variable_mapping?: VariableMapping
      manual_checkpoint?: boolean
      step_type?: ChainStepType
      conditions?: ChainConditions
    }) => {
      const body: Record<string, unknown> = {
        name: data.name ?? "",
        variable_mapping: data.variable_mapping,
        manual_checkpoint: data.manual_checkpoint ?? false,
      }
      if (data.step_type) body.step_type = data.step_type
      if (data.conditions) body.conditions = data.conditions
      return api<ChainStep>(`/chains/${chainId}/steps/${stepId}`, {
        method: "PUT",
        body: JSON.stringify(body),
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: chainKey(chainId) })
    },
  })
}

export function useRemoveStep(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (stepId: number) => apiVoid(`/chains/${chainId}/steps/${stepId}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: chainKey(chainId) })
    },
  })
}

export function useReorderSteps(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (stepIds: number[]) =>
      api<ChainDetail>(`/chains/${chainId}/reorder`, {
        method: "POST",
        body: JSON.stringify({ step_ids: stepIds }),
      }),
    onSuccess: (data) => {
      qc.setQueryData(chainKey(chainId), data)
    },
  })
}

export function useMoveStepUp(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (stepId: number) =>
      apiVoid(`/chains/${chainId}/steps/${stepId}/move-up`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: chainKey(chainId) })
    },
  })
}

export function useMoveStepDown(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (stepId: number) =>
      apiVoid(`/chains/${chainId}/steps/${stepId}/move-down`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: chainKey(chainId) })
    },
  })
}

export function useStartExecution(chainId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (initialVars: Record<string, string> = {}) =>
      api<ChainExecution>(`/chains/${chainId}/executions`, {
        method: "POST",
        body: JSON.stringify({ initial_vars: initialVars }),
      }),
    onSuccess: (exec) => {
      qc.setQueryData(executionKey(exec.id), exec)
    },
  })
}

export function useExecution(id: number) {
  return useQuery({
    queryKey: executionKey(id),
    queryFn: () => api<ChainExecution>(`/executions/${id}`),
    enabled: id > 0,
  })
}

/** Список последних запусков цепочки. RBAC — read-access к chain (как GetByID). */
export function useChainExecutions(chainId: number, limit = 50) {
  return useQuery({
    queryKey: ["chain-executions", chainId, limit] as const,
    queryFn: () => api<ChainExecutionListResponse>(`/chains/${chainId}/executions?limit=${limit}`),
    enabled: chainId > 0,
  })
}

export function useAdvanceStep(executionId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { stepOutput: string; chosenBranchIndex?: number }) => {
      const body: Record<string, unknown> = { step_output: input.stepOutput }
      if (input.chosenBranchIndex !== undefined) body.chosen_branch_index = input.chosenBranchIndex
      return api<ChainExecution>(`/executions/${executionId}/advance`, {
        method: "POST",
        body: JSON.stringify(body),
      })
    },
    onSuccess: (exec) => {
      qc.setQueryData(executionKey(exec.id), exec)
    },
  })
}
