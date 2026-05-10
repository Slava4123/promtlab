// query-client.ts — singleton QueryClient, чтобы auth-store мог его очистить
// при logout (MJ-9 в REVIEW_2026-05-07.md). До этого fix'а на shared device
// после logout'а юзера A в TanStack Query cache оставались его данные —
// юзер B видел их до полного reload страницы (data leak between users).
//
// Раньше queryClient создавался прямо в App.tsx — создавал circular import
// если auth-store попытается его импортировать.
import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query"
import { captureException } from "@sentry/react"

import { ApiError } from "@/api/client"

// Только 5xx (ApiError.status >= 500) или non-API (network) ошибки шлём в Sentry —
// 4xx это user-faced (validation, auth, quota), они не индикатор сервер-баги.
function shouldReport(error: unknown): boolean {
  if (error instanceof ApiError) {
    return error.status >= 500
  }
  return true
}

export const queryClient = new QueryClient({
  // Captures query errors в Sentry на уровне cache — ловит все failed queries
  // централизованно, без необходимости добавлять обработчики в каждый хук.
  queryCache: new QueryCache({
    onError: (error, query) => {
      if (!shouldReport(error)) return
      captureException(error, {
        tags: {
          query_key: JSON.stringify(query.queryKey),
          source: "tanstack_query",
        },
      })
    },
  }),
  // MN-57: до фикса failed mutations требовали explicit onError в каждом хуке,
  // иначе 5xx терялся (toast.error показывался, но Sentry не получал stacktrace).
  // Теперь централизованный capture аналогично queryCache.
  mutationCache: new MutationCache({
    onError: (error, _variables, _context, mutation) => {
      if (!shouldReport(error)) return
      captureException(error, {
        tags: {
          mutation_key: mutation.options.mutationKey
            ? JSON.stringify(mutation.options.mutationKey)
            : "unknown",
          source: "tanstack_mutation",
        },
      })
    },
  }),
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 5 * 60 * 1000,
    },
  },
})
