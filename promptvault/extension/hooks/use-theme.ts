import { useEffect } from 'react';
import type { Theme } from '../lib/storage';

/**
 * Применяет выбранную тему к <html> элементу через `dark` класс (Tailwind dark: variant).
 */
export function useApplyTheme(theme: Theme | null): void {
  useEffect(() => {
    if (!theme) return undefined;
    const root = document.documentElement;

    const apply = (t: Theme) => {
      let dark: boolean;
      if (t === 'system') {
        dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      } else {
        dark = t === 'dark';
      }
      root.classList.toggle('dark', dark);
    };

    apply(theme);

    if (theme === 'system') {
      const mq = window.matchMedia('(prefers-color-scheme: dark)');
      const listener = () => apply('system');
      mq.addEventListener('change', listener);
      return () => mq.removeEventListener('change', listener);
    }
    return undefined;
  }, [theme]);
}
