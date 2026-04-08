import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api } from "@/api/client"
import { useAuthStore } from "@/stores/auth-store"
import type {
  StarterCatalog,
  CompleteOnboardingRequest,
  CompleteOnboardingResponse,
  User,
} from "@/api/types"

/**
 * useStarterCatalog — загрузка каталога starter промптов из embedded JSON.
 *
 * Кэшируется на 5 минут. Каталог встроен в бинарник и обновляется только
 * при деплое — но 5 минут защищают от ситуации, когда юзер открыл /welcome
 * до деплоя и держит таб открытым: после ре-фетча получит свежие template_id
 * и не словит 400 ErrUnknownTemplate на install.
 */
export function useStarterCatalog() {
  return useQuery({
    queryKey: ["starter", "catalog"],
    queryFn: () => api<StarterCatalog>("/starter/catalog"),
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * useCompleteOnboarding — финиш wizard'а: создаёт промпты в БД юзера +
 * маркирует онбординг пройденным.
 *
 * onSuccess:
 *  - инвалидирует ["prompts"] — чтобы dashboard сразу показал новые промпты
 *  - обновляет user в auth-store локально (без лишнего fetchMe), чтобы
 *    ProtectedRoute мгновенно перестал редиректить на /welcome
 */
export function useCompleteOnboarding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CompleteOnboardingRequest) =>
      api<CompleteOnboardingResponse>("/starter/complete", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: (resp) => {
      qc.invalidateQueries({ queryKey: ["prompts"] })
      qc.invalidateQueries({ queryKey: ["collections"] })

      // Optimistic update авторизованного юзера: сразу прописываем
      // onboarding_completed_at, чтобы ProtectedRoute не послал на /welcome.
      const store = useAuthStore.getState()
      if (store.user) {
        const updated: User = {
          ...store.user,
          onboarding_completed_at: resp.onboarding_completed_at,
        }
        useAuthStore.setState({ user: updated })
      }
    },
  })
}
