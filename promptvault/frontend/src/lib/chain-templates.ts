// Phase 16 UI polish: 4 hardcoded шаблона цепочек для empty-state галереи на
// странице /chains. Клик по карточке шаблона создаёт N промптов и цепочку
// в активном workspace через applyTemplate.
//
// Шаблоны намеренно простые и общие — идея не «дать готовое решение», а
// показать структуру: «output одного шага → variable следующего». Юзер
// после applyTemplate видит готовый граф в /chains/{id}/edit и может
// переписать prompt-content под свой стек.
//
// Если шаблонов станет больше 6-8 — переехать на backend table starter_chains
// (пока YAGNI, hardcode достаточно).

import { api } from "@/api/client"
import type { Chain, Prompt } from "@/api/types"

/**
 * Шаблон одного шага цепочки.
 *
 * variableMapping — отображение `{переменная_в_промпте: VariableSource}`.
 * Для первого шага обычно пусто (юзер вводит вручную в run-mode).
 * Для последующих — берёт `step_output` предыдущего шага по `stepIndex` (0-based
 * в template; реальный step_id resolves'ится при создании в applyTemplate).
 */
export interface ChainTemplateStep {
  /** Имя шага (отображается в Canvas/Run UI). */
  name: string
  /** Контент промпта с шаблонными переменными `{{var_name}}`. */
  promptContent: string
  /**
   * Маппинг переменных из output'ов предыдущих шагов.
   * Ключ — имя переменной (как в `{{var_name}}` в promptContent).
   * Значение — индекс предыдущего шага (0-based) или null для manual ввода.
   *
   * applyTemplate перепишет `null` → `{type:"manual"}` и `number` →
   * `{type:"step_output", step_id: <реальный id>}`.
   */
  variableMapping: Record<string, number | null>
}

export interface ChainTemplate {
  /** Стабильный slug (для tracking/aria-id). */
  id: string
  /** Эмодзи в карточке галереи. */
  emoji: string
  /** Заголовок «Идея → PRD → Tests». */
  title: string
  /** 1-line описание use case'а. */
  description: string
  /** Шаги (последовательно). Должно быть ≥2 шагов. */
  steps: ChainTemplateStep[]
}

export const CHAIN_TEMPLATES: ChainTemplate[] = [
  {
    id: "idea-prd-tests",
    emoji: "💡",
    title: "Идея → PRD → Tests",
    description: "Превратите идею в спецификацию и набор тестов на её основе.",
    steps: [
      {
        name: "Идея",
        promptContent:
          "Опишите вашу продуктовую идею:\n- Что за проблема?\n- Кто целевой пользователь?\n- Чем это лучше существующих решений?\n\nИдея: {{idea}}",
        variableMapping: { idea: null },
      },
      {
        name: "PRD на основе идеи",
        promptContent:
          "На основе следующей идеи составь PRD (product requirements document):\n\n{{idea_response}}\n\nСтруктура: Цели, Метрики успеха, User stories, Out of scope, Зависимости, Риски.",
        variableMapping: { idea_response: 0 },
      },
      {
        name: "Tests на основе PRD",
        promptContent:
          "На основе PRD ниже сгенерируй чек-лист функциональных тестов:\n\n{{prd}}\n\nКаждый тест: что делаем (action), что ожидаем (expected), edge case.",
        variableMapping: { prd: 1 },
      },
    ],
  },
  {
    id: "code-review-tests-docs",
    emoji: "🔍",
    title: "Code review → Tests → Docs",
    description: "После написания фичи: ревью кода, генерация unit-тестов, обновление документации.",
    steps: [
      {
        name: "Code review",
        promptContent:
          "Проведи code review для следующего diff'а. Найди bugs, security issues, code smells.\n\n{{diff}}\n\nФормат: пронумерованный список замечаний с указанием severity (low/medium/high).",
        variableMapping: { diff: null },
      },
      {
        name: "Tests на изменённый код",
        promptContent:
          "На основе review ниже напиши unit-тесты для покрытия исправленных мест:\n\n{{review}}\n\nИспользуй стек проекта (Go test / vitest — спроси если непонятно).",
        variableMapping: { review: 0 },
      },
      {
        name: "Обновление документации",
        promptContent:
          "На основе review и тестов ниже обнови README/CHANGELOG раздел:\n\n{{review_and_tests}}\n\nКратко: что изменилось, почему, как использовать.",
        variableMapping: { review_and_tests: 1 },
      },
    ],
  },
  {
    id: "research-outline-draft",
    emoji: "📝",
    title: "Research → Outline → Draft",
    description: "Контентная цепочка: поиск источников, план статьи, черновик.",
    steps: [
      {
        name: "Research темы",
        promptContent:
          "Соберись research для темы «{{topic}}»:\n- 5-7 ключевых тезисов\n- Известные авторы / источники\n- Контр-аргументы\n- Свежие тренды (последние 12 месяцев)",
        variableMapping: { topic: null },
      },
      {
        name: "Outline статьи",
        promptContent:
          "На основе research ниже составь план статьи (~1500 слов):\n\n{{research}}\n\nСтруктура: Hook (зачем читать), 4-6 разделов с подзаголовками, Заключение, CTA.",
        variableMapping: { research: 0 },
      },
      {
        name: "Draft статьи",
        promptContent:
          "Напиши черновик статьи по плану ниже:\n\n{{outline}}\n\nСтиль: разговорный, с примерами. Каждый раздел — 200-300 слов.",
        variableMapping: { outline: 1 },
      },
    ],
  },
  {
    id: "bug-repro-fix-docs",
    emoji: "🐛",
    title: "Bug → Repro → Fix → Docs",
    description: "Bug-fix workflow: воспроизведение, патч, обновление доков.",
    steps: [
      {
        name: "Описание бага",
        promptContent:
          "Опиши баг детально:\n- Симптом\n- Контекст (версия, окружение)\n- Что ожидалось vs что произошло\n\nБаг: {{bug}}",
        variableMapping: { bug: null },
      },
      {
        name: "Шаги воспроизведения",
        promptContent:
          "На основе описания бага ниже напиши минимальные шаги воспроизведения:\n\n{{bug_description}}\n\nКаждый шаг — одна команда/клик. Включи expected vs actual на финальном шаге.",
        variableMapping: { bug_description: 0 },
      },
      {
        name: "Fix (патч)",
        promptContent:
          "На основе репро ниже предложи фикс:\n\n{{repro_steps}}\n\nФормат: краткое объяснение root cause + diff кода. Если есть несколько вариантов — сравни trade-off'ы.",
        variableMapping: { repro_steps: 1 },
      },
      {
        name: "Обновление документации",
        promptContent:
          "На основе фикса ниже обнови документацию (CHANGELOG, README, troubleshooting раздел):\n\n{{fix}}\n\nКратко: что было сломано, как теперь работает.",
        variableMapping: { fix: 2 },
      },
    ],
  },
]

/**
 * Применяет шаблон цепочки в активном workspace: создаёт N промптов и
 * связывает их в цепочку через step_output mapping.
 *
 * Возвращает созданную Chain (без шагов в response — UI редиректит в
 * `/chains/{id}/edit` где они догружаются).
 *
 * Если активна команда (`teamId !== null`), промпты и цепочка создаются с
 * этим team_id. Backend проверит canWrite (owner/editor) — viewer получит
 * 403, и frontend должен скрыть галерею для viewer'ов до клика.
 */
export async function applyTemplate(
  template: ChainTemplate,
  teamId: number | null,
): Promise<Chain> {
  // Шаг 1: создать промпты
  const promptIds: number[] = []
  for (const step of template.steps) {
    const prompt = await api<Prompt>(`/prompts`, {
      method: "POST",
      body: JSON.stringify({
        title: `${template.emoji} ${step.name}`,
        content: step.promptContent,
        team_id: teamId ?? undefined,
      }),
    })
    promptIds.push(prompt.id)
  }

  // Шаг 2: создать цепочку
  const chain = await api<Chain>(`/chains`, {
    method: "POST",
    body: JSON.stringify({
      name: template.title,
      description: template.description,
      team_id: teamId ?? undefined,
    }),
  })

  // Шаг 3: добавить шаги в цепочку. variable_mapping ссылается на step_id
  // ранее созданных шагов (по их index в template.steps[]).
  const stepIds: number[] = []
  for (let i = 0; i < template.steps.length; i++) {
    const tStep = template.steps[i]
    const variableMapping: Record<string, { type: string; step_id?: number }> = {}
    for (const [varName, sourceIdx] of Object.entries(tStep.variableMapping)) {
      if (sourceIdx === null) {
        variableMapping[varName] = { type: "manual" }
      } else if (typeof sourceIdx === "number" && sourceIdx >= 0 && sourceIdx < stepIds.length) {
        variableMapping[varName] = { type: "step_output", step_id: stepIds[sourceIdx] }
      }
    }
    const created = await api<{ id: number }>(`/chains/${chain.id}/steps`, {
      method: "POST",
      body: JSON.stringify({
        prompt_id: promptIds[i],
        name: tStep.name,
        variable_mapping: variableMapping,
        step_type: "prompt",
      }),
    })
    stepIds.push(created.id)
  }

  return chain
}
