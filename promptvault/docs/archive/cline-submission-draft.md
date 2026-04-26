# Cline MCP Marketplace — draft submission

Draft для подачи в https://github.com/cline/mcp-marketplace.

## Процедура подачи

1. Открыть https://github.com/cline/mcp-marketplace/issues/new?template=mcp-server-submission.yml
2. Заполнить форму значениями ниже (поля точно соответствуют их YAML-форме).
3. Отправить. Ответ от мейнтейнеров обычно за несколько дней.

---

## Поля формы

### GitHub Repository URL

```
https://github.com/Slava4123/promtlab
```

### Logo Image (400×400 PNG)

```
https://promtlabs.ru/logo-mcp-400.png
```

Можно также перетащить в поле файл
`promptvault/frontend/public/logo-mcp-400.png` из этого репозитория.

### Installation Testing (checkboxes — оба должны быть ✅)

- ✅ I have tested that Cline can successfully set up this server using only the README.md and/or llms-install.md file
- ✅ The server is stable and ready for public use

(Тестовая установка: Cline по `promptvault/llms-install.md` понимает
что это remote HTTP-сервер, ничего не клонирует, просто добавляет
URL + Authorization header в `cline_mcp_settings.json`.)

### Additional Information

```
ПромтЛаб — self-hosted AI prompt library для частных пользователей и
команд. MCP-сервер работает поверх Streamable HTTP на
https://promtlabs.ru/mcp — Cline подключается через URL + Bearer
token, никакого клонирования или npm-install не требуется (см.
promptvault/llms-install.md).

Возможности (30 tools):
- CRUD для промптов / коллекций / тегов
- Поиск и автодополнение, закреплённые и недавние
- Версионирование каждого изменения с откатом (prompt_revert)
- Корзина на 30 дней с list/restore/purge
- Командные пространства (list_teams)
- Публичные share-ссылки (share_create / share_deactivate)
- Идентификация текущего аккаунта (whoami)
- Extract переменных {{...}} + use_prompt для подстановки

Каждый tool имеет safety-аннотации (ReadOnlyHint, DestructiveHint,
IdempotentHint). Scoped API-keys с allowed_tools whitelist для
минимальных прав. Квотирование 13 платных write/destructive операций
(Free 5/день, Pro 30/день, Max безлимит).

Уже опубликованы в Official MCP Registry:
https://registry.modelcontextprotocol.io/v0/servers?search=promtlab

Документация: https://promtlabs.ru/help/mcp
Server manifest: https://github.com/Slava4123/promtlab/blob/main/promptvault/server.json

Аудитория — русскоязычные разработчики, для которых self-hosting
удобнее западных SaaS. Интерфейс на русском, платёжка через Т-Банк.
Для Cline-пользователей, работающих с prompt engineering — одно
место для библиотеки промптов, доступной из любого MCP-клиента.
```
