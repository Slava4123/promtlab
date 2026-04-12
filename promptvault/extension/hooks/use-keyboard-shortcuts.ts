import { useEffect } from 'react';

export interface Shortcut {
  /** key ровно как в event.key, чувствительно к регистру */
  key: string;
  /** требует Ctrl или Cmd (cross-platform) */
  ctrlOrCmd?: boolean;
  /** требует Shift */
  shift?: boolean;
  /** callback, вызывается с preventDefault если handler вернул true или undefined */
  handler: (e: KeyboardEvent) => boolean | void;
}

export function useKeyboardShortcuts(shortcuts: Shortcut[], enabled = true): void {
  useEffect(() => {
    if (!enabled) return;

    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      // Игнорируем события из input/textarea если key — печатный символ
      // (чтобы не перехватывать обычное набирание текста)
      const inInput =
        target?.tagName === 'INPUT' ||
        target?.tagName === 'TEXTAREA' ||
        target?.isContentEditable;

      for (const sc of shortcuts) {
        const keyMatch = e.key === sc.key;
        const ctrlMatch = sc.ctrlOrCmd ? e.ctrlKey || e.metaKey : !e.ctrlKey && !e.metaKey;
        const shiftMatch = sc.shift ? e.shiftKey : !e.shiftKey;

        if (!keyMatch || !ctrlMatch || !shiftMatch) continue;

        // Если пользователь пишет в инпут — разрешаем только модификаторные shortcut'ы
        if (inInput && !sc.ctrlOrCmd && sc.key.length === 1) continue;

        const result = sc.handler(e);
        if (result !== false) {
          e.preventDefault();
          e.stopPropagation();
          return;
        }
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [shortcuts, enabled]);
}
