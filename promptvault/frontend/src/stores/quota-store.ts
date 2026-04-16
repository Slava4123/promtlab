import { create } from "zustand"
import { devtools } from "zustand/middleware"

export interface QuotaPayload {
  quotaType: string
  message: string
  used?: number
  limit?: number
  plan?: string
}

interface QuotaState {
  open: boolean
  quotaType: string | null
  message: string | null
  used: number | null
  limit: number | null
  plan: string | null
  show: (payload: QuotaPayload) => void
  dismiss: () => void
}

export const useQuotaStore = create<QuotaState>()(
  devtools(
    (set) => ({
      open: false,
      quotaType: null,
      message: null,
      used: null,
      limit: null,
      plan: null,
      show: ({ quotaType, message, used, limit, plan }) =>
        set({
          open: true,
          quotaType,
          message,
          used: used ?? null,
          limit: limit ?? null,
          plan: plan ?? null,
        }),
      dismiss: () =>
        set({ open: false, quotaType: null, message: null, used: null, limit: null, plan: null }),
    }),
    { name: "quota-store" },
  ),
)
