import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from '../../components/app';
import { initSentry } from '../../lib/sentry';

initSentry({
  enabled: import.meta.env.VITE_SENTRY_ENABLED === 'true',
  release: chrome.runtime.getManifest?.().version,
});

const root = document.getElementById('root');
if (root) {
  ReactDOM.createRoot(root).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );
}
