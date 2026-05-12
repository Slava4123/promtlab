import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://chat.qwen.ai/*'],
  runAt: 'document_idle',
  main() {
    installContentHandler('chat.qwen.ai');
  },
});
