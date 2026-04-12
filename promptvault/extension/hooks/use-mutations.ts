import { useMutation, useQueryClient } from '@tanstack/react-query';
import { sendBg } from '../lib/bg-client';

/**
 * Favorite/Pin mutations — после success invalidate-им все `['prompts']` queries,
 * TanStack Query сам перезапросит актуальные данные (Home list, pinned, recent, one).
 * Оптимистичного update нет — это надёжнее, учитывая что:
 *   - pin endpoint возвращает `{pinned, team_wide}`, а не полный Prompt
 *   - главный список использует useInfiniteQuery с формой `{pages: [...]}`, которую
 *     сложно patch-ить вручную без багов
 * Задержка ~100-200ms незаметна при таких операциях.
 */

export function useToggleFavorite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (promptId: number) => sendBg({ type: 'api.toggleFavorite', promptId }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['prompts'] });
    },
  });
}

export function useTogglePin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (promptId: number) => sendBg({ type: 'api.togglePin', promptId }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['prompts'] });
    },
  });
}
