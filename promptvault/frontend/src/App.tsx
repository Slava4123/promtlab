import { lazy, Suspense, useEffect } from "react"
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { QueryClient, QueryCache, QueryClientProvider } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { Loader2 } from "lucide-react"

import { ErrorBoundary } from "@/components/error-boundary"
import { useAuthStore } from "@/stores/auth-store"
import ProtectedRoute from "@/components/auth/protected-route"
import AppLayout from "@/components/layout/app-layout"
import { ApiError } from "@/api/client"
import { captureException } from "@/lib/sentry"
import { captureReferralFromURL } from "@/lib/referral"

// Eager-loaded (public, lightweight)
import SignIn from "@/pages/sign-in"
import SignUp from "@/pages/sign-up"
import OAuthCallback from "@/pages/oauth-callback"
import VerifyEmail from "@/pages/verify-email"
import ForgotPassword from "@/pages/forgot-password"
import SharedPrompt from "@/pages/shared-prompt"
import PublicPrompt from "@/pages/public-prompt"

// P-13: Landing — lazy, он не на hot-path для logged-in юзеров, а initial
// bundle ценнее держать лёгким (первая загрузка чаще /sign-in или /dashboard).
const Landing = lazy(() => import("@/pages/landing"))

// Lazy-loaded (protected, heavier)
const Dashboard = lazy(() => import("@/pages/dashboard"))
const PromptEditor = lazy(() => import("@/pages/prompt-editor"))
const Collections = lazy(() => import("@/pages/collections"))
const CollectionView = lazy(() => import("@/pages/collection-view"))
const Versions = lazy(() => import("@/pages/versions"))
const SettingsPage = lazy(() => import("@/pages/settings"))
const Teams = lazy(() => import("@/pages/teams"))
const TeamView = lazy(() => import("@/pages/team-view"))
const Pricing = lazy(() => import("@/pages/pricing"))
const Trash = lazy(() => import("@/pages/trash"))
const History = lazy(() => import("@/pages/history"))
const Welcome = lazy(() => import("@/pages/welcome"))
const Changelog = lazy(() => import("@/pages/changelog"))
const Badges = lazy(() => import("@/pages/badges"))

// Admin pages
const AdminLayout = lazy(() => import("@/pages/admin/layout"))
const AdminUsers = lazy(() => import("@/pages/admin/users"))
const AdminUserDetail = lazy(() => import("@/pages/admin/user-detail"))
const AdminAuditLog = lazy(() => import("@/pages/admin/audit-log"))
const AdminHealth = lazy(() => import("@/pages/admin/health"))
const AdminTOTPEnroll = lazy(() => import("@/pages/admin/totp-enroll"))
const ExtensionPrivacy = lazy(() => import("@/pages/legal/extension-privacy"))
const Terms = lazy(() => import("@/pages/legal/terms"))
const Privacy = lazy(() => import("@/pages/legal/privacy"))
const Offer = lazy(() => import("@/pages/legal/offer"))

function PageFallback() {
  return (
    <div className="flex h-[60vh] items-center justify-center">
      <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
    </div>
  )
}

const queryClient = new QueryClient({
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

function AppRoutes() {
  const restoreSession = useAuthStore((s) => s.restoreSession)

  useEffect(() => {
    captureReferralFromURL()
    restoreSession()
  }, [restoreSession])

  return (
    <Routes>
      {/* public */}
      <Route path="/" element={<Suspense fallback={<PageFallback />}><Landing /></Suspense>} />
      <Route path="/sign-in" element={<SignIn />} />
      <Route path="/sign-up" element={<SignUp />} />
      <Route path="/oauth/callback" element={<OAuthCallback />} />
      <Route path="/verify-email" element={<VerifyEmail />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/s/:token" element={<SharedPrompt />} />
      <Route path="/p/:slug" element={<PublicPrompt />} />
      <Route path="/legal/extension-privacy" element={<Suspense fallback={<PageFallback />}><ExtensionPrivacy /></Suspense>} />
      <Route path="/legal/terms" element={<Suspense fallback={<PageFallback />}><Terms /></Suspense>} />
      <Route path="/legal/privacy" element={<Suspense fallback={<PageFallback />}><Privacy /></Suspense>} />
      <Route path="/legal/offer" element={<Suspense fallback={<PageFallback />}><Offer /></Suspense>} />

      {/* protected — with layout */}
      <Route element={<ProtectedRoute />}>
        {/* Onboarding wizard — full-screen, без AppLayout */}
        <Route path="/welcome" element={<Suspense fallback={<PageFallback />}><Welcome /></Suspense>} />

        <Route element={<AppLayout />}>
          <Route path="/dashboard" element={<Suspense fallback={<PageFallback />}><Dashboard /></Suspense>} />
          <Route path="/prompts/new" element={<Suspense fallback={<PageFallback />}><PromptEditor /></Suspense>} />
          <Route path="/prompts/:id" element={<Suspense fallback={<PageFallback />}><PromptEditor /></Suspense>} />
          <Route path="/prompts/:id/versions" element={<Suspense fallback={<PageFallback />}><Versions /></Suspense>} />
          <Route path="/collections" element={<Suspense fallback={<PageFallback />}><Collections /></Suspense>} />
          <Route path="/collections/:id" element={<Suspense fallback={<PageFallback />}><CollectionView /></Suspense>} />
          <Route path="/teams" element={<Suspense fallback={<PageFallback />}><Teams /></Suspense>} />
          <Route path="/teams/:slug" element={<Suspense fallback={<PageFallback />}><TeamView /></Suspense>} />
          <Route path="/settings" element={<Suspense fallback={<PageFallback />}><SettingsPage /></Suspense>} />
          <Route path="/history" element={<Suspense fallback={<PageFallback />}><History /></Suspense>} />
          <Route path="/trash" element={<Suspense fallback={<PageFallback />}><Trash /></Suspense>} />
          <Route path="/pricing" element={<Suspense fallback={<PageFallback />}><Pricing /></Suspense>} />
          <Route path="/changelog" element={<Suspense fallback={<PageFallback />}><Changelog /></Suspense>} />
          <Route path="/badges" element={<Suspense fallback={<PageFallback />}><Badges /></Suspense>} />

          {/* Admin routes — гвардятся useAdminGuard внутри AdminLayout */}
          <Route path="/admin" element={<Suspense fallback={<PageFallback />}><AdminLayout /></Suspense>}>
            <Route index element={<Suspense fallback={<PageFallback />}><AdminUsers /></Suspense>} />
            <Route path="users" element={<Suspense fallback={<PageFallback />}><AdminUsers /></Suspense>} />
            <Route path="users/:id" element={<Suspense fallback={<PageFallback />}><AdminUserDetail /></Suspense>} />
            <Route path="audit" element={<Suspense fallback={<PageFallback />}><AdminAuditLog /></Suspense>} />
            <Route path="health" element={<Suspense fallback={<PageFallback />}><AdminHealth /></Suspense>} />
            <Route path="totp" element={<Suspense fallback={<PageFallback />}><AdminTOTPEnroll /></Suspense>} />
          </Route>
        </Route>
      </Route>
    </Routes>
  )
}

export default function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
