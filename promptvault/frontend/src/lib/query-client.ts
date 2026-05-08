// query-client.ts — singleton QueryClient, чтобы auth-store мог его очистить
// при logout (MJ-9 в REVIEW_2026-05-07.md). До этого fix'а на shared device
// после logout'а юзера A в TanStack Query cache оставались его данные —
// юзер B видел их до полного reload страницы (data leak between users).
//
// Раньше queryClient создавался прямо в App.tsx — создавал circular import
// если auth-store попытается его импортировать.
import { QueryCache, QueryClient } from "@tanstack/react-query"
import { captureException } from "@sentry/react"

import { ApiError } from "@/api/client"

export const queryClient = new QueryClient({
  // Captures query errors в Sentry на уровне cache — ловит все failed queries
  // централизованно, без необходимости добавлять обработчики в каждый хук.
  // Только 5xx (ApiError) + non-ApiError (network errors) отправляются,
  // 4xx пропускаются как expected user errors.
  queryCache: new QueryCache({
    onError: (error, query) => {
      const isApiError = error instanceof ApiError
      if (isApiError && error.status < 500) {
        return
      }
      captureException(error, {
        tags: {
          query_key: JSON.stringify(query.queryKey),
          source: "tanstack_query",
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
