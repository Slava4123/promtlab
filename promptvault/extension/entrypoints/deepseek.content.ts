import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://chat.deepseek.com/*'],
  runAt: 'document_idle',
  main() {
    installContentHandler('chat.deepseek.com');
  },
});
