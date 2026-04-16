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
    // Разрешаем cloudflare/ngrok tunnels для dev-тестирования платежей
    // (T-Bank требует публичный https Success/Fail URL).
    allowedHosts: [".trycloudflare.com", ".ngrok-free.app", ".ngrok.io"],
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
    rollupOptions: {
      output: {
        // Code splitting: выносим тяжёлые vendor-пакеты в отдельные chunks.
        // Цель — landing/login-страницы тянут только core, остальное грузится
        // при навигации в app (P-7). Замеряй bundle через `npm run build`.
        // Vite 8 / Rolldown требует ManualChunksFunction вместо объекта.
        manualChunks(id) {
          if (!id.includes("node_modules")) return undefined
          if (id.includes("/react-router-dom/") || id.includes("/react-router/") ||
              id.includes("/react/") || id.includes("/react-dom/") ||
              id.includes("/scheduler/")) {
            return "vendor-react"
          }
          if (id.includes("/@tanstack/react-query")) return "vendor-query"
          if (id.includes("/@sentry/")) return "vendor-sentry"
          if (id.includes("/react-hook-form/") || id.includes("/@hookform/") || id.includes("/zod/")) {
            return "vendor-forms"
          }
          if (id.includes("/@base-ui/") || id.includes("/cmdk/") || id.includes("/sonner/") ||
              id.includes("/class-variance-authority/") || id.includes("/clsx/") ||
              id.includes("/tailwind-merge/") || id.includes("/tw-animate-css/")) {
            return "vendor-ui"
          }
          if (id.includes("/lucide-react/")) return "vendor-icons"
          if (id.includes("/react-markdown/") || id.includes("/remark-") || id.includes("/rehype-")) {
            return "vendor-markdown"
          }
          if (id.includes("/react-diff-viewer-continued/") || id.includes("/diff/")) {
            return "vendor-diff"
          }
          return undefined
        },
      },
    },
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    css: false,
  },
})
