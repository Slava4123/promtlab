# Фичи: Prompt Chains + Conditional Chains

**Создано:** 2026-04-28
**Статус:** реализованы (dark launch за `CHAINS_ENABLED=false`); pending — UI-полировка и MCP multi-step эксперимент.
**Связано с:** `BUSINESS_RESEARCH.md` (R8 limit rebalance), будущая Phase 0 launch

> **История версий (читай сверху вниз; нижестоящие отменяют верхние):**
>
> - **v1** (это базовый текст ниже) — Conditional Chains как DSL-эвалюатор: matchers (`contains`/`regex`/`equals`/…), AND/OR/NOT, MaxConditionDepth=10, ReDoS-защита. Миграция `000054`.
> - **v2** (миграция `000055_chain_fork`, 2026-04-28) — DSL удалён, термин `conditional` переименован в `fork`. Юзер вручную выбирает ветку в run-mode по её Label через `chosen_branch_index`. Добавлен Canvas (xyflow + Dagre) как визуализатор графа.
> - **v3** (миграция `000056_chain_explicit_next`, 2026-04-29) — переходы между шагами стали явными: у prompt-шагов колонка `next_step_id`, у fork-шагов остались `branches[].next_step_id`. `position` остаётся только UI-сортировкой. Editor (`pages/chains/editor.tsx`) переписан как inline-tree: рекурсивный `StepNode` рендерит линейные участки и развилки на любой глубине, кнопки «+ Шаг» / «+ Развилка» в каждом листе/пустой ветке. AddStep API расширен полями `after_step_id`, `parent_fork_id`, `branch_index`. Tier-gate Max сохранён (`ErrForkRequiresMax` 403; UI делает кнопку disabled с tooltip). Удаление шага «зашивает» граф — `Service.RemoveStep` re-link'ает prompt-предшественников и fork-ветки. Подробности — секция [v3: Inline-tree editor](#v3-inline-tree-editor) в конце.

---

## Контекст и обоснование выбора

### История решения
- В рамках brainstorm'а non-AI фич для усиления Pro/Max-tier'ов было предложено 30+ идей
- Большинство отвергнуты как «не цепляют» или «дополнения, не большая фича»
- **Зашли только 2 фичи:** Prompt Chains и Conditional Chains
- Решение по 3-й фиче **отложено на post-launch** (через 4-6 недель собрать реальный feedback от первых 50 юзеров и выбрать осознанно, а не угадать)

### Зачем эти 2 фичи

Текущий PromptVault — **менеджер отдельных промптов**. Юзер хранит библиотеку, использует по одному.

С Chains + Conditional — PromptVault становится **AI workflow platform**:
- Связывать несколько промптов в последовательность (output одного → input следующего)
- Условные ветвления (if output содержит «X» → запустить промпт A, иначе → B)

**Уникальность:** ни один из конкурентов в РФ-нише prompt-managers этого не делает. В мире — только Latitude (open-source, developer-focused). PromptVault может занять позицию **«Zapier для промптов»** в нише.

**Принцип «без AI»:** обе фичи **детерминистические** — никаких LLM-вызовов с нашей стороны. Сохраняется маржа ~92%.

---

## Часть 1: Prompt Chains

### Что это в одном предложении
Последовательность из нескольких промптов, где **output каждого шага** становится **переменной** для следующего.

### Конкретный пример

Цепочка **«PRD Generator»** (5 шагов):

```
Шаг 1: «Идея продукта»
  - Промпт: «Опиши идею: {{идея}}»
  - User вводит: «приложение для food-delivery в малых городах»
  - Output: расширенное описание идеи

Шаг 2: «User Stories»
  - Промпт: «На основе идеи: {{output_шаг_1}}, напиши 5 user stories»
  - {{output_шаг_1}} автоматически подставляется
  - Output: список user stories

Шаг 3: «MVP Scope»
  - Промпт: «На основе user stories: {{output_шаг_2}}, выдели MVP scope»
  
Шаг 4: «Test Cases»
  - Промпт: «На основе MVP: {{output_шаг_3}}, придумай 10 test cases»

Шаг 5: «Estimate»
  - Промпт: «На основе MVP {{output_шаг_3}} и тестов {{output_шаг_4}}, оцени effort»
```

Юзер запускает chain → проходит шаги → получает структурированный PRD-документ за один workflow.

### Боль которую решает

Сейчас юзер делает это **руками**:
1. Открыть промпт «User Stories» → ввести описание идеи → запустить → скопировать результат
2. Открыть промпт «MVP Scope» → вставить user stories → запустить → скопировать
3. Повторить ещё 3 раза
4. Скомпилировать всё в один документ

С Chains: **один запуск → пройти шаги → получить готовый pipeline output**.

### User stories

```
Как разработчик
Я хочу запустить «Code Review pipeline»
  (анализ → найденные баги → план фикса → коммит-message)
Чтобы пройти через все шаги одной комбинацией клавиш
  без копирования между промптами

Как контент-маркетолог
Я хочу запустить «Контент-pipeline для VK»
  (идея → пост → комментарии → ответы на возражения)
Чтобы за один раз создать готовый набор контента
  для публикации
```

### UI / UX (текстовое описание)

**В разделе «Цепочки» (новая вкладка в sidebar):**
- Список цепочек со счётчиком запусков, last-run timestamp, тегами
- Кнопка «Создать новую цепочку»

**Editor цепочки:**
- Список шагов в порядке выполнения (drag-drop reorder)
- Каждый шаг — карточка:
  - Название шага
  - Selector «выбрать промпт из библиотеки» (dropdown с поиском)
  - Маппинг переменных промпта на: ручной ввод / output предыдущего шага / переменную цепочки
  - Опционально: «после этого шага сделать паузу для review» (manual checkpoint)
- Кнопка «Добавить шаг»
- Кнопка «Запустить цепочку» (run mode)
- Кнопка «Сохранить как шаблон»

**Run mode:**
- Шаг 1 показывается → юзер заполняет переменные → клик «Запустить» → копирует результат от Claude/GPT → вставляет в поле «Output» → клик «Далее»
- Шаг 2 автоматически подставляет output_шаг_1 → юзер заполняет оставшиеся переменные → выполняет
- ... до конца
- Финальный экран: все шаги + outputs collapsed; кнопка «Скопировать всё», «Сохранить запуск»

### Технические детали

#### Изменения в БД

```sql
-- Новая таблица: цепочка
CREATE TABLE prompt_chains (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  team_id BIGINT REFERENCES teams(id) ON DELETE CASCADE, -- nullable
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ -- soft delete
);
CREATE INDEX idx_prompt_chains_user_id ON prompt_chains (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_prompt_chains_team_id ON prompt_chains (team_id) WHERE deleted_at IS NULL;

-- Шаги цепочки
CREATE TABLE prompt_chain_steps (
  id BIGSERIAL PRIMARY KEY,
  chain_id BIGINT NOT NULL REFERENCES prompt_chains(id) ON DELETE CASCADE,
  position INTEGER NOT NULL, -- порядок выполнения, 1, 2, 3...
  prompt_id BIGINT NOT NULL REFERENCES prompts(id),
  name VARCHAR(255), -- опциональное название шага
  variable_mapping JSONB NOT NULL DEFAULT '{}', -- {var_name: {source: 'manual'|'step_output'|'chain_var', step_id?, var_name?}}
  manual_checkpoint BOOLEAN DEFAULT FALSE, -- остановиться для review после этого шага
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_prompt_chain_steps_chain ON prompt_chain_steps (chain_id, position);

-- Сохранённые запуски цепочки (для resume и аналитики, опционально в Phase 2)
CREATE TABLE prompt_chain_executions (
  id BIGSERIAL PRIMARY KEY,
  chain_id BIGINT NOT NULL REFERENCES prompt_chains(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id),
  current_step INTEGER DEFAULT 1,
  variables JSONB DEFAULT '{}', -- значения переменных цепочки
  step_outputs JSONB DEFAULT '{}', -- {step_id: output_text}
  status VARCHAR(20) DEFAULT 'in_progress', -- in_progress | completed | abandoned
  started_at TIMESTAMPTZ DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Backend (Go)

Новые места:
- `internal/models/prompt_chain.go` — модели Chain, ChainStep, ChainExecution
- `internal/interface/repository/prompt_chain.go` — интерфейс репозитория
- `internal/infrastructure/postgres/repository/prompt_chain_repo.go` — реализация
- `internal/usecases/chain/` — service: Create, AddStep, RemoveStep, ReorderSteps, StartExecution, AdvanceStep, CompleteExecution
- `internal/delivery/http/chain/` — handler:
  - `POST /api/chains` — create
  - `GET /api/chains` — list (с pagination)
  - `GET /api/chains/:id` — detail
  - `PUT /api/chains/:id` — update meta
  - `DELETE /api/chains/:id` — soft delete
  - `POST /api/chains/:id/steps` — add step
  - `PUT /api/chains/:id/steps/:step_id` — update step
  - `DELETE /api/chains/:id/steps/:step_id` — remove step
  - `POST /api/chains/:id/reorder` — reorder steps
  - `POST /api/chains/:id/executions` — start new execution
  - `GET /api/chains/:id/executions/:exec_id` — get execution state
  - `POST /api/executions/:exec_id/advance` — submit output for current step → advance to next

#### MCP integration (важно для Cursor / Claude Code)

Новые MCP tools:
- `list_chains(team_id?)` → массив цепочек
- `get_chain(id)` → структура цепочки + шаги
- `start_chain_execution(chain_id, initial_vars)` → execution_id + first step rendered prompt
- `advance_chain_step(execution_id, step_output)` → next step rendered prompt OR final result

**Важно:** MCP-клиент (Cursor / Claude Code) должен **сам вызывать LLM** — наш бэкенд только хранит, рендерит и навигирует. Клиент в цикле:
1. `start_chain_execution(...)` → получает render для шага 1
2. Отправляет в LLM (Claude API через свой ключ)
3. Получает output
4. `advance_chain_step(exec_id, output)` → render для шага 2
5. Повторяет до status='completed'

#### Frontend (React)

Новые страницы:
- `pages/chains/index.tsx` — список цепочек
- `pages/chains/new.tsx` — создание
- `pages/chains/:id/edit.tsx` — редактор (drag-drop steps)
- `pages/chains/:id/run.tsx` — run mode (пошаговый wizard)

Новые хуки:
- `useChains(filter)` — list with TanStack Query
- `useChainDetail(id)`
- `useStartExecution(chainId)`
- `useAdvanceStep(executionId)`

Новые компоненты:
- `ChainStepCard` — карточка шага с promтpt selector + variable mapping UI
- `ChainExecutionWizard` — пошаговый UI выполнения

### Tier mapping

| Лимит | Free | Pro | Max | Team |
|---|---|---|---|---|
| Цепочек создать | 1 (попробовать) | 5 | unlimited | unlimited |
| Шагов в одной цепочке | 3 | 10 | unlimited | unlimited |
| Запусков сохранять | 0 (одноразовый) | 10 | unlimited | unlimited |

### Effort

- **Backend:** 5-7 дней (модели, repo, service, HTTP, MCP)
- **Frontend:** 5-7 дней (3 страницы, drag-drop, wizard)
- **Тесты:** 2 дня (unit + integration)
- **Документация:** 1 день (CLAUDE.md, MCP.md update)

**Итого:** ~2 недели соло-разработки

### Риски и mitigations

| Риск | Вероятность | Mitigation |
|---|---|---|
| MCP-клиенты (Cursor/Cline/Claude Code) не сразу подхватят новые tools | средняя | Обновить MCP-документацию, написать example workflow, оповестить через MCP-чаты |
| UI run-mode сложен для не-техничных юзеров | высокая | Onboarding tutorial при первом запуске, video walkthrough, готовые starter chains в стартер-паке |
| Concurrent execution conflicts (один юзер запустил 2 chains параллельно) | низкая | Execution имеет уникальный ID, не блокирует chain в целом |
| Variable mapping syntax confusing | средняя | Visual UI для мапинга (dropdown «откуда брать значение»), не free-form text |

---

## Часть 2: Conditional Chains

### Зависит от
Prompt Chains должен быть полностью реализован первым. Conditional Chains — расширение поверх него.

### Что это в одном предложении
В цепочке можно добавить **условные шаги** — следующий шаг выбирается на основе содержимого output'а предыдущего.

### Конкретный пример

Цепочка **«Smart Code Review»**:

```
Шаг 1: «Анализ кода»
  → output: текст анализа

Шаг 2 (CONDITIONAL):
  ЕСЛИ output_шаг_1 contains "критический"
    → запустить промпт «Debug Session»
  ИНАЧЕ ЕСЛИ output_шаг_1 contains "warning"
    → запустить промпт «Suggest Refactoring»
  ИНАЧЕ
    → запустить промпт «Approve & Commit Message»

Шаг 3: «Финальный отчёт»
  → output: суммирующий отчёт
```

### Условия (DSL)

Простой DSL для условий — детерминистические matcher'ы (без LLM-judge):

| Условие | Описание | Пример |
|---|---|---|
| `contains(text)` | output содержит подстроку | `contains("error")` |
| `not_contains(text)` | не содержит | `not_contains("ok")` |
| `regex(pattern)` | совпадает regex | `regex("^ERROR:")` |
| `length_gt(n)` | длина > N символов | `length_gt(500)` |
| `length_lt(n)` | длина < N | `length_lt(100)` |
| `equals(text)` | точное совпадение | `equals("APPROVED")` |
| `starts_with(text)` | начинается с | `starts_with("BUG:")` |
| `ends_with(text)` | заканчивается на | `ends_with(".")` |

Логические операторы: `AND`, `OR`, `NOT`.

UI: visual condition builder (dropdown'ы), не free-form code. Например:
```
[output_шаг_1] [contains] ["критический"] AND [length_gt] [200]
```

### Технические детали

#### Расширение БД

```sql
-- Расширить prompt_chain_steps
ALTER TABLE prompt_chain_steps
  ADD COLUMN step_type VARCHAR(20) DEFAULT 'prompt', -- 'prompt' | 'conditional'
  ADD COLUMN conditions JSONB; -- если step_type = 'conditional'

-- Структура conditions JSONB:
-- {
--   "branches": [
--     {
--       "condition": {
--         "operator": "AND",
--         "rules": [
--           {"source": "step_1_output", "matcher": "contains", "value": "критический"},
--           {"source": "step_1_output", "matcher": "length_gt", "value": "200"}
--         ]
--       },
--       "next_step_prompt_id": 42
--     },
--     {
--       "condition": {"matcher": "default"},
--       "next_step_prompt_id": 56
--     }
--   ]
-- }
```

#### Service layer

В `internal/usecases/chain/`:
- `EvaluateCondition(condition Condition, ctx ExecutionContext) bool`
- `ResolveNextStep(step Step, ctx ExecutionContext) (*Step, error)`

Алгоритм при advance_chain_step:
1. Получить текущий шаг
2. Если `step_type == 'prompt'` — вернуть следующий по `position`
3. Если `step_type == 'conditional'` — для каждой branch вычислить condition, выбрать первую matching, вернуть её `next_step_prompt_id`
4. Если ни одна branch не matched и есть `default` — взять default
5. Если ничего нет — завершить цепочку

#### Frontend

- В Chain editor добавить новый тип шага «Conditional» (другой цвет / иконка)
- Visual condition builder UI (dropdown matcher + value input)
- В run mode visual flow: показать какая ветвь выбралась («✓ branch 2: contains "критический"»)

### Tier mapping

| Лимит | Free | Pro | Max | Team |
|---|---|---|---|---|
| Conditional шагов в цепочке | 0 | 0 (не доступно) | unlimited | unlimited |
| Branches per condition | — | — | unlimited | unlimited |
| Сложные conditions (AND/OR) | — | — | ✓ | ✓ |

**Решение:** Conditional Chains — **строго Max-only**. Это создаёт сильный Pro→Max driver: «обычные linear цепочки в Pro, smart logic — Max».

### Effort

- **Backend:** 3-4 дня (расширение модели, condition evaluator, resolver)
- **Frontend:** 3-4 дня (condition builder UI, type-toggle для шагов)
- **Тесты:** 1-2 дня (особенно edge-cases для evaluator)

**Итого:** ~1 неделя поверх готового Chains

### Риски

| Риск | Вероятность | Mitigation |
|---|---|---|
| Сложность UI condition builder | средняя | Простые matcher'ы (contains/regex), pre-set examples, tooltips |
| Регрессия в Chains при изменении модели | средняя | Default `step_type='prompt'`, все существующие шаги остаются как раньше |
| Цикл (бранч ведёт обратно) | низкая | При сохранении проверять граф на циклы, отказывать |

---

## Implementation Plan

### Phase A — базовые Chains (~2 недели)

| Шаг | Что | Effort |
|---|---|---|
| A1 | Миграция БД (3 таблицы) | 0.5 дня |
| A2 | Модели Go + Repository interface | 1 день |
| A3 | Postgres Repository implementation + tests | 1 день |
| A4 | Use-case Service (CRUD + Execution) + tests | 2 дня |
| A5 | HTTP handlers + DTO + routes | 1.5 дня |
| A6 | MCP tools (4 новых) + integration | 1 день |
| A7 | Frontend: API + хуки | 1 день |
| A8 | Frontend: List page | 1 день |
| A9 | Frontend: Editor page (drag-drop steps) | 2 дня |
| A10 | Frontend: Run mode wizard | 2 дня |
| A11 | Tier limits enforcement (через quota service) | 0.5 дня |
| A12 | E2E тестирование через Playwright | 1 день |

### Phase B — Conditional Chains (~1 неделя)

| Шаг | Что | Effort |
|---|---|---|
| B1 | Расширение БД (`step_type`, `conditions` JSONB) | 0.5 дня |
| B2 | Condition evaluator + resolver + tests | 1 день |
| B3 | Service: extend AdvanceStep с conditional logic | 0.5 дня |
| B4 | HTTP DTO для conditional steps | 0.5 дня |
| B5 | Frontend: condition builder component | 2 дня |
| B6 | Frontend: visual indicator branches в run mode | 1 день |
| B7 | E2E тестирование conditional flows | 1 день |

### Phase C — Sourcing данных (отдельно)

После A+B запустить:
- Запустить 3 готовых демо-цепочки в стартер-паке («PRD», «Code Review», «Контент-pipeline»)
- Документировать в CLAUDE.md паттерн использования
- Обновить MCP.md

---

## MCP Multi-Step Loop — manual test checklist

> **Обязательно ДО включения `CHAINS_ENABLED=true` в prod.** Проверяет, что MCP-клиенты (Cursor, Claude Code) корректно проходят полный цикл: `start_chain_execution` → клиент сам вызывает LLM по полученному prompt → `advance_chain_step` с output → повтор до `completed`.

### Подготовка

1. Backend: `docker compose -f docker-compose.dev.yml up -d --build` с `CHAINS_ENABLED=true` в `.env`.
2. Frontend: создать тестовую цепочку из 3 шагов:
   - Step 1: prompt «Сгенерируй идею для блог-поста по теме `{{topic}}`»
   - Step 2: prompt «Напиши outline на основе: `{{step_1_output}}`»
   - Step 3: prompt «Сгенерируй intro по outline: `{{step_2_output}}`»
3. Получить `chain_id` через UI → settings.

### Cursor / Claude Code setup

`.mcp.json` (в корне рабочей директории Cursor / `~/.config/claude-code/mcp.json`):

```json
{
  "mcpServers": {
    "promptvault-local": {
      "type": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer pvlt_<test_api_key_with_chains_scope>"
      }
    }
  }
}
```

### Manual run-through

В Cursor / Claude Code (новая чат-сессия):

1. **Шаг 1 — start.** Промпт юзера:
   > Запусти chain `<chain_id>` с переменными `{"topic": "TypeScript"}`.

   Ожидание: клиент вызывает `start_chain_execution(chain_id, {"topic": "TypeScript"})`. Получает обратно `{execution_id, current_step: {prompt: "..."}}`.

2. **Шаг 2 — LLM call.** Клиент должен **сам** вызвать LLM с полученным promptом (не звать tool). Проверяем что LLM получает rendered template (с подставленным `topic`).

3. **Шаг 3 — advance.** После получения LLM-ответа клиент вызывает `advance_chain_step(execution_id, step_output: "<ответ>")`. Получает следующий шаг.

4. **Цикл.** Повтор шагов 2-3 до получения `status: "completed"`.

### Что проверить

- [ ] **Template rendering** — `{{step_1_output}}` корректно заменяется на actual ответ Step 1 (видно в `current_step.prompt`).
- [ ] **Initiator-only** — попробовать `advance_chain_step` с другого MCP-клиента (другая API-key того же команды) → должен получить 403.
- [ ] **Snapshot** — отредактировать chain в UI после Start (изменить prompt Step 2). Запустить advance_step → должен использовать **старую** версию prompt'а из snapshot.
- [ ] **Conditional fork** (если есть Phase B) — chain с fork-step требует `chosen_branch_idx` в advance. Проверить что без него возвращается ошибка.
- [ ] **MCP quota** — каждый advance_step списывает MCP-квоту. После 13 successful advances Free-юзер должен получить 429.
- [ ] **UX через Cursor**: насколько естественно агент понимает loop? Не зацикливается ли (вызывает start снова вместо advance)?
- [ ] **Claude Code completion**: при `status: "completed"` агент должен корректно завершить цепочку и вернуть финальный output юзеру.

### Что записать в результате

Создать `docs/MCP_CHAIN_LOOP_REPORT.md` с:
- Дата теста + версия Cursor/Claude Code
- Использованная модель LLM (claude-sonnet-4-6 / gpt-4-turbo / etc)
- Прошёл ли цикл из 3 шагов без вмешательства человека
- Сколько advance_step вызовов потребовалось (ожидание: ровно 3)
- Найденные edge cases (зацикливание / неправильный context / etc)

### Решение по результату

| Результат | Действие |
|---|---|
| Цикл проходит чисто, агент сам понимает loop | `CHAINS_ENABLED=true` в prod, ship |
| Агент периодически зацикливается | Доработать system prompt в MCP server (добавить hint в `start_chain_execution` description) |
| Template rendering ломается | Регрессионный тест в `chain_test.go` + fix в `chain.AdvanceStep` |
| LLM игнорирует context из step_outputs | Усилить description полей `current_step` в MCP tool schema |

---

## Open Questions (что решить перед стартом)

1. **Resume executions** — делать ли уже в Phase A или отложить? Без этого юзер не может прервать длинную цепочку и вернуться. **Рекомендация:** включить — сохранение state в `prompt_chain_executions` уже спроектировано.

2. **Chain templates / starter** — добавлять ли 3-5 готовых цепочек в стартер-пак? **Рекомендация:** да, сильно поможет onboarding'у.

3. **Public sharing chains** — позволять ли шарить цепочки публично (как сейчас промпты)? **Рекомендация:** Phase 2, не сейчас.

4. **MCP integration scope** — Cursor / Claude Code умеют ли работать с многоступенчатыми вызовами? **Нужно проверить экспериментально** перед merging A6.

5. **Conditional matchers** — какой минимум поддержать в Phase B? **Рекомендация:** `contains`, `not_contains`, `regex`, `equals`, `starts_with`, `length_gt`/`length_lt` — этих 7 хватит для 90% use-cases.

---

## Связь с другими решениями (из BUSINESS_RESEARCH.md)

- **R1 (fix Max AI/MCP):** должен быть сделан до Phase A — иначе странно «Max получает unlimited chains, но меньше AI/день чем Pro»
- **R8 (limit rebalance):** должен быть сделан вместе с Phase A — лимиты chains встраиваются в Plan struct
- **R3 (Annual + Lifetime):** может идти параллельно — не блокирует
- **R7 (Public repo + ProductHunt):** Chains + Conditional — отличный wow-материал для launch'а на ProductHunt

---

## Дальше (когда вернёшься)

1. **Подтвердить план implementation** (Phase A → Phase B по этапам)
2. **Решить open questions** (5 вопросов выше)
3. **Запустить Phase A1** (миграция БД) — самый первый шаг кода
4. **Параллельно решить про 3-ю фичу:** оставить решение на post-launch (через 4-6 недель собрать реальный feedback) ИЛИ начать новый цикл brainstorm'а

**Когда вернёшься завтра:** прочитай этот документ + `BUSINESS_RESEARCH.md` §8 R8 → у тебя будет полный контекст где остановились.

---

## v3: Inline-tree editor

**Дата:** 2026-04-29. **Миграция:** `000056_chain_explicit_next`.

### Что поменялось

В v1/v2 порядок выполнения цепочки был задан **позицией шага** (`position+1`). Это работало для линейной цепочки, но при tree-структуре (ветки независимы) ломалось: последний шаг ветки A влетал в первый шаг ветки B по позиции — «протечка». UI создания развилок прямо в редакторе требовал явный граф.

В v3 переходы стали **явными**:

| Тип шага | Переход |
|---|---|
| `prompt` | `step.next_step_id` (`NULL` = конец ветки/цепочки) |
| `fork`   | `conditions.branches[chosen_branch_index].next_step_id` (как в v2) |

`position` остаётся как стабильная сортировка для UI и tie-break при поиске «root» шага.

### Schema (ALTER TABLE)

```sql
ALTER TABLE prompt_chain_steps
    ADD COLUMN next_step_id BIGINT REFERENCES prompt_chain_steps(id) ON DELETE SET NULL;

CREATE INDEX idx_prompt_chain_steps_next ON prompt_chain_steps (next_step_id)
    WHERE next_step_id IS NOT NULL;

-- Backfill: для существующих prompt-шагов сохраняем v2-поведение 1:1
UPDATE prompt_chain_steps p
   SET next_step_id = (
       SELECT n.id FROM prompt_chain_steps n
        WHERE n.chain_id = p.chain_id AND n.position = p.position + 1
        LIMIT 1
   )
 WHERE p.step_type = 'prompt';
```

### API

`POST /api/chains/:id/steps` — добавлены три взаимоисключающих поля для выбора места вставки:

| Поле | Семантика |
|---|---|
| `after_step_id` | новый шаг встаёт **после** указанного prompt-шага. Старая `after.next_step_id` становится `next_step_id` нового; `after.next_step_id = новый.id`. После fork-шага вставлять нельзя — `ErrCannotInsertAfterFork`. |
| `parent_fork_id` + `branch_index` | новый шаг становится **первым** шагом указанной ветки fork-шага. Старая `branch.next_step_id` становится `next_step_id` нового; `branch.next_step_id = новый.id`. |
| ничего | tail-mode: добавить в конец главной линии (после последнего prompt-шага без `next_step_id`). Если шагов нет — root. |

Если новый шаг сам — `fork`, у anchor не должно быть хвоста (`next` ≠ NULL) — иначе `ErrInsertForkLosesTail` (хвост ветки потерялся бы при подмене на fork).

`Service.RemoveStep` теперь «зашивает» граф:
- `RelinkPromptPredecessors(chainID, T.id, T.next_step_id)` — UPDATE one-shot.
- Для всех fork-шагов цепочки: если `branch.next_step_id == T.id`, то `branch.next_step_id = T.next_step_id`. Перезаписывается `conditions` JSONB.
- `DELETE FROM prompt_chain_steps WHERE id = T.id`.

`StepResponse` отдаёт поле `next_step_id`. Frontend строит дерево обходом из root (шаг, на который никто не ссылается) через `next_step_id` и `conditions.branches[].next_step_id`.

### Frontend

`pages/chains/editor.tsx` переписан полностью:
- `buildTree(steps)` — реконструирует дерево из плоского списка шагов; root = шаг без incoming references; рекурсивный обход с visited-set против петель.
- `<StepNode>` рендерит шаг и продолжение:
  - prompt → карточка + либо рекурсивный `<StepNode>` (если есть `next_step_id`), либо `<AddRow>` в листе.
  - fork → карточка-ветвление + nested список веток с собственными `<StepNode>` или «Ветка пуста» + `<AddRow>` для пустых веток.
- `<AddStepDialog>` — выбор промпта + опциональное имя шага.
- `<AddForkDialog>` — выбор промпта fork-шага + название + N веток (≥2, уникальные label'ы, можно добавлять/удалять). При сохранении шлёт `step_type=fork`, `conditions={branches:[{label, next_step_id: null}, …]}` + место вставки.
- Tier-gate Max: `<AddRow>` делает кнопку «+ Развилка» disabled с tooltip «Развилки доступны на тарифе Max» если `plan_id ∉ {max, max_yearly}`. Бэк всё равно проверяет (`ErrForkRequiresMax`) — defence-in-depth.
- Кнопка «Граф» в header ведёт на `/chains/:id/canvas` — Canvas остаётся как read-only визуализатор того же дерева.

### Что сознательно НЕ сделали в v3

- **Join после fork** (общий хвост, сходящийся после ветвления). В реальных сценариях ветки расходятся в листья. Если станет нужно — `branch.next_step_id` уже умеет указать на любой шаг, включая «общий» — добавится в UI без миграции.
- **DnD reorder** в tree-режиме. Линейный DnD из v1/v2 убран: с tree-структурой простая sortable-list-семантика противоречит явному графу. Добавится отдельной фичей с custom-логикой «переподшить next_step_id у соседей».
- **Insert step между двумя существующими** (например между A и B в линейной подцепочке). Сейчас можно только в конец листа. Если попросят — добавится `before_step_id` поле, которое подменит `prev.next_step_id` на новый, а `новый.next_step_id` = старое.
- **Visual condition builder для fork** (Phase 1 polish из v2 spec). С manual fork это уже не нужно — есть простая форма branch-labels.

### Известные недоделки

- `useUpdateStep` не принимает `step_type` / `conditions` / `next_step_id` — менять fork↔prompt или ребранчевать существующий fork через UI нельзя. Workaround: удалить и создать заново.
- При cycle-detect ошибке во время AddStep новый шаг откатывается, а изменение anchor (`after.next_step_id` или `branch.next_step_id`) — нет. Это редкий путь, но требует или транзакционности на repo-уровне, или двухфазного flow на сервисе.
- Editor не показывает `prompt.title` в карточке fork-шага (показывает «Промпт #ID»), хотя бэк это отдаёт. Дозаполнить.

