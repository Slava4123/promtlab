package app

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/email"
	pgrepo "promptvault/internal/infrastructure/postgres/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/middleware/ratelimit"
	sentrymw "promptvault/internal/middleware/sentry"

	aihttp "promptvault/internal/delivery/http/ai"
	authhttp "promptvault/internal/delivery/http/auth"
	collhttp "promptvault/internal/delivery/http/collection"
	prompthttp "promptvault/internal/delivery/http/prompt"
	searchhttp "promptvault/internal/delivery/http/search"
	starterhttp "promptvault/internal/delivery/http/starter"
	taghttp "promptvault/internal/delivery/http/tag"
	teamhttp "promptvault/internal/delivery/http/team"
	userhttp "promptvault/internal/delivery/http/user"
	"promptvault/internal/infrastructure/openrouter"
	aiuc "promptvault/internal/usecases/ai"
	authuc "promptvault/internal/usecases/auth"
	colluc "promptvault/internal/usecases/collection"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
	starteruc "promptvault/internal/usecases/starter"
	taguc "promptvault/internal/usecases/tag"
	teamuc "promptvault/internal/usecases/team"
	useruc "promptvault/internal/usecases/user"
)

type App struct {
	cfg            *config.Config
	authSvc        *authuc.Service
	tokenValidator authmw.TokenValidator
	authHandler    *authhttp.Handler
	oauthHandler      *authhttp.OAuthHandler
	promptHandler     *prompthttp.Handler
	collectionHandler *collhttp.Handler
	tagHandler        *taghttp.Handler
	aiHandler         *aihttp.Handler
	searchHandler     *searchhttp.Handler
	teamHandler       *teamhttp.Handler
	userHandler       *userhttp.Handler
	starterHandler    *starterhttp.Handler
}

func New(cfg *config.Config, db *gorm.DB) *App {
	userRepo := pgrepo.NewUserRepository(db)
	linkedAccountRepo := pgrepo.NewLinkedAccountRepository(db)
	verificationRepo := pgrepo.NewVerificationRepository(db)
	emailSvc := email.NewService(&cfg.SMTP)

	promptRepo := pgrepo.NewPromptRepository(db)
	tagRepo := pgrepo.NewTagRepository(db)
	collectionRepo := pgrepo.NewCollectionRepository(db)
	versionRepo := pgrepo.NewVersionRepository(db)
	starterRepo := pgrepo.NewStarterRepository(db)

	authSvc := authuc.NewService(cfg, userRepo, linkedAccountRepo, verificationRepo, emailSvc)
	oauthSvc := authuc.NewOAuthService(cfg, userRepo, linkedAccountRepo, authSvc)
	// Teams
	teamRepo := pgrepo.NewTeamRepository(db)
	teamSvc := teamuc.NewService(teamRepo, userRepo)
	teamSvc.SetEmail(emailSvc)

	promptSvc := promptuc.NewService(promptRepo, tagRepo, collectionRepo, versionRepo, teamRepo)
	collectionSvc := colluc.NewService(collectionRepo, teamRepo)
	tagSvc := taguc.NewService(tagRepo, teamRepo)

	// AI
	orClient := openrouter.NewClient(cfg.AI.OpenRouterAPIKey, cfg.AI.OpenRouterBaseURL, time.Duration(cfg.AI.OpenRouterTimeoutSec)*time.Second)
	aiSvc := aiuc.NewService(orClient, &cfg.AI)

	// Search
	searchSvc := searchuc.NewService(promptRepo, collectionRepo, tagRepo)

	// Starter templates (onboarding wizard)
	starterSvc, err := starteruc.NewService(starterRepo, userRepo)
	if err != nil {
		// Fail-fast: catalog.json встроен в бинарник, ошибка парсинга = bug в
		// коде или JSON. Логируем структурно (slog → Sentry/JSON output) перед
		// паникой, чтобы alert-ы знали почему сервис не стартанул.
		slog.Error("starter.catalog.load_failed", "error", err)
		panic(fmt.Sprintf("starter catalog load failed: %v", err))
	}

	return &App{
		cfg:               cfg,
		authSvc:           authSvc,
		tokenValidator:    authSvc,
		authHandler:       authhttp.NewHandler(authSvc, cfg.Server.SecureCookies),
		oauthHandler:      authhttp.NewOAuthHandler(oauthSvc, cfg.Server.FrontendURL, cfg.JWT.Secret, cfg.Server.SecureCookies),
		promptHandler:     prompthttp.NewHandler(promptSvc),
		collectionHandler: collhttp.NewHandler(collectionSvc),
		tagHandler:        taghttp.NewHandler(tagSvc),
		aiHandler:         aihttp.NewHandler(aiSvc),
		searchHandler:     searchhttp.NewHandler(searchSvc),
		teamHandler:       teamhttp.NewHandler(teamSvc),
		userHandler:       userhttp.NewHandler(useruc.NewService(userRepo)),
		starterHandler:    starterhttp.NewHandler(starterSvc),
	}
}

// Shutdown waits for background tasks to complete.
func (a *App) Shutdown(timeout time.Duration) {
	a.authSvc.WaitBackground(timeout)
}

func (a *App) MountRoutes(r chi.Router) {
	authMiddleware := authmw.Middleware(a.tokenValidator)

	r.Route("/api", func(r chi.Router) {
		// public — auth (rate limited: 20 req/min per IP)
		r.Route("/auth", func(r chi.Router) {
			r.Use(ratelimit.ByIP(20))
			r.Post("/register", a.authHandler.Register)
			r.Post("/login", a.authHandler.Login)
			r.Post("/refresh", a.authHandler.Refresh)
			r.Post("/verify-email", a.authHandler.VerifyEmail)
			r.Post("/resend-code", a.authHandler.ResendCode)
			r.Post("/forgot-password", a.authHandler.ForgotPassword)
			r.Post("/reset-password", a.authHandler.ResetPassword)

			// OAuth
			r.Get("/oauth/github", a.oauthHandler.GitHubRedirect)
			r.Get("/oauth/github/callback", a.oauthHandler.GitHubCallback)
			r.Get("/oauth/google", a.oauthHandler.GoogleRedirect)
			r.Get("/oauth/google/callback", a.oauthHandler.GoogleCallback)
			r.Get("/oauth/yandex", a.oauthHandler.YandexRedirect)
			r.Get("/oauth/yandex/callback", a.oauthHandler.YandexCallback)
		})

		// protected
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			// UserContext вешает sentry.User{ID} на Hub из JWT claims, чтобы
			// ошибки внутри protected handlers атрибутировались к конкретному
			// юзеру в GlitchTip UI. No-op если Sentry не инициализирован.
			r.Use(sentrymw.UserContext)
			r.Use(ratelimit.ByIP(60))
			r.Get("/auth/me", a.authHandler.Me)
			r.Post("/auth/set-password/initiate", a.authHandler.InitiateSetPassword)
			r.Post("/auth/set-password/confirm", a.authHandler.ConfirmSetPassword)
			r.Put("/auth/profile", a.authHandler.UpdateProfile)
			r.Put("/auth/password", a.authHandler.ChangePassword)
			r.Get("/auth/linked-accounts", a.authHandler.LinkedAccounts)
			r.Delete("/auth/unlink/{provider}", a.authHandler.UnlinkProvider)
			r.Post("/auth/link/{provider}", a.oauthHandler.InitiateLink)
			r.Post("/auth/logout", a.authHandler.Logout)

			// Search
			r.Get("/search", a.searchHandler.Search)

			// Users
			r.Get("/users/search", a.userHandler.Search)

			// Collections
			r.Route("/collections", func(r chi.Router) {
				r.Get("/", a.collectionHandler.List)
				r.Post("/", a.collectionHandler.Create)
				r.Get("/{id}", a.collectionHandler.GetByID)
				r.Put("/{id}", a.collectionHandler.Update)
				r.Delete("/{id}", a.collectionHandler.Delete)
			})

			// Tags
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", a.tagHandler.List)
				r.Post("/", a.tagHandler.Create)
				r.Delete("/{id}", a.tagHandler.Delete)
			})

			// AI
			r.Route("/ai", func(r chi.Router) {
				r.Get("/models", a.aiHandler.Models)
				r.Post("/enhance", a.aiHandler.Enhance)
				r.Post("/rewrite", a.aiHandler.Rewrite)
				r.Post("/analyze", a.aiHandler.Analyze)
				r.Post("/variations", a.aiHandler.Variations)
			})

			// Teams
			r.Route("/teams", func(r chi.Router) {
				r.Get("/", a.teamHandler.List)
				r.Post("/", a.teamHandler.Create)
				r.Route("/{slug}", func(r chi.Router) {
					r.Get("/", a.teamHandler.GetBySlug)
					r.Put("/", a.teamHandler.Update)
					r.Delete("/", a.teamHandler.Delete)
					r.Put("/members/{userId}", a.teamHandler.UpdateMemberRole)
					r.Delete("/members/{userId}", a.teamHandler.RemoveMember)
					r.Post("/invitations", a.teamHandler.InviteMember)
					r.Get("/invitations", a.teamHandler.ListTeamInvitations)
					r.Delete("/invitations/{invitationId}", a.teamHandler.CancelInvitation)
				})
			})

			// Invitations (глобальные — для текущего пользователя)
			r.Route("/invitations", func(r chi.Router) {
				r.Get("/", a.teamHandler.ListMyInvitations)
				r.Post("/{invitationId}/accept", a.teamHandler.AcceptInvitation)
				r.Post("/{invitationId}/decline", a.teamHandler.DeclineInvitation)
			})

			// Prompts
			r.Route("/prompts", func(r chi.Router) {
				r.Get("/", a.promptHandler.List)
				r.Post("/", a.promptHandler.Create)
				r.Get("/{id}", a.promptHandler.GetByID)
				r.Put("/{id}", a.promptHandler.Update)
				r.Delete("/{id}", a.promptHandler.Delete)
				r.Post("/{id}/favorite", a.promptHandler.ToggleFavorite)
				r.Post("/{id}/use", a.promptHandler.IncrementUsage)
				r.Get("/{id}/versions", a.promptHandler.ListVersions)
				r.Post("/{id}/revert/{versionId}", a.promptHandler.RevertToVersion)
			})

			// Starter templates (onboarding wizard)
			r.Route("/starter", func(r chi.Router) {
				r.Get("/catalog", a.starterHandler.Catalog)
				r.Post("/complete", a.starterHandler.Complete)
			})
		})
	})
}
