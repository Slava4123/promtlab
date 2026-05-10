import { test, expect } from '@playwright/test'

test.describe('Max tier — smoke', () => {
  test('test_max юзер видит /pricing с 3 prod-карточками', async ({ page }) => {
    await page.goto('/pricing')

    await expect(page.getByRole('heading', { name: 'Тарифы', level: 1 })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByRole('heading', { name: 'Free', level: 3 })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Pro', level: 3 })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Max', level: 3 })).toBeVisible()
  })
})
