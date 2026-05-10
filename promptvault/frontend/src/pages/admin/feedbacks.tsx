import { useState, useEffect } from "react"
import {
  Loader2,
  Search,
  ChevronRight,
  X,
  Bug,
  Lightbulb,
  MessageSquare,
  Trash2,
  ExternalLink,
} from "lucide-react"
import { toast } from "sonner"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet"
import { ActionDialog } from "@/components/admin/action-dialog"
import { useIsMobile } from "@/hooks/use-mobile"
import {
  useAdminFeedbacks,
  useAdminFeedbackDetail,
  useUpdateFeedbackStatus,
  useDeleteFeedback,
} from "@/hooks/admin/use-admin-feedbacks"
import type {
  AdminFeedbackItem,
  AdminFeedbacksFilter,
  FeedbackStatus,
  FeedbackType,
} from "@/api/admin/feedbacks"

const PAGE_SIZE = 20

const TYPE_LABEL: Record<FeedbackType, string> = {
  bug: "Баг",
  feature: "Идея",
  other: "Другое",
}

const STATUS_LABEL: Record<FeedbackStatus, string> = {
  new: "Новый",
  read: "Прочитан",
  archived: "В архиве",
}

function TypeBadge({ type }: { type: FeedbackType }) {
  const map: Record<FeedbackType, { cls: string; Icon: typeof Bug }> = {
    bug: {
      cls: "bg-rose-500/15 text-rose-300",
      Icon: Bug,
    },
    feature: {
      cls: "bg-amber-500/15 text-amber-300",
      Icon: Lightbulb,
    },
    other: {
      cls: "bg-slate-500/15 text-slate-300",
      Icon: MessageSquare,
    },
  }
  const { cls, Icon } = map[type]
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[0.65rem] font-medium ${cls}`}
    >
      <Icon className="h-3 w-3" />
      {TYPE_LABEL[type]}
    </span>
  )
}

function StatusBadge({ status }: { status: FeedbackStatus }) {
  const cls: Record<FeedbackStatus, string> = {
    new: "bg-violet-500/15 text-violet-300",
    read: "bg-emerald-500/15 text-emerald-300",
    archived: "bg-slate-500/15 text-slate-400",
  }
  return (
    <span
      className={`inline-flex items-center rounded-md px-1.5 py-0.5 text-[0.65rem] font-medium ${cls[status]}`}
    >
      {STATUS_LABEL[status]}
    </span>
  )
}

export default function AdminFeedbacksPage() {
  const isMobile = useIsMobile()
  const [query, setQuery] = useState("")
  const [debouncedQuery, setDebouncedQuery] = useState("")
  const [type, setType] = useState<AdminFeedbacksFilter["type"]>("")
  const [status, setStatus] =
    useState<AdminFeedbacksFilter["status"]>("")
  const [page, setPage] = useState(1)
  const [openId, setOpenId] = useState<number | null>(null)

  // Debounce — паттерн идентичен users.tsx:46-49.
  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(t)
  }, [query])

  const { data, isLoading, error } = useAdminFeedbacks({
    q: debouncedQuery,
    type,
    status,
    page,
    page_size: PAGE_SIZE,
  })

  const totalPages = data
    ? Math.max(1, Math.ceil(data.total / data.page_size))
    : 1

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
        <div className="relative w-full sm:flex-1 sm:min-w-[220px]">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setPage(1)
            }}
            placeholder="Поиск по тексту или email..."
            className="pl-8"
          />
        </div>
        <div className="flex gap-2">
          <select
            value={type}
            onChange={(e) => {
              setType(e.target.value as AdminFeedbacksFilter["type"])
              setPage(1)
            }}
            className="flex-1 rounded-md border border-border bg-background px-2 py-1.5 text-sm sm:flex-none"
          >
            <option value="">Все типы</option>
            <option value="bug">Баг</option>
            <option value="feature">Идея</option>
            <option value="other">Другое</option>
          </select>
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value as AdminFeedbacksFilter["status"])
              setPage(1)
            }}
            className="flex-1 rounded-md border border-border bg-background px-2 py-1.5 text-sm sm:flex-none"
          >
            <option value="">Все статусы</option>
            <option value="new">Новые</option>
            <option value="read">Прочитанные</option>
            <option value="archived">Архив</option>
          </select>
        </div>
      </div>

      {isLoading && (
        <div className="flex h-40 items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
          Не удалось загрузить отзывы
        </div>
      )}

      {data && (
        <>
          {isMobile ? (
            <MobileFeedbackList
              items={data.items}
              onSelect={setOpenId}
            />
          ) : (
            <DesktopFeedbackTable
              items={data.items}
              onSelect={setOpenId}
            />
          )}

          {totalPages > 1 && (
            <div className="flex items-center justify-between gap-2">
              <p className="text-xs text-muted-foreground">
                Всего {data.total} · стр {data.page} из {totalPages}
              </p>
              <div className="flex gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  Назад
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => p + 1)}
                  disabled={!data.has_more}
                >
                  Вперёд
                </Button>
              </div>
            </div>
          )}
        </>
      )}

      <FeedbackDetailSheet
        feedbackId={openId}
        onClose={() => setOpenId(null)}
      />
    </div>
  )
}

// ===== Desktop table =====

function DesktopFeedbackTable({
  items,
  onSelect,
}: {
  items: AdminFeedbackItem[]
  onSelect: (id: number) => void
}) {
  if (items.length === 0) {
    return <EmptyState />
  }
  return (
    <div className="rounded-xl border border-border">
      <table className="w-full text-sm">
        <thead className="bg-muted/20 text-[0.72rem] uppercase tracking-wider text-muted-foreground">
          <tr>
            <th className="px-3 py-2 text-left font-medium">От</th>
            <th className="px-3 py-2 text-left font-medium">Тип</th>
            <th className="px-3 py-2 text-left font-medium">Сообщение</th>
            <th className="px-3 py-2 text-left font-medium">Статус</th>
            <th className="px-3 py-2 text-left font-medium">Дата</th>
          </tr>
        </thead>
        <tbody>
          {items.map((f) => (
            <tr
              key={f.id}
              onClick={() => onSelect(f.id)}
              className="cursor-pointer border-t border-border hover:bg-muted/20"
            >
              <td className="px-3 py-2">
                <div className="font-medium">{f.user_email || "—"}</div>
                {f.user_name && (
                  <div className="text-[0.7rem] text-muted-foreground">
                    {f.user_name}
                  </div>
                )}
              </td>
              <td className="px-3 py-2">
                <TypeBadge type={f.type} />
              </td>
              <td className="max-w-[420px] px-3 py-2">
                <p className="truncate text-[0.78rem] text-muted-foreground">
                  {f.message}
                </p>
              </td>
              <td className="px-3 py-2">
                <StatusBadge status={f.status} />
              </td>
              <td className="px-3 py-2 text-[0.72rem] text-muted-foreground">
                {new Date(f.created_at).toLocaleDateString("ru-RU")}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ===== Mobile cards =====

function MobileFeedbackList({
  items,
  onSelect,
}: {
  items: AdminFeedbackItem[]
  onSelect: (id: number) => void
}) {
  if (items.length === 0) {
    return <EmptyState />
  }
  return (
    <ul className="space-y-2">
      {items.map((f) => (
        <li key={f.id}>
          <button
            onClick={() => onSelect(f.id)}
            className="flex w-full items-center gap-3 rounded-xl border border-border bg-background p-3 text-left transition-colors hover:bg-muted/20"
          >
            <div className="min-w-0 flex-1 space-y-1.5">
              <div className="flex items-center gap-2">
                <TypeBadge type={f.type} />
                <StatusBadge status={f.status} />
              </div>
              <p className="line-clamp-2 text-sm">{f.message}</p>
              <div className="flex items-center justify-between text-[0.7rem] text-muted-foreground">
                <span className="truncate">{f.user_email}</span>
                <span>{new Date(f.created_at).toLocaleDateString("ru-RU")}</span>
              </div>
            </div>
            <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
          </button>
        </li>
      ))}
    </ul>
  )
}

function EmptyState() {
  return (
    <div className="rounded-xl border border-border px-3 py-10 text-center text-sm text-muted-foreground">
      Отзывов нет
    </div>
  )
}

// ===== Detail Sheet =====

function FeedbackDetailSheet({
  feedbackId,
  onClose,
}: {
  feedbackId: number | null
  onClose: () => void
}) {
  const open = feedbackId !== null
  const id = feedbackId ?? 0
  const { data, isLoading } = useAdminFeedbackDetail(id)
  const updateStatus = useUpdateFeedbackStatus()
  const deleteFeedback = useDeleteFeedback()

  // Локальный state для confirm-dialogs (TOTP).
  const [pendingStatus, setPendingStatus] = useState<FeedbackStatus | null>(
    null,
  )
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handleStatusConfirm = async (totpCode?: string) => {
    if (!totpCode || !data || !pendingStatus) return
    await updateStatus.mutateAsync({
      id: data.id,
      status: pendingStatus,
      totpCode,
    })
    toast.success(`Статус: ${STATUS_LABEL[pendingStatus]}`)
    setPendingStatus(null)
  }

  const handleDeleteConfirm = async (totpCode?: string) => {
    if (!totpCode || !data) return
    await deleteFeedback.mutateAsync({ id: data.id, totpCode })
    toast.success("Отзыв удалён")
    setConfirmDelete(false)
    onClose()
  }

  return (
    <>
      <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
        <SheetContent className="flex w-full flex-col gap-0 sm:max-w-lg">
          <SheetHeader>
            <SheetTitle>Отзыв #{feedbackId}</SheetTitle>
            <SheetDescription>
              Полный текст и действия администратора
            </SheetDescription>
          </SheetHeader>

          {isLoading || !data ? (
            <div className="flex flex-1 items-center justify-center">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="flex-1 space-y-4 overflow-y-auto px-4 pb-4">
              <div className="flex flex-wrap items-center gap-2">
                <TypeBadge type={data.type} />
                <StatusBadge status={data.status} />
                <span className="text-[0.7rem] text-muted-foreground">
                  {new Date(data.created_at).toLocaleString("ru-RU")}
                </span>
              </div>

              <div className="space-y-1">
                <p className="text-[0.7rem] uppercase tracking-wider text-muted-foreground">
                  От
                </p>
                <p className="text-sm">
                  <span className="font-medium">{data.user_email}</span>
                  {data.user_name && (
                    <span className="text-muted-foreground"> · {data.user_name}</span>
                  )}
                </p>
              </div>

              {data.page_url && (
                <div className="space-y-1">
                  <p className="text-[0.7rem] uppercase tracking-wider text-muted-foreground">
                    Страница
                  </p>
                  <a
                    href={data.page_url.startsWith("http") ? data.page_url : `${window.location.origin}${data.page_url}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 break-all text-sm text-violet-300 hover:underline"
                  >
                    {data.page_url}
                    <ExternalLink className="h-3 w-3 shrink-0" />
                  </a>
                </div>
              )}

              <div className="space-y-1">
                <p className="text-[0.7rem] uppercase tracking-wider text-muted-foreground">
                  Сообщение
                </p>
                <p className="whitespace-pre-wrap rounded-md border border-border bg-muted/20 p-3 text-sm">
                  {data.message}
                </p>
              </div>

              <div className="space-y-2 border-t border-border pt-4">
                <p className="text-[0.7rem] uppercase tracking-wider text-muted-foreground">
                  Действия
                </p>
                <div className="flex flex-wrap gap-2">
                  {(["new", "read", "archived"] as FeedbackStatus[])
                    .filter((s) => s !== data.status)
                    .map((s) => (
                      <Button
                        key={s}
                        size="sm"
                        variant="outline"
                        onClick={() => setPendingStatus(s)}
                      >
                        В «{STATUS_LABEL[s]}»
                      </Button>
                    ))}
                  <Button
                    size="sm"
                    variant="destructive-solid"
                    onClick={() => setConfirmDelete(true)}
                  >
                    <Trash2 className="mr-1 h-3.5 w-3.5" />
                    Удалить
                  </Button>
                </div>
              </div>
            </div>
          )}

          <div className="border-t border-border p-3">
            <Button variant="outline" size="sm" onClick={onClose} className="w-full">
              <X className="mr-1 h-3.5 w-3.5" />
              Закрыть
            </Button>
          </div>
        </SheetContent>
      </Sheet>

      <ActionDialog
        open={pendingStatus !== null}
        onOpenChange={(v) => !v && setPendingStatus(null)}
        title="Изменить статус отзыва"
        description={
          pendingStatus
            ? `Перевести отзыв в статус «${STATUS_LABEL[pendingStatus]}». Действие будет залогировано.`
            : ""
        }
        confirmLabel="Изменить"
        requireTOTP
        onConfirm={handleStatusConfirm}
      />

      <ActionDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Удалить отзыв"
        description="Отзыв будет удалён навсегда. Это действие нельзя отменить."
        confirmLabel="Удалить"
        requireTOTP
        onConfirm={handleDeleteConfirm}
      />
    </>
  )
}
