import { useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { ChevronDown, LifeBuoy, Mail, Search } from "lucide-react"

interface FaqItem {
  question: string
  answer: string
  tags: string[]
}

const FAQ: FaqItem[] = [
  {
    question: "Что такое ПромтЛаб?",
    answer:
      "Self-hosted хранилище AI-промптов с версионированием, AI-улучшением, командной работой, расширением для браузера и MCP-сервером для Claude.",
    tags: ["общее"],
  },
  {
    question: "Чем отличаются тарифы Free, Pro и Max?",
    answer:
      "Free: 50 промптов, 3 коллекции, 5 AI-запросов на весь период. Pro (599₽/мес): 500 промптов, безлимит коллекций, 10 AI-запросов в день. Max (1299₽/мес): безлимит во всём, 15 AI-запросов в день. Годовой план — скидка 10%.",
    tags: ["тарифы", "оплата"],
  },
  {
    question: "Как отменить подписку?",
    answer:
      "В Настройках → Подписка → «Отменить подписку». Вам предложат выбрать причину (это помогает улучшить продукт) и подтвердить. Доступ сохранится до конца оплаченного периода.",
    tags: ["подписка", "оплата"],
  },
  {
    question: "Можно ли поставить подписку на паузу?",
    answer:
      "Да. В Настройках → Подписка → «Приостановить» можно заморозить подписку на 1, 2 или 3 месяца. Оставшиеся дни сохранятся. Во время паузы аккаунт работает как Free, в конце паузы автоматически активируется обратно.",
    tags: ["подписка"],
  },
  {
    question: "Как пригласить друга и получить бонус?",
    answer:
      "В Настройках → «Пригласить друзей» возьмите свой код или ссылку и отправьте другу. После его первой платной подписки вам продлят Pro на 30 дней. Ограничений на количество приглашений нет.",
    tags: ["подписка", "реферал"],
  },
  {
    question: "Что такое публичные промпты?",
    answer:
      "В редакторе промпта можно поставить «Публичный промпт» — он становится доступен по ссылке /p/<slug> без авторизации. Хорошо для шаринга в блоге или соцсетях и для SEO.",
    tags: ["промпты"],
  },
  {
    question: "Как подключить расширение для Chrome/Firefox?",
    answer:
      "Установите из магазина по ссылке в Настройках → Расширение. После входа в аккаунт горячей клавишей можно искать промпты и вставлять их в ChatGPT, Claude и другие веб-чаты.",
    tags: ["расширение"],
  },
  {
    question: "Как подключить MCP-сервер для Claude Desktop/Code?",
    answer:
      "В Настройках → API keys создайте ключ, затем в Claude Desktop/Code пропишите MCP-сервер с HTTP-транспортом и этим ключом. Подробная инструкция с примером конфига — в FAQ статье «MCP integration».",
    tags: ["mcp", "claude"],
  },
  {
    question: "Что делать, если не пришёл код подтверждения?",
    answer:
      "Проверьте «Спам»/«Промоакции». Код валиден 15 минут. Если не пришёл — нажмите «Отправить ещё раз». Если проблема повторяется — напишите на slava0gpt@gmail.com.",
    tags: ["авторизация", "email"],
  },
  {
    question: "Забыл пароль. Что делать?",
    answer:
      "На странице входа нажмите «Забыли пароль» и следуйте инструкции — мы отправим код восстановления на email. Также можно войти через GitHub/Google/Yandex, если вы настраивали привязку.",
    tags: ["авторизация"],
  },
  {
    question: "Безопасно ли хранить промпты на сервере?",
    answer:
      "Мы self-hosted в России, шифруем соединение, пароли храним как bcrypt-хеши, JWT-сессии с nonce (invalidate по выходу). API-ключи хранятся хеш-функцией, показываются только один раз. Подробности — в Политике конфиденциальности.",
    tags: ["безопасность"],
  },
  {
    question: "Как отключить автопродление подписки?",
    answer:
      "В Настройках → Подписка снимите галочку «Автопродление». Подписка отработает до конца оплаченного периода и не будет списана автоматически.",
    tags: ["подписка", "оплата"],
  },
]

export default function Help() {
  const [query, setQuery] = useState("")
  const [openIdx, setOpenIdx] = useState<number | null>(null)

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return FAQ
    return FAQ.filter(
      (f) =>
        f.question.toLowerCase().includes(q) ||
        f.answer.toLowerCase().includes(q) ||
        f.tags.some((t) => t.includes(q)),
    )
  }, [query])

  return (
    <div className="mx-auto max-w-3xl px-4 py-10">
      <header className="mb-8">
        <div className="flex items-center gap-3">
          <LifeBuoy className="h-6 w-6 text-violet-400" aria-hidden="true" />
          <h1 className="text-2xl font-bold tracking-tight">Центр поддержки</h1>
        </div>
        <p className="mt-2 text-sm text-muted-foreground">
          Частые вопросы. Не нашли ответ — напишите на{" "}
          <a href="mailto:slava0gpt@gmail.com" className="underline underline-offset-4 hover:text-foreground">
            slava0gpt@gmail.com
          </a>
          , обычно отвечаем в течение рабочего дня.
        </p>
      </header>

      <div className="relative mb-6">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Поиск по вопросам…"
          className="h-11 w-full rounded-lg border border-border bg-background pl-10 pr-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
        />
      </div>

      {filtered.length === 0 ? (
        <div className="rounded-xl border border-border bg-card px-6 py-8 text-center">
          <p className="text-sm text-muted-foreground">
            Ничего не нашли по запросу. Напишите нам — разберёмся.
          </p>
          <a
            href="mailto:slava0gpt@gmail.com"
            className="mt-4 inline-flex items-center gap-1.5 text-sm font-medium text-violet-400 hover:text-violet-300"
          >
            <Mail className="h-4 w-4" aria-hidden="true" />
            slava0gpt@gmail.com
          </a>
        </div>
      ) : (
        <ul className="space-y-2">
          {filtered.map((item, i) => {
            const isOpen = openIdx === i
            return (
              <li key={item.question} className="rounded-xl border border-border bg-card">
                <button
                  type="button"
                  onClick={() => setOpenIdx(isOpen ? null : i)}
                  aria-expanded={isOpen}
                  className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left text-sm font-medium text-foreground transition-colors hover:bg-muted/30"
                >
                  <span className="flex-1">{item.question}</span>
                  <ChevronDown
                    className={`h-4 w-4 shrink-0 text-muted-foreground transition-transform ${isOpen ? "rotate-180" : ""}`}
                    aria-hidden="true"
                  />
                </button>
                {isOpen && (
                  <div className="border-t border-border px-4 py-3 text-[0.88rem] leading-relaxed text-muted-foreground">
                    {item.answer}
                  </div>
                )}
              </li>
            )
          })}
        </ul>
      )}

      <footer className="mt-10 rounded-xl border border-border bg-muted/20 px-4 py-4 text-sm">
        <p className="text-foreground">Нужна ещё помощь?</p>
        <p className="mt-1 text-muted-foreground">
          Напишите на{" "}
          <a href="mailto:slava0gpt@gmail.com" className="font-medium text-foreground underline underline-offset-4">
            slava0gpt@gmail.com
          </a>{" "}
          — отвечаем в течение рабочего дня. Если вопрос по оплате, приложите номер платежа.
        </p>
        <p className="mt-2 text-[0.8rem] text-muted-foreground">
          Смотрите также:{" "}
          <Link to="/legal/terms" className="underline underline-offset-4 hover:text-foreground">
            Условия использования
          </Link>
          ,{" "}
          <Link to="/legal/privacy" className="underline underline-offset-4 hover:text-foreground">
            Политика конфиденциальности
          </Link>
          ,{" "}
          <Link to="/legal/offer" className="underline underline-offset-4 hover:text-foreground">
            Публичная оферта
          </Link>
          .
        </p>
      </footer>
    </div>
  )
}
