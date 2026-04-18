import { useState } from "react"
import { Link } from "react-router-dom"
import {
  ArrowLeft,
  Check,
  Copy,
  KeyRound,
  Plug,
  Sparkles,
  ExternalLink,
  Search,
  PenSquare,
  Trash2,
  Terminal,
  AlertTriangle,
  Users,
  Zap,
  Clock,
} from "lucide-react"
import { toast } from "sonner"

import { cn } from "@/lib/utils"
import { AIShareBlock } from "@/components/help/ai-share-block"

// --- Конфиги клиентов ---

type ClientId = "claude-code" | "claude-desktop" | "cursor" | "windsurf" | "generic"
type ClientConfig = {
  id: ClientId
  label: string
  command?: { lang: "bash" | "json"; code: string; description?: string }[]
  steps?: string[]
}

const HOST_PLACEHOLDER = "https://promtlabs.ru"
const KEY_PLACEHOLDER = "pvlt_ваш_ключ"

const CLIENTS: ClientConfig[] = [
  {
    id: "claude-code",
    label: "Claude Code",
    command: [
      {
        lang: "bash",
        description: "В терминале — одна команда:",
        code: `claude mcp add promptvault --transport http ${HOST_PLACEHOLDER}/mcp \\
  --header "Authorization: Bearer ${KEY_PLACEHOLDER}"`,
      },
    ],
  },
  {
    id: "claude-desktop",
    label: "Claude Desktop",
    command: [
      {
        lang: "json",
        description: "Откройте файл конфига и добавьте сервер:",
        code: `{
  "mcpServers": {
    "promptvault": {
      "url": "${HOST_PLACEHOLDER}/mcp",
      "headers": {
        "Authorization": "Bearer ${KEY_PLACEHOLDER}"
      }
    }
  }
}`,
      },
    ],
    steps: [
      "macOS: ~/Library/Application Support/Claude/claude_desktop_config.json",
      "Windows: %APPDATA%\\Claude\\claude_desktop_config.json",
      "После сохранения — перезапустите Claude Desktop",
    ],
  },
  {
    id: "cursor",
    label: "Cursor",
    steps: [
      "Откройте Settings → MCP Servers → Add Server",
      "Name: promptvault",
      `URL: ${HOST_PLACEHOLDER}/mcp`,
      `Headers: Authorization: Bearer ${KEY_PLACEHOLDER}`,
    ],
  },
  {
    id: "windsurf",
    label: "Windsurf",
    steps: [
      "Откройте Settings → MCP → Add Custom Server",
      "Name: promptvault",
      `URL: ${HOST_PLACEHOLDER}/mcp`,
      "Transport: HTTP / Streamable",
      `Header: Authorization: Bearer ${KEY_PLACEHOLDER}`,
    ],
  },
  {
    id: "generic",
    label: "Любой MCP-клиент",
    steps: [
      `Transport: HTTP (Streamable)`,
      `URL: ${HOST_PLACEHOLDER}/mcp`,
      `Auth header: Authorization: Bearer ${KEY_PLACEHOLDER}`,
      "Для локальной разработки замените домен на http://localhost:8080",
    ],
  },
]

// --- Tools ---

const TOOLS_READ = [
  ["search_prompts", "Поиск по промптам, коллекциям, тегам"],
  ["search_suggest", "Автодополнение по префиксу"],
  ["list_prompts", "Список промптов с фильтрами (коллекция, теги, избранное)"],
  ["get_prompt", "Получить промпт по ID с полным содержимым"],
  ["prompt_list_pinned", "Список закреплённых промптов"],
  ["prompt_list_recent", "Список недавно использованных промптов"],
  ["list_collections", "Список коллекций с количеством промптов"],
  ["collection_get", "Получить коллекцию по ID"],
  ["list_tags", "Список тегов"],
  ["get_prompt_versions", "История версий промпта"],
] as const

const TOOLS_WRITE = [
  ["create_prompt", "Создать промпт"],
  ["update_prompt", "Обновить промпт (создаёт новую версию)"],
  ["prompt_favorite", "Переключить статус избранного"],
  ["prompt_pin", "Закрепить/открепить промпт (team_wide для команды)"],
  ["prompt_revert", "Откатить промпт к предыдущей версии"],
  ["prompt_increment_usage", "Отметить использование промпта (для аналитики)"],
  ["share_create", "Создать публичную ссылку на промпт"],
  ["collection_update", "Обновить название/описание/цвет/иконку коллекции"],
  ["create_tag", "Создать тег"],
  ["create_collection", "Создать коллекцию для организации промптов"],
] as const

const TOOLS_DELETE = [
  ["delete_prompt", "Удалить промпт (в корзину на 30 дней)"],
  ["delete_collection", "Удалить коллекцию (промпты внутри не затрагиваются)"],
  ["tag_delete", "Удалить тег (промпты не затрагиваются)"],
  ["share_deactivate", "Деактивировать публичную ссылку"],
] as const

const RESOURCES = [
  ["promptvault://collections", "Все коллекции (контекст для LLM)"],
  ["promptvault://tags", "Все теги (контекст для LLM)"],
  ["promptvault://prompts/{id}", "Конкретный промпт по ID"],
] as const

// --- Каталоги-моки ---

type Catalog = {
  name: string
  description: string
  url: string | null
  status: "ready" | "soon"
}

const CATALOGS: Catalog[] = [
  {
    name: "modelcontextprotocol.io",
    description: "Официальный сайт MCP-протокола — инструкции, спецификация, список клиентов",
    url: "https://modelcontextprotocol.io",
    status: "ready",
  },
  {
    name: "Smithery",
    description: "Маркетплейс MCP-серверов с one-click установкой в Claude/Cursor",
    url: null,
    status: "soon",
  },
  {
    name: "MCP Registry",
    description: "Централизованный реестр публичных MCP-серверов",
    url: null,
    status: "soon",
  },
  {
    name: "Awesome MCP Servers",
    description: "GitHub-список курируемых MCP-серверов комьюнити",
    url: null,
    status: "soon",
  },
  {
    name: "Cline Marketplace",
    description: "Каталог расширений и MCP-серверов для VS Code Cline",
    url: null,
    status: "soon",
  },
  {
    name: "PulseMCP",
    description: "Discovery-платформа MCP-серверов с рейтингами и обзорами",
    url: null,
    status: "soon",
  },
]

// --- Component ---

export default function HelpMCPPage() {
  const [activeClient, setActiveClient] = useState<ClientId>("claude-code")
  const current = CLIENTS.find((c) => c.id === activeClient)!

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10">
        <Link
          to="/help"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Назад в Помощь
        </Link>

        <header className="mb-8 space-y-2">
          <div className="flex items-center gap-2 text-sm text-brand-muted-foreground">
            <Plug className="h-4 w-4" />
            MCP-интеграция
          </div>
          <h1 className="text-3xl font-semibold tracking-tight text-foreground md:text-4xl">
            Подключение ПромтЛаб как MCP-сервера
          </h1>
          <p className="text-base text-muted-foreground">
            ПромтЛаб реализует <a href="https://modelcontextprotocol.io/" target="_blank" rel="noreferrer" className="underline decoration-dotted hover:text-foreground">Model Context Protocol</a> — открытый стандарт для подключения данных к AI-клиентам. Через MCP ваши промпты, коллекции и теги становятся доступны прямо из Claude Code, Claude Desktop, Cursor, Windsurf и других совместимых клиентов.
          </p>
        </header>

        <AIShareBlock
          mdUrl="/help/mcp.md"
          topic="настроить MCP-сервер ПромтЛаб для Claude Code, Claude Desktop, Cursor или Windsurf"
          className="mb-8"
        />

        {/* --- Быстрый старт --- */}
        <Section icon={Sparkles} title="Быстрый старт">
          <ol className="space-y-3">
            <Step
              number={1}
              icon={KeyRound}
              title="Создайте API-ключ"
              body={
                <>
                  Откройте <Link to="/settings/integrations" className="text-brand-muted-foreground underline-offset-4 hover:underline">Настройки → API-ключи</Link>, нажмите «Создать», задайте название (например, «Claude Code на ноутбуке»). Ключ показывается <strong>один раз</strong> — скопируйте сразу.
                </>
              }
            />
            <Step
              number={2}
              icon={Plug}
              title="Подключите MCP-сервер в клиенте"
              body={<>Выберите свой клиент ниже и скопируйте команду или конфиг — поставьте свой ключ вместо <code className="rounded bg-muted px-1 text-[0.78em]">{KEY_PLACEHOLDER}</code>.</>}
            />
            <Step
              number={3}
              icon={Sparkles}
              title="Начните пользоваться"
              body={<>В чате клиента: <em>«Найди мой промпт про код-ревью»</em> или <em>«Создай промпт для рефакторинга»</em>. Полный список команд — в разделе ниже.</>}
            />
          </ol>
        </Section>

        {/* --- Подключение клиентов --- */}
        <Section icon={Plug} title="Подключение клиентов">
          <div className="mb-4 -mx-4 flex gap-1.5 overflow-x-auto px-4 pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            {CLIENTS.map((c) => (
              <button
                key={c.id}
                onClick={() => setActiveClient(c.id)}
                className={cn(
                  "flex shrink-0 items-center rounded-full border px-3 py-1.5 text-[0.8rem] transition-colors min-h-[36px]",
                  activeClient === c.id
                    ? "border-brand/40 bg-brand-muted text-brand-muted-foreground font-medium"
                    : "border-border bg-background text-muted-foreground hover:text-foreground",
                )}
              >
                {c.label}
              </button>
            ))}
          </div>

          <div className="space-y-4">
            {current.command?.map((cmd, i) => (
              <div key={i}>
                {cmd.description && (
                  <p className="mb-2 text-sm text-muted-foreground">{cmd.description}</p>
                )}
                <CodeBlock lang={cmd.lang} code={cmd.code} />
              </div>
            ))}
            {current.steps && (
              <ol className="space-y-2">
                {current.steps.map((s, i) => (
                  <li key={i} className="flex gap-2 text-sm text-foreground">
                    <span className="shrink-0 text-muted-foreground">{i + 1}.</span>
                    <span>{s}</span>
                  </li>
                ))}
              </ol>
            )}
          </div>
        </Section>

        {/* --- Что доступно: Tools --- */}
        <Section icon={Terminal} title="Что MCP умеет (24 tool)">
          <p className="mb-4 text-sm text-muted-foreground">
            Все операции работают и в личном пространстве, и в команде (через параметр <code className="rounded bg-muted px-1 text-[0.78em]">team_id</code>). Запись недоступна для роли <strong>viewer</strong>.
          </p>
          <ToolGroup icon={Search} title="Чтение" colorClass="text-emerald-500" tools={TOOLS_READ} viewerOk />
          <ToolGroup icon={PenSquare} title="Запись" colorClass="text-amber-500" tools={TOOLS_WRITE} />
          <ToolGroup icon={Trash2} title="Удаление" colorClass="text-red-500" tools={TOOLS_DELETE} />
        </Section>

        {/* --- Resources & Prompts --- */}
        <Section icon={Sparkles} title="Resources и Prompts">
          <div className="mb-5">
            <h3 className="mb-2 text-sm font-semibold text-foreground">Resources</h3>
            <p className="mb-3 text-xs text-muted-foreground">URI, которые LLM подгружает как контекст без явной команды.</p>
            <div className="space-y-1">
              {RESOURCES.map(([uri, desc]) => (
                <div key={uri} className="flex flex-col gap-0.5 rounded-lg border border-border bg-card/50 px-3 py-2 sm:flex-row sm:items-center sm:gap-3">
                  <code className="text-[0.78rem] text-brand-muted-foreground">{uri}</code>
                  <span className="text-xs text-muted-foreground">{desc}</span>
                </div>
              ))}
            </div>
          </div>
          <div>
            <h3 className="mb-2 text-sm font-semibold text-foreground">Prompts</h3>
            <p className="mb-3 text-xs text-muted-foreground">Готовые шаблоны, которые клиент может вставить как сообщение.</p>
            <div className="rounded-lg border border-border bg-card/50 px-3 py-2">
              <code className="text-[0.78rem] text-brand-muted-foreground">use_prompt</code>
              <span className="ml-3 text-xs text-muted-foreground">Загрузить промпт из библиотеки и отформатировать для LLM</span>
            </div>
          </div>
        </Section>

        {/* --- Команды и роли --- */}
        <Section icon={Users} title="Работа в командах">
          <p className="mb-4 text-sm text-muted-foreground">
            Все tools принимают необязательный <code className="rounded bg-muted px-1 text-[0.78em]">team_id</code>. Без него — личное пространство, с ним — командное (если у вас есть доступ).
          </p>
          <CodeBlock lang="bash" code={`"Найди мой промпт для код-ревью в команде"
→ search_prompts(query="код-ревью", team_id=2)

"Создай промпт в командном пространстве"
→ create_prompt(title="...", content="...", team_id=2)`} />

          <div className="mt-5 overflow-hidden rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-xs uppercase tracking-wider text-muted-foreground">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">Роль</th>
                  <th className="px-3 py-2 text-left font-medium">Чтение</th>
                  <th className="px-3 py-2 text-left font-medium">Запись</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                <tr><td className="px-3 py-2 font-medium">owner</td><td className="px-3 py-2 text-emerald-500">✓</td><td className="px-3 py-2 text-emerald-500">✓</td></tr>
                <tr><td className="px-3 py-2 font-medium">editor</td><td className="px-3 py-2 text-emerald-500">✓</td><td className="px-3 py-2 text-emerald-500">✓</td></tr>
                <tr><td className="px-3 py-2 font-medium">viewer</td><td className="px-3 py-2 text-emerald-500">✓</td><td className="px-3 py-2 text-muted-foreground">—</td></tr>
              </tbody>
            </table>
          </div>
        </Section>

        {/* --- Примеры --- */}
        <Section icon={Sparkles} title="Примеры запросов">
          <Example
            title="Поиск и получение промпта"
            code={`"Найди промпты про TypeScript"
→ search_prompts(query="TypeScript")
→ get_prompt(id=42)`}
          />
          <Example
            title="Создание промпта с тегами и коллекцией"
            code={`"Создай промпт для рефакторинга кода"
→ list_tags()                  # доступные теги
→ list_collections()           # доступные коллекции
→ create_prompt(
    title="Рефакторинг кода",
    content="Ты — эксперт по рефакторингу...",
    tag_ids=[1, 3],
    collection_ids=[2]
  )`}
          />
          <Example
            title="История версий и откат"
            code={`"Покажи историю изменений промпта #10"
→ get_prompt_versions(prompt_id=10)

"Откати промпт #10 к версии #3"
→ prompt_revert(prompt_id=10, version_id=3)`}
          />
          <Example
            title="Закреплённые промпты команды"
            code={`"Закрепи промпт #5 для всей команды"
→ prompt_pin(id=5, team_wide=true)

"Покажи все закреплённые промпты"
→ prompt_list_pinned()`}
          />
          <Example
            title="Шаринг публичной ссылкой"
            code={`"Поделись промптом #5"
→ share_create(prompt_id=5)
# → { url: "https://promtlabs.ru/s/abc123" }`}
          />
        </Section>

        {/* --- Лимиты --- */}
        <Section icon={Zap} title="Лимиты">
          <div className="grid gap-3 sm:grid-cols-3">
            <LimitCard label="API-ключей на пользователя" value="до 5" />
            <LimitCard label="Запросов в минуту с IP" value="120" />
            <LimitCard label="Запросов в минуту на пользователя" value="60" />
          </div>
          <p className="mt-3 text-xs text-muted-foreground">
            Максимум 100 записей на страницу в list-операциях. При превышении — ответ <code className="rounded bg-muted px-1 text-[0.78em]">429</code> с заголовком <code className="rounded bg-muted px-1 text-[0.78em]">Retry-After</code>.
          </p>
        </Section>

        {/* --- Troubleshooting --- */}
        <Section icon={AlertTriangle} title="Если что-то пошло не так">
          <Trouble
            title="«unauthorized» / 401"
            body={
              <ul className="list-disc pl-5 space-y-1">
                <li>Заголовок должен быть точно <code className="rounded bg-muted px-1 text-[0.78em]">Authorization: Bearer pvlt_…</code></li>
                <li>Ключ не отозван — проверьте в <Link to="/settings/integrations" className="underline">Настройки → API-ключи</Link></li>
                <li>MCP включён на сервере (<code className="rounded bg-muted px-1 text-[0.78em]">MCP_ENABLED=true</code>)</li>
              </ul>
            }
          />
          <Trouble
            title="«read-only access» / 403 на write-операции"
            body={<>Ваша роль в команде — <strong>viewer</strong>. Попросите owner или editor повысить вам права.</>}
          />
          <Trouble
            title="«too many requests» / 429"
            body={<>Превышен лимит. Заголовок <code className="rounded bg-muted px-1 text-[0.78em]">Retry-After: 60</code> — ждите указанное число секунд.</>}
          />
          <Trouble
            title="Нет подключения"
            body={
              <ul className="list-disc pl-5 space-y-1">
                <li>URL должен быть <code className="rounded bg-muted px-1 text-[0.78em]">/mcp</code> (не <code className="rounded bg-muted px-1 text-[0.78em]">/api/mcp</code>)</li>
                <li>Локально: <code className="rounded bg-muted px-1 text-[0.78em]">http://localhost:8080/mcp</code></li>
                <li>Прод: убедитесь что HTTPS, не HTTP</li>
              </ul>
            }
          />
        </Section>

        {/* --- Каталоги (мок) --- */}
        <Section icon={ExternalLink} title="Найти ПромтЛаб в каталогах MCP">
          <p className="mb-4 text-sm text-muted-foreground">
            Места, где можно открыть для себя MCP-серверы — мы туда добавимся в ближайшее время. Пока ссылки приведены для общего ориентира.
          </p>
          <div className="grid gap-3 sm:grid-cols-2">
            {CATALOGS.map((c) => (
              <CatalogCard key={c.name} catalog={c} />
            ))}
          </div>
        </Section>
      </div>
    </div>
  )
}

// --- Components ---

function Section({
  icon: Icon,
  title,
  children,
}: {
  icon: typeof Plug
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="mb-10">
      <h2 className="mb-4 flex items-center gap-2 text-lg font-semibold text-foreground">
        <Icon className="h-4 w-4 text-brand-muted-foreground" />
        {title}
      </h2>
      <div>{children}</div>
    </section>
  )
}

function Step({
  number,
  icon: Icon,
  title,
  body,
}: {
  number: number
  icon: typeof Plug
  title: string
  body: React.ReactNode
}) {
  return (
    <li className="flex gap-3 rounded-xl border border-border bg-card/50 p-4">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand-muted text-sm font-semibold text-brand-muted-foreground">
        {number}
      </div>
      <div className="min-w-0 flex-1">
        <div className="mb-1 flex items-center gap-2 text-sm font-semibold text-foreground">
          <Icon className="h-3.5 w-3.5 text-brand-muted-foreground" />
          {title}
        </div>
        <div className="text-sm text-muted-foreground">{body}</div>
      </div>
    </li>
  )
}

function CodeBlock({ lang, code }: { lang: "bash" | "json"; code: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      toast.success("Скопировано")
      setTimeout(() => setCopied(false), 1500)
    } catch {
      toast.error("Не удалось скопировать")
    }
  }
  return (
    <div className="relative overflow-hidden rounded-lg border border-border bg-muted/40">
      <div className="flex items-center justify-between border-b border-border px-3 py-1.5">
        <span className="text-[0.7rem] uppercase tracking-wider text-muted-foreground">{lang}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 rounded-md px-2 py-1 text-[0.72rem] text-muted-foreground hover:bg-foreground/[0.04] hover:text-foreground"
          aria-label="Скопировать код"
        >
          {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
          {copied ? "Готово" : "Копировать"}
        </button>
      </div>
      <pre className="overflow-x-auto px-3 py-3 text-[0.78rem] leading-relaxed text-foreground">
        <code>{code}</code>
      </pre>
    </div>
  )
}

function ToolGroup({
  icon: Icon,
  title,
  colorClass,
  tools,
  viewerOk,
}: {
  icon: typeof Plug
  title: string
  colorClass: string
  tools: readonly (readonly [string, string])[]
  viewerOk?: boolean
}) {
  return (
    <div className="mb-5 last:mb-0">
      <div className="mb-2 flex items-center gap-2 text-sm font-semibold text-foreground">
        <Icon className={cn("h-4 w-4", colorClass)} />
        {title}
        <span className="rounded-full border border-border px-1.5 text-[0.65rem] text-muted-foreground">{tools.length}</span>
        {viewerOk && (
          <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-[0.65rem] font-medium text-emerald-500">
            доступно viewer
          </span>
        )}
      </div>
      <div className="overflow-hidden rounded-lg border border-border">
        <ul className="divide-y divide-border">
          {tools.map(([name, desc]) => (
            <li key={name} className="flex flex-col gap-0.5 px-3 py-2 sm:flex-row sm:items-center sm:gap-3">
              <code className="shrink-0 text-[0.78rem] text-brand-muted-foreground sm:w-56">{name}</code>
              <span className="text-xs text-muted-foreground">{desc}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

function Example({ title, code }: { title: string; code: string }) {
  return (
    <div className="mb-4 last:mb-0">
      <h3 className="mb-2 text-sm font-medium text-foreground">{title}</h3>
      <CodeBlock lang="bash" code={code} />
    </div>
  )
}

function LimitCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-card/50 p-3">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Clock className="h-3 w-3" />
        {label}
      </div>
      <div className="mt-1 text-lg font-semibold text-foreground">{value}</div>
    </div>
  )
}

function Trouble({ title, body }: { title: string; body: React.ReactNode }) {
  return (
    <div className="mb-3 last:mb-0 rounded-lg border border-border bg-card/30 p-4">
      <div className="mb-2 text-sm font-semibold text-foreground">{title}</div>
      <div className="text-sm text-muted-foreground">{body}</div>
    </div>
  )
}

function CatalogCard({ catalog }: { catalog: Catalog }) {
  const isReady = catalog.status === "ready" && catalog.url
  const Tag = isReady ? "a" : "div"
  const props = isReady
    ? { href: catalog.url!, target: "_blank", rel: "noreferrer" }
    : { "aria-disabled": true as const }
  return (
    <Tag
      {...props}
      className={cn(
        "group flex flex-col gap-1.5 rounded-xl border p-4 transition-colors",
        isReady
          ? "border-border bg-card/50 hover:border-brand/30 hover:bg-card cursor-pointer"
          : "border-dashed border-border bg-card/20 cursor-not-allowed",
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <span className="font-medium text-foreground">{catalog.name}</span>
        {isReady ? (
          <ExternalLink className="h-3.5 w-3.5 text-muted-foreground group-hover:text-foreground" />
        ) : (
          <span className="rounded-full bg-amber-500/10 px-2 py-0.5 text-[0.65rem] font-medium text-amber-500">
            Скоро
          </span>
        )}
      </div>
      <p className="text-xs text-muted-foreground">{catalog.description}</p>
    </Tag>
  )
}
