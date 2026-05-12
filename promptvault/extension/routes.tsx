import { lazy, Suspense } from "react"
import { Route, Routes } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { AppShell } from "./components/layout"
import { AuthGate } from "./components/auth-gate"
import { PlaceholderPage } from "./pages/_placeholder"
import { NotFoundPage } from "./pages/not-found"

const DashboardPage = lazy(() =>
  import("./pages/dashboard").then((m) => ({ default: m.DashboardPage })),
)
const UsePromptPage = lazy(() =>
  import("./pages/use-prompt").then((m) => ({ default: m.UsePromptPage })),
)
const SignInPage = lazy(() =>
  import("./pages/sign-in").then((m) => ({ default: m.SignInPage })),
)
const SettingsPage = lazy(() =>
  import("./pages/settings").then((m) => ({ default: m.SettingsPage })),
)
const PromptEditorPage = lazy(() =>
  import("./pages/prompts/editor-page").then((m) => ({ default: m.PromptEditorPage })),
)
const PromptDetailPage = lazy(() =>
  import("./pages/prompts/detail-page").then((m) => ({ default: m.PromptDetailPage })),
)
const VersionsPage = lazy(() =>
  import("./pages/prompts/versions-page").then((m) => ({ default: m.VersionsPage })),
)
const TrashPage = lazy(() =>
  import("./pages/trash-page").then((m) => ({ default: m.TrashPage })),
)
const CollectionsPage = lazy(() =>
  import("./pages/collections-page").then((m) => ({ default: m.CollectionsPage })),
)
const CollectionDetailPage = lazy(() =>
  import("./pages/collection-detail-page").then((m) => ({ default: m.CollectionDetailPage })),
)
const TagsPage = lazy(() =>
  import("./pages/tags-page").then((m) => ({ default: m.TagsPage })),
)
const TagDetailPage = lazy(() =>
  import("./pages/tag-detail-page").then((m) => ({ default: m.TagDetailPage })),
)
const BadgesPage = lazy(() =>
  import("./pages/badges-page").then((m) => ({ default: m.BadgesPage })),
)
const ChangelogPage = lazy(() =>
  import("./pages/changelog-page").then((m) => ({ default: m.ChangelogPage })),
)
const HistoryPage = lazy(() =>
  import("./pages/history-page").then((m) => ({ default: m.HistoryPage })),
)
const AnalyticsPage = lazy(() =>
  import("./pages/analytics-page").then((m) => ({ default: m.AnalyticsPage })),
)
const ChainsIndexPage = lazy(() =>
  import("./pages/chains/index-page").then((m) => ({ default: m.ChainsIndexPage })),
)
const ChainRunPage = lazy(() =>
  import("./pages/chains/run-page").then((m) => ({ default: m.ChainRunPage })),
)
const ChainRunsPage = lazy(() =>
  import("./pages/chains/runs-page").then((m) => ({ default: m.ChainRunsPage })),
)
const TeamsIndexPage = lazy(() =>
  import("./pages/teams/index-page").then((m) => ({ default: m.TeamsIndexPage })),
)
const TeamDetailPage = lazy(() =>
  import("./pages/teams/detail-page").then((m) => ({ default: m.TeamDetailPage })),
)
const ProfilePage = lazy(() =>
  import("./pages/settings/profile-page").then((m) => ({ default: m.ProfilePage })),
)
const IntegrationsPage = lazy(() =>
  import("./pages/settings/integrations-page").then((m) => ({ default: m.IntegrationsPage })),
)
const SubscriptionSettingsPage = lazy(() =>
  import("./pages/settings/subscription-page").then((m) => ({ default: m.SubscriptionPage })),
)
const AppearancePage = lazy(() =>
  import("./pages/settings/appearance-page").then((m) => ({ default: m.AppearancePage })),
)
const PricingPage = lazy(() =>
  import("./pages/pricing-page").then((m) => ({ default: m.PricingPage })),
)

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center">
      <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
    </div>
  )
}

export function AppRoutes() {
  return (
    <Suspense fallback={<PageLoader />}>
      <Routes>
        <Route path="/sign-in" element={<SignInPage />} />
        <Route element={<AuthGate />}>
          <Route element={<AppShell />}>
            <Route index element={<DashboardPage />} />
            <Route path="/prompts" element={<DashboardPage />} />
            {/* Phase 1: Prompts CRUD */}
            <Route path="/prompts/new" element={<PromptEditorPage />} />
            <Route path="/prompts/:id" element={<PromptDetailPage />} />
            <Route path="/prompts/:id/edit" element={<PromptEditorPage />} />
            <Route path="/prompts/:id/versions" element={<VersionsPage />} />
            <Route path="/prompts/:id/use" element={<UsePromptPage />} />
            <Route path="/trash" element={<TrashPage />} />
            {/* Phase 2: Organization */}
            <Route path="/collections" element={<CollectionsPage />} />
            <Route path="/collections/:id" element={<CollectionDetailPage />} />
            <Route path="/tags" element={<TagsPage />} />
            <Route path="/tags/:id" element={<TagDetailPage />} />
            {/* Phase 3: Chains */}
            <Route path="/chains" element={<ChainsIndexPage />} />
            <Route
              path="/chains/new"
              element={
                <PlaceholderPage
                  title="Редактор цепочки"
                  description="Создавайте цепочки в веб-приложении. Запускайте здесь."
                  phase="Phase 3 polish"
                />
              }
            />
            <Route
              path="/chains/:id"
              element={
                <PlaceholderPage
                  title="Детали цепочки"
                  description="Используйте 'Запустить' для запуска цепочки."
                  phase="Phase 3 polish"
                  webPath="/chains"
                />
              }
            />
            <Route
              path="/chains/:id/edit"
              element={
                <PlaceholderPage
                  title="Редактор цепочки"
                  description="Редактирование в веб-приложении."
                  phase="Phase 3 polish"
                />
              }
            />
            <Route path="/chains/:id/run" element={<ChainRunPage />} />
            <Route path="/chains/:id/runs" element={<ChainRunsPage />} />
            <Route
              path="/chains/:id/canvas"
              element={
                <PlaceholderPage
                  title="Canvas цепочки"
                  description="DAG-визуализация в веб-приложении."
                  phase="Phase 3 polish"
                />
              }
            />
            {/* Phase 4: Teams */}
            <Route path="/teams" element={<TeamsIndexPage />} />
            <Route path="/teams/:slug" element={<TeamDetailPage />} />
            <Route
              path="/teams/:slug/branding"
              element={
                <PlaceholderPage
                  title="Брендинг"
                  description="Логотип и цвета команды — редактирование в веб-приложении."
                  phase="Phase 4 polish"
                />
              }
            />
            <Route
              path="/teams/:slug/analytics"
              element={
                <PlaceholderPage
                  title="Аналитика команды"
                  description="Метрики команды — в веб-приложении."
                  phase="Phase 2 polish"
                />
              }
            />
            <Route
              path="/teams/:slug/activity"
              element={
                <PlaceholderPage
                  title="Активность команды"
                  description="События в команде — в веб-приложении."
                  phase="Phase 2 polish"
                />
              }
            />
            <Route path="/analytics" element={<AnalyticsPage />} />
            <Route path="/history" element={<HistoryPage />} />
            <Route path="/badges" element={<BadgesPage />} />
            <Route
              path="/notifications"
              element={
                <PlaceholderPage
                  title="Уведомления"
                  description="Настройте email-уведомления в веб-приложении."
                  phase="Phase 5"
                  webPath="/settings/notifications"
                />
              }
            />
            <Route path="/changelog" element={<ChangelogPage />} />
            <Route path="/pricing" element={<PricingPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/settings/profile" element={<ProfilePage />} />
            <Route
              path="/settings/security"
              element={
                <PlaceholderPage
                  title="Безопасность"
                  description="2FA, активные сессии — управляйте в веб-приложении."
                  phase="Phase 5 polish"
                />
              }
            />
            <Route
              path="/settings/accounts"
              element={
                <PlaceholderPage
                  title="Подключённые аккаунты"
                  description="OAuth-привязка — в веб-приложении."
                  phase="Phase 5 polish"
                />
              }
            />
            <Route
              path="/settings/notifications"
              element={
                <PlaceholderPage
                  title="Уведомления"
                  description="Настройки email-уведомлений — в веб-приложении."
                  phase="Phase 5 polish"
                />
              }
            />
            <Route path="/settings/subscription" element={<SubscriptionSettingsPage />} />
            <Route
              path="/settings/referral"
              element={
                <PlaceholderPage
                  title="Реферальная программа"
                  description="Реферальный код и приглашения — в веб-приложении."
                  phase="Phase 5 polish"
                />
              }
            />
            <Route path="/settings/integrations" element={<IntegrationsPage />} />
            <Route path="/settings/appearance" element={<AppearancePage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Route>
      </Routes>
    </Suspense>
  )
}
