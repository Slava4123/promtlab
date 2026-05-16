// Согласие на регулярные списания — обязательный UI-блок для активации
// рекуррентного эквайринга T-Bank. Требование acq_help@tbank.ru
// (диалог 16.05.2026): покупатель ОБЯЗАН явно отметить сумму, период и
// подтвердить согласие на безакцептные списания. Без этого T-Bank не
// активирует метод Charge — что приводит к ошибке `error_code=10
// Метод Charge заблокирован для данного терминала` при автопродлении
// (инцидент с sub_id=2 14-16.05.2026, см. fix renewal Receipt + это
// согласие).
//
// Используется в трёх местах входа в checkout: pricing.tsx,
// quota-exceeded-dialog.tsx, sign-in.tsx (после login если intent
// сохранён без consent). Кнопка checkout в каждом месте обязана быть
// disabled пока checked=false.

import type { Plan } from "@/api/types"

interface RecurrentConsentProps {
  plan: Plan
  checked: boolean
  onChange: (checked: boolean) => void
  // idSuffix добавляется к htmlFor, чтобы на одной странице несколько
  // карточек тарифов (pricing.tsx показывает Pro+Max) имели уникальные
  // id чек-боксов — без этого второй label кликал по первому input.
  idSuffix?: string
}

export function RecurrentConsent({ plan, checked, onChange, idSuffix = "" }: RecurrentConsentProps) {
  const price = (plan.price_kop / 100).toLocaleString("ru-RU")
  const period = plan.period_days >= 300 ? "год" : "месяц"
  const id = `recurrent-consent-${plan.id}${idSuffix ? `-${idSuffix}` : ""}`

  return (
    <label
      htmlFor={id}
      className="mb-3 flex cursor-pointer items-start gap-2 rounded-lg border border-border bg-muted/20 p-3 text-[0.72rem] leading-relaxed text-muted-foreground transition-colors hover:bg-muted/30"
    >
      <input
        id={id}
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="mt-0.5 h-3.5 w-3.5 shrink-0 cursor-pointer rounded border-border accent-violet-600"
      />
      <span>
        Согласен на регулярное автоматическое списание{" "}
        <strong className="text-foreground">{price}&nbsp;₽</strong> за каждый {period}.
        Подписка продлевается автоматически до отмены в{" "}
        <a href="/settings/subscription" className="underline hover:text-foreground">
          Настройки&nbsp;→&nbsp;Подписка
        </a>
        . Условия —{" "}
        <a
          href="/legal/offer"
          target="_blank"
          rel="noopener noreferrer"
          className="underline hover:text-foreground"
        >
          публичная оферта
        </a>
        .
      </span>
    </label>
  )
}
