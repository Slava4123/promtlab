# Публикация MCP-сервера PromptLab в реестрах

## Статус: запланировано

---

## 1. Official MCP Registry (главный приоритет)

**URL:** https://registry.modelcontextprotocol.io/
**Охват:** Все MCP-клиенты (Claude, Cursor, Windsurf, Cline и др.)

Старый репозиторий `modelcontextprotocol/servers` больше НЕ принимает PR. Всё через CLI `mcp-publisher`.

### Процесс

```bash
# 1. Установить CLI
curl -fsSL https://registry.modelcontextprotocol.io/install | sh

# 2. Сгенерировать server.json
mcp-publisher init

# 3. Авторизоваться (GitHub OAuth → namespace io.github.username/*)
mcp-publisher login github

# 4. Или DNS/HTTP-верификация домена → namespace ru.promptvault/*
# mcp-publisher login domain

# 5. Опубликовать
mcp-publisher publish
```

### server.json (шаблон для PromptLab)

```json
{
  "name": "com.promptvault/server",
  "description": "AI prompt management server — search, organize, version your prompts from any MCP client",
  "remotes": [{
    "type": "streamable-http",
    "url": "https://{domain}/mcp",
    "configSchema": {
      "properties": {
        "domain": {
          "type": "string",
          "description": "Your PromptLab instance domain (e.g. app.promptlab.ru)"
        }
      },
      "required": ["domain"]
    },
    "headers": [{
      "name": "Authorization",
      "description": "API key from Settings → API Keys (format: pvlt_...)",
      "isRequired": true,
      "isSecret": true
    }]
  }],
  "packages": [{
    "registryType": "oci",
    "name": "ghcr.io/youruser/promptvault"
  }]
}
```

---

## 2. Anthropic Connectors Directory (Claude.com)

**URL:** https://claude.com/connectors
**Охват:** Десятки миллионов пользователей Claude
**Подача:** https://forms.gle/tyiAZvch1kDADKoP9

### Требования

- Минимум 3 рабочих примера использования
- Все tools ДОЛЖНЫ иметь safety annotations (ReadOnlyHint, DestructiveHint) — уже сделано
- Соответствие Anthropic MCP Directory Policy
- Подача НЕ гарантирует включение — кураторский отбор

---

## 3. Smithery.ai

**URL:** https://smithery.ai/
**Охват:** Популярный хаб MCP-серверов

```bash
smithery auth login
smithery mcp publish "https://ваш-домен/mcp" -n yourorg/promptvault
```

Для self-hosted HTTP-серверов не нужна smithery.yaml.

---

## 4. Cline MCP Marketplace

**URL:** https://github.com/cline/mcp-marketplace
**Охват:** Миллионы пользователей VS Code / Cline

Подача: GitHub Issue с URL репозитория + логотип 400x400 PNG.

---

## 5. PulseMCP

**URL:** https://www.pulsemcp.com/submit
**Охват:** 14,000+ серверов

Веб-форма. Автоматически подхватывает из Official MCP Registry.

---

## 6. Glama.ai

**URL:** https://glama.ai/mcp/servers
**Охват:** 20,000+ серверов с security-грейдингом

Автоиндексация GitHub. Кнопка "Add Server". Авторы могут claim серверы.

---

## 7. Прочие

| Площадка | URL | Подача |
|----------|-----|--------|
| mcp.so | mcp.so | GitHub Issue |
| mcpservers.org | mcpservers.org/submit | Веб-форма |
| OpenTools | opentools.com/registry | Автоимпорт из Official Registry |
| MCP Market | mcpmarket.com | Каталог |

---

## Порядок действий

| Шаг | Действие | Усилия | Отдача |
|-----|----------|--------|--------|
| 1 | Official MCP Registry (mcp-publisher) | 30 мин | Все MCP-клиенты |
| 2 | Anthropic Connectors (Google-форма) | 15 мин | Десятки млн Claude |
| 3 | smithery mcp publish | 5 мин | Smithery-аудитория |
| 4 | GitHub Issue в Cline Marketplace | 10 мин | VS Code / Cline |
| 5 | PulseMCP + Glama + mcp.so | 15 мин | SEO + discoverability |

---

## Заметки

- **stdio-прокси НЕ нужен** — Streamable HTTP поддерживается всеми основными клиентами
- Если кому-то нужен stdio-мост, есть `github.com/sparfenyuk/mcp-proxy`
- После публикации в Official Registry, OpenTools и PulseMCP подхватывают автоматически
- Tool Annotations (ReadOnlyHint, DestructiveHint) — обязательны для Anthropic Directory, уже реализованы
