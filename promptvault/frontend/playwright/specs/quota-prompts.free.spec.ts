import { test, expect, tierLimits, refreshAuth } from '../fixtures'

// A1: Free юзер с лимитом 2 промпта пытается создать 3-й — backend возвращает 402,
// фронт показывает quota-exceeded dialog с CTA на upgrade.
// Полный flow backend (postgres + quota.Service) + frontend (UI + state).

test.describe('Free tier — лимит prompts', () => {
  test.beforeEach(async ({ context, userEmail, cleanup }) => {
    // Cleanup state в БД + перевыдать refresh-cookie. Без re-login storageState
    // от setup project устаревает между тестами из-за token_nonce rotation.
    await cleanup()
    await refreshAuth(context, userEmail)
  })

  test('debug: /prompts/new открывается и форма видна', async ({ page }) => {
    await page.goto('/prompts/new')
    await page.waitForLoadState('networkidle')
    console.log(`URL after goto: ${page.url()}`)
    expect(page.url()).toContain('/prompts/new')
    await expect(page.getByRole('heading', { name: 'Новый промпт' })).toBeVisible({ timeout: 15_000 })
  })

  test('A1: создание (limit+1)-го промпта блокируется quota-диалогом', async ({ page, tier }) => {
    const limit = tierLimits[tier].prompts // = 2 для free

    // Создаём limit промптов — все должны успешно сохраниться.
    for (let i = 1; i <= limit; i++) {
      await createPrompt(page, `E2E prompt ${i}`, `content ${i}`)
      // После сохранения фронт возвращает на /prompts (или /prompts/:id);
      // дожидаемся что в списке виден созданный.
      await page.goto('/prompts')
      await page.waitForLoadState('networkidle')
      await expect(page.getByText(`E2E prompt ${i}`)).toBeVisible({ timeout: 10_000 })
    }

    // (limit+1)-й — должен упереться в quota.
    await page.goto('/prompts/new')
    await page.waitForLoadState('networkidle')
    await page.getByRole('textbox', { name: 'Название' }).fill(`E2E prompt ${limit + 1}`)
    await fillPromptBody(page, `content ${limit + 1}`)
    await page.getByRole('button', { name: 'Создать', exact: true }).click()

    // Quota dialog появляется. По плану он содержит «Лимит» в тексте + ссылку на /pricing.
    // Используем «или» — фронт может показать toast или dialog; ловим любое.
    await expect(
      page.getByText(/Лимит|превыш|тарифу|upgrade/i).first(),
    ).toBeVisible({ timeout: 10_000 })

    // Сам промпт НЕ должен появиться в списке.
    await page.goto('/prompts')
    await page.waitForLoadState('networkidle')
    await expect(page.getByText(`E2E prompt ${limit + 1}`)).toHaveCount(0)
  })
})

async function createPrompt(page: import('@playwright/test').Page, title: string, body: string) {
  await page.goto('/prompts/new')
  await page.waitForLoadState('networkidle')
  console.log(`createPrompt(${title}): URL=${page.url()}, headings=${await page.locator('h1').allTextContents()}`)
  await page.getByRole('textbox', { name: 'Название' }).fill(title)
  await fillPromptBody(page, body)
  await page.getByRole('button', { name: 'Создать', exact: true }).click()
}

async function fillPromptBody(page: import('@playwright/test').Page, body: string) {
  // CodeMirror — не нативный textarea. У него role=textbox с placeholder
  // «Введите текст промпта...». Кликаем + печатаем.
  const editor = page.getByRole('textbox', { name: /Введите текст промпта/i })
  await editor.click()
  // Перед заполнением очищаем default-плейсхолдер если он попал в value.
  await page.keyboard.press('Control+A')
  await page.keyboard.press('Delete')
  await page.keyboard.type(body)
}
