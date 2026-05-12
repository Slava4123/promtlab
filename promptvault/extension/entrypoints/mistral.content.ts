import { defineContentScript } from 'wxt/utils/define-content-script';
import { installContentHandler } from '../lib/content-handler';

export default defineContentScript({
  matches: ['https://chat.mistral.ai/*', 'https://le-chat.mistral.ai/*'],
  runAt: 'document_idle',
  main() {
    const host = location.host as 'chat.mistral.ai' | 'le-chat.mistral.ai';
    installContentHandler(host);
  },
});
