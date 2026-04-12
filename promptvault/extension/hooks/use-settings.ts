import { useEffect, useState } from 'react';
import { getSettings, onSettingsChanged, type StoredSettings } from '../lib/storage';

export function useSettings(): StoredSettings | null {
  const [settings, setSettings] = useState<StoredSettings | null>(null);

  useEffect(() => {
    let mounted = true;
    void getSettings().then((s) => {
      if (mounted) setSettings(s);
    });
    const off = onSettingsChanged((s) => {
      if (mounted) setSettings(s);
    });
    return () => {
      mounted = false;
      off();
    };
  }, []);

  return settings;
}
