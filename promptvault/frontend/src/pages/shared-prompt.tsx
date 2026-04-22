import { useParams, Link } from "react-router-dom"
import { Copy, Check, ExternalLink, Loader2, Sparkles, BookOpen, Users, Zap } from "lucide-react"
import { useState } from "react"
import { toast, Toaster } from "sonner"

import { usePublicPrompt } from "@/hooks/use-share"
import { Button } from "@/components/ui/button"
import { ApiError } from "@/api/client"
import { PromptView } from "@/components/prompts/prompt-view"
import { BrandedHeader } from "@/components/teams/branded-header"

export default function SharedPrompt() {
  const { token } = useParams<{ token: string }>()
  const { data: prompt, isLoading, isError, error } = usePublicPrompt(token ?? "")
  const [copied, setCopied] = useState(false)

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-3">
          <div className="relative">
            <div className="absolute inset-0 animate-ping rounded-full bg-violet-500/20" />
            <Loader2 className="relative h-8 w-8 animate-spin text-violet-500" />
          </div>
          <span className="text-sm text-muted-foreground">Загрузка промпта...</span>
        </div>
      </div>
    )
  }

  const isNotFound = error instanceof ApiError && error.status === 404

  if (isNotFound || (!isError && !prompt)) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-background px-4">
        <div className="flex flex-col items-center gap-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-violet-500/10">
            <BookOpen className="h-8 w-8 text-violet-400" />
          </div>
          <div className="text-center">
            <h1 className="text-2xl font-bold text-foreground">Промпт не найден</h1>
            <p className="mt-2 max-w-sm text-muted-foreground">
              Ссылка недействительна или срок действия истёк
            </p>
          </div>
        </div>
        <Link to="/sign-up">
          <Button size="lg">Создать свою библиотеку промптов</Button>
        </Link>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-background px-4">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground">Не удалось загрузить промпт</h1>
          <p className="mt-2 text-muted-foreground">Попробуйте обновить страницу</p>
        </div>
        <Button onClick={() => window.location.reload()}>Обновить</Button>
      </div>
    )
  }

  if (!prompt) return null

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(prompt.content)
      setCopied(true)
      toast.success("Промпт скопирован", { description: "Вставьте в чат с AI" })
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error("clipboard write failed:", err)
      toast.error("Не удалось скопировать. Выделите текст и нажмите Ctrl+C")
    }
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="sticky top-0 z-10 border-b border-border/50 bg-background/80 backdrop-blur-lg">
        <div className="mx-auto flex h-14 max-w-4xl items-center justify-between px-4 sm:px-6">
          <Link to="/" className="flex items-center gap-2 text-lg font-bold text-foreground transition-colors hover:text-violet-400">
            <Sparkles className="h-5 w-5 text-violet-500" />
            ПромтЛаб
          </Link>
          <div className="flex items-center gap-2">
            <Link to="/sign-in">
              <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground">
                Войти
              </Button>
            </Link>
            <Link to="/sign-up">
              <Button size="sm" className="bg-violet-500 text-white hover:bg-violet-600">Регистрация</Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Phase 14 D: Branded header — только если Max-владелец настроил brand. */}
      {prompt.branding && (
        <div className="mx-auto max-w-4xl px-4 pt-6 sm:px-6">
          <BrandedHeader branding={prompt.branding} />
        </div>
      )}

      {/* Hero / Title Section */}
      <div className="border-b border-border/50 bg-gradient-to-b from-violet-500/[0.07] to-transparent">
        <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 sm:py-12">
          <div className="flex flex-col gap-4">
            {/* Author + Meta */}
            <div className="flex items-center gap-3">
              {prompt.author.avatar_url ? (
                <img
                  src={prompt.author.avatar_url}
                  alt={prompt.author.name}
                  className="h-10 w-10 rounded-full ring-2 ring-violet-500/20"
                />
              ) : (
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-violet-500/15 text-sm font-semibold text-violet-400 ring-2 ring-violet-500/20">
                  {prompt.author.name.charAt(0).toUpperCase()}
                </div>
              )}
              <div>
                <span className="text-sm font-medium text-foreground">{prompt.author.name}</span>
                <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <span>{new Date(prompt.created_at).toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" })}</span>
                  {prompt.model && (
                    <>
                      <span className="text-border">·</span>
                      <span className="rounded bg-violet-500/10 px-1.5 py-0.5 text-[10px] font-medium text-violet-400">
                        {prompt.model.split("/").pop()}
                      </span>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* Title */}
            <h1 className="text-2xl font-bold leading-tight text-foreground sm:text-3xl">
              {prompt.title}
            </h1>

            {/* Tags */}
            {prompt.tags.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {prompt.tags.map((tag) => (
                  <span
                    key={tag.name}
                    className="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
                    style={{
                      backgroundColor: tag.color + "15",
                      borderColor: tag.color + "30",
                      color: tag.color,
                    }}
                  >
                    {tag.name}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Content */}
      <main className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <div className="space-y-8">
          {/* Prompt Content Card */}
          <div className="group relative overflow-hidden rounded-xl border border-border/60 bg-card shadow-sm transition-shadow hover:shadow-md">
            {/* Gradient accent top */}
            <div className="h-1 bg-gradient-to-r from-violet-500 via-purple-500 to-fuchsia-500" />

            <div className="p-5 sm:p-6">
              <PromptView content={prompt.content} storageKey="shared-prompt-view" />
            </div>

            {/* Copy button */}
            <div className="flex items-center justify-between border-t border-border/40 bg-muted/30 px-5 py-3 sm:px-6">
              <span className="text-xs text-muted-foreground">
                {prompt.content.length} символов
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={handleCopy}
                className={`gap-2 transition-all ${copied ? "border-green-500/50 text-green-500" : ""}`}
              >
                {copied ? (
                  <>
                    <Check className="h-3.5 w-3.5" />
                    Скопировано
                  </>
                ) : (
                  <>
                    <Copy className="h-3.5 w-3.5" />
                    Скопировать промпт
                  </>
                )}
              </Button>
            </div>
          </div>

          {/* CTA Section */}
          <div className="relative overflow-hidden rounded-xl border border-violet-500/20 bg-gradient-to-br from-violet-500/10 via-purple-500/5 to-fuchsia-500/10 p-8 sm:p-10">
            {/* Background decoration */}
            <div className="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-violet-500/10 blur-3xl" />
            <div className="absolute -bottom-10 -left-10 h-40 w-40 rounded-full bg-fuchsia-500/10 blur-3xl" />

            <div className="relative text-center">
              <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-violet-500/15">
                <Sparkles className="h-6 w-6 text-violet-400" />
              </div>
              <h2 className="text-xl font-bold text-foreground sm:text-2xl">
                Создайте свою библиотеку промптов
              </h2>
              <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">
                Сохраняйте, версионируйте и делитесь промптами — прямо из Claude, Cursor и других клиентов через MCP.
              </p>

              <Link to="/sign-up">
                <Button size="lg" className="mt-6 bg-violet-500 px-8 text-white hover:bg-violet-600">
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Попробовать бесплатно
                </Button>
              </Link>

              {/* Feature pills */}
              <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
                <div className="flex items-center gap-1.5 rounded-full border border-border/50 bg-background/50 px-3 py-1.5 text-xs text-muted-foreground">
                  <Zap className="h-3 w-3 text-amber-400" />
                  MCP-сервер
                </div>
                <div className="flex items-center gap-1.5 rounded-full border border-border/50 bg-background/50 px-3 py-1.5 text-xs text-muted-foreground">
                  <BookOpen className="h-3 w-3 text-blue-400" />
                  Версионирование
                </div>
                <div className="flex items-center gap-1.5 rounded-full border border-border/50 bg-background/50 px-3 py-1.5 text-xs text-muted-foreground">
                  <Users className="h-3 w-3 text-green-400" />
                  Команды
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      <Toaster
        theme="dark"
        richColors
        position="bottom-center"
        closeButton
        duration={4000}
        toastOptions={{
          className: "relative overflow-hidden",
        }}
      />

      {/* Footer */}
      <footer className="border-t border-border/30 py-6">
        <div className="mx-auto max-w-4xl px-4 text-center text-xs text-muted-foreground sm:px-6">
          <Link to="/" className="transition-colors hover:text-violet-400">ПромтЛаб</Link>
          {" "}· Менеджер промптов с MCP
        </div>
      </footer>
    </div>
  )
}
