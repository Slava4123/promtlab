import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from '../../components/app';
import { initSentry } from '../../lib/sentry';

initSentry({
  enabled: import.meta.env.WXT_SENTRY_DSN ? true : false,
  release: chrome.runtime.getManifest?.().version,
  dsn: import.meta.env.WXT_SENTRY_DSN,
});

const root = document.getElementById('root');
if (root) {
  ReactDOM.createRoot(root).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );
}
