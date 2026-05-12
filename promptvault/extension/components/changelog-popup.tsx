import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { BookOpen, X } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import { sendBg } from "../lib/bg-client"

const STORAGE_KEY = "pv.changelogToastSeen"

// Popup-toast при первом запуске после release, если у юзера has_unread=true.
// Показываем один раз — после dismiss flag в localStorage предотвращает повтор
// до следующего release (когда appears новая запись и has_unread снова true).
//
// localStorage значение = timestamp последнего dismiss. Поскольку backend
// возвращает has_unread исходя из last_seen vs latest_release_at — после
// нашего dismiss-без-MarkRead на backend юзер всё равно увидит popup в
// другой день (это OK для UX). MarkRead дёргается только при клике CTA.
export function ChangelogPopup() {
  const navigate = useNavigate()
  const [dismissed, setDismissed] = useState(false)
  const [show, setShow] = useState(false)

  const changelogQuery = useQuery({
    queryKey: ["changelog", "popup-check"],
    queryFn: () => sendBg({ type: "api.getChangelog" }),
    staleTime: 5 * 60_000,
    refetchOnWindowFocus: false,
  })

  useEffect(() => {
    if (!changelogQuery.data?.has_unread) return
    if (dismissed) return
    // Проверяем localStorage: если за последние 24 часа уже dismissed —
    // не показываем. Это предотвращает спам при reload extension.
    try {
      const last = localStorage.getItem(STORAGE_KEY)
      if (last && Date.now() - Number(last) < 24 * 60 * 60 * 1000) {
        return
      }
    } catch {
      // localStorage недоступен — продолжаем
    }
    setShow(true)
  }, [changelogQuery.data?.has_unread, dismissed])

  const latest = changelogQuery.data?.entries?.[0]

  function dismiss() {
    setShow(false)
    setDismissed(true)
    try {
      localStorage.setItem(STORAGE_KEY, String(Date.now()))
    } catch {
      // ignore
    }
  }

  function openChangelog() {
    dismiss()
    navigate("/changelog")
  }

  if (!show || !latest) return null

  return (
    <div className="fixed bottom-16 right-3 z-50 w-72 rounded-lg border border-(--color-primary)/40 bg-(--color-background) shadow-xl animate-in slide-in-from-bottom-2">
      <div className="flex items-start gap-2 p-3">
        <div className="rounded-md bg-(--color-primary)/15 p-1.5">
          <BookOpen className="h-3.5 w-3.5 text-(--color-primary)" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            <span className="text-[10px] font-medium uppercase tracking-wide text-(--color-primary)">
              Что нового
            </span>
            <span className="rounded bg-(--color-muted) px-1 py-px text-[9px] font-mono">
              {latest.version}
            </span>
          </div>
          <h4 className="mt-0.5 text-xs font-semibold">{latest.title}</h4>
          {latest.description && (
            <p className="mt-0.5 line-clamp-2 text-[10px] text-(--color-muted-foreground)">
              {latest.description}
            </p>
          )}
          <div className="mt-1.5 flex gap-1.5">
            <button
              type="button"
              onClick={openChangelog}
              className="rounded bg-(--color-primary) px-2 py-1 text-[10px] font-medium text-(--color-primary-foreground) hover:opacity-90"
            >
              Посмотреть
            </button>
            <button
              type="button"
              onClick={dismiss}
              className="text-[10px] text-(--color-muted-foreground) hover:underline"
            >
              Позже
            </button>
          </div>
        </div>
        <button
          type="button"
          onClick={dismiss}
          className="rounded p-0.5 text-(--color-muted-foreground) hover:bg-(--color-muted)"
          aria-label="Закрыть"
        >
          <X className="h-3 w-3" />
        </button>
      </div>
    </div>
  )
}
