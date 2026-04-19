Adds **PromptVault (ПромтЛаб)** — a self-hosted AI prompt library — to the 🧠 Knowledge & Memory category.

### About the server
- Remote MCP server via **Streamable HTTP** (`https://promtlabs.ru/mcp`).
- Go backend, React frontend, PostgreSQL storage.
- Already published in the [Official MCP Registry](https://registry.modelcontextprotocol.io/v0/servers?search=promtlab) as `ru.promtlabs/promptvault`.
- 30 tools covering prompts CRUD, versions, collections, tags, teams, sharing and trash.
- Scoped API keys (read-only flag + `allowed_tools` whitelist + per-team).
- Quota-aware: per-plan daily limits for write/destructive tools.

### Compliance with CONTRIBUTING.md
- Alphabetical placement inside `🧠 Knowledge & Memory`.
- Emojis: 🏎️ (Go codebase), ☁️ (cloud / remote hosted).
- Linked to the upstream GitHub repo and working remote URL.
- Verified on [Glama](https://glama.ai/mcp/servers?query=promtlab).

Generated automatically on each stable release via `.github/workflows/mcp-publish.yml`.
