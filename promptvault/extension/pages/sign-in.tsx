import { useNavigate } from "react-router-dom"
import { useEffect } from "react"
import { ApiKeySetup } from "../components/api-key-setup"
import { useSettings } from "../hooks/use-settings"

// Страница sign-in (API-key setup для текущего flow). В Phase 6 расширим
// до OAuth + email/password.
export function SignInPage() {
  const navigate = useNavigate()
  const settings = useSettings()

  useEffect(() => {
    if (settings?.apiKey) navigate("/", { replace: true })
  }, [settings?.apiKey, navigate])

  if (!settings) return null
  return <ApiKeySetup initialBase={settings.apiBase} />
}
