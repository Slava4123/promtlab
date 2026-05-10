// Страница истории запусков цепочки. Phase 16: Pro=10, Max=1000 последних
// сохранённых execution'ов (max_saved_executions из плана).
// RBAC: всем у кого есть read-access к chain (owner личной; owner/editor/viewer
// команды). Это team-property — в отличие от resume/advance (initiator-only).

import { Link, useParams } from "react-router-dom"
import { ArrowLeft, ArrowRight, CheckCircle2, Clock, XCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useChain, useChainExecutions } from "@/hooks/use-chains"
import { useAuthStore } from "@/stores/auth-store"
import type { ChainExecutionSummary } from "@/api/types"

function formatDate(s: string): string {
  return new Date(s).toLocaleString("ru-RU", {
    day: "2-digit", month: "2-digit", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  })
}

function formatDuration(start: string, end?: string | null): string {
  if (!end) return "—"
  const ms = new Date(end).getTime() - new Date(start).getTime()
  if (ms < 0) return "—"
  if (ms < 1000) return `${ms} мс`
  const sec = Math.round(ms / 1000)
  if (sec < 60) return `${sec} с`
  const min = Math.floor(sec / 60)
  const rem = sec % 60
  return `${min} мин ${rem} с`
}

function StatusBadge({ status }: { status: ChainExecutionSummary["status"] }) {
  if (status === "completed") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/15 px-2 py-0.5 text-[0.7rem] font-medium text-emerald-600 dark:text-emerald-400">
        <CheckCircle2 className="h-3 w-3" />
        Завершена
      </span>
    )
  }
  if (status === "in_progress") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/15 px-2 py-0.5 text-[0.7rem] font-medium text-amber-600 dark:text-amber-400">
        <Clock className="h-3 w-3" />
        В процессе
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-[0.7rem] font-medium text-muted-foreground">
      <XCircle className="h-3 w-3" />
      Прервана
    </span>
  )
}

export default function ChainRunsPage() {
  const { id } = useParams<{ id: string }>()
  const chainID = id ? Number(id) : 0
  const currentUserID = useAuthStore((s) => s.user?.id)

  const { data: chain, isLoading: chainLoading } = useChain(chainID)
  const { data, isLoading } = useChainExecutions(chainID, 50)

  return (
    <div className="container mx-auto p-6">
      <div className="mb-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={`/chains/${chainID}/edit`}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            К редактору
          </Link>
        </Button>
      </div>

      <div className="mb-6">
        <h1 className="text-2xl font-semibold">
          {chainLoading ? <Skeleton className="h-7 w-64" /> : `История запусков — ${chain?.name ?? "цепочка"}`}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Сохраняются последние запуски согласно лимиту тарифа.
          Сводка без полных данных шагов — открыть запуск, чтобы увидеть детали.
        </p>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-20" />)}
        </div>
      )}

      {!isLoading && data && data.items.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Clock className="mb-4 h-12 w-12 text-muted-foreground" />
            <p className="mb-4 text-muted-foreground">У этой цепочки пока нет завершённых запусков.</p>
            <Button asChild>
              <Link to={`/chains/${chainID}/run`}>Запустить</Link>
            </Button>
          </CardContent>
        </Card>
      )}

      {!isLoading && data && data.items.length > 0 && (
        <div className="space-y-2">
          {data.items.map((exec) => {
            const isOwn = exec.user_id === currentUserID
            const canResume = exec.status === "in_progress" && isOwn
            return (
              <Card key={exec.id}>
                <CardHeader className="flex flex-row items-start justify-between gap-3 space-y-0 pb-2">
                  <div className="flex items-center gap-3">
                    <CardTitle className="text-base">Запуск №{exec.id}</CardTitle>
                    <StatusBadge status={exec.status} />
                  </div>
                  <div className="flex gap-2">
                    {canResume && (
                      <Button size="sm" asChild>
                        <Link to={`/chains/${chainID}/run?resume=${exec.id}`}>
                          Продолжить
                          <ArrowRight className="ml-1 h-4 w-4" />
                        </Link>
                      </Button>
                    )}
                  </div>
                </CardHeader>
                <CardContent className="grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <div>
                    <span className="text-muted-foreground">Начат:</span>{" "}
                    <span>{formatDate(exec.started_at)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">
                      {exec.status === "in_progress" ? "Текущий шаг:" : "Длительность:"}
                    </span>{" "}
                    <span>
                      {exec.status === "in_progress"
                        ? `${exec.current_step + 1}`
                        : formatDuration(exec.started_at, exec.completed_at)}
                    </span>
                  </div>
                  <div className="text-muted-foreground">
                    {isOwn ? "Запустили вы" : `Инициатор: id ${exec.user_id}`}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
