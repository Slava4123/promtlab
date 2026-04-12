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
├── backend/       # Go API (Chi + GORM + PostgreSQL), Clean Architecture
├── frontend/      # React 19 SPA (Vite + shadcn/ui + TanStack Query + Zustand)
├── docs/          # PLAN.md, TODO.md, FEATURES.md, DEPLOY.md
└── CLAUDE.md      # ← полная документация проекта
```
