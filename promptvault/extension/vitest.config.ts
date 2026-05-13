import { defineConfig } from 'vitest/config';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  test: {
    globals: true,
    environment: 'happy-dom',
    include: ['tests/**/*.test.ts'],
    exclude: ['node_modules', '.output', '.wxt'],
    setupFiles: ['./tests/setup.ts'],
  },
  resolve: {
    alias: {
      '@pv/shared': path.resolve(__dirname, '../shared/src'),
    },
  },
});
