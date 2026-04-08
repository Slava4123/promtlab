import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { initSentry } from '@/lib/sentry'

// Init Sentry до рендера — иначе ранние ошибки не будут пойманы.
// Noop если VITE_SENTRY_ENABLED !== 'true'.
initSentry()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
