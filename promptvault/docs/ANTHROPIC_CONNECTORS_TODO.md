# Anthropic Connectors Directory — TODO

## Статус: Отложено (нужен OAuth 2.0)

## Что нужно реализовать

### 1. OAuth 2.0 Authorization Code Flow
- Добавить OAuth2-сервер в backend (authorize, token, refresh endpoints)
- Callback URLs которые нужно поддержать:
  - `http://localhost:6274/oauth/callback` (Claude Code)
  - `http://localhost:6274/oauth/callback/debug`
  - `https://claude.ai/api/mcp/auth_callback` (Claude.ai)
  - `https://claude.com/api/mcp/auth_callback`
- PKCE (S256) обязателен
- Scopes: read, write, admin (опционально)

### 2. Allowlist Claude IP-адресов
- Список: https://docs.claude.com/en/api/ip-addresses
- Настроить на уровне nginx или firewall на VPS
- Обязательно для claude.ai и Claude Desktop

### 3. Подготовить 3 use-case примера

**Use Case 1 — Быстрая вставка промпта:**
```
Пользователь: "Найди мой промпт для code review"
Claude вызывает: search_prompts(query="code review")
Claude возвращает: название, содержимое с {{переменными}}, теги
Пользователь: "Обнови его — добавь проверку безопасности"
Claude вызывает: update_prompt(id=42, content="...обновлённый...")
```

**Use Case 2 — Организация библиотеки:**
```
Пользователь: "Создай коллекцию DevOps и добавь теги docker, kubernetes"
Claude вызывает: create_collection(name="DevOps", color="#06b6d4")
Claude вызывает: create_tag(name="docker"), create_tag(name="kubernetes")
Пользователь: "Покажи все промпты с тегом docker"
Claude вызывает: list_prompts(tag_ids=[5])
```

**Use Case 3 — Командная работа:**
```
Пользователь: "Покажи последние промпты команды и закрепи самый используемый"
Claude вызывает: list_prompts(team_id=3, sort="usage_count")
Claude вызывает: prompt_pin(id=15, team_wide=true)
Пользователь: "Создай ссылку чтобы поделиться этим промптом"
Claude вызывает: share_create(prompt_id=15)
```

### 4. Форма заявки
- URL: https://forms.gle/tyiAZvch1kDADKoP9
- Тест-аккаунт с данными
- Документация: docs/MCP.md
- Privacy Policy: https://promtlabs.ru/legal/extension-privacy
- Поддержка: slava0gpt@gmail.com

### 5. Чеклист перед подачей
- [ ] OAuth 2.0 реализован и протестирован
- [ ] Claude IP allowlisted
- [ ] 3 use-case работают через claude.ai
- [ ] Документация обновлена
- [ ] Тест-аккаунт создан с demo-данными
- [ ] Server не помечен как beta
