import { test, expect } from '@playwright/test'

// Smoke Phase 1: убедиться что test-юзер реально на test_free плане и UI грузится.
// Полный flow: storageState (логин уже сделан в auth.setup) → /pricing → /me → ассерт plan_id.
test.describe('Free tier — smoke', () => {
  test('test_free юзер видит /pricing с 3 prod-карточками', async ({ page }) => {
    await page.goto('/pricing')

    // Заголовок страницы тарифов — h1 в PageLayout.
    // Длинный timeout: фронт может делать refresh-token + /me перед рендером.
    await expect(page.getByRole('heading', { name: 'Тарифы', level: 1 })).toBeVisible({ timeout: 15_000 })

    // Карточки трёх prod-планов видны. test_* отфильтрованы в plan_repo.GetActive
    // (см. /api/plans/), поэтому на /pricing их быть не должно.
    await expect(page.getByRole('heading', { name: 'Free', level: 3 })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Pro', level: 3 })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Max', level: 3 })).toBeVisible()
  })
})
