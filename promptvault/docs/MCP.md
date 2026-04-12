# MCP-сервер ПромтЛаб

ПромтЛаб работает как [MCP-сервер](https://modelcontextprotocol.io/) — ваши промпты доступны прямо из ИИ-клиентов (Claude Code, Cursor, Windsurf и др.).

## Быстрый старт

### 1. Создайте API-ключ

Откройте **Настройки → API-ключи → Создать**. Скопируйте ключ — он показывается один раз.

### 2. Подключите MCP-сервер

#### Claude Code

```bash
claude mcp add promptvault --transport http https://ваш-домен/mcp --header "Authorization: Bearer pvlt_ваш_ключ"
```

Или добавьте в `~/.claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "promptvault": {
      "url": "https://ваш-домен/mcp",
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
- URL: `https://ваш-домен/mcp`
- Headers: `Authorization: Bearer pvlt_ваш_ключ`

#### Другие клиенты

Любой клиент с поддержкой MCP Streamable HTTP:

- URL: `https://ваш-домен/mcp`
- Аутентификация: `Authorization: Bearer pvlt_ваш_ключ`

Для локальной разработки замените `https://ваш-домен` на `http://localhost:8080`.

## Возможности

### Tools (12 шт.)

#### Чтение

| Tool | Описание | Viewer |
|------|----------|--------|
| `search_prompts` | Поиск по промптам, коллекциям, тегам | ✅ |
| `list_prompts` | Список промптов с фильтрами (коллекция, теги, избранное) | ✅ |
| `get_prompt` | Получить промпт по ID с полным содержимым | ✅ |
| `list_collections` | Список коллекций с количеством промптов | ✅ |
| `list_tags` | Список тегов | ✅ |
| `get_prompt_versions` | История версий промпта | ✅ |

#### Запись

| Tool | Описание | Viewer |
|------|----------|--------|
| `create_prompt` | Создать промпт | ❌ |
| `update_prompt` | Обновить промпт (создаёт новую версию) | ❌ |
| `delete_prompt` | Удалить промпт (в корзину на 30 дней) | ❌ |
| `create_tag` | Создать тег | ❌ |
| `create_collection` | Создать коллекцию для организации промптов | ❌ |
| `delete_collection` | Удалить коллекцию (промпты внутри не затрагиваются) | ❌ |

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

Viewer имеет доступ только к read-tools: `search_prompts`, `list_prompts`, `get_prompt`, `list_collections`, `list_tags`, `get_prompt_versions`.

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

### История версий

```
"Покажи историю изменений промпта #10"
→ get_prompt_versions(prompt_id=10)
```

## API-ключи

- Максимум **5 ключей** на пользователя
- Ключ показывается **один раз** при создании
- Управление: **Настройки → API-ключи**
- Формат: `pvlt_` + 43 символа

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
