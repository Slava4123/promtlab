import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://giga.chat/*', 'https://developers.sber.ru/*'],
  runAt: 'document_idle',
  main() {
    const host = location.host as 'giga.chat' | 'developers.sber.ru';
    installContentHandler(host);
  },
});
