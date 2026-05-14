import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from '../../components/app';
import { initSentry } from '../../lib/sentry';

initSentry({
  enabled: import.meta.env.WXT_SENTRY_DSN ? true : false,
  release: chrome.runtime.getManifest?.().version,
  dsn: import.meta.env.WXT_SENTRY_DSN,
});

// Чистим chunk-reload flag после успешного mount —
// если стартанули нормально, значит recovery сработало, и при следующей
// ошибке можно снова auto-recover (см. ErrorBoundary).
try {
  sessionStorage.removeItem('pv.chunkErrorReloaded');
} catch {
  // ignore
}

const root = document.getElementById('root');
if (root) {
  ReactDOM.createRoot(root).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );
}
