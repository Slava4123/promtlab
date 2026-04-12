/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly WXT_API_BASE?: string;
  readonly VITE_SENTRY_ENABLED?: string;
  readonly VITE_SENTRY_DSN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
