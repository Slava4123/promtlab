// M-7: Реферальный код — capture и хранение на фронте.

const COOKIE_NAME = "pv_ref"
const COOKIE_MAX_AGE_DAYS = 30

// captureReferralFromURL читает ?ref=XXXXXXXX из window.location и сохраняет
// в cookie на 30 дней. Валидация: 8 символов alphanumeric (под серверный
// формат base32). Вызывается при загрузке приложения.
export function captureReferralFromURL(): void {
  if (typeof window === "undefined") return
  const params = new URLSearchParams(window.location.search)
  const ref = (params.get("ref") ?? "").toUpperCase().trim()
  if (!/^[A-Z0-9]{8}$/.test(ref)) return
  const maxAge = COOKIE_MAX_AGE_DAYS * 24 * 60 * 60
  document.cookie = `${COOKIE_NAME}=${ref}; path=/; max-age=${maxAge}; SameSite=Lax`
}

// readReferralCookie возвращает сохранённый реферальный код или "".
// Используется на Register и редиректе на OAuth-провайдер.
export function readReferralCookie(): string {
  if (typeof document === "undefined") return ""
  const parts = document.cookie.split(";").map((s) => s.trim())
  for (const p of parts) {
    const [k, v] = p.split("=")
    if (k === COOKIE_NAME && v) {
      const code = decodeURIComponent(v).toUpperCase()
      if (/^[A-Z0-9]{8}$/.test(code)) return code
    }
  }
  return ""
}

// clearReferralCookie — вызываем после успешного применения (registration/OAuth complete),
// чтобы реф не "переносился" между разными юзерами на одной машине.
export function clearReferralCookie(): void {
  if (typeof document === "undefined") return
  document.cookie = `${COOKIE_NAME}=; path=/; max-age=0; SameSite=Lax`
}
