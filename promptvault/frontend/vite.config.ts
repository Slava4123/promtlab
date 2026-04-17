import path from "path"
import { defineConfig } from "vitest/config"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

// Vite по умолчанию отдаёт статику из public/ с Content-Type без charset.
// Для UTF-8 текстов (llms.txt, *.md, robots.txt) браузер угадывает кодировку
// и для кириллицы получает mojibake. Принудительно ставим charset=utf-8.
const utf8TextPlugin = {
  name: "utf8-text-charset",
  configureServer(server: { middlewares: { use: (fn: (req: { url?: string }, res: { setHeader: (k: string, v: string) => void }, next: () => void) => void) => void } }) {
    server.middlewares.use((req, res, next) => {
      const url = req.url?.split("?")[0] ?? ""
      if (url.endsWith(".txt") || url.endsWith(".md")) {
        const type = url.endsWith(".md") ? "text/markdown" : "text/plain"
        res.setHeader("Content-Type", `${type}; charset=utf-8`)
      }
      next()
    })
  },
}

export default defineConfig({
  plugins: [react(), tailwindcss(), utf8TextPlugin],
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
