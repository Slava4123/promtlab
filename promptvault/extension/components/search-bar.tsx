import { Search, X } from 'lucide-react';
import { forwardRef, type Ref } from 'react';
import { Input } from './ui/input';
import { cn } from '../lib/utils';

interface Props {
  value: string;
  onChange: (v: string) => void;
  className?: string;
  inputRef?: Ref<HTMLInputElement>;
}

export const SearchBar = forwardRef<HTMLInputElement, Props>(function SearchBar(
  { value, onChange, className, inputRef },
  _ref,
) {
  return (
    <div className={cn('relative', className)}>
      <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-(--color-muted-foreground)" />
      <Input
        ref={inputRef}
        type="search"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Поиск (⌘K)"
        className="pl-8 pr-8"
      />
      {value ? (
        <button
          type="button"
          onClick={() => onChange('')}
          aria-label="Очистить"
          className="absolute right-2 top-1/2 -translate-y-1/2 text-(--color-muted-foreground) hover:text-(--color-foreground)"
        >
          <X className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
});
