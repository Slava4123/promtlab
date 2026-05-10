import { test as base, expect, type APIRequestContext, type Page, type BrowserContext } from '@playwright/test'

// Per-tier mapping для cleanup-helper и POM-методов.
// Берётся из storageState — playwright.config.ts задаёт его per-project.
type Tier = 'free' | 'pro' | 'max'

const tierEmail: Record<Tier, string> = {
  free: 'e2e-free@test.local',
  pro: 'e2e-pro@test.local',
  max: 'e2e-max@test.local',
}

/** Лимиты test_-планов. Должны совпадать с scripts/seed-test-data.sql. */
export const tierLimits: Record<Tier, {
  prompts: number
  collections: number
  teams: number
  teamMembers: number
  shareLinks: number
  dailyShares: number
  extDaily: number
  mcpDaily: number
  chains: number
  stepsPerChain: number
  savedExecutions: number
}> = {
  // ВАЖНО: цифры синхронизированы с scripts/seed-test-data.sql.
  // Если меняешь там — меняй здесь, иначе specs упадут.
  free: { prompts: 1, collections: 1, teams: 1, teamMembers: 1, shareLinks: 1, dailyShares: 1, extDaily: 1, mcpDaily: 2, chains: 1, stepsPerChain: 2, savedExecutions: 0 },
  pro:  { prompts: 2, collections: 2, teams: 2, teamMembers: 2, shareLinks: 2, dailyShares: 2, extDaily: 2, mcpDaily: 3, chains: 2, stepsPerChain: 3, savedExecutions: 1 },
  max:  { prompts: 3, collections: 3, teams: 3, teamMembers: 3, shareLinks: 3, dailyShares: 3, extDaily: 3, mcpDaily: 5, chains: 3, stepsPerChain: 4, savedExecutions: 2 },
}

/** Хелпер: вычисляет тир из storageState. Использует email в cookies/local — */
function tierFromStorageState(stateFile: string | undefined): Tier {
  if (!stateFile) throw new Error('storageState не задан в проекте Playwright')
  if (stateFile.includes('free')) return 'free'
  if (stateFile.includes('pro'))  return 'pro'
  if (stateFile.includes('max'))  return 'max'
  throw new Error(`не удалось определить tier из storageState: ${stateFile}`)
}

type Fixtures = {
  /** Текущий тир теста (free/pro/max) — определяется по storageState проекта. */
  tier: Tier
  /** Email текущего юзера — для cleanup и UI ассертов. */
  userEmail: string
  /**
   * Хелпер cleanup. Вызывает POST /api/test/cleanup?email=...
   * Используется в beforeEach каждого quota-spec, чтобы тесты не влияли друг на друга.
   * Endpoint существует только при SERVER_ENVIRONMENT=development.
   */
  cleanup: (email?: string) => Promise<void>
  /** Открыть /prompts и подождать пока интерфейс полностью загрузится (включая token-refresh). */
  gotoPrompts: () => Promise<void>
}

/**
 * Перевыдать refresh-cookie через прямой POST /api/auth/login.
 *
 * Зачем: backend ротирует `users.token_nonce` при /api/auth/refresh. После 1-го
 * теста в spec'е storageState-cookie становится невалидным (nonce в БД новый),
 * и фронт следующего теста редиректит на /sign-in. Re-login исправляет это
 * без UI-навигации.
 *
 * Не override'им context fixture (это конфликтует с trace recording — см.
 * ENOENT артефакты в Playwright). Вместо этого вызываем функцию в beforeEach
 * каждого spec'а через extended test, и работаем с обычным context от Playwright.
 */
export async function refreshAuth(context: BrowserContext, email: string) {
  const password = process.env.E2E_TEST_PASSWORD
  if (!password) throw new Error('E2E_TEST_PASSWORD не установлен — проверь frontend/.env.test')

  // Удалить старый refresh-cookie перед re-login: backend Set-Cookie заменяет,
  // но clearCookies гарантирует чистое состояние и не даёт подцепить мёртвый nonce.
  await context.clearCookies()

  const resp = await context.request.post('/api/auth/login', {
    data: { email, password },
  })
  if (!resp.ok()) {
    throw new Error(`re-login ${email} failed HTTP ${resp.status()}: ${await resp.text()}`)
  }
}

export const test = base.extend<Fixtures>({
  tier: async ({}, use, testInfo) => {
    const projectName = testInfo.project.name as Tier
    if (!['free', 'pro', 'max'].includes(projectName)) {
      throw new Error(`fixtures.ts ожидают project ∈ {free,pro,max}, получили "${projectName}"`)
    }
    await use(projectName)
  },
  userEmail: async ({ tier }, use) => {
    await use(tierEmail[tier])
  },
  cleanup: async ({ playwright, baseURL, userEmail }, use) => {
    const apiContext: APIRequestContext = await playwright.request.newContext({ baseURL })
    const fn = async (email?: string) => {
      const target = email ?? userEmail
      const resp = await apiContext.post('/api/test/cleanup', {
        params: { email: target },
      })
      expect(resp.status(), `cleanup ${target} должен вернуть 200`).toBe(200)
    }
    await use(fn)
    await apiContext.dispose()
  },
  gotoPrompts: async ({ page }, use) => {
    const fn = async () => {
      await page.goto('/prompts')
      await page.waitForLoadState('networkidle')
    }
    await use(fn)
  },
})

export { expect, type Page }

// Экспорт служебных функций для не-fixture использования (например, в setup-проекте).
export { tierEmail, tierFromStorageState }
