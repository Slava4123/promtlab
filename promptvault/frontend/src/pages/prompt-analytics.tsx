import { useParams, Link, useNavigate } from "react-router-dom"
import { ArrowLeft, BarChart3 } from "lucide-react"
import { buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { usePromptAnalytics } from "@/hooks/use-analytics"
import { usePrompt } from "@/hooks/use-prompts"
import { UsageChart } from "@/components/analytics/usage-chart"
import { MetricCard } from "@/components/analytics/metric-card"

// Per-prompt analytics — timeline использования и просмотров share-ссылок
// для одного конкретного промпта. Access check на backend (владелец или
// член команды промпта).
export default function PromptAnalyticsPage() {
  const { id: idParam } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const promptId = Number(idParam)

  const { data: prompt, isLoading: promptLoading, isError: promptError } = usePrompt(promptId)
  const { data, isLoading, isError } = usePromptAnalytics(promptId)

  if (promptError || isError) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="mb-4 text-2xl font-bold">Аналитика промпта</h1>
        <p className="text-destructive">Не удалось загрузить данные. Попробуйте обновить страницу.</p>
      </div>
    )
  }

  const totalUses = data?.usage_per_day.reduce((s, p) => s + p.count, 0) ?? 0
  const totalViews = data?.share_views_per_day.reduce((s, p) => s + p.count, 0) ?? 0

  return (
    <div className="container mx-auto max-w-4xl space-y-6 px-4 py-8">
      <div className="flex items-center gap-3">
        <Link
          to={`/prompts/${promptId}`}
          className={buttonVariants({ variant: "ghost", size: "sm" })}
        >
          <ArrowLeft className="size-4" />
        </Link>
        <div className="flex flex-1 items-center gap-2">
          <BarChart3 className="size-5 text-muted-foreground" />
          <div>
            <h1 className="text-2xl font-bold">{prompt?.title ?? "Аналитика промпта"}</h1>
            <p className="text-sm text-muted-foreground">
              Использование и просмотры за последнее окно
            </p>
          </div>
        </div>
      </div>

      {promptLoading || isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-28 w-full" />
          ))}
        </div>
      ) : data ? (
        <>
          <div className="grid gap-4 sm:grid-cols-2">
            <MetricCard
              title="Всего использований"
              value={totalUses.toLocaleString("ru")}
              subtitle="по всей истории окна"
            />
            <MetricCard
              title="Просмотров публичной ссылки"
              value={totalViews.toLocaleString("ru")}
              subtitle="если ссылка создана и активна"
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Использование по дням</CardTitle>
            </CardHeader>
            <CardContent>
              <UsageChart title="" data={data.usage_per_day} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Просмотры публичной ссылки по дням</CardTitle>
            </CardHeader>
            <CardContent>
              <UsageChart title="" data={data.share_views_per_day} />
            </CardContent>
          </Card>
        </>
      ) : (
        <p className="text-muted-foreground">Нет данных.</p>
      )}

      {/* Скрытая кнопка удалить — для Esc/back */}
      <button
        onClick={() => navigate(-1)}
        className="sr-only"
        aria-label="Назад"
      />
    </div>
  )
}
