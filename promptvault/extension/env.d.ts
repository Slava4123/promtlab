/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly WXT_API_BASE?: string;
  readonly WXT_SENTRY_DSN?: string;
  readonly WXT_SENTRY_ENABLED?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
