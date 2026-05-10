import { test as setup, expect } from '@playwright/test'
import path from 'path'
import fs from 'fs'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// Логиним каждый тестовый тир и сохраняем cookie+localStorage в storageState.
// После этого specs c project=free|pro|max используют готовое состояние,
// не повторяя UI-логин в каждом тесте — экономит ~2 сек на спек.
//
// Тестовые юзеры создаются через scripts/seed-test-data.sql. Пароль один — TestPass2026!
// (см. .env.test). Если seed не запущен — тест упадёт на первом expect.

const authDir = path.resolve(__dirname, '.auth')
fs.mkdirSync(authDir, { recursive: true })

const tiers = [
  { tier: 'free', email: 'e2e-free@test.local' },
  { tier: 'pro',  email: 'e2e-pro@test.local'  },
  { tier: 'max',  email: 'e2e-max@test.local'  },
] as const

for (const u of tiers) {
  setup(`authenticate ${u.tier}`, async ({ page }) => {
    const password = process.env.E2E_TEST_PASSWORD
    if (!password) {
      throw new Error('E2E_TEST_PASSWORD не установлен — проверь frontend/.env.test')
    }

    await page.goto('/sign-in')
    await page.getByRole('textbox', { name: 'Email' }).fill(u.email)
    await page.getByRole('textbox', { name: 'Пароль' }).fill(password)

    // Login response содержит user.plan_id — самый надёжный способ проверить
    // что юзер реально на test_-плане (без reload-race и token-refresh-race).
    const loginRespPromise = page.waitForResponse(
      (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
      { timeout: 10_000 },
    )
    // exact:true — иначе матчатся «Войти через GitHub/Google/Яндекс».
    await page.getByRole('button', { name: 'Войти', exact: true }).click()
    const loginResp = await loginRespPromise
    expect(loginResp.status(), `login ${u.tier} должен вернуть 200`).toBe(200)
    const loginBody = await loginResp.json()
    expect(loginBody.user.email).toBe(u.email)
    expect(loginBody.user.plan_id, `plan_id ${u.tier} должен начинаться с test_`).toMatch(/^test_/)

    // Дожидаемся редиректа — это подтверждает что cookie + access_token приняты фронтом.
    await page.waitForURL(/\/(welcome|dashboard|prompts)/, { timeout: 10_000 })

    await page.context().storageState({ path: path.join(authDir, `${u.tier}.json`) })
  })
}
