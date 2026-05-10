import js from '@eslint/js'
import globals from 'globals'
import reactPlugin from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import jsxA11y from 'eslint-plugin-jsx-a11y'
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
      // MN-54: React + jsx-a11y. recommended config — sane defaults.
      // react-in-jsx-scope OFF (React 19 не требует import React).
      reactPlugin.configs.flat.recommended,
      reactPlugin.configs.flat['jsx-runtime'],
      jsxA11y.flatConfigs.recommended,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    settings: {
      react: { version: 'detect' },
    },
    rules: {
      // React 19 не требует import React.
      'react/react-in-jsx-scope': 'off',
      // prop-types — тип-чек делает TypeScript, plugin-react ругается на отсутствующие
      // propTypes даже когда есть TS-типы. Вырубаем.
      'react/prop-types': 'off',
      // cmdk lib использует data-атрибуты типа `cmdk-input-wrapper=""` для CSS targeting
      // (см. https://cmdk.paco.me). Это валидный data-attr, не unknown-property.
      'react/no-unknown-property': ['error', { ignore: ['cmdk-input-wrapper', 'cmdk-input', 'cmdk-list', 'cmdk-item'] }],
      // jsx-a11y/no-autofocus — autofocus в PromptLab используется ТОЛЬКО внутри
      // Dialog'ов / Modal'ов / Wizard-форм (TOTP code, sign-in TOTP, edit-team
      // dialog, file-import choice, invite-dialog), где это стандартное и оправданное
      // UX-поведение (focus-trap context, юзер ждёт ввода). На страницах верхнего
      // уровня autoFocus не используется — поэтому правило отключено глобально.
      'jsx-a11y/no-autofocus': 'off',
      // MN-54 baseline: shadcn `<Label>` — обёртка без htmlFor (по дизайну shadcn —
      // Label и input живут в одном flex-container, не nested и не связаны htmlFor).
      // Правило false-positive'ит на эти случаи; downgrade до warn — реальные нарушения
      // видны как warnings, но не блокируют build. Per-file fix отдельным PR.
      'jsx-a11y/label-has-associated-control': 'warn',
      // div с onClick — иногда нужно (overlay, swipe target). Реально проблема в
      // нескольких местах (member-list, prompt-card row) — фиксим отдельно.
      'jsx-a11y/click-events-have-key-events': 'warn',
      'jsx-a11y/no-static-element-interactions': 'warn',
      'jsx-a11y/no-noninteractive-element-interactions': 'warn',
      // tablist + interactive-supports-focus — landing/product-demo показывает
      // декоративные tab pills, не реальный ARIA tablist; downgrade до warn.
      'jsx-a11y/interactive-supports-focus': 'warn',
    },
  },
])
