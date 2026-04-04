import { useEffect } from "react"
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"

import { ErrorBoundary } from "@/components/error-boundary"
import { useAuthStore } from "@/stores/auth-store"
import ProtectedRoute from "@/components/auth/protected-route"
import AppLayout from "@/components/layout/app-layout"
import SignIn from "@/pages/sign-in"
import SignUp from "@/pages/sign-up"
import OAuthCallback from "@/pages/oauth-callback"
import VerifyEmail from "@/pages/verify-email"
import Dashboard from "@/pages/dashboard"
import PromptEditor from "@/pages/prompt-editor"
import Collections from "@/pages/collections"
import CollectionView from "@/pages/collection-view"
import Versions from "@/pages/versions"
import SettingsPage from "@/pages/settings"
import ForgotPassword from "@/pages/forgot-password"
import Landing from "@/pages/landing"
import Teams from "@/pages/teams"
import TeamView from "@/pages/team-view"

const queryClient = new QueryClient({
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
    restoreSession()
  }, [restoreSession])

  return (
    <Routes>
      {/* public */}
      <Route path="/" element={<Landing />} />
      <Route path="/sign-in" element={<SignIn />} />
      <Route path="/sign-up" element={<SignUp />} />
      <Route path="/oauth/callback" element={<OAuthCallback />} />
      <Route path="/verify-email" element={<VerifyEmail />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />

      {/* protected — with layout */}
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/prompts/new" element={<PromptEditor />} />
          <Route path="/prompts/:id" element={<PromptEditor />} />
          <Route path="/prompts/:id/versions" element={<Versions />} />
          <Route path="/collections" element={<Collections />} />
          <Route path="/collections/:id" element={<CollectionView />} />
          <Route path="/teams" element={<Teams />} />
          <Route path="/teams/:slug" element={<TeamView />} />
          <Route path="/settings" element={<SettingsPage />} />
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
