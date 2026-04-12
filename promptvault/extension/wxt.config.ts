import { defineConfig } from 'wxt';
import tailwindcss from '@tailwindcss/vite';

// See https://wxt.dev/api/config.html
export default defineConfig({
  modules: ['@wxt-dev/module-react'],
  srcDir: '.',
  outDir: '.output',
  manifest: {
    name: 'ПромтЛаб — библиотека AI-промптов',
    short_name: 'ПромтЛаб',
    description:
      'Быстрый доступ к вашей библиотеке промптов прямо в ChatGPT, Claude, Gemini, Perplexity. Требует аккаунт promtlabs.ru.',
    version: '0.1.0',
    minimum_chrome_version: '116',
    permissions: ['sidePanel', 'storage', 'activeTab', 'scripting'],
    host_permissions: [
      'https://chatgpt.com/*',
      'https://claude.ai/*',
      'https://gemini.google.com/*',
      'https://www.perplexity.ai/*',
      'https://promtlabs.ru/*',
      'https://*.promtlabs.ru/*',
      'http://localhost:8080/*',
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
  }),
});
