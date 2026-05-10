# Chains UI Redesign — Empty state с шаблонами + filled state mini-graph

**Дата:** 2026-05-10
**Статус:** Design approved, готов к implementation plan
**Связан с:** Phase 16 chains feature, REVIEW_2026-05-07 (post-launch UX polish)

---

## Контекст

После создания тестового пользователя обнаружено, что страница `/chains` в empty-state выглядит как огромный пустой rectangle с маленькой иконкой `Link2` и общим текстом «У вас пока нет ни одной цепочки». Пользователь не понимает, что такое цепочка и какие use-case'ы за ней стоят. В filled-state карточки несут 5 кнопок (Запустить/Дерево/История/Редактор/Удалить) — visual overload, нет preview структуры.

Данный документ фиксирует вариант **A + C** из брейншторма: образовательный empty state с готовыми шаблонами + улучшенные карточки в filled state.

## Цель

1. **Уменьшить time-to-first-chain** — новый пользователь должен понять «что это и зачем» за 5 секунд и создать первую цепочку одним кликом.
2. **Обогатить карточку цепочки** preview-схемой структуры и метаданными — пользователь видит «3 шага, 1 ветвление, 12 запусков» без перехода в редактор.
3. **Снизить визуальный шум** — 5 кнопок в карточке → 2 главные + menu для редких action'ов.
4. **Не затронуть** реальный функционал run/edit/canvas/runs страниц — это исключительно index page polish.

## Решение

### Часть 1 — Empty state с галереей шаблонов (Вариант A)

#### Layout

```
┌──────────────────────────────────────────────────────────┐
│  Цепочки промптов                                        │
│  Связывайте промпты в pipeline — output одного шага      │
│  становится переменной для следующего.                   │
├──────────────────────────────────────────────────────────┤
│  Начните с шаблона                                       │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ 💡 Идея  │  │ 🔍 Code  │  │ 📝 Текст │  │ 🐛 Баг   │ │
│  │ → PRD    │  │ review → │  │ research │  │ → repro  │ │
│  │ → Tests  │  │ tests →  │  │ → outline│  │ → fix →  │ │
│  │          │  │ docs     │  │ → draft  │  │ docs     │ │
│  │ [▢]→[▢]  │  │ [▢]→[▢]  │  │ [▢]→[▢]  │  │ [▢]→[▢]  │ │
│  │  →[▢]    │  │  →[▢]    │  │  →[▢]    │  │  →[▢]→[▢]│ │
│  │ 3 шага   │  │ 3 шага   │  │ 3 шага   │  │ 4 шага   │ │
│  │ Использ→ │  │ Использ→ │  │ Использ→ │  │ Использ→ │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
│                                                          │
│  ─── или ───                                             │
│                                                          │
│  [+ Создать с нуля]                                      │
└──────────────────────────────────────────────────────────┘
```

#### 4 шаблона (hardcoded в `frontend/src/lib/chain-templates.ts`)

| Иконка | Название | Шаги | Use case |
|--------|----------|------|----------|
| 💡 | Идея → PRD → Tests | Идея, PRD на основе идеи, Tests на основе PRD | Продуктовая разработка фич |
| 🔍 | Code review → Tests → Docs | Анализ diff'а, генерация unit-тестов, документация | После написания кода |
| 📝 | Research → Outline → Draft | Поиск источников, создание плана, черновик | Контентные статьи, посты |
| 🐛 | Bug → Repro → Fix → Docs | Описание бага, шаги воспроизведения, фикс, обновление документации | Bug-fix workflow для разработчиков |

Каждый шаблон — это TS-объект:

```ts
interface ChainTemplate {
  id: string                  // 'idea-prd-tests'
  emoji: string               // '💡'
  title: string               // 'Идея → PRD → Tests'
  description: string         // 1-line use case
  steps: {
    name: string              // 'Идея'
    promptContent: string     // raw промпт с {{variables}}
    variableMapping?: Record<string, string>  // mapping из step_outputs предыдущих шагов
  }[]
}
```

Клик по шаблону → frontend создаёт цепочку через **2 API-вызова**:

1. `POST /api/prompts` ×N — создать N промптов с `promptContent` шаблона (по одному на шаг)
2. `POST /api/chains` — создать цепочку с `steps[]`, ссылающимися на свежие prompt_id + variable_mapping

Затем редирект на `/chains/{id}/edit` — пользователь видит готовую структуру и может изменить названия/содержимое.

**Важно по workspace:** шаблоны создают промпты и цепочку в **активном workspace** из `useWorkspaceStore`. Если выбрано «Личное пространство» (`team` is null) — промпты приватные. Если выбрана команда — промпты создаются с `team_id` команды; в этом случае требуется роль `owner`/`editor` в команде (`canWrite=true` в `useCurrentTeamRole`). Если активна команда и пользователь viewer — галерея шаблонов скрывается, остаётся только сообщение «У команды пока нет цепочек» + блок «попросите owner'а/editor'а создать».

Промпты создаются как обычные private/team prompts (не public, не shared), которые пользователь может потом редактировать или удалять как любые свои.

### Часть 2 — Filled state с mini-graph и actions menu (Вариант C)

#### Layout карточки

```
┌──────────────────────────────────────────┐
│  📝 Code review → Tests → Docs       ⋯  │
│  Объясняет diff, генерит unit-tests…    │
│                                          │
│  [▢]──→[◇]──→[▢]──→[▢]                  │
│   1    2     3     4                    │
│                                          │
│  3 шага · 1 ветвление · 12 запусков     │
│                                          │
│  [▶ Запустить]              [✏ Редактор]│
└──────────────────────────────────────────┘
```

#### Изменения относительно текущего

- **5 кнопок → 2 главные + ⋯ menu**
  - Главные: `[▶ Запустить]`, `[✏ Редактор]` (для viewer'а — `Просмотр`)
  - В menu: `Дерево` (canvas), `История` (запуски), `Удалить` (только canWrite)
- **Mini-graph SVG** — рендерится из `steps_preview` (см. backend extension):
  - Prompt-шаг = квадрат 8×8 с border
  - Fork-шаг = ромб 8×8
  - Стрелки между шагами (`→`)
  - Если шагов > 5 — рисуем 4 + бейдж `+N more` справа
  - Высота ~40px, full-width карточки
- **Badges metadata** — одна строка с разделителями `·`:
  - `{step_count} шага` (правильное склонение: 1 шаг / 2-4 шага / 5+ шагов через `pluralize.ts`)
  - `{N} ветвление` (только если `has_branching`, склонение: 1 ветвление / 2-4 ветвления / 5+ ветвлений) — N = количество fork-шагов в `steps_preview`
  - `{saved_runs_count} запусков` (склонение: 1 запуск / 2-4 запуска / 5+ запусков)

### Часть 3 — Backend DTO extension

Расширяем response `GET /api/chains` дополнительными полями. Один SELECT без N+1, через correlated subqueries или LATERAL.

#### TypeScript interface

```ts
export interface Chain {
  id: number
  user_id: number
  team_id?: number
  name: string
  description: string
  created_at: string
  updated_at: string
  // NEW (Phase 16 UI polish):
  step_count: number               // total steps в цепочке
  has_branching: boolean           // EXISTS step с step_type='fork'
  saved_runs_count: number         // COUNT executions со status='completed'
  steps_preview: ChainStepPreview[] // первые 5 шагов, только types и position для mini-graph
}

export interface ChainStepPreview {
  position: number
  step_type: 'prompt' | 'fork'
}
```

#### Go DTO + repository

Изменение в `internal/delivery/http/chain/response.go`:

```go
type ChainListItem struct {
    ID              uint                `json:"id"`
    UserID          uint                `json:"user_id"`
    TeamID          *uint               `json:"team_id,omitempty"`
    Name            string              `json:"name"`
    Description     string              `json:"description"`
    CreatedAt       time.Time           `json:"created_at"`
    UpdatedAt       time.Time           `json:"updated_at"`
    // NEW
    StepCount       int                 `json:"step_count"`
    HasBranching    bool                `json:"has_branching"`
    SavedRunsCount  int                 `json:"saved_runs_count"`
    StepsPreview    []ChainStepPreview  `json:"steps_preview"`
}

type ChainStepPreview struct {
    Position int    `json:"position"`
    StepType string `json:"step_type"`
}
```

Изменение в `internal/infrastructure/postgres/repository/prompt_chain_repo.go`:

Метод `ListByUserID` (и `ListByTeamID`) дополняется через LATERAL для каждой цепочки:

```sql
SELECT c.*,
       cs.step_count,
       cs.has_branching,
       cr.runs_count,
       sp.steps_preview_json
FROM prompt_chains c
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS step_count,
           BOOL_OR(step_type = 'fork') AS has_branching
    FROM prompt_chain_steps
    WHERE chain_id = c.id
) cs ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS runs_count
    FROM prompt_chain_executions
    WHERE chain_id = c.id AND status = 'completed'
) cr ON true
LEFT JOIN LATERAL (
    SELECT json_agg(json_build_object('position', position, 'step_type', step_type) ORDER BY position) AS steps_preview_json
    FROM (
        SELECT position, step_type
        FROM prompt_chain_steps
        WHERE chain_id = c.id
        ORDER BY position
        LIMIT 5
    ) s
) sp ON true
WHERE c.user_id = $1 AND c.deleted_at IS NULL
ORDER BY c.updated_at DESC
LIMIT $2 OFFSET $3
```

LATERAL'ы уже применяются в `team_repo` (см. MN-38), pattern принят. Один SELECT — нет N+1.

### Часть 4 — Header cleanup

- **Empty state:** удалить верхнюю CTA `[+ Создать цепочку]` в header. Только subtitle с описанием. CTA `[+ Создать с нуля]` живёт в body под галереей шаблонов.
- **Filled state:** оставить header CTA `[+ Создать цепочку]` как сейчас.

---

## Out of scope

- **Backend chain_templates table** — для 4 шаблонов hardcoded на frontend достаточно. Если шаблонов будет 20+ — пересмотреть.
- **Drag-drop reorder в карточке** — это уже фича canvas/edit pages.
- **Chain duplicate / import-export** — отдельная фича, не часть этого редизайна.
- **Localization шаблонов** — пока только русский (всё приложение — РФ-only). Если в будущем будет EN — шаблоны выносятся в i18n.
- **A/B test шаблонов** — нет инфраструктуры для эксперимента; ship как есть.

## Файлы изменений

**Frontend:**
- `frontend/src/lib/chain-templates.ts` (NEW) — 4 ChainTemplate объекта + helper `applyTemplate(template, teamId)`
- `frontend/src/components/chains/chain-template-card.tsx` (NEW) — карточка одного шаблона в empty state
- `frontend/src/components/chains/chain-mini-graph.tsx` (NEW) — SVG-схема steps_preview
- `frontend/src/components/chains/chain-card.tsx` (NEW) — карточка цепочки в filled state (extracted из index.tsx)
- `frontend/src/pages/chains/index.tsx` — переписан empty/filled state'ы
- `frontend/src/api/types.ts` — расширить `Chain` interface (step_count, has_branching, saved_runs_count, steps_preview)
- `frontend/src/lib/pluralize.ts` — добавить case'ы для «шаг», «ветвление», «запуск» (если их там нет)

**Backend:**
- `backend/internal/delivery/http/chain/response.go` — `ChainListItem` + `ChainStepPreview`
- `backend/internal/infrastructure/postgres/repository/prompt_chain_repo.go` — `ListByUserID` + `ListByTeamID` через LATERAL
- `backend/internal/usecases/chain/types.go` — domain type для preview, если нужен
- `backend/internal/usecases/chain/chain_test.go` — обновить mock'и

**Тесты:**
- `frontend/src/components/chains/chain-template-card.test.tsx` — smoke render
- `frontend/src/components/chains/chain-mini-graph.test.tsx` — рендер 1/3/5/8 шагов (последний — с `+N more`)
- `frontend/src/lib/chain-templates.test.ts` — 4 шаблона валидны (имеют ≥2 шагов, prompt content не пустой)
- `backend/internal/infrastructure/postgres/repository/prompt_chain_repo_test.go` — расширить с проверкой LATERAL результата (step_count/has_branching/saved_runs_count/steps_preview)

## Верификация

После реализации — manual QA на dev stack (`docker compose -f docker-compose.dev.yml up`):

1. **Empty state:** свежий аккаунт → `/chains` → видит 4 шаблона + CTA «Создать с нуля». Layout responsive (mobile = stack column, tablet = 2×2 grid, desktop = 4×1 row).
2. **Создать через шаблон (личное):** клик по «💡 Идея → PRD → Tests» при активном «Личном пространстве» → создаются 3 промпта user_id=current + цепочка user_id=current, team_id=null → редирект в `/chains/{id}/edit` с готовой структурой.
3. **Создать через шаблон (команда, owner/editor):** переключиться на команду в workspace-switcher → `/chains` → клик по шаблону → промпты и цепочка создаются с `team_id`.
4. **Команда + viewer:** залогиниться viewer'ом в команду → `/chains` → галерея шаблонов скрыта, видно empty-state «команда пока без цепочек» с подсказкой к owner/editor'у.
5. **Filled state:** после создания нескольких цепочек → `/chains` показывает карточки с mini-graph (включая fork-вариант) + badges + 2 главные кнопки + menu.
6. **Pluralization:** проверить корректные склонения для 1/2/5/21 (`шаг`/`шага`/`шагов`).
7. **Tests:** `npm run test`, `go test -short ./...`, `npm run lint`, `npm run build`, `golangci-lint run` — все green.

Manual QA endpoint'ы:
- `curl http://localhost:8080/api/chains -H "Authorization: Bearer $TOKEN"` — должен вернуть chain list со всеми новыми полями (step_count, has_branching, saved_runs_count, steps_preview).
