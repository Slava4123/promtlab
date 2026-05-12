import { defineConfig } from 'wxt';
import tailwindcss from '@tailwindcss/vite';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const sharedDir = path.resolve(__dirname, '../shared/src');

// See https://wxt.dev/api/config.html
export default defineConfig({
  modules: ['@wxt-dev/module-react'],
  srcDir: '.',
  outDir: '.output',
  manifest: {
    name: 'ПромтЛаб — библиотека AI-промптов',
    short_name: 'ПромтЛаб',
    description:
      'Полный клиент ПромтЛаба: библиотека, цепочки, команды и подписка на 9 AI-сайтах (ChatGPT, Claude, Gemini, Perplexity, Yandex GPT, GigaChat, DeepSeek, Mistral, Qwen). Требует аккаунт promtlabs.ru.',
    version: '1.0.0',
    minimum_chrome_version: '116',
    permissions: ['sidePanel', 'storage', 'activeTab', 'scripting', 'contextMenus'],
    host_permissions: [
      'https://chatgpt.com/*',
      'https://claude.ai/*',
      'https://gemini.google.com/*',
      'https://www.perplexity.ai/*',
      'https://alice.yandex.ru/*',
      'https://ya.ru/*',
      'https://yandex.ru/alice*',
      'https://giga.chat/*',
      'https://developers.sber.ru/*',
      'https://chat.deepseek.com/*',
      'https://chat.mistral.ai/*',
      'https://le-chat.mistral.ai/*',
      'https://chat.qwen.ai/*',
      'https://promtlabs.ru/*',
      'https://*.promtlabs.ru/*',
      // GlitchTip endpoint для Sentry envelope (см. lib/sentry-envelope.ts).
      // Если DSN указывает на другой хост — обновить здесь и пересобрать.
      'https://glitchtip.promtlabs.ru/*',
    ],
    icons: {
      16: 'icon/16.png',
      32: 'icon/32.png',
      48: 'icon/48.png',
      128: 'icon/128.png',
    },
    action: {
      default_title: 'Открыть ПромтЛаб',
      default_icon: {
        16: 'icon/16.png',
        32: 'icon/32.png',
      },
    },
    side_panel: {
      default_path: 'sidepanel.html',
    },
    commands: {
      _execute_action: {
        suggested_key: {
          default: 'Ctrl+Shift+K',
          mac: 'Command+Shift+K',
        },
        description: 'Открыть боковую панель ПромтЛаба',
      },
    },
  },
  vite: () => ({
    plugins: [tailwindcss()],
    resolve: {
      alias: {
        '@pv/shared': sharedDir,
      },
    },
  }),
});
