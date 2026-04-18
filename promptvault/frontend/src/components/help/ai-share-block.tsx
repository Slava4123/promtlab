import { useState } from "react"
import { Copy, Check, ExternalLink, Sparkles } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type Props = {
  /** Абсолютный или site-relative URL к markdown-зеркалу страницы. */
  mdUrl: string
  /** Что AI должен помочь сделать (вставляется в промпт после «Помоги мне …»). */
  topic: string
  /** Компактный режим — одна узкая карточка для legal-страниц. */
  compact?: boolean
  className?: string
}

const SITE_ORIGIN = "https://promtlabs.ru"

function buildPrompt(topic: string, mdUrl: string): string {
  const absoluteUrl = mdUrl.startsWith("http") ? mdUrl : `${SITE_ORIGIN}${mdUrl}`
  return `Помоги мне ${topic}. Открой документ ${absoluteUrl} и проведи меня по шагам — задавай уточняющие вопросы, если чего-то не хватает.`
}

export function AIShareBlock({ mdUrl, topic, compact, className }: Props) {
  const [copied, setCopied] = useState(false)
  const prompt = buildPrompt(topic, mdUrl)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(prompt)
      setCopied(true)
      toast.success("Промпт скопирован — вставьте в любой AI-чат")
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error("Не удалось скопировать")
    }
  }

  const claudeUrl = `https://claude.ai/new?q=${encodeURIComponent(prompt)}`
  const chatgptUrl = `https://chatgpt.com/?q=${encodeURIComponent(prompt)}`

  return (
    <div
      className={cn(
        "rounded-xl border border-border bg-card/50",
        compact ? "p-3" : "p-4 sm:p-5",
        className,
      )}
    >
      <div className={cn("flex items-start gap-3", compact && "gap-2")}>
        <div
          className={cn(
            "flex shrink-0 items-center justify-center rounded-full bg-brand-muted text-brand-muted-foreground",
            compact ? "h-7 w-7" : "h-9 w-9",
          )}
        >
          <Sparkles className={cn(compact ? "h-3.5 w-3.5" : "h-4 w-4")} />
        </div>
        <div className="min-w-0 flex-1">
          <h3
            className={cn(
              "font-semibold text-foreground",
              compact ? "text-sm" : "text-base",
            )}
          >
            Не разобрался? Спроси AI
          </h3>
          <p
            className={cn(
              "text-muted-foreground",
              compact ? "mt-0.5 text-xs" : "mt-1 text-sm",
            )}
          >
            AI прочитает{" "}
            <a
              href={mdUrl}
              target="_blank"
              rel="noreferrer"
              className="underline decoration-dotted hover:text-foreground"
            >
              инструкцию
            </a>{" "}
            и проведёт по шагам — выберите чат или скопируйте готовый промпт.
          </p>

          <div className={cn("mt-3 flex flex-wrap gap-2", compact && "mt-2")}>
            <Button
              variant="default"
              size={compact ? "xs" : "sm"}
              onClick={handleCopy}
            >
              {copied ? <Check /> : <Copy />}
              {copied ? "Скопировано" : "Скопировать промпт"}
            </Button>
            <Button
              variant="outline"
              size={compact ? "xs" : "sm"}
              nativeButton={false}
              render={<a href={claudeUrl} target="_blank" rel="noreferrer" />}
            >
              Открыть в Claude
              <ExternalLink data-icon="inline-end" />
            </Button>
            <Button
              variant="outline"
              size={compact ? "xs" : "sm"}
              nativeButton={false}
              render={<a href={chatgptUrl} target="_blank" rel="noreferrer" />}
            >
              Открыть в ChatGPT
              <ExternalLink data-icon="inline-end" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
