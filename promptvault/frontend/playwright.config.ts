import { defineConfig, devices } from '@playwright/test'
import dotenv from 'dotenv'
import path from 'path'
import { fileURLToPath } from 'url'

// ESM-безопасный аналог __dirname.
const __dirname = path.dirname(fileURLToPath(import.meta.url))

// .env.test содержит E2E_TEST_PASSWORD и BASE_URL для test-стека.
// Файл коммитим — пароль действует только для test-юзеров в dev-БД.
dotenv.config({ path: path.resolve(__dirname, '.env.test') })

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:5173'

export default defineConfig({
  testDir: './playwright',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  // 1 retry даже локально: фронт может делать refresh-token + /me на холодный старт,
  // первая загрузка изредка падает по таймауту /me. После Phase 2 fixture с явным
  // waitForLoadState retry должен стать ненужным.
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 1 : 1,
  reporter: process.env.CI ? [['html', { open: 'never' }], ['list']] : 'list',
  // 60s — quota-specs делают по 5-7 page-навигаций (создать N промптов до лимита).
  // Каждый goto + networkidle ≈ 1-2с в dev-стеке.
  timeout: 90_000,
  expect: { timeout: 10_000 },

  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'setup',
      testMatch: /auth\.setup\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'free',
      testMatch: /\.free\.spec\.ts$/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Chrome'], storageState: 'playwright/.auth/free.json' },
    },
    {
      name: 'pro',
      testMatch: /\.pro\.spec\.ts$/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Chrome'], storageState: 'playwright/.auth/pro.json' },
    },
    {
      name: 'max',
      testMatch: /\.max\.spec\.ts$/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Chrome'], storageState: 'playwright/.auth/max.json' },
    },
  ],
})
