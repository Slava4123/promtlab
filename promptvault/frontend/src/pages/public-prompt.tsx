import { useEffect } from "react"
import { useParams, Link } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import { Copy, Check, Loader2, Sparkles } from "lucide-react"
import { useState } from "react"
import { publicApi } from "@/api/client"
import { Button } from "@/components/ui/button"
import type { Prompt } from "@/api/types"

/**
 * /p/:slug — публичный SEO-индексируемый просмотр промпта.
 * Без авторизации. Цель: привлечение через органику + продуктовое демо.
 *
 * Без SSR → OpenGraph теги вставляем через document.title и meta в useEffect;
 * SEO-движки Yandex/Google рендерят JS и считывают их. Если нужен полноценный
 * SSR — миграция на Next/Astro отдельной задачей.
 */
function fetchPublic(slug: string): Promise<Prompt> {
  return publicApi<Prompt>(`/public/prompts/${encodeURIComponent(slug)}`)
}

export default function PublicPrompt() {
  const { slug = "" } = useParams<{ slug: string }>()
  const { data, isLoading, error } = useQuery({
    queryKey: ["public-prompt", slug],
    queryFn: () => fetchPublic(slug),
    retry: false,
  })
  const [copied, setCopied] = useState(false)

  // Meta-теги для OpenGraph. Set/reset при unmount — иначе title/desc
  // протекают между роутами.
  useEffect(() => {
    if (!data) return
    const prevTitle = document.title
    document.title = `${data.title} — ПромтЛаб`

    const description = data.content.length > 160 ? data.content.slice(0, 157) + "..." : data.content
    setMeta("description", description)
    setOG("og:title", data.title)
    setOG("og:description", description)
    setOG("og:type", "article")
    setOG("og:url", window.location.href)

    return () => {
      document.title = prevTitle
    }
  }, [data])

  const handleCopy = async () => {
    if (!data) return
    try {
      await navigator.clipboard.writeText(data.content)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      /* noop */
    }
  }

  if (isLoading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }
  if (error || !data) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-center">
        <h1 className="text-xl font-semibold">Промпт не найден</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Возможно, владелец снял публикацию или вы ошиблись в ссылке.
        </p>
        <Link to="/" className="mt-4 inline-block text-sm text-violet-400 underline">
          На главную
        </Link>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-12">
      <header className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight text-foreground">{data.title}</h1>
        {data.model && (
          <p className="mt-1 text-[0.8rem] text-muted-foreground">
            Модель: <span className="font-medium text-foreground">{data.model}</span>
          </p>
        )}
      </header>

      <pre className="whitespace-pre-wrap rounded-xl border border-border bg-card px-5 py-4 text-[0.9rem] leading-relaxed text-foreground">
        {data.content}
      </pre>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button size="sm" onClick={handleCopy}>
          {copied ? <Check className="mr-2 h-3.5 w-3.5" /> : <Copy className="mr-2 h-3.5 w-3.5" />}
          Скопировать промпт
        </Button>
        <Button asChild size="sm" variant="outline">
          <Link to="/sign-up">
            <Sparkles className="mr-2 h-3.5 w-3.5" />
            Сохранить в ПромтЛаб
          </Link>
        </Button>
      </div>

      <footer className="mt-12 border-t border-border pt-6 text-[0.75rem] text-muted-foreground">
        Опубликовано через{" "}
        <Link to="/" className="font-medium text-foreground underline">
          ПромтЛаб
        </Link>
        {" — хранилище AI-промптов с командной работой, AI-улучшением и MCP-интеграцией."}
      </footer>
    </div>
  )
}

function setMeta(name: string, content: string) {
  let el = document.head.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)
  if (!el) {
    el = document.createElement("meta")
    el.name = name
    document.head.appendChild(el)
  }
  el.content = content
}

function setOG(property: string, content: string) {
  let el = document.head.querySelector<HTMLMetaElement>(`meta[property="${property}"]`)
  if (!el) {
    el = document.createElement("meta")
    el.setAttribute("property", property)
    document.head.appendChild(el)
  }
  el.content = content
}
