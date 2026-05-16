// Confirmation dialog — открывается перед редиректом на T-Bank. Заменил
// inline-чек-бокс в pricing.tsx (был шумно, дублировался для Pro и Max).
//
// Поток: юзер кликает «Получить Pro» → открывается этот диалог со сводкой
// тарифа, датой следующего списания, top features и чек-боксом согласия на
// recurrent → клик «Перейти к оплате» → запуск checkout.mutate → T-Bank.
//
// Из quota-exceeded-dialog в этот flow попадаем через
// navigate('/pricing?upgrade=pro') — pricing.tsx считывает query и сам
// открывает этот диалог. Это даёт юзеру шанс увидеть сравнение тарифов
// перед окончательным согласием.

import { useState } from "react"
import { Crown, Check, Loader2, Sparkles, type LucideIcon } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { RecurrentConsent } from "./recurrent-consent"
import type { Plan, PlanID } from "@/api/types"

const planIcons: Record<string, LucideIcon> = {
  pro: Sparkles,
  pro_yearly: Sparkles,
  max: Crown,
  max_yearly: Crown,
}

const planColors: Record<string, string> = {
  pro: "#8b5cf6",
  pro_yearly: "#8b5cf6",
  max: "#f59e0b",
  max_yearly: "#f59e0b",
}

interface CheckoutConfirmDialogProps {
  // plan === null → диалог закрыт. Когда юзер кликает «Получить Pro»,
  // setConfirmPlan(plan) → диалог открывается. onClose → setConfirmPlan(null).
  plan: Plan | null
  // features — top-5 фич для отображения в диалоге. Передаются из pricing.tsx
  // через planFeatures(plan).slice(0, 5) чтобы dialog не дублировал длинный
  // список (он уже виден в карточке тарифа на pricing).
  features: string[]
  onClose: () => void
  onConfirm: (consent: boolean) => void
  isPending: boolean
}

// Дата следующего списания = сегодня + period_days. Точная дата вызывает
// больше доверия чем абстрактное «каждый месяц», особенно для подписок
// с recurrent — T-Bank-партнёры рекомендуют показывать конкретное число.
function nextBillingDate(periodDays: number): string {
  const next = new Date(Date.now() + periodDays * 24 * 60 * 60 * 1000)
  return next.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
  })
}

export function CheckoutConfirmDialog({
  plan,
  features,
  onClose,
  onConfirm,
  isPending,
}: CheckoutConfirmDialogProps) {
  const [consent, setConsent] = useState(false)

  // Reset consent при закрытии — следующее открытие должно начинаться
  // с пустого чек-бокса (юзер должен заново подтвердить, особенно если
  // прошлый раз отказался).
  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setConsent(false)
      onClose()
    }
  }

  // Render-guard: plan=null значит закрыт. Возвращаем nothing, чтобы
  // не держать DialogContent в DOM зря.
  if (!plan) return null

  const price = (plan.price_kop / 100).toLocaleString("ru-RU")
  const period = plan.period_days >= 300 ? "в год" : "в месяц"
  const nextDate = nextBillingDate(plan.period_days)
  const planKey = plan.id as PlanID
  const Icon = planIcons[planKey] ?? Sparkles
  const color = planColors[planKey] ?? "#8b5cf6"

  return (
    <Dialog open onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div
            className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-xl"
            style={{ background: `${color}15`, boxShadow: `inset 0 0 0 1px ${color}25` }}
          >
            <Icon className="h-6 w-6" style={{ color }} />
          </div>
          <DialogTitle className="text-center">Оформить подписку {plan.name}</DialogTitle>
          <DialogDescription className="text-center">
            Подтвердите тариф и согласие на регулярные списания.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2 rounded-lg border border-border bg-muted/30 p-4">
          <div className="flex items-baseline justify-between">
            <span className="text-[0.8rem] text-muted-foreground">Стоимость</span>
            <span className="text-lg font-semibold text-foreground">
              {price}&nbsp;₽ {period}
            </span>
          </div>
          <div className="flex items-baseline justify-between">
            <span className="text-[0.8rem] text-muted-foreground">Следующее списание</span>
            <span className="text-[0.8rem] font-medium text-foreground">{nextDate}</span>
          </div>
        </div>

        {features.length > 0 && (
          <div className="space-y-1.5">
            <p className="text-[0.78rem] font-medium text-foreground">Что включено:</p>
            <ul className="space-y-1.5">
              {features.slice(0, 5).map((f) => (
                <li
                  key={f}
                  className="flex items-start gap-2 text-[0.78rem] text-muted-foreground"
                >
                  <Check className="mt-0.5 h-3.5 w-3.5 shrink-0" style={{ color }} />
                  <span>{f}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        <RecurrentConsent
          plan={plan}
          checked={consent}
          onChange={setConsent}
          idSuffix="checkout-confirm"
        />

        <DialogFooter className="gap-2 sm:gap-2">
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isPending}
          >
            Отмена
          </Button>
          <Button
            disabled={!consent || isPending}
            onClick={() => onConfirm(consent)}
            style={{ background: "var(--brand-gradient)" }}
            className="text-white"
          >
            {isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              "Перейти к оплате"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
