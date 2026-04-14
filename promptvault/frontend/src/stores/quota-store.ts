import { create } from "zustand"
import { devtools } from "zustand/middleware"

interface QuotaState {
  open: boolean
  quotaType: string | null
  message: string | null
  show: (quotaType: string, message: string) => void
  dismiss: () => void
}

export const useQuotaStore = create<QuotaState>()(
  devtools(
    (set) => ({
      open: false,
      quotaType: null,
      message: null,
      show: (quotaType, message) => set({ open: true, quotaType, message }),
      dismiss: () => set({ open: false, quotaType: null, message: null }),
    }),
    { name: "quota-store" },
  ),
)
