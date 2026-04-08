import path from "path"
import { defineConfig } from "vitest/config"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: process.env.VITE_API_URL || "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  build: {
    // "hidden" — source maps генерируются в dist/assets/*.map, но bundle
    // НЕ содержит ссылки //# sourceMappingURL. Это значит:
    // 1. Браузеры не загружают maps автоматически (защита от утечек исходников)
    // 2. sentry-cli загружает maps в GlitchTip по артефактам release
    // 3. GlitchTip матчит maps к stack traces через release + file name
    sourcemap: "hidden",
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    css: false,
  },
})
