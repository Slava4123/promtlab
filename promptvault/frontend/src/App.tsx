import { lazy, Suspense, useEffect } from "react"
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { QueryClientProvider } from "@tanstack/react-query"
import { Loader2 } from "lucide-react"

// queryClient — singleton, экспортирован из lib/query-client.ts чтобы
// auth-store мог его очистить при logout (MJ-9 data leak fix).
import { queryClient } from "@/lib/query-client"

// MJ-18: ReactQueryDevtools должен быть подключён ТОЛЬКО в dev (~37-40 KB
// gzip иначе попадает в production bundle бесполезным грузом). React.lazy
// + import.meta.env.DEV gating вырезается tree-shaker'ом из prod build'а.
const ReactQueryDevtools = import.meta.env.DEV
  ? lazy(() =>
      import("@tanstack/react-query-devtools").then((m) => ({
        default: m.ReactQueryDevtools,
      })),
    )
  : () => null

import { ErrorBoundary } from "@/components/error-boundary"
import { useAuthStore } from "@/stores/auth-store"
import ProtectedRoute from "@/components/auth/protected-route"
import AppLayout from "@/components/layout/app-layout"
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
const Chains = lazy(() => import("@/pages/chains"))
const ChainEditor = lazy(() => import("@/pages/chains/editor"))
const ChainRun = lazy(() => import("@/pages/chains/run"))
const ChainRuns = lazy(() => import("@/pages/chains/runs"))
const ChainCanvas = lazy(() => import("@/pages/chains/canvas"))
const Versions = lazy(() => import("@/pages/versions"))
const PromptAnalytics = lazy(() => import("@/pages/prompt-analytics"))
// /settings/* — nested routes. Layout — lazy (грузится один раз при заходе),
// sub-страницы — eager: формы лёгкие, per-section split дал бы 8 микро-чанков
// и мерцание Suspense fallback при каждом переключении nav.
const SettingsLayout = lazy(() => import("@/pages/settings/layout"))
import SettingsProfile from "@/pages/settings/profile"
import SettingsSecurity from "@/pages/settings/security"
import SettingsAccounts from "@/pages/settings/accounts"
import SettingsNotifications from "@/pages/settings/notifications"
import SettingsSubscription from "@/pages/settings/subscription"
import SettingsReferral from "@/pages/settings/referral"
import SettingsIntegrations from "@/pages/settings/integrations"
import SettingsAppearance from "@/pages/settings/appearance"
const Teams = lazy(() => import("@/pages/teams"))
const TeamView = lazy(() => import("@/pages/team-view"))
const Pricing = lazy(() => import("@/pages/pricing"))
const Analytics = lazy(() => import("@/pages/analytics"))
const TeamAnalytics = lazy(() => import("@/pages/team-analytics"))
const TeamActivity = lazy(() => import("@/pages/team-activity"))
const TeamBranding = lazy(() => import("@/pages/team-branding"))
const Trash = lazy(() => import("@/pages/trash"))
const History = lazy(() => import("@/pages/history"))
const Welcome = lazy(() => import("@/pages/welcome"))
const Changelog = lazy(() => import("@/pages/changelog"))
const Badges = lazy(() => import("@/pages/badges"))
const Help = lazy(() => import("@/pages/help"))
const HelpMCP = lazy(() => import("@/pages/help/mcp"))

// Admin pages
const AdminLayout = lazy(() => import("@/pages/admin/layout"))
const AdminUsers = lazy(() => import("@/pages/admin/users"))
const AdminUserDetail = lazy(() => import("@/pages/admin/user-detail"))
const AdminAuditLog = lazy(() => import("@/pages/admin/audit-log"))
const AdminHealth = lazy(() => import("@/pages/admin/health"))
const AdminTOTPEnroll = lazy(() => import("@/pages/admin/totp-enroll"))
const AdminFeedbacks = lazy(() => import("@/pages/admin/feedbacks"))
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

// queryClient вынесен в @/lib/query-client.ts — см. import выше.

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
      <Route path="/help" element={<Suspense fallback={<PageFallback />}><Help /></Suspense>} />
      <Route path="/help/mcp" element={<Suspense fallback={<PageFallback />}><HelpMCP /></Suspense>} />
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
          <Route path="/prompts/:id/analytics" element={<Suspense fallback={<PageFallback />}><PromptAnalytics /></Suspense>} />
          <Route path="/collections" element={<Suspense fallback={<PageFallback />}><Collections /></Suspense>} />
          <Route path="/collections/:id" element={<Suspense fallback={<PageFallback />}><CollectionView /></Suspense>} />
          {/* Phase 16: routes только при VITE_CHAINS_ENABLED=true. */}
          {import.meta.env.VITE_CHAINS_ENABLED === "true" && (
            <>
              <Route path="/chains" element={<Suspense fallback={<PageFallback />}><Chains /></Suspense>} />
              <Route path="/chains/new" element={<Suspense fallback={<PageFallback />}><ChainEditor /></Suspense>} />
              <Route path="/chains/:id/edit" element={<Suspense fallback={<PageFallback />}><ChainEditor /></Suspense>} />
              <Route path="/chains/:id/run" element={<Suspense fallback={<PageFallback />}><ChainRun /></Suspense>} />
              <Route path="/chains/:id/runs" element={<Suspense fallback={<PageFallback />}><ChainRuns /></Suspense>} />
              <Route path="/chains/:id/canvas" element={<Suspense fallback={<PageFallback />}><ChainCanvas /></Suspense>} />
            </>
          )}
          <Route path="/teams" element={<Suspense fallback={<PageFallback />}><Teams /></Suspense>} />
          <Route path="/teams/:slug" element={<Suspense fallback={<PageFallback />}><TeamView /></Suspense>} />
          <Route path="/teams/:slug/analytics" element={<Suspense fallback={<PageFallback />}><TeamAnalytics /></Suspense>} />
          <Route path="/teams/:slug/activity" element={<Suspense fallback={<PageFallback />}><TeamActivity /></Suspense>} />
          <Route path="/teams/:slug/branding" element={<Suspense fallback={<PageFallback />}><TeamBranding /></Suspense>} />
          <Route path="/settings" element={<Suspense fallback={<PageFallback />}><SettingsLayout /></Suspense>}>
            <Route index element={<Navigate to="profile" replace />} />
            <Route path="profile" element={<SettingsProfile />} />
            <Route path="security" element={<SettingsSecurity />} />
            <Route path="accounts" element={<SettingsAccounts />} />
            <Route path="notifications" element={<SettingsNotifications />} />
            <Route path="subscription" element={<SettingsSubscription />} />
            <Route path="referral" element={<SettingsReferral />} />
            <Route path="integrations" element={<SettingsIntegrations />} />
            {/* Backward-compat: старые URL после реорганизации */}
            <Route path="extension" element={<Navigate to="/settings/integrations" replace />} />
            <Route path="api-keys" element={<Navigate to="/settings/integrations" replace />} />
            <Route path="appearance" element={<SettingsAppearance />} />
          </Route>
          <Route path="/history" element={<Suspense fallback={<PageFallback />}><History /></Suspense>} />
          <Route path="/trash" element={<Suspense fallback={<PageFallback />}><Trash /></Suspense>} />
          <Route path="/pricing" element={<Suspense fallback={<PageFallback />}><Pricing /></Suspense>} />
          <Route path="/analytics" element={<Suspense fallback={<PageFallback />}><Analytics /></Suspense>} />
          <Route path="/changelog" element={<Suspense fallback={<PageFallback />}><Changelog /></Suspense>} />
          <Route path="/badges" element={<Suspense fallback={<PageFallback />}><Badges /></Suspense>} />

          {/* Admin routes — гвардятся useAdminGuard внутри AdminLayout */}
          <Route path="/admin" element={<Suspense fallback={<PageFallback />}><AdminLayout /></Suspense>}>
            <Route index element={<Suspense fallback={<PageFallback />}><AdminUsers /></Suspense>} />
            <Route path="users" element={<Suspense fallback={<PageFallback />}><AdminUsers /></Suspense>} />
            <Route path="users/:id" element={<Suspense fallback={<PageFallback />}><AdminUserDetail /></Suspense>} />
            <Route path="feedbacks" element={<Suspense fallback={<PageFallback />}><AdminFeedbacks /></Suspense>} />
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
        <Suspense fallback={null}>
          <ReactQueryDevtools initialIsOpen={false} />
        </Suspense>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
