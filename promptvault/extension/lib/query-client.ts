// Singleton QueryClient — выделен из components/app.tsx чтобы быть доступным
// из любого hook/store без передачи через props.

import { QueryClient } from "@tanstack/react-query"
import { ApiError } from "./types"

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (
          error instanceof ApiError &&
          error.status >= 400 &&
          error.status < 500 &&
          error.status !== 0
        ) {
          return false
        }
        return failureCount < 2
      },
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 3000),
      refetchOnWindowFocus: true,
      staleTime: 30_000,
    },
  },
})
