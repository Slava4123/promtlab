import { useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { Loader2, Check, ArrowLeft, Sparkles } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { useStarterCatalog, useCompleteOnboarding } from "@/hooks/use-starter"
import { ApiError } from "@/api/client"
import { useAuthStore } from "@/stores/auth-store"
import type { StarterCategory, StarterTemplate } from "@/api/types"

type Step = "category" | "templates"

export default function WelcomePage() {
  const navigate = useNavigate()
  const { data: catalog, isLoading, error } = useStarterCatalog()
  const complete = useCompleteOnboarding()

  const [step, setStep] = useState<Step>("category")
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  // Set ID'шников промптов для install — изначально все из выбранной категории.
  const [selectedTemplateIds, setSelectedTemplateIds] = useState<Set<string>>(new Set())

  const categoryTemplates = useMemo<StarterTemplate[]>(() => {
    if (!catalog || !selectedCategory) return []
    return catalog.templates.filter((t) => t.category === selectedCategory)
  }, [catalog, selectedCategory])

  function handleCategorySelect(category: StarterCategory) {
    setSelectedCategory(category.id)
    if (catalog) {
      const ids = catalog.templates
        .filter((t) => t.category === category.id)
        .map((t) => t.id)
      setSelectedTemplateIds(new Set(ids))
    }
    setStep("templates")
  }

  function toggleTemplate(id: string) {
    setSelectedTemplateIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  // Единый flow: и "Готово", и "Пропустить" — это POST /api/starter/complete,
  // отличие только в payload. Объединяем чтобы 409-handling и navigation
  // не дублировались.
  async function completeOnboarding(ids: string[]) {
    try {
      const result = await complete.mutateAsync({ install: ids })
      const count = result.installed.length
      if (count > 0) {
        toast.success(`Готово! Добавлено промптов: ${count}`)
      } else {
        toast.success("Готово! Можно создавать свои промпты")
      }
      navigate("/dashboard", { replace: true })
    } catch (err) {
      // 409 = онбординг уже завершён в другой вкладке / на другом устройстве.
      // Не показываем error-toast — это не ошибка для юзера. Синкуем user state
      // через /auth/me и редиректим на dashboard как обычный success.
      if (err instanceof ApiError && err.status === 409) {
        try {
          await useAuthStore.getState().fetchMe()
        } catch {
          // если fetchMe тоже упал — навигация всё равно даст ProtectedRoute
          // ещё одну попытку через restoreSession
        }
        navigate("/dashboard", { replace: true })
        return
      }
      const msg = err instanceof Error ? err.message : "Не удалось завершить онбординг"
      toast.error(msg)
    }
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !catalog) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <div className="max-w-md text-center">
          <p className="text-base font-medium text-foreground">Не удалось загрузить шаблоны</p>
          <p className="mt-2 text-sm text-muted-foreground">
            Попробуйте обновить страницу. Можно пропустить онбординг и перейти к dashboard.
          </p>
          <Button
            className="mt-6"
            variant="outline"
            onClick={() => completeOnboarding([])}
            disabled={complete.isPending}
          >
            Пропустить
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-4xl px-4 py-10 sm:py-16">
        {/* Header */}
        <div className="mb-10 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-violet-500/10 ring-1 ring-violet-500/20">
            <Sparkles className="h-7 w-7 text-violet-400" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
            Добро пожаловать в ПромтЛаб!
          </h1>
          <p className="mx-auto mt-2 max-w-xl text-sm text-muted-foreground">
            {step === "category"
              ? "Выберите направление — мы добавим в вашу библиотеку готовые промпты под него. Можно пропустить и создать свои."
              : "Снимите галочки с тех, что не нужны. Промпты будут добавлены в вашу личную библиотеку."}
          </p>
        </div>

        {/* Step 1: Category selection */}
        {step === "category" && (
          <div className="grid gap-3 sm:grid-cols-2">
            {catalog.categories.map((cat) => {
              const count = catalog.templates.filter((t) => t.category === cat.id).length
              return (
                <button
                  key={cat.id}
                  type="button"
                  onClick={() => handleCategorySelect(cat)}
                  className="group relative flex flex-col items-start gap-2 rounded-xl border border-border bg-card p-5 text-left transition-colors hover:border-violet-500/40 hover:bg-violet-500/[0.03] focus:outline-none focus-visible:ring-2 focus-visible:ring-violet-500/40"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">{cat.icon}</span>
                    <h2 className="text-base font-semibold text-foreground">{cat.name}</h2>
                  </div>
                  <p className="text-[0.8rem] text-muted-foreground">{cat.description}</p>
                  <ul className="mt-1 space-y-0.5 text-[0.75rem] text-muted-foreground">
                    {cat.use_cases.slice(0, 4).map((uc) => (
                      <li key={uc} className="flex items-start gap-1.5">
                        <span className="mt-1 h-1 w-1 rounded-full bg-violet-500/60" />
                        <span>{uc}</span>
                      </li>
                    ))}
                  </ul>
                  <div className="mt-2 inline-flex items-center gap-1 rounded-full bg-violet-500/10 px-2 py-0.5 text-[0.7rem] font-medium text-violet-400">
                    {count} промптов
                  </div>
                </button>
              )
            })}
          </div>
        )}

        {/* Step 2: Templates checklist */}
        {step === "templates" && (
          <div className="space-y-4">
            <div className="rounded-xl border border-border bg-card">
              <div className="border-b border-border px-5 py-3">
                <p className="text-[0.8rem] font-medium text-muted-foreground">
                  Выбрано: {selectedTemplateIds.size} из {categoryTemplates.length}
                </p>
              </div>
              <ul className="divide-y divide-border">
                {categoryTemplates.map((tpl) => {
                  const checked = selectedTemplateIds.has(tpl.id)
                  return (
                    <li key={tpl.id}>
                      <label className="flex cursor-pointer items-start gap-3 px-5 py-3 transition-colors hover:bg-muted/30">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={() => toggleTemplate(tpl.id)}
                          className="mt-1 h-4 w-4 cursor-pointer accent-violet-500"
                        />
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium text-foreground">{tpl.title}</p>
                          <p className="mt-0.5 line-clamp-2 text-[0.75rem] text-muted-foreground">
                            {tpl.content.slice(0, 200)}
                          </p>
                        </div>
                      </label>
                    </li>
                  )
                })}
              </ul>
            </div>
          </div>
        )}

        {/* Footer actions */}
        <div className="mt-8 flex flex-col-reverse items-stretch justify-between gap-3 sm:flex-row sm:items-center">
          <div>
            {step === "templates" && (
              <Button
                variant="ghost"
                onClick={() => setStep("category")}
                disabled={complete.isPending}
              >
                <ArrowLeft className="h-4 w-4" />
                Назад к категориям
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => completeOnboarding([])}
              disabled={complete.isPending}
            >
              Пропустить
            </Button>
            {step === "templates" && (
              <Button
                onClick={() => completeOnboarding(Array.from(selectedTemplateIds))}
                disabled={complete.isPending}
              >
                {complete.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Check className="h-4 w-4" />
                )}
                {selectedTemplateIds.size > 0
                  ? `Добавить ${selectedTemplateIds.size}`
                  : "Готово"}
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
