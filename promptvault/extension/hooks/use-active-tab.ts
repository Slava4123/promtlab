import { useEffect, useState } from 'react';
import { sendBg } from '../lib/bg-client';
import { isSupportedHost } from '../lib/messages';

export interface ActiveTabState {
  host: string | null;
  supported: boolean;
}

/**
 * Возвращает информацию об активной вкладке (host и поддерживается ли extension'ом).
 * Обновляется на событиях chrome.tabs.onActivated и onUpdated.
 */
export function useActiveTab(): ActiveTabState {
  const [state, setState] = useState<ActiveTabState>({ host: null, supported: false });

  useEffect(() => {
    let cancelled = false;

    const refresh = async () => {
      try {
        const data = await sendBg({ type: 'cmd.getActiveHost' });
        if (!cancelled) {
          setState({
            host: data.host,
            supported: isSupportedHost(data.host),
          });
        }
      } catch {
        if (!cancelled) setState({ host: null, supported: false });
      }
    };

    void refresh();

    const onActivated = () => void refresh();
    const onUpdated = (
      _tabId: number,
      changeInfo: chrome.tabs.TabChangeInfo,
      tab: chrome.tabs.Tab,
    ) => {
      if (tab.active && (changeInfo.url || changeInfo.status === 'complete')) {
        void refresh();
      }
    };

    chrome.tabs.onActivated.addListener(onActivated);
    chrome.tabs.onUpdated.addListener(onUpdated);

    return () => {
      cancelled = true;
      chrome.tabs.onActivated.removeListener(onActivated);
      chrome.tabs.onUpdated.removeListener(onUpdated);
    };
  }, []);

  return state;
}
