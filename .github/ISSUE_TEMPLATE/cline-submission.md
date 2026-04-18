---
name: MCP Marketplace submission (Cline) — draft
about: Шпаргалка для подачи ПромтЛаб в github.com/cline/mcp-marketplace
title: "[MCP Marketplace] Add PromptLab — AI prompt library with teams, tags, versioning"
labels: submission
---

<!--
Это не issue для этого репозитория, а шаблон текста для подачи в
https://github.com/cline/mcp-marketplace/issues/new

Скопируйте секции ниже в новую issue в репозитории cline/mcp-marketplace.
Меняйте только README / Demo GIF / Logo если нужно.
-->

### GitHub Repo URL

https://github.com/Slava4123/promtlab

### Logo (400×400 PNG)

https://promtlabs.ru/logo-mcp-400.png

### Brief Description

ПромтЛаб — self-hosted AI prompt library for individuals and teams. Remote MCP server over Streamable HTTP with 30 tools: CRUD для promts/collections/tags, поиск с autocomplete, версионирование и откат, корзина (delete/restore/purge), команды (list_teams), share-ссылки. Авторизация API-ключом. Русскоязычный интерфейс.

### Why are you submitting this server?

- Покрывает end-to-end workflow работы с промптами прямо из AI-клиента: `use_prompt id=X` — подстановка `{{переменных}}` и вставка как сообщение.
- Self-hosted, данные остаются у пользователя. Хостинг в России, что важно для аудитории РФ.
- 30 tools с safety annotations (ReadOnlyHint/DestructiveHint), scoped API-keys с allowed_tools whitelist, квотирование 13 платных операций.

### Testing instructions

1. Зайти на https://promtlabs.ru, создать аккаунт.
2. Settings → API-ключи → Создать → скопировать `pvlt_...`.
3. В Cline / Claude Desktop добавить MCP-сервер:
   - URL: `https://promtlabs.ru/mcp`
   - Transport: Streamable HTTP
   - Header: `Authorization: Bearer pvlt_ВАШ_КЛЮЧ`
4. Вызовы: `whoami`, `list_teams`, `list_prompts`, `create_prompt(title=..., content=...)`.

### Server manifest (server.json)

https://github.com/Slava4123/promtlab/blob/main/promptvault/server.json

### Documentation

- User-facing: https://promtlabs.ru/help/mcp
- Developer: https://github.com/Slava4123/promtlab/blob/main/promptvault/docs/MCP.md
- Published in Official MCP Registry: https://registry.modelcontextprotocol.io/v0/servers?search=promtlab

### Author

- GitHub: [@Slava4123](https://github.com/Slava4123)
- Contact: slava0gpt@gmail.com
