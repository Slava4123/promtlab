import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import { useWorkspaceStore } from "../stores/workspace-store"
import type {
  ChainListResponse,
  ChainDetail,
  ChainExecution,
  ChainExecutionListResponse,
} from "../lib/types"

export function useChains() {
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  return useQuery<ChainListResponse>({
    queryKey: [...qk.chains, teamId],
    queryFn: () => sendBg({ type: "api.listChains", teamId }),
    staleTime: 60_000,
  })
}

export function useChain(id: number | null) {
  return useQuery<ChainDetail>({
    queryKey: id ? qk.chain(id) : ["chain", "none"],
    queryFn: () => {
      if (!id) throw new Error("no chain id")
      return sendBg({ type: "api.getChain", id })
    },
    enabled: id !== null,
    staleTime: 30_000,
  })
}

export function useStartExecution() {
  const qc = useQueryClient()
  return useMutation<ChainExecution, Error, { chainId: number; initialVars: Record<string, string> }>(
    {
      mutationFn: ({ chainId, initialVars }) =>
        sendBg({ type: "api.startChainExecution", chainId, initialVars }),
      onSuccess: (exec) => {
        void qc.invalidateQueries({ queryKey: qk.executions(exec.chain_id) })
        qc.setQueryData(qk.execution(exec.id), exec)
      },
    },
  )
}

export function useExecution(execId: number | null) {
  return useQuery<ChainExecution>({
    queryKey: execId ? qk.execution(execId) : ["execution", "none"],
    queryFn: () => {
      if (!execId) throw new Error("no exec id")
      return sendBg({ type: "api.getExecution", execId })
    },
    enabled: execId !== null,
    staleTime: 10_000,
  })
}

export function useAdvanceStep(execId: number) {
  const qc = useQueryClient()
  return useMutation<
    ChainExecution,
    Error,
    { stepOutput: string; chosenBranchIndex?: number }
  >({
    mutationFn: ({ stepOutput, chosenBranchIndex }) =>
      sendBg({ type: "api.advanceChainStep", execId, stepOutput, chosenBranchIndex }),
    onSuccess: (exec) => {
      qc.setQueryData(qk.execution(execId), exec)
      void qc.invalidateQueries({ queryKey: qk.executions(exec.chain_id) })
    },
  })
}

export function useExecutions(chainId: number | null) {
  return useQuery<ChainExecutionListResponse>({
    queryKey: chainId ? qk.executions(chainId) : ["executions", "none"],
    queryFn: () => {
      if (!chainId) throw new Error("no chain id")
      return sendBg({ type: "api.listExecutions", chainId })
    },
    enabled: chainId !== null,
    staleTime: 30_000,
  })
}
