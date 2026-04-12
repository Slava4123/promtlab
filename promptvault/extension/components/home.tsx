import { useEffect, useMemo, useRef, useState } from 'react';
import { RefreshCw, Settings as SettingsIcon } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { Button } from './ui/button';
import { SearchBar } from './search-bar';
import { PromptList } from './prompt-list';
import { PromptListSkeleton } from './prompt-list-skeleton';
import { EmptyState } from './empty-state';
import { ActiveTabBadge } from './active-tab-badge';
import { WorkspaceSelector } from './workspace-selector';
import { CollectionSelector } from './collection-selector';
import { StreakBadge } from './streak-badge';
import {
  useInfinitePromptList,
  usePinned,
  useRecent,
  useSearch,
} from '../hooks/use-prompts';
import { useDebounced } from '../hooks/use-debounced';
import { useActiveTab } from '../hooks/use-active-tab';
import { useKeyboardShortcuts } from '../hooks/use-keyboard-shortcuts';
import { useWorkspace } from '../hooks/use-workspace';
import type { PromptFilterMessage } from '../lib/messages';
import type { Prompt } from '../lib/types';

type Tab = 'all' | 'pinned' | 'recent' | 'favorites';

interface Props {
  onSelect: (p: Prompt) => void;
  onOpenSettings: () => void;
  highlightedId?: number | null;
}

export function Home({ onSelect, onOpenSettings, highlightedId }: Props) {
  const [query, setQuery] = useState('');
  const [tab, setTab] = useState<Tab>('all');
  const [focusIdx, setFocusIdx] = useState(0);
  const debouncedQuery = useDebounced(query, 300);
  const searching = debouncedQuery.trim().length > 0;
  const searchInputRef = useRef<HTMLInputElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();

  const activeTab = useActiveTab();
  const workspace = useWorkspace();

  const filter: PromptFilterMessage = useMemo(
    () => ({
      teamId: workspace.workspaceId,
      collectionId: workspace.collectionId,
    }),
    [workspace.workspaceId, workspace.collectionId],
  );

  const list = useInfinitePromptList(!searching, filter);
  const searchResult = useSearch(debouncedQuery, searching, filter);
  const pinned = usePinned(!searching, filter);
  const recent = useRecent(!searching, filter);

  const allPrompts: Prompt[] = useMemo(() => {
    if (searching) {
      return (searchResult.data?.prompts ?? []).map((r) => ({
        id: r.id,
        title: r.title,
        content: r.description,
        favorite: false,
        pinned_personal: false,
        pinned_team: false,
        usage_count: 0,
        tags: [],
        collections: [],
        created_at: '',
        updated_at: '',
      }));
    }
    return list.data?.pages.flatMap((p) => p.items) ?? [];
  }, [searching, searchResult.data, list.data]);

  const visiblePrompts = useMemo(() => {
    if (searching) return allPrompts;
    switch (tab) {
      case 'pinned':
        return pinned.data ?? [];
      case 'recent':
        return recent.data ?? [];
      case 'favorites':
        return allPrompts.filter((p) => p.favorite);
      default:
        return allPrompts;
    }
  }, [searching, tab, allPrompts, pinned.data, recent.data]);

  const isLoading = searching ? searchResult.isPending : list.isPending;
  const showInitialSkeleton = isLoading && visiblePrompts.length === 0;

  useEffect(() => {
    if (searching) return;
    const el = scrollRef.current;
    if (!el) return;
    const onScroll = () => {
      if (list.hasNextPage && !list.isFetchingNextPage) {
        const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 80;
        if (nearBottom) void list.fetchNextPage();
      }
    };
    el.addEventListener('scroll', onScroll);
    return () => el.removeEventListener('scroll', onScroll);
  }, [searching, list.hasNextPage, list.isFetchingNextPage, list.fetchNextPage, list]);

  useEffect(() => {
    setFocusIdx(0);
  }, [visiblePrompts.length, tab, searching]);

  useKeyboardShortcuts([
    {
      key: 'k',
      ctrlOrCmd: true,
      handler: () => {
        searchInputRef.current?.focus();
      },
    },
    {
      key: 'ArrowDown',
      handler: () => {
        setFocusIdx((i) => Math.min(i + 1, visiblePrompts.length - 1));
      },
    },
    {
      key: 'ArrowUp',
      handler: () => {
        setFocusIdx((i) => Math.max(i - 1, 0));
      },
    },
    {
      key: 'Enter',
      handler: () => {
        const p = visiblePrompts[focusIdx];
        if (p) {
          onSelect(p);
          return true;
        }
        return false;
      },
    },
    {
      key: 'Escape',
      handler: () => {
        if (query) {
          setQuery('');
          return true;
        }
        return false;
      },
    },
    {
      key: 'r',
      ctrlOrCmd: true,
      handler: () => {
        void queryClient.invalidateQueries({ queryKey: ['prompts'] });
      },
    },
  ]);

  const focusedId = visiblePrompts[focusIdx]?.id ?? null;

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex items-center gap-2 border-b border-(--color-border) p-3">
        <WorkspaceSelector
          workspaceId={workspace.workspaceId}
          teams={workspace.teams}
          onChange={workspace.setWorkspaceId}
        />
        <ActiveTabBadge state={activeTab} />
        <StreakBadge />
        <div className="flex-1" />
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => {
            void queryClient.invalidateQueries({ queryKey: ['prompts'] });
          }}
          aria-label="Обновить"
          title="Обновить (⌘R)"
        >
          <RefreshCw className="h-4 w-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={onOpenSettings}
          aria-label="Настройки"
          title="Настройки"
        >
          <SettingsIcon className="h-4 w-4" />
        </Button>
      </div>

      {/* Search + Collection selector */}
      <div className="space-y-2 border-b border-(--color-border) p-3">
        <SearchBar value={query} onChange={setQuery} inputRef={searchInputRef} />
        {workspace.collections.length > 0 ? (
          <CollectionSelector
            collectionId={workspace.collectionId}
            collections={workspace.collections}
            onChange={workspace.setCollectionId}
          />
        ) : null}
      </div>

      {/* Tabs */}
      {!searching ? (
        <div className="flex gap-1 border-b border-(--color-border) px-3 pt-2">
          <TabButton active={tab === 'all'} onClick={() => setTab('all')}>
            Все
          </TabButton>
          <TabButton active={tab === 'pinned'} onClick={() => setTab('pinned')}>
            Закреплённые
          </TabButton>
          <TabButton active={tab === 'recent'} onClick={() => setTab('recent')}>
            Недавние
          </TabButton>
          <TabButton active={tab === 'favorites'} onClick={() => setTab('favorites')}>
            ⭐ Избранное
          </TabButton>
        </div>
      ) : null}

      {/* Content */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-3">
        {showInitialSkeleton ? (
          <PromptListSkeleton />
        ) : visiblePrompts.length === 0 ? (
          <EmptyState
            title={searching ? 'Ничего не найдено' : emptyMessageForTab(tab)}
            description={
              searching ? 'Попробуйте другой запрос' : emptyDescriptionForTab(tab)
            }
            showOpenWebLink={!searching && tab === 'all' && allPrompts.length === 0}
          />
        ) : (
          <>
            <PromptList
              prompts={visiblePrompts}
              onSelect={onSelect}
              highlightedId={highlightedId}
              focusedId={focusedId}
            />
            {!searching && list.isFetchingNextPage ? (
              <div className="mt-3">
                <PromptListSkeleton count={2} />
              </div>
            ) : null}
          </>
        )}
      </div>
    </div>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={
        'relative px-2.5 pb-2 text-xs font-medium transition-colors ' +
        (active
          ? 'text-(--color-foreground)'
          : 'text-(--color-muted-foreground) hover:text-(--color-foreground)')
      }
    >
      {children}
      {active ? (
        <span className="absolute bottom-0 left-0 right-0 h-0.5 rounded-t bg-(--color-primary)" />
      ) : null}
    </button>
  );
}

function emptyMessageForTab(tab: Tab): string {
  switch (tab) {
    case 'pinned':
      return 'Нет закреплённых промптов';
    case 'recent':
      return 'Нет недавно использованных';
    case 'favorites':
      return 'Нет избранных промптов';
    default:
      return 'Нет промптов';
  }
}

function emptyDescriptionForTab(tab: Tab): string {
  switch (tab) {
    case 'pinned':
      return 'Закрепите промпт через иконку 📌 в списке';
    case 'recent':
      return 'Вставьте промпт чтобы он появился здесь';
    case 'favorites':
      return 'Отметьте промпт звёздочкой ⭐ в списке';
    default:
      return 'Создайте первый промпт в веб-интерфейсе ПромтЛаба';
  }
}
