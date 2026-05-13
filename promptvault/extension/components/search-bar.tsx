import { Search, X } from 'lucide-react';
import { forwardRef, useEffect, useState, type Ref } from 'react';
import { Input } from './ui/input';
import { cn } from '../lib/utils';

interface Props {
  value: string;
  onChange: (v: string) => void;
  className?: string;
  inputRef?: Ref<HTMLInputElement>;
}

// Platform-aware shortcut. On Mac — ⌘K (Cmd+K), на Windows/Linux — Ctrl+K.
// useState + useEffect — `navigator` доступен только в браузере (SSR-safe
// если когда-нибудь будем делать pre-render).
function useShortcutHint(): string {
  const [hint, setHint] = useState('Ctrl+K');
  useEffect(() => {
    const isMac =
      typeof navigator !== 'undefined' &&
      /Mac|iPhone|iPad/i.test(navigator.platform || navigator.userAgent);
    setHint(isMac ? '⌘K' : 'Ctrl+K');
  }, []);
  return hint;
}

export const SearchBar = forwardRef<HTMLInputElement, Props>(function SearchBar(
  { value, onChange, className, inputRef },
  _ref,
) {
  const shortcut = useShortcutHint();
  return (
    <div className={cn('relative', className)}>
      <Search
        className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-(--color-muted-foreground)"
        aria-hidden
      />
      <Input
        ref={inputRef}
        type="search"
        role="searchbox"
        autoComplete="off"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={`Поиск (${shortcut})`}
        aria-label={`Поиск промптов (${shortcut})`}
        className="pl-8 pr-8"
      />
      {value ? (
        <button
          type="button"
          onClick={() => onChange('')}
          aria-label="Очистить поиск"
          className="absolute right-2 top-1/2 -translate-y-1/2 text-(--color-muted-foreground) hover:text-(--color-foreground)"
        >
          <X className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
});
