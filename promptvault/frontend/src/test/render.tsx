// MN-14 — общий render helper для page tests.
//
// Зачем: каждая страница использует TanStack Query + react-router + auth-store.
// Без обёртки в каждом тесте — 30+ строк boilerplate. Здесь готовый
// `<TestProviders>` с в-тестовым QueryClient (no retry, no cache) и MemoryRouter.
//
// Use:
//
//   import { renderWithProviders } from "@/test/render"
//   renderWithProviders(<SignInPage />, { route: "/sign-in" })
import { type ReactNode } from "react"
import { render, type RenderOptions } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router-dom"

interface ProvidersProps {
  children: ReactNode
  route?: string
}

// eslint-disable-next-line react-refresh/only-export-components
function TestProviders({ children, route = "/" }: ProvidersProps) {
  // Свежий QueryClient на каждый тест — нет shared state между прогонами.
  // retry: false ускоряет фейл-кейсы (по умолчанию 3 retry × ms задержки).
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  return (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
    </QueryClientProvider>
  )
}

export function renderWithProviders(
  ui: ReactNode,
  { route, ...options }: { route?: string } & Omit<RenderOptions, "wrapper"> = {},
) {
  return render(ui, {
    wrapper: ({ children }) => <TestProviders route={route}>{children}</TestProviders>,
    ...options,
  })
}
