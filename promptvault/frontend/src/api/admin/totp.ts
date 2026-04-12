import { api } from "@/api/client"
import type { TOTPEnrollResponse, TOTPStatusResponse } from "@/api/types"

export function totpStatus() {
  return api<TOTPStatusResponse>("/admin/totp/status")
}

export function totpEnroll() {
  return api<TOTPEnrollResponse>("/admin/totp/enroll", { method: "POST" })
}

export function totpConfirmEnrollment(code: string) {
  return api<{ confirmed: boolean }>("/admin/totp/verify-enrollment", {
    method: "POST",
    body: JSON.stringify({ code }),
  })
}

export function totpRegenBackupCodes() {
  return api<{ backup_codes: string[] }>("/admin/totp/backup-codes/regenerate", {
    method: "POST",
  })
}
