import { Navigate, useNavigate } from "react-router-dom"
import { SettingsView } from "../components/settings-view"
import { useSettings } from "../hooks/use-settings"
import { Loader2 } from "lucide-react"

// Главная страница настроек (текущий SettingsView). В Phase 5 заменим
// на full /settings/* sub-routes (profile/security/etc.).
export function SettingsPage() {
  const navigate = useNavigate()
  const settings = useSettings()

  if (!settings) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  // Декларативный редирект — без вызова navigate() во время render
  // (последнее вызывает infinite loop и Chrome throttle navigation).
  if (!settings.apiKey) {
    return <Navigate to="/sign-in" replace />
  }

  return (
    <SettingsView
      apiKey={settings.apiKey}
      apiBase={settings.apiBase}
      theme={settings.theme}
      onBack={() => navigate("/")}
    />
  )
}
