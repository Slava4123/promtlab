import { Skeleton } from './ui/skeleton';

// Общий skeleton для list-страниц (collections, tags, teams, chains, history).
// Соответствует разметке list-row: title + optional subtitle/badge. Используем
// вместо Loader2 spinner при initial loading, чтобы не было flash-of-empty
// state и пользователь видел структуру до получения данных.
export function ListSkeleton({
  count = 5,
  showSubtitle = true,
  showBadge = false,
}: {
  count?: number;
  showSubtitle?: boolean;
  showBadge?: boolean;
}) {
  return (
    <div
      className="flex flex-col gap-2"
      role="status"
      aria-live="polite"
      aria-label="Загружаем данные"
    >
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="flex items-center gap-3 rounded-md border border-(--color-border) bg-(--color-card) p-3"
        >
          <Skeleton className="h-8 w-8 shrink-0 rounded-md" />
          <div className="min-w-0 flex-1 space-y-1.5">
            <Skeleton className="h-4 w-3/5" />
            {showSubtitle ? <Skeleton className="h-3 w-2/5" /> : null}
          </div>
          {showBadge ? <Skeleton className="h-5 w-12 rounded-full" /> : null}
        </div>
      ))}
    </div>
  );
}

// Узкий skeleton для history-time-list (date-grouped rows). 32px высота,
// padding L = время | R = title.
export function RowSkeleton({ count = 8 }: { count?: number }) {
  return (
    <ul
      className="space-y-0.5"
      role="status"
      aria-live="polite"
      aria-label="Загружаем историю"
    >
      {Array.from({ length: count }).map((_, i) => (
        <li key={i} className="flex items-center gap-2 px-2 py-1.5">
          <Skeleton className="h-3 w-10 shrink-0" />
          <Skeleton className="h-3 w-3/4" />
        </li>
      ))}
    </ul>
  );
}
