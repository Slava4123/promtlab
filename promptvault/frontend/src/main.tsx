import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { initSentry } from '@/lib/sentry'

// Init Sentry до рендера — иначе ранние ошибки не будут пойманы.
// Noop если VITE_SENTRY_ENABLED !== 'true'.
initSentry()

// Сбрасываем флаг chunk-reload при каждом новом mount'е: если приложение
// успешно загрузилось — старые chunks больше не проблема, и при следующем
// деплое ErrorBoundary снова сможет один раз перезагрузить страницу.
sessionStorage.removeItem("promptvault_chunk_reload_attempt")

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
