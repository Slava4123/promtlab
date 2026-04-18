# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Основной проект — `promptvault/`. Подробная документация по архитектуре, стеку и правилам разработки: [promptvault/CLAUDE.md](promptvault/CLAUDE.md).

## Быстрый старт

```bash
# Backend (из promptvault/backend/)
go run ./cmd/server                                        # dev-сервер
go test -short ./...                                       # unit-тесты
go test ./...                                              # + integration (нужен Docker — testcontainers)
golangci-lint run                                          # lint

# Frontend (из promptvault/frontend/)
npm run dev                                                # Vite dev server
npx vitest run                                             # тесты
npm run lint                                               # ESLint

# Docker (из promptvault/)
docker compose -f docker-compose.dev.yml up                # full dev stack
```

## Структура репозитория

```
promptvault/
├── backend/              # Go API (Chi + GORM + PostgreSQL), Clean Architecture
├── frontend/             # React 19 SPA (Vite + shadcn/ui + TanStack Query + Zustand)
├── docs/                 # PLAN.md, FEATURES.md, DEPLOY.md, MCP.md, … (+ archive/)
├── server.json           # MCP server manifest для Official MCP Registry
├── llms-install.md       # install guide для MCP install agents (Cline и др.)
└── CLAUDE.md             # ← полная документация проекта
```

## MCP server

Встроенный в backend MCP-сервер v1.2.0 опубликован в [Official MCP Registry](https://registry.modelcontextprotocol.io/v0/servers?search=promtlab) как `ru.promtlabs/promptvault`.

**Релиз новой версии:**
```bash
# 1. bump в promptvault/server.json (и опционально в internal/mcpserver/server.go)
# 2. commit + push в main
# 3. теги:
git tag v1.3.0 && git push origin v1.3.0
# → workflow .github/workflows/mcp-publish.yml публикует в Registry (DNS-auth)
# → создаётся GitHub Release
```

Подробнее в [promptvault/docs/MCP-PUBLISHING.md](promptvault/docs/MCP-PUBLISHING.md).
