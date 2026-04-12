import { create } from "zustand"
import { devtools } from "zustand/middleware"

interface AdminState {
  // totpVerifiedAt — время последней успешной TOTP-проверки. Используется
  // для UX hint'ов («скоро потребуется перепроверка»), но real check
  // происходит на backend при каждом destructive action через sudo mode.
  totpVerifiedAt: number | null
  markTOTPVerified: () => void
  clear: () => void
}

export const useAdminStore = create<AdminState>()(
  devtools(
    (set) => ({
      totpVerifiedAt: null,
      markTOTPVerified: () => set({ totpVerifiedAt: Date.now() }),
      clear: () => set({ totpVerifiedAt: null }),
    }),
    { name: "admin-store" },
  ),
)
