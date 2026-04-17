import { useState } from "react"
import { Link } from "react-router-dom"
import { Plug, ArrowRight, Copy, Check } from "lucide-react"
import { toast } from "sonner"

import { APIKeysSection } from "@/components/settings/api-keys-section"
import { ExtensionPromoSection } from "@/components/settings/extension-promo-section"
import { SectionHeader } from "./_section-header"

const HOST_PLACEHOLDER = "https://promtlabs.ru"
const KEY_PLACEHOLDER = "pvlt_ваш_ключ"
const MCP_COMMAND = `claude mcp add promptvault --transport http ${HOST_PLACEHOLDER}/mcp \\
  --header "Authorization: Bearer ${KEY_PLACEHOLDER}"`

export default function SettingsIntegrationsPage() {
  return (
    <section className="space-y-5">
      <SectionHeader
        title="Интеграции"
        description="Подключите ПромтЛаб к AI-клиентам, IDE и браузеру"
      />

      <MCPCard />
      <ExtensionPromoSection />
      <APIKeysSection />
    </section>
  )
}

function MCPCard() {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(MCP_COMMAND)
      setCopied(true)
      toast.success("Скопировано")
      setTimeout(() => setCopied(false), 1500)
    } catch {
      toast.error("Не удалось скопировать")
    }
  }

  return (
    <div className="rounded-xl border border-border bg-card p-5 overflow-hidden">
      <div className="mb-4 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Plug className="h-4 w-4 text-brand-muted-foreground" />
          <h2 className="text-sm font-semibold text-foreground">MCP-сервер для Claude / Cursor / Windsurf</h2>
        </div>
      </div>

      <p className="text-xs text-muted-foreground mb-4">
        ПромтЛаб реализует Model Context Protocol — ваши промпты, коллекции и теги становятся доступны
        прямо в Claude Code, Claude Desktop, Cursor, Windsurf и других совместимых клиентах.
      </p>

      <div className="overflow-hidden rounded-lg border border-border bg-muted/40">
        <div className="flex items-center justify-between border-b border-border px-3 py-1.5">
          <span className="text-[0.7rem] uppercase tracking-wider text-muted-foreground">claude code</span>
          <button
            onClick={handleCopy}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-[0.72rem] text-muted-foreground hover:bg-foreground/[0.04] hover:text-foreground"
            aria-label="Скопировать команду"
          >
            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
            {copied ? "Готово" : "Копировать"}
          </button>
        </div>
        <pre className="overflow-x-auto px-3 py-3 text-[0.78rem] leading-relaxed text-foreground">
          <code>{MCP_COMMAND}</code>
        </pre>
      </div>
      <p className="mt-2 text-[11px] text-muted-foreground">
        Подставьте свой ключ из раздела «API-ключи» ниже.
      </p>

      <div className="mt-4">
        <Link
          to="/help/mcp"
          className="inline-flex items-center gap-1 text-[0.78rem] text-brand-muted-foreground underline-offset-4 hover:underline"
        >
          Подробная инструкция со всеми клиентами и командами
          <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>
    </div>
  )
}
