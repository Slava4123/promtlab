# MCP-сервер ПромтЛаб

ПромтЛаб работает как [MCP-сервер](https://modelcontextprotocol.io/) — ваши промпты доступны прямо из ИИ-клиентов (Claude Code, Cursor, Windsurf и др.).

## Быстрый старт

### 1. Создайте API-ключ

Откройте **Настройки → API-ключи → Создать**. Скопируйте ключ — он показывается один раз.

### 2. Подключите MCP-сервер

#### Claude Code

```bash
claude mcp add promptvault --transport http https://promtlabs.ru/mcp --header "Authorization: Bearer pvlt_ваш_ключ"
```

Или добавьте в `~/.claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "promptvault": {
      "url": "https://promtlabs.ru/mcp",
      "headers": {
        "Authorization": "Bearer pvlt_ваш_ключ"
      }
    }
  }
}
```

#### Cursor

Откройте **Settings → MCP Servers → Add Server**:

- Name: `promptvault`
- URL: `https://promtlabs.ru/mcp`
- Headers: `Authorization: Bearer pvlt_ваш_ключ`

#### Другие клиенты

Любой клиент с поддержкой MCP Streamable HTTP:

- URL: `https://promtlabs.ru/mcp`
- Аутентификация: `Authorization: Bearer pvlt_ваш_ключ`

Для локальной разработки замените `https://promtlabs.ru` на `http://localhost:8080`.

## Возможности

### Tools (30 шт.)

#### Чтение (14)

| Tool | Описание | Viewer |
|------|----------|--------|
| `whoami` | Текущий пользователь (id, email, plan, default_model) | ✅ |
| `search_prompts` | Поиск по промптам, коллекциям, тегам | ✅ |
| `search_suggest` | Автодополнение по префиксу | ✅ |
| `list_prompts` | Список промптов с фильтрами (коллекция, теги, избранное) | ✅ |
| `get_prompt` | Получить промпт по ID с полным содержимым | ✅ |
| `list_prompt_vars` | Извлечь `{{переменные}}` из промпта | ✅ |
| `prompt_list_pinned` | Список закреплённых промптов | ✅ |
| `prompt_list_recent` | Список недавно использованных промптов | ✅ |
| `list_collections` | Список коллекций с количеством промптов | ✅ |
| `collection_get` | Получить коллекцию по ID | ✅ |
| `list_tags` | Список тегов | ✅ |
| `list_teams` | Список команд пользователя (id, role, member_count) | ✅ |
| `list_trash` | Содержимое корзины (soft-deleted промпты) | ✅ |
| `get_prompt_versions` | История версий промпта | ✅ |

#### Запись (11)

| Tool | Описание | Viewer | Ест квоту |
|------|----------|--------|-----------|
| `create_prompt` | Создать промпт | ❌ | ✅ |
| `update_prompt` | Обновить промпт (создаёт новую версию) | ❌ | ✅ |
| `prompt_favorite` | Переключить статус избранного | ❌ | — |
| `prompt_pin` | Закрепить/открепить промпт (team_wide для команды) | ❌ | — |
| `prompt_revert` | Откатить промпт к предыдущей версии | ❌ | ✅ |
| `prompt_increment_usage` | Отметить использование промпта (для аналитики) | ❌ | — |
| `share_create` | Создать публичную ссылку на промпт | ❌ | ✅ |
| `restore_prompt` | Восстановить промпт из корзины | ❌ | ✅ |
| `collection_update` | Обновить название/описание/цвет/иконку коллекции | ❌ | ✅ |
| `create_tag` | Создать тег | ❌ | ✅ |
| `create_collection` | Создать коллекцию для организации промптов | ❌ | ✅ |

#### Удаление (5)

| Tool | Описание | Viewer | Ест квоту |
|------|----------|--------|-----------|
| `delete_prompt` | Удалить промпт (в корзину на 30 дней) | ❌ | ✅ |
| `delete_collection` | Удалить коллекцию (промпты внутри не затрагиваются) | ❌ | ✅ |
| `tag_delete` | Удалить тег (промпты не затрагиваются) | ❌ | ✅ |
| `share_deactivate` | Деактивировать публичную ссылку | ❌ | ✅ |
| `purge_prompt` | Удалить промпт навсегда (из корзины, необратимо) | ❌ | ✅ |

### Resources

| URI | Описание |
|-----|----------|
| `promptvault://collections` | Все коллекции (контекст для LLM) |
| `promptvault://tags` | Все теги (контекст для LLM) |
| `promptvault://prompts/{id}` | Конкретный промпт по ID |

### Prompts

| Имя | Описание |
|-----|----------|
| `use_prompt` | Загрузить промпт из библиотеки и отформатировать для использования LLM |

## Работа с командами

Все tools поддерживают параметр `team_id` для работы в командном пространстве. Без `team_id` — личное пространство.

```
"Найди мой промпт для код-ревью в команде"
→ search_prompts(query="код-ревью", team_id=2)

"Создай промпт в командном пространстве"
→ create_prompt(title="...", content="...", team_id=2)
```

### Ролевые ограничения

| Роль | Чтение | Запись |
|------|--------|--------|
| **owner** | ✅ | ✅ |
| **editor** | ✅ | ✅ |
| **viewer** | ✅ | ❌ |

Viewer имеет доступ ко всем 14 read-tools (включая `whoami`, `list_teams`, `list_trash`, `list_prompt_vars`). Любые write/destructive операции для viewer запрещены.

## Примеры использования

### Поиск и получение промпта

```
"Найди промпты про TypeScript"
→ search_prompts(query="TypeScript")
→ get_prompt(id=42)
```

### Создание промпта с тегами и коллекцией

```
"Создай промпт для рефакторинга кода"
→ list_tags()                        # получить доступные теги
→ list_collections()                 # получить доступные коллекции
→ create_prompt(
    title="Рефакторинг кода",
    content="Ты — эксперт по рефакторингу...",
    tag_ids=[1, 3],
    collection_ids=[2]
  )
```

### Использование prompt-ресурса

```
"Используй мой промпт #5 для текущей задачи"
→ use_prompt(id="5")
# LLM получает отформатированный промпт: "# Заголовок\n\nСодержимое"
```

### История версий и откат

```
"Покажи историю изменений промпта #10"
→ get_prompt_versions(prompt_id=10)

"Откати промпт #10 к версии #3"
→ prompt_revert(prompt_id=10, version_id=3)
```

### Управление избранным и закреплением

```
"Добавь промпт #5 в избранное"
→ prompt_favorite(id=5)

"Закрепи промпт #5 для всей команды"
→ prompt_pin(id=5, team_wide=true)

"Покажи мои закреплённые промпты"
→ prompt_list_pinned()
```

### Шаринг

```
"Поделись промптом #5"
→ share_create(prompt_id=5)
# → { url: "https://promtlabs.ru/s/abc123" }

"Отключи ссылку на промпт #5"
→ share_deactivate(prompt_id=5)
```

## API-ключи

- Максимум **5 ключей** на пользователя
- Ключ показывается **один раз** при создании
- Управление: **Настройки → API-ключи**
- Формат: `pvlt_` + 43 символа

## Квотирование MCP

ПромтЛаб тарифицирует только реальные изменения в БД. Из 30 инструментов:

- **Бесплатны** (14 read + 3 UX-toggle + `use_prompt`): `whoami`, `search_prompts`, `search_suggest`, `list_prompts`, `list_collections`, `list_tags`, `list_teams`, `list_trash`, `list_prompt_vars`, `get_prompt`, `get_prompt_versions`, `prompt_list_pinned`, `prompt_list_recent`, `collection_get`, `prompt_favorite`, `prompt_pin`, `prompt_increment_usage`.
- **Едят дневную MCP-квоту** (13 write/destructive): `create_prompt`, `update_prompt`, `delete_prompt`, `restore_prompt`, `purge_prompt`, `prompt_revert`, `create_collection`, `collection_update`, `delete_collection`, `create_tag`, `tag_delete`, `share_create`, `share_deactivate`.

Лимиты по тарифам:

| Тариф | Платных вызовов в день |
|-------|-----------------------|
| Free | 5 |
| Pro | 30 |
| Max | Безлимит |

При превышении backend возвращает `402 Payment Required` c сообщением «MCP quota exceeded».

## Лимиты

- **120 запросов/мин** на IP
- **60 запросов/мин** на пользователя
- Максимум **100 записей** на страницу в list-операциях

## Переменные окружения

```bash
MCP_ENABLED=true          # включить MCP-сервер (по умолчанию false)
MCP_MAX_KEYS_PER_USER=5   # лимит ключей на пользователя
```

## Troubleshooting

### "unauthorized"

- Проверьте формат заголовка: `Authorization: Bearer pvlt_ваш_ключ`
- Убедитесь, что ключ не отозван (Настройки → API-ключи)
- Проверьте, что MCP включён (`MCP_ENABLED=true`)

### "read-only access"

- У вас роль **viewer** в команде — запись недоступна
- Попросите owner/editor повысить вашу роль

### "too many requests"

- Превышен лимит: 120 req/мин на IP или 60 req/мин на пользователя
- Заголовок `Retry-After: 60` указывает время ожидания

### Нет подключения

- Проверьте URL: `/mcp` (не `/api/mcp`)
- Убедитесь, что протокол правильный (HTTP vs HTTPS)
- Для локальной разработки: `http://localhost:8080/mcp`
