import { HashRouter } from "react-router-dom"
import { QueryClientProvider } from "@tanstack/react-query"
import { ErrorBoundary } from "./error-boundary"
import { ToasterProvider } from "./ui/toaster"
import { OnboardingOverlay } from "./onboarding-overlay"
import { AppRoutes } from "../routes"
import { queryClient } from "../lib/query-client"

export function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ToasterProvider>
          <HashRouter>
            <AppRoutes />
          </HashRouter>
          <OnboardingOverlay />
        </ToasterProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
