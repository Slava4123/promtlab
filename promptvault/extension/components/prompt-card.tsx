import { useState, type ReactNode } from 'react';
import { Pin, Star } from 'lucide-react';
import type { Prompt } from '../lib/types';
import { cn } from '../lib/utils';
import { useToggleFavorite, useTogglePin } from '../hooks/use-mutations';

// Подсветка {{var}} в preview карточки. Тот же regex что в shared/template/parse
// (Unicode-aware, identifier = letter+digits+underscore). Возвращает React-узлы,
// чтобы избежать dangerouslySetInnerHTML.
const VAR_REGEX = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu;

function highlightVariables(text: string): ReactNode[] {
  const parts: ReactNode[] = [];
  let lastIndex = 0;
  for (const match of text.matchAll(VAR_REGEX)) {
    const idx = match.index!;
    if (idx > lastIndex) parts.push(text.slice(lastIndex, idx));
    parts.push(
      <span
        key={`${idx}-${match[1]}`}
        className="rounded-sm bg-(--color-brand-muted) px-1 font-mono text-[10px] text-(--color-brand-muted-foreground)"
      >
        {match[0]}
      </span>,
    );
    lastIndex = idx + match[0].length;
  }
  if (lastIndex < text.length) parts.push(text.slice(lastIndex));
  return parts;
}

interface Props {
  prompt: Prompt;
  onClick: () => void;
  highlighted?: boolean;
  focused?: boolean;
}

export function PromptCard({ prompt, onClick, highlighted, focused }: Props) {
  const toggleFav = useToggleFavorite();
  const togglePin = useTogglePin();
  const [hoverPreview, setHoverPreview] = useState(false);

  const preview =
    prompt.content.length > 120 ? prompt.content.slice(0, 120) + '…' : prompt.content;

  const isPinned = prompt.pinned_personal || prompt.pinned_team;

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onClick();
        }
      }}
      onMouseEnter={() => setHoverPreview(true)}
      onMouseLeave={() => setHoverPreview(false)}
      data-prompt-id={prompt.id}
      className={cn(
        'group cv-auto relative flex w-full cursor-pointer flex-col items-start gap-1 rounded-md border bg-(--color-card) p-3 text-left transition-colors duration-150',
        focused
          ? 'border-(--color-brand) ring-2 ring-(--color-brand)/35'
          : 'border-(--color-border)',
        highlighted
          ? 'border-(--color-success) bg-(--color-brand-muted) ring-2 ring-(--color-brand)/25'
          : 'hover:border-(--color-brand)/45 hover:bg-(--color-brand-muted)',
      )}
    >
      <div className="flex w-full items-center gap-2">
        <span className="flex-1 truncate text-sm font-medium text-(--color-card-foreground)">
          {prompt.title}
        </span>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            togglePin.mutate(prompt.id);
          }}
          aria-label={isPinned ? 'Открепить' : 'Закрепить'}
          className={cn(
            'shrink-0 rounded p-0.5 transition-colors',
            isPinned
              ? // Когда оба маркера активны — pin приглушаем (звезда яркая amber
                // достаточно акцентируется, два rich-цвета рядом перегружают).
                prompt.favorite
                ? 'text-(--color-brand)/65 hover:text-(--color-brand)'
                : 'text-(--color-brand)'
              : 'text-(--color-muted-foreground) opacity-0 group-hover:opacity-100 hover:text-(--color-foreground)',
          )}
        >
          <Pin
            className={cn('h-3.5 w-3.5', isPinned && 'fill-current')}
            aria-hidden
          />
        </button>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            toggleFav.mutate(prompt.id);
          }}
          aria-label={prompt.favorite ? 'Убрать из избранного' : 'В избранное'}
          className={cn(
            'shrink-0 rounded p-0.5 transition-colors',
            prompt.favorite
              ? 'text-amber-500'
              : 'text-(--color-muted-foreground) opacity-0 group-hover:opacity-100 hover:text-amber-500',
          )}
        >
          <Star className={cn('h-3.5 w-3.5', prompt.favorite && 'fill-current')} />
        </button>
      </div>
      <p className="line-clamp-2 text-xs text-(--color-muted-foreground)">
        {highlightVariables(preview)}
      </p>
      {prompt.tags.length > 0 ? (
        <div className="mt-1 flex flex-wrap gap-1">
          {prompt.tags.slice(0, 4).map((t) => (
            <span
              key={t.id}
              className="rounded-sm border px-1.5 py-0.5 text-[10px]"
              style={{
                backgroundColor: t.color ? `${t.color}22` : 'var(--color-secondary)',
                borderColor: t.color ? `${t.color}55` : 'transparent',
                color: t.color || 'var(--color-secondary-foreground)',
              }}
            >
              {t.name}
            </span>
          ))}
          {prompt.tags.length > 4 ? (
            <span className="text-[10px] text-(--color-muted-foreground)">
              +{prompt.tags.length - 4}
            </span>
          ) : null}
        </div>
      ) : null}

      {/* Hover preview — полный content при задержке hover 500ms */}
      {hoverPreview && prompt.content.length > 120 ? (
        <div
          className="pointer-events-none absolute left-0 right-0 top-full z-20 mt-1 max-h-56 overflow-y-auto rounded-md border border-(--color-border) bg-(--color-card) p-3 text-xs shadow-lg opacity-0 transition-opacity duration-150 group-hover:opacity-100"
          style={{ transitionDelay: '400ms' }}
        >
          <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
            Полный текст
          </div>
          <div className="whitespace-pre-wrap text-(--color-foreground)">{prompt.content}</div>
        </div>
      ) : null}
    </div>
  );
}
