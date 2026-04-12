import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://claude.ai/*'],
  runAt: 'document_idle',
  main() {
    installContentHandler('claude.ai');
  },
});
