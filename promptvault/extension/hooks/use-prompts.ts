import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { sendBg } from '../lib/bg-client';
import type { PromptFilterMessage } from '../lib/messages';
import type { PaginatedPrompts, Prompt } from '../lib/types';

const PAGE_SIZE = 50;

export function useInfinitePromptList(enabled: boolean, filter: PromptFilterMessage) {
  return useInfiniteQuery({
    queryKey: ['prompts', 'list', filter],
    enabled,
    initialPageParam: 1,
    queryFn: ({ pageParam }) =>
      sendBg({
        type: 'api.fetchPrompts',
        page: pageParam,
        pageSize: PAGE_SIZE,
        filter,
      }),
    getNextPageParam: (lastPage: PaginatedPrompts, allPages) => {
      if (!lastPage.has_more) return undefined;
      if (lastPage.page) return lastPage.page + 1;
      return allPages.length + 1;
    },
    staleTime: 60_000,
  });
}

export function useSearch(q: string, enabled: boolean, filter: PromptFilterMessage) {
  return useQuery({
    queryKey: ['prompts', 'search', q, filter],
    queryFn: () => sendBg({ type: 'api.searchPrompts', q, filter }),
    enabled: enabled && q.trim().length > 0,
    staleTime: 30_000,
  });
}

export function usePinned(enabled: boolean, filter: PromptFilterMessage) {
  return useQuery<Prompt[]>({
    queryKey: ['prompts', 'pinned', filter],
    queryFn: () => sendBg({ type: 'api.getPinned', limit: 10, filter }),
    enabled,
    staleTime: 60_000,
  });
}

export function useRecent(enabled: boolean, filter: PromptFilterMessage) {
  return useQuery<Prompt[]>({
    queryKey: ['prompts', 'recent', filter],
    queryFn: () => sendBg({ type: 'api.getRecent', limit: 10, filter }),
    enabled,
    staleTime: 60_000,
  });
}

export function usePrompt(id: number | null) {
  return useQuery<Prompt>({
    queryKey: ['prompts', 'one', id],
    queryFn: () => sendBg({ type: 'api.getPrompt', id: id! }),
    enabled: id !== null,
  });
}
