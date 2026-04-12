import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://www.perplexity.ai/*'],
  runAt: 'document_idle',
  main() {
    installContentHandler('www.perplexity.ai');
  },
});
