# Anthropic Connectors Directory — submission blueprint

Готовые значения полей для формы подачи: https://clau.de/mcp-directory-submission

Справочник по требованиям: https://claude.com/docs/connectors/building/submission

---

## 1. Server basics

| Поле | Значение |
|------|----------|
| **Name** | ПромтЛаб (PromptVault) |
| **Server URL (MCP endpoint)** | `https://promtlabs.ru/mcp` |
| **Tagline** | Self-hosted AI prompt library with team collaboration and versioning |
| **Description** | ПромтЛаб — self-hosted менеджер AI-промптов: организация (коллекции/теги/версии/корзина), команды с ролями (owner/editor/viewer), публичный шаринг и MCP-интеграция. 30 tools, quota-aware, scoped API keys. На русском, self-hosted в РФ. |
| **Homepage** | https://promtlabs.ru |
| **Category (если спросят)** | Productivity / Knowledge Management |

## 2. Connection details

| Поле | Значение |
|------|----------|
| **Transport** | Streamable HTTP |
| **Authentication** | OAuth 2.1 Authorization Code + PKCE S256 (RFC 7591 Dynamic Client Registration). Дополнительно — static API keys (pvlt_*) для self-hosted клиентов. |
| **OAuth discovery** | `https://promtlabs.ru/.well-known/oauth-protected-resource` (RFC 9728) + `https://promtlabs.ru/.well-known/oauth-authorization-server` (RFC 8414) |
| **Scopes** | `mcp:read`, `mcp:write` |
| **PKCE** | S256 only (plain отклоняется) |
| **Resource indicator** | `https://promtlabs.ru/mcp` (RFC 8707 audience) |

## 3. Tools (30)

Все имеют `title`, `readOnlyHint`/`destructiveHint`/`idempotentHint` annotations. Полная таблица — в https://github.com/Slava4123/promtlab/blob/main/promptvault/docs/MCP.md.

Разбиение:
- **Read (14):** search_prompts, list_prompts, get_prompt, list_collections, collection_get, list_tags, list_teams, list_trash, get_prompt_versions, prompt_list_pinned, prompt_list_recent, search_suggest, whoami, (+4 resources)
- **Write (11):** create_prompt, update_prompt, create_tag, create_collection, collection_update, share_create, prompt_revert, restore_prompt, prompt_favorite (idempotent), prompt_pin (idempotent), prompt_increment_usage (idempotent)
- **Delete (5):** delete_prompt (soft), delete_collection, purge_prompt (hard), share_deactivate, tag_delete

Все write/delete кроме idempotent UX-toggle'ов едят дневную MCP-квоту.

## 4. Data & compliance

- **Data residency:** Россия (VPS Timeweb, DB managed PostgreSQL в РФ).
- **Data handling:** пользовательские промпты/коллекции/теги хранятся только в нашей БД. MCP-клиенты получают только то, что принадлежит авторизованному user_id + team_ids из scope.
- **Third-party:** OpenRouter (только для AI-улучшения промптов по явному действию пользователя, текст промпта отправляется в API Anthropic через OpenRouter). Glitchtip (self-hosted Sentry-clone, PII scrubbing в SDK).
- **Payment:** T-Bank (Тинькофф) рекуррент. Webhook с HMAC-подписью и IP-allowlist.
- **Encryption:** HTTPS везде (Let's Encrypt), JWT HS256 для сессий, bcrypt для паролей.
- **Retention:** soft-delete 30 дней, потом cron purge. Аккаунт + все данные удаляются по запросу на slava0gpt@gmail.com.

## 5. Support information

| Поле | Значение |
|------|----------|
| **Documentation** | https://github.com/Slava4123/promtlab/blob/main/promptvault/docs/MCP.md |
| **Privacy Policy** | https://promtlabs.ru/legal/extension-privacy |
| **Support channel** | slava0gpt@gmail.com |
| **Security disclosure** | https://github.com/Slava4123/promtlab/blob/main/SECURITY.md |
| **Source code** | https://github.com/Slava4123/promtlab (FSL 1.1) |
| **Registry listing** | https://registry.modelcontextprotocol.io/v0/servers?search=promtlab |

## 6. Branding

- **Logo (400×400 PNG):** https://promtlabs.ru/logo-mcp-400.png (размер 16KB, validated 2026-04-20)
- **Favicon:** https://promtlabs.ru/favicon.svg (работает)
- **MCP Apps screenshots (min 1000px, PNG):** _TODO — если попросят в форме, сделать 3-5 скринов_ `/dashboard`, `/prompts/:id` (редактор), `/settings/integrations` через Chrome DevTools mobile:off 1440×900.

## 7. Demo / test account

Reviewer'ам нужен тест-аккаунт с наполненными данными.

### Чеклист создания (ручной шаг перед подачей)

- [ ] Зарегистрировать на promtlabs.ru: `connector-review@promtlabs.ru` (или временный mail.ru, email от Anthropic для reviewers лучше узнать отдельно).
- [ ] Включить Pro-план через админку (`/admin/users/<id>/subscription`) на 30 дней без оплаты.
- [ ] Заполнить промптами (минимум 10-15):
  - 2-3 code review промпта с `{{переменными}}`
  - 2-3 writing assistant
  - 1-2 SQL / data analysis
  - 1 marketing email template
- [ ] Создать 2-3 коллекции: «Code Review», «Writing», «DevOps».
- [ ] Создать 3-4 тега: `python`, `sql`, `docs`, `email`.
- [ ] Создать тестовую команду с 2 участниками (сам + dummy).
- [ ] Поделиться 1-2 промптами через public link.
- [ ] **Credentials в форму:**
  - Email: _TODO_
  - Password: _TODO_
  - Step-by-step: «Login → Settings → Integrations → Create API key 'Claude Reviewer' → copy pvlt_…»

## 8. Три use-case для reviewer демонстрации

### UC-1: Быстрая вставка промпта
```
User: Найди мой промпт для code review, добавь параметр language=Python
Claude:
  1. search_prompts(query="code review") → получает список
  2. get_prompt(id=42) → возвращает содержимое с {{language}}
  3. prompt_increment_usage(id=42)
  Claude вставляет промпт с подставленным language=Python.
```

### UC-2: Организация библиотеки
```
User: Создай коллекцию «DevOps» с тегами docker и kubernetes,
      перенеси туда 3 последних промпта.
Claude:
  1. create_collection(name="DevOps", color="#06b6d4")
  2. create_tag(name="docker"), create_tag(name="kubernetes")
  3. prompt_list_recent(limit=3)
  4. update_prompt(id=X, collection_id=..., tag_ids=[...])
```

### UC-3: Командная работа
```
User: Покажи последние промпты команды Marketing,
      закрепи самый используемый для всех.
Claude:
  1. list_teams() → находит Marketing team
  2. list_prompts(team_id=3, sort="usage_count")
  3. prompt_pin(id=15, team_wide=true)
  4. share_create(prompt_id=15) для external доступа.
```

## 9. Policy confirmations

- ✅ Tools имеют title + readOnlyHint/destructiveHint annotations (все 30 проверены).
- ✅ OAuth 2.1 PKCE S256 реализован.
- ✅ RFC 9728 Protected Resource Metadata.
- ✅ RFC 8414 Authorization Server Metadata.
- ✅ RFC 8707 Resource Indicators.
- ✅ RFC 7591 Dynamic Client Registration.
- ✅ RFC 7009 Token Revocation.
- ✅ HTTPS-only (Let's Encrypt).
- ✅ Privacy policy присутствует.
- ✅ Сервер НЕ в beta (v1.3.4+ в production).
- ✅ Responsible disclosure через SECURITY.md.

## 10. Claude outbound IP

Per https://platform.claude.com/docs/en/api/ip-addresses (актуально на 2026-04-20):

| Тип | CIDR |
|-----|------|
| Inbound (Claude → API) | `160.79.104.0/23` (IPv4) + `2607:6bc0::/48` (IPv6) |
| **Outbound (Claude → наш MCP)** | **`160.79.104.0/21` (IPv4)** |
| Phased out (не использовать) | 34.162.46.92/32, 34.162.102.82/32, 34.162.136.91/32, 34.162.142.92/32, 34.162.183.95/32 |

**Allowlist на nginx не обязателен** для approval, но хорошая практика. Если добавлять — не блокировать, а использовать как audit-signal (логировать `Claude-Source: true` для запросов из `160.79.104.0/21`).

## 11. Submission workflow

1. Прогнать чек-лист §7 (test account).
2. Прогнать smoke-тест OAuth flow через MCP Inspector:
   ```bash
   npx @modelcontextprotocol/inspector
   # URL: https://promtlabs.ru/mcp
   # Inspector должен получить 401 → прочитать WWW-Authenticate
   # → дойти до oauth-authorization-server → запросить authorize → получить token → tools/list
   ```
3. Добавить https://claude.ai как custom connector и прогнать 3 use-case.
4. Открыть форму: https://clau.de/mcp-directory-submission.
5. Заполнить полями из §1-§9.
6. Submit. Review времена варьируются, нет expedited track.

## 12. Post-approval: снятие `continue-on-error` с jobs

После approval:
- Перевести Anthropic Connectors в `ready` в `frontend/src/pages/help/mcp.tsx`.
- Добавить ссылку на наш connector в Claude.ai directory.

---

**Статус:** backend полностью готов. Блокер — ручные шаги §7 (test account) + §11.3 (тест через Claude.ai UI). После — submit.
