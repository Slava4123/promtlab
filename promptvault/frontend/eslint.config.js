import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  // playwright/* — это E2E тесты, не React. `use` там — это Playwright fixture
  // API (https://playwright.dev/docs/test-fixtures), не React Hook. Изначально
  // playwright-spec'и попадали в react-hooks/rules-of-hooks правило и ругались
  // на каждый { use } destructure → 4 false-positive errors.
  globalIgnores(['dist', 'playwright/**', 'tests/e2e/**']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
  },
])
