import { useEffect, useRef, useState } from "react"
import { NavLink } from "react-router-dom"
import {
  FolderOpen,
  Trash2,
  Award,
  CreditCard,
  Bell,
  BookOpen,
  MessageSquare,
  Settings as SettingsIcon,
  X,
} from "lucide-react"
import { cn } from "../../lib/utils"
import { FeedbackDialog } from "../feedback-dialog"
import { useUnreadCount } from "../../hooks/use-unread-count"

interface DrawerLink {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
}

// Меню «остальное» — вторичные страницы, не помещающиеся в bottom-tabs.
// История — bottom-tab. Теги — создаются inline в prompt-editor.
const DRAWER_LINKS: DrawerLink[] = [
  { to: "/collections", label: "Коллекции", icon: FolderOpen },
  { to: "/trash", label: "Корзина", icon: Trash2 },
  { to: "/badges", label: "Достижения", icon: Award },
  { to: "/settings/notifications", label: "Уведомления", icon: Bell },
  { to: "/changelog", label: "Что нового", icon: BookOpen },
  { to: "/pricing", label: "Тарифы", icon: CreditCard },
  { to: "/settings", label: "Настройки", icon: SettingsIcon },
]

interface DrawerProps {
  open: boolean
  onClose: () => void
}

export function Drawer({ open, onClose }: DrawerProps) {
  const [feedbackOpen, setFeedbackOpen] = useState(false)
  const unreadCount = useUnreadCount()
  const closeButtonRef = useRef<HTMLButtonElement | null>(null)

  // M6: Esc закрывает drawer. Backdrop-click был, keyboard — не было.
  useEffect(() => {
    if (!open) return
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose()
    }
    document.addEventListener("keydown", handleKey)
    return () => document.removeEventListener("keydown", handleKey)
  }, [open, onClose])

  // Фокус на close-кнопку при открытии — попадаем в drawer для Tab-навигации.
  useEffect(() => {
    if (open) closeButtonRef.current?.focus()
  }, [open])

  function openFeedback() {
    setFeedbackOpen(true)
    onClose()
  }

  // Drawer overlay рендерится conditionally, НО FeedbackDialog держим
  // mounted вне условия — иначе onClose() при openFeedback() unmount'ил
  // весь компонент (включая <FeedbackDialog />), и dialog не успевал
  // показаться. FeedbackDialog сам рендерит null когда open=false.
  return (
    <>
      {open && (
      <div className="fixed inset-0 z-50">
        <div
          className="absolute inset-0 bg-black/40 backdrop-blur-sm"
          onClick={onClose}
          aria-hidden
        />
        <aside
          role="dialog"
          aria-modal="true"
          aria-label="Меню"
          className="absolute left-0 top-0 flex h-full w-64 flex-col border-r border-(--color-border) bg-(--color-background) shadow-xl"
        >
          <header className="flex items-center justify-between border-b border-(--color-border) px-3 py-2">
            <h2 className="text-sm font-semibold">Меню</h2>
            <button
              ref={closeButtonRef}
              type="button"
              onClick={onClose}
              className="rounded-md p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
              aria-label="Закрыть меню"
            >
              <X className="h-4 w-4" />
            </button>
          </header>
          <ul className="flex flex-1 flex-col gap-0.5 overflow-y-auto p-2">
            {DRAWER_LINKS.map((link) => {
              const showBadge =
                link.to === "/settings/notifications" && unreadCount > 0
              return (
                <li key={link.to}>
                  <NavLink
                    to={link.to}
                    onClick={onClose}
                    className={({ isActive }) =>
                      cn(
                        "flex items-center gap-2 rounded-md px-2 py-2 text-sm transition-colors",
                        isActive
                          ? "bg-(--color-brand-muted) text-(--color-brand)"
                          : "text-(--color-foreground) hover:bg-(--color-muted)",
                      )
                    }
                  >
                    <link.icon className="h-4 w-4 text-(--color-muted-foreground)" />
                    <span className="flex-1">{link.label}</span>
                    {showBadge && (
                      <span
                        className="ml-auto flex h-4 min-w-4 items-center justify-center rounded-full bg-(--color-brand) px-1 text-[10px] font-bold text-(--color-brand-foreground)"
                        aria-label={`${unreadCount} непрочитанных`}
                      >
                        {unreadCount > 99 ? "99+" : unreadCount}
                      </span>
                    )}
                  </NavLink>
                </li>
              )
            })}
          </ul>
          <footer className="border-t border-(--color-border) p-2">
            <button
              type="button"
              onClick={openFeedback}
              className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm text-(--color-foreground) hover:bg-(--color-muted)"
            >
              <MessageSquare className="h-4 w-4 text-(--color-muted-foreground)" />
              <span>Обратная связь</span>
            </button>
          </footer>
        </aside>
      </div>
      )}
      <FeedbackDialog open={feedbackOpen} onClose={() => setFeedbackOpen(false)} />
    </>
  )
}

export function useDrawer() {
  const [open, setOpen] = useState(false)
  return {
    open,
    openDrawer: () => setOpen(true),
    closeDrawer: () => setOpen(false),
    toggleDrawer: () => setOpen((v) => !v),
  }
}
