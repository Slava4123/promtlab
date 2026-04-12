import { useMemo } from "react"
import { Puzzle, ExternalLink, Zap, Lock, Shield } from "lucide-react"
import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"

// Extension URL — заменить на реальный после CWS publish.
// TODO(cws-launch): подменить на `https://chromewebstore.google.com/detail/<EXTENSION_ID>` когда расширение пройдёт review.
const CHROME_WEB_STORE_URL = "" // пусто = кнопка в "Скоро" состоянии
const PUBLIC_EXTENSION_RELEASES_URL =
  "https://github.com/slava4123/promtlab/releases" // fallback для early adopters через unpacked dev mode

interface BrowserInfo {
  name: string
  isChromium: boolean
  isFirefox: boolean
  isSafari: boolean
}

function detectBrowser(): BrowserInfo {
  if (typeof navigator === "undefined") {
    return { name: "unknown", isChromium: false, isFirefox: false, isSafari: false }
  }
  const ua = navigator.userAgent
  const isEdge = /Edg\//.test(ua)
  const isChrome = /Chrome\//.test(ua) && !isEdge
  const isFirefox = /Firefox\//.test(ua)
  const isSafari = /Safari\//.test(ua) && !isChrome && !isEdge

  if (isEdge) return { name: "Edge", isChromium: true, isFirefox: false, isSafari: false }
  if (isChrome) return { name: "Chrome", isChromium: true, isFirefox: false, isSafari: false }
  if (isFirefox) return { name: "Firefox", isChromium: false, isFirefox: true, isSafari: false }
  if (isSafari) return { name: "Safari", isChromium: false, isFirefox: false, isSafari: true }
  return { name: ua, isChromium: false, isFirefox: false, isSafari: false }
}

export function ExtensionPromoSection() {
  const browser = useMemo(() => detectBrowser(), [])
  const isLive = Boolean(CHROME_WEB_STORE_URL)

  return (
    <div className="rounded-xl border border-border bg-card p-5 overflow-hidden">
      <div className="mb-4 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Puzzle className="h-4 w-4 text-brand-muted-foreground" />
          <h2 className="text-sm font-semibold text-foreground">Chrome-расширение</h2>
        </div>
        {isLive ? null : (
          <span className="rounded-full border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-500">
            Скоро
          </span>
        )}
      </div>

      <p className="text-xs text-muted-foreground mb-4">
        Боковая панель в браузере, которая вставляет ваши промпты прямо в ChatGPT, Claude, Gemini и Perplexity
        одним кликом. Поддерживает {'{{переменные}}'}, поиск (⌘K), закреплённые и избранные промпты.
      </p>

      <ul className="mb-4 space-y-1.5 text-xs text-muted-foreground">
        <li className="flex items-start gap-2">
          <Zap className="mt-0.5 h-3 w-3 shrink-0 text-brand" />
          <span>Вставка через горячую клавишу (⌘⇧K) без переключения вкладок</span>
        </li>
        <li className="flex items-start gap-2">
          <Lock className="mt-0.5 h-3 w-3 shrink-0 text-brand" />
          <span>
            Подключение через ваш API-ключ — промпты никогда не уходят на сторонние сервера
          </span>
        </li>
        <li className="flex items-start gap-2">
          <Shield className="mt-0.5 h-3 w-3 shrink-0 text-brand" />
          <span>
            Открытый код, оффлайн-кэш, отсутствие аналитики и трекеров сторонних сервисов
          </span>
        </li>
      </ul>

      {/* Кнопки в зависимости от браузера и статуса публикации */}
      <div className="flex flex-col gap-2 sm:flex-row">
        {browser.isChromium || !browser.name ? (
          <Button
            variant={isLive ? "default" : "outline"}
            disabled={!isLive}
            className="flex-1"
            onClick={() => {
              if (isLive) {
                window.open(CHROME_WEB_STORE_URL, "_blank", "noopener,noreferrer")
              }
            }}
          >
            <Puzzle className="mr-2 h-4 w-4" />
            {isLive
              ? `Установить в ${browser.name === "Edge" ? "Edge" : "Chrome"}`
              : `Установить в ${browser.name === "Edge" ? "Edge" : "Chrome"} — скоро`}
            {isLive ? <ExternalLink className="ml-2 h-3 w-3" /> : null}
          </Button>
        ) : null}

        {browser.isFirefox ? (
          <Button variant="outline" disabled className="flex-1">
            Firefox — в разработке
          </Button>
        ) : null}

        {browser.isSafari ? (
          <Button variant="outline" disabled className="flex-1">
            Safari не поддерживается
          </Button>
        ) : null}
      </div>

      {!isLive ? (
        <p className="mt-3 text-[11px] text-muted-foreground">
          Идёт подготовка к публикации в Chrome Web Store. Ранний доступ для разработчиков —{" "}
          <a
            href={PUBLIC_EXTENSION_RELEASES_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="text-brand hover:underline"
          >
            последний релиз на GitHub
          </a>
          .{" "}
          <Link to="/legal/extension-privacy" className="text-brand hover:underline">
            Политика конфиденциальности
          </Link>
          .
        </p>
      ) : (
        <p className="mt-3 text-[11px] text-muted-foreground">
          <Link to="/legal/extension-privacy" className="text-brand hover:underline">
            Политика конфиденциальности расширения
          </Link>
        </p>
      )}
    </div>
  )
}
