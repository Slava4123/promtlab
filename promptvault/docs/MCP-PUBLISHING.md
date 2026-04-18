# Публикация MCP-сервера PromptLab в реестрах

## Статус

- ✅ Official MCP Registry — v1.2.0 опубликована 2026-04-18, автопубликация новых версий по git-тегу настроена (см. §8).
- 🔒 Cline Marketplace — **отложено до перевода репо в public**. Cline требует открытый GitHub-репозиторий для ревью кода; пока `Slava4123/promtlab` приватный, подача не имеет смысла. Артефакты (логотип, `llms-install.md`, draft текста) оставлены в репо — если репо будет сделан публичным, подача делается одним шагом.
- ⏳ Остальные (Smithery, Anthropic Connectors, PulseMCP, Glama) — не подано.

---

## 0. DNS setup для автопубликации (one-time)

Чтобы GitHub Actions мог публиковать в namespace `ru.promtlabs/*`, нужна DNS-верификация домена:

```bash
# 1. Сгенерировать Ed25519 keypair (локально)
openssl genpkey -algorithm Ed25519 -out mcp-dns-key.pem

# 2. Public key → base64 для DNS TXT
openssl pkey -in mcp-dns-key.pem -pubout -outform DER | tail -c 32 | base64

# 3. Private key → 64-символьная hex-строка для GitHub Secret
openssl pkey -in mcp-dns-key.pem -noout -text | grep -A3 "priv:" | tail -n +2 | tr -d ' :\n'
```

Дальше:
- Добавить в DNS зону `promtlabs.ru` TXT-запись:
  ```
  promtlabs.ru. IN TXT "v=MCPv1; k=ed25519; p=<BASE64_PUBLIC_KEY>"
  ```
  Проверка: `dig TXT promtlabs.ru +short` должен содержать строку с `v=MCPv1`.
- В GitHub → Settings → Secrets → Actions создать secret `MCP_DNS_PRIVATE_KEY` с 64-hex private key.
- Удалить локальный `mcp-dns-key.pem` — private key хранится только в GitHub Secret.

---

## 8. Автопубликация через GitHub Actions

Workflow: `.github/workflows/mcp-publish.yml`.

### Триггеры
- `push` тега `vX.Y.Z` — публикация + создание GitHub Release.
- `workflow_dispatch` — ручной запуск из Actions UI (на случай failed run'а).

### Procedure релиза

```bash
# 1. Bump версии в server.json
cd promptvault
jq '.version = "1.3.0"' server.json > server.json.tmp && mv server.json.tmp server.json

# 2. Commit + push в main
cd ..
git add promptvault/server.json
git commit -m "chore(mcp): bump v1.2.0 → v1.3.0"
git push

# 3. Тег + push тега
git tag v1.3.0
git push --tags
```

Дальше всё делает workflow:
- Проверяет что tag_version == server.json.version.
- Устанавливает `mcp-publisher`.
- Логинится через DNS (secret `MCP_DNS_PRIVATE_KEY`).
- `mcp-publisher publish`.
- Проверяет что новая версия реально появилась в реестре.
- Создаёт GitHub Release с автогенерируемыми release notes.

### PR-guard `mcp-validate.yml`

Отдельный workflow `.github/workflows/mcp-validate.yml` запускается на PR, если `promptvault/server.json` изменён:
- `mcp-publisher validate` — schema-check.
- Сравнение `server.json.version` с текущей версией в реестре — не разрешает merge, если версия не bump'нута.

---

## 1. Official MCP Registry (главный приоритет)

**URL:** https://registry.modelcontextprotocol.io/
**Охват:** Все MCP-клиенты (Claude, Cursor, Windsurf, Cline и др.)

Старый репозиторий `modelcontextprotocol/servers` больше НЕ принимает PR. Всё через CLI `mcp-publisher`.

### Процесс (ручной, для первой публикации)

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

Для нас используется DNS-верификация (см. §0), namespace `ru.promtlabs/*`. Первичная публикация v1.0.0 была сделана вручную; обновления идут через workflow (см. §8).

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

Подача: GitHub Issue с URL репозитория + логотип 400×400 PNG.

### Готовые артефакты (в репо)

- **Логотип:** `promptvault/frontend/public/logo-mcp-400.png` (400×400 PNG, ~16 KB), публично доступен как `https://promtlabs.ru/logo-mcp-400.png`.
- **Install guide для Cline agent'а:** `promptvault/llms-install.md` — объясняет LLM'у, что сервер remote и клонировать не нужно, как только прописать URL + Bearer token.
- **Draft формы подачи:** `promptvault/docs/cline-submission-draft.md` — готовые значения полей под YAML-форму Cline.

### Процедура подачи (ручная)

1. Открыть https://github.com/cline/mcp-marketplace/issues/new?template=mcp-server-submission.yml
2. Заполнить поля значениями из `promptvault/docs/cline-submission-draft.md`:
   - GitHub Repository URL
   - Logo Image (URL или drag-and-drop `logo-mcp-400.png`)
   - Installation Testing (оба чекбокса)
   - Additional Information (длинный текст с описанием возможностей)
3. Отправить issue. Ревью мейнтейнеров обычно идёт несколько дней.

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
