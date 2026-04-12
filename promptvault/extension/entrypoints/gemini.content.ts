import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://gemini.google.com/*'],
  runAt: 'document_idle',
  main() {
    installContentHandler('gemini.google.com');
  },
});
