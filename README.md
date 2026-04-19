# ПромтЛаб (PromptLab / PromptVault)

> **Self-hosted библиотека AI-промптов** для соло и команд.
> Self-hosted в России. Работает с Claude через встроенный MCP-сервер.

[🌐 promtlabs.ru](https://promtlabs.ru) · [📖 Документация](promptvault/docs/) · [🔌 MCP](promptvault/docs/MCP.md) · [📝 Changelog](https://promtlabs.ru/changelog)

---

## Что это

ПромтЛаб — это self-hosted веб-приложение для управления библиотекой промптов
к Claude (Opus / Sonnet / Haiku). Решает три задачи:

1. **Организация** — коллекции, теги, иерархия, версии, корзина, поиск.
2. **Командная работа** — расшаривание промптов команде, роли (owner / editor / viewer),
   общие коллекции.
3. **Интеграция с Claude** — встроенный MCP-сервер позволяет Claude искать, читать
   и обновлять ваши промпты прямо из чата.

## Особенности

- 🇷🇺 Интерфейс на русском, self-hosted в России (хостинг Timeweb)
- 🔌 **MCP-сервер v1.2+** — 30 tools, опубликован в
  [Official MCP Registry](https://registry.modelcontextprotocol.io/v0/servers?search=promtlab)
- 👥 Команды с ролями (owner / editor / viewer), 2FA для админов
- 💳 Тарифы: Free (0₽), Pro (599₽/мес), Max (1299₽/мес), оплата через Т-Банк
- 🧩 Интеграция с Claude Code, Claude Desktop, Cursor, Windsurf
- 🔐 OAuth 2.0 + PKCE для GitHub / Google / Yandex, TOTP 2FA для админов
- 📊 Версионирование промптов, публичные ссылки, шаринг
- 🤖 AI-улучшение промптов (Claude Sonnet 4) с квотами по тарифам
- 🌐 PWA, Telegram-бот, браузерное расширение

## Стек

**Backend:** Go 1.25 · Chi v5 · GORM v2 · PostgreSQL 18 · slog · koanf
**Frontend:** React 19 · Vite 8 · shadcn/ui · Tailwind 4 · TanStack Query · Zustand
**Infra:** Docker Compose · GlitchTip (self-hosted Sentry) · GitHub Actions CI/CD
**MCP:** `modelcontextprotocol/go-sdk` v1.5, Streamable HTTP transport

Подробнее — [`promptvault/CLAUDE.md`](promptvault/CLAUDE.md) (архитектура,
Clean Architecture слои, конвенции).

## Быстрый старт (dev)

```bash
git clone https://github.com/Slava4123/promtlab.git
cd promtlab/promptvault
cp .env.example .env  # заполнить DATABASE_*, JWT_SECRET, OPENROUTER_API_KEY
docker compose -f docker-compose.dev.yml up
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- MCP endpoint: http://localhost:8080/mcp

Создать первого админа:

```bash
docker compose exec api go run ./cmd/create-admin --email=you@example.com
```

## Подключение к Claude (MCP)

```bash
claude mcp add --transport http promtlab https://promtlabs.ru/mcp \
  --header "Authorization: Bearer pvlt_YOUR_API_KEY"
```

API-ключи создаются в `https://promtlabs.ru/settings/integrations`.

Полный гайд и список всех 30 tools — [`promptvault/docs/MCP.md`](promptvault/docs/MCP.md).

## Лицензия

[**Functional Source License 1.1** (Apache-2.0 Future License)](LICENSE.md)

- ✅ Можно self-host'ить, форкать, модифицировать, использовать внутри компании
- ❌ Нельзя делать конкурирующий SaaS в течение 2 лет
- 🔓 Через 2 года каждый релиз автоматически становится Apache 2.0

Вопросы по коммерческим использованиям: slava0gpt@gmail.com.

## Безопасность

Нашли уязвимость? Пишите на slava0gpt@gmail.com или через
[GitHub Security Advisory](https://github.com/Slava4123/promtlab/security/advisories/new).
Подробности — [`SECURITY.md`](SECURITY.md).

## Контакты

- 🌐 **Сайт:** https://promtlabs.ru
- 📧 **Email:** slava0gpt@gmail.com
- 📬 **Telegram:** @promtlabs

---

*Сделано в России с ❤️ для тех, кто много работает с AI.*
