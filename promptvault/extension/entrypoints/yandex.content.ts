import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: [
    'https://alice.yandex.ru/*',
    'https://ya.ru/*',
    'https://yandex.ru/alice*',
  ],
  runAt: 'document_idle',
  main() {
    const host = location.host as 'alice.yandex.ru' | 'ya.ru' | 'yandex.ru';
    installContentHandler(host);
  },
});
