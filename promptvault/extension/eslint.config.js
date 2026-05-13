import js from '@eslint/js'
import globals from 'globals'
import reactPlugin from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

// Mirror frontend/eslint.config.js с WXT-адаптациями:
// - игнор .output (build) и .wxt (codegen) вместо dist
// - shared/ workspace включён как тот же фронтовый пакет
// - webextensions globals (chrome) для background и content-scripts
export default defineConfig([
  globalIgnores([
    '.output/**',
    '.wxt/**',
    'node_modules/**',
    // shared/ имеет собственный tsconfig — lint'им из его пути если нужно
  ]),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactPlugin.configs.flat.recommended,
      reactPlugin.configs.flat['jsx-runtime'],
      jsxA11y.flatConfigs.recommended,
    ],
    languageOptions: {
      ecmaVersion: 2022,
      globals: {
        ...globals.browser,
        ...globals.webextensions,
      },
    },
    settings: {
      react: { version: 'detect' },
    },
    rules: {
      // React 19 не требует import React.
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      // cmdk пробрасывает свои data-атрибуты для CSS targeting.
      'react/no-unknown-property': [
        'error',
        { ignore: ['cmdk-input-wrapper', 'cmdk-input', 'cmdk-list', 'cmdk-item'] },
      ],
      // autoFocus используется в auth-pages и dialog-wizard'ах (focus-trap context).
      'jsx-a11y/no-autofocus': 'off',
      // shadcn <Label> без hthmlFor — обёртка по дизайну shadcn; downgrade до warn.
      'jsx-a11y/label-has-associated-control': 'warn',
      'jsx-a11y/click-events-have-key-events': 'warn',
      'jsx-a11y/no-static-element-interactions': 'warn',
      'jsx-a11y/no-noninteractive-element-interactions': 'warn',
      'jsx-a11y/interactive-supports-focus': 'warn',
      // React 19 + react-hooks v7 ввели два правила, которые широко срабатывают
      // на легитимных паттернах в существующем коде (форма-state из user/data,
      // useEffect для синхронизации с server-state из TanStack Query). Они дают
      // ценные сигналы — оставляем как warn, фиксим точечно отдельным PR.
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/static-components': 'warn',
      // Игнорировать unused vars с префиксом `_` (стандарт TS-проектов).
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_', caughtErrorsIgnorePattern: '^_' },
      ],
    },
  },
  // entrypoints/*.content.ts — content scripts. Те же globals что и для sidepanel.
  // Background — отдельно: только webextensions globals (нет window/document на старте).
  {
    files: ['entrypoints/background.ts'],
    languageOptions: {
      globals: {
        ...globals.webextensions,
        ...globals.serviceworker,
      },
    },
  },
])
