import { useState, useEffect } from "react"
import { Outlet } from "react-router-dom"
import { Toaster } from "sonner"
import { Search, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { TooltipProvider } from "@/components/ui/tooltip"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { CommandPalette } from "@/components/command-palette"
import { NotificationCenter } from "@/components/notifications/notification-center"
import { QuotaExceededDialog } from "@/components/subscription/quota-exceeded-dialog"
import { useRefreshSubscription } from "@/hooks/use-subscription"

export default function AppLayout() {
  const refreshSubscription = useRefreshSubscription()

  // checking=true пока polling идёт — показывает баннер "Проверяем оплату".
  // Polling длится до 2 минут (40 × 3 сек), т.к. T-Bank иногда шлёт webhook
  // с задержкой 30-60 сек. При timeout показываем мягкое сообщение с советом
  // обновить страницу позже (webhook дойдёт, план обновится асинхронно).
  const [checking, setChecking] = useState(false)

  // Обработка возврата после оплаты T-Bank.
  // 1) ?payment=success/failure — если T-Bank использует наш SuccessURL/FailURL
  // 2) sessionStorage pending_checkout — если T-Bank DEMO терминал редиректит на promtlabs.ru
  //    и юзер вручную возвращается в приложение
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const payment = params.get("payment")

    const runCheck = async () => {
      setChecking(true)
      try {
        const result = await refreshSubscription()
        if (result === "updated" || result === "already_pro") {
          toast.success("Подписка оформлена!")
        } else {
          toast.warning(
            "Оплата получена, но подтверждение от банка задерживается. Подписка активируется автоматически — обновите страницу через минуту.",
            { duration: 8000 },
          )
        }
      } finally {
        setChecking(false)
      }
    }

    if (payment === "success") {
      window.history.replaceState({}, "", window.location.pathname)
      sessionStorage.removeItem("pending_checkout")
      void runCheck()
      return
    }

    if (payment === "failure" || payment === "cancel") {
      window.history.replaceState({}, "", window.location.pathname)
      sessionStorage.removeItem("pending_checkout")
      // Явно сообщаем: списания не было, банк отклонил или юзер отменил.
      // Без такой формулировки юзер боится что деньги «зависли».
      toast.error("Оплата отклонена", {
        description:
          "Средства не списаны. Попробуйте другую карту или свяжитесь с банком.",
        duration: 8000,
      })
      return
    }

    if (sessionStorage.getItem("pending_checkout")) {
      sessionStorage.removeItem("pending_checkout")
      void runCheck()
    }
  }, [refreshSubscription])

  return (
    <TooltipProvider>
      <SidebarProvider>
        <div className="flex min-h-screen w-full overflow-x-hidden">
          <a href="#main-content" className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-background focus:px-4 focus:py-2 focus:text-sm focus:text-foreground focus:shadow-lg focus:ring-2 focus:ring-brand">
            Перейти к содержимому
          </a>
          {checking && (
            <div
              role="status"
              aria-live="polite"
              className="fixed inset-x-0 top-0 z-50 flex items-center justify-center gap-2 bg-brand/95 px-4 py-2 text-sm font-medium text-brand-foreground shadow-md"
            >
              <Loader2 className="h-4 w-4 animate-spin" />
              Проверяем оплату… подписка активируется после подтверждения банка.
            </div>
          )}
          <AppSidebar />
          <div className="flex min-w-0 flex-1 flex-col">
            <header role="banner" className="flex h-14 items-center justify-between px-4">
              <div className="lg:hidden">
                <SidebarTrigger />
              </div>
              <div className="ml-auto flex items-center gap-2">
                <NotificationCenter />

                {/* Search button */}
                <button
                  type="button"
                  onClick={() =>
                    window.dispatchEvent(
                      new KeyboardEvent("keydown", { key: "k", metaKey: true }),
                    )
                  }
                  className="flex h-11 min-w-11 cursor-pointer items-center gap-2 rounded-lg border border-border bg-muted/20 px-3 text-[0.8rem] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                >
                  <Search className="h-4 w-4" />
                  <span className="hidden sm:inline">Поиск...</span>
                  <kbd className="hidden rounded border border-border bg-muted/30 px-1 py-px text-[9px] sm:inline">
                    ⌘K
                  </kbd>
                </button>
              </div>
            </header>
            <main id="main-content" role="main" className="flex-1 overflow-x-hidden px-4 py-5 sm:px-8 sm:py-7">
              <Outlet />
            </main>
          </div>
        </div>
        <CommandPalette />
        <QuotaExceededDialog />
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
      </SidebarProvider>
    </TooltipProvider>
  )
}
