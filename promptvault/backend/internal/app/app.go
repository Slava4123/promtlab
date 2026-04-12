package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/email"
	pgrepo "promptvault/internal/infrastructure/postgres/repository"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/middleware/ratelimit"
	sentrymw "promptvault/internal/middleware/sentry"

	adminhttp "promptvault/internal/delivery/http/admin"
	adminauthhttp "promptvault/internal/delivery/http/adminauth"
	aihttp "promptvault/internal/delivery/http/ai"
	apikeyhttp "promptvault/internal/delivery/http/apikey"
	authhttp "promptvault/internal/delivery/http/auth"
	badgehttp "promptvault/internal/delivery/http/badge"
	changeloghttp "promptvault/internal/delivery/http/changelog"
	collhttp "promptvault/internal/delivery/http/collection"
	feedbackhttp "promptvault/internal/delivery/http/feedback"
	prompthttp "promptvault/internal/delivery/http/prompt"
	searchhttp "promptvault/internal/delivery/http/search"
	streakhttp "promptvault/internal/delivery/http/streak"
	sharehttp "promptvault/internal/delivery/http/share"
	starterhttp "promptvault/internal/delivery/http/starter"
	taghttp "promptvault/internal/delivery/http/tag"
	teamhttp "promptvault/internal/delivery/http/team"
	trashhttp "promptvault/internal/delivery/http/trash"
	userhttp "promptvault/internal/delivery/http/user"
	"promptvault/internal/infrastructure/openrouter"
	"promptvault/internal/mcpserver"
	adminmw "promptvault/internal/middleware/admin"
	adminuc "promptvault/internal/usecases/admin"
	adminauthuc "promptvault/internal/usecases/adminauth"
	aiuc "promptvault/internal/usecases/ai"
	apikeyuc "promptvault/internal/usecases/apikey"
	auditsvc "promptvault/internal/usecases/audit"
	authuc "promptvault/internal/usecases/auth"
	badgeuc "promptvault/internal/usecases/badge"
	changeloguc "promptvault/internal/usecases/changelog"
	colluc "promptvault/internal/usecases/collection"
	feedbackuc "promptvault/internal/usecases/feedback"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
	streakuc "promptvault/internal/usecases/streak"
	shareuc "promptvault/internal/usecases/share"
	starteruc "promptvault/internal/usecases/starter"
	taguc "promptvault/internal/usecases/tag"
	teamuc "promptvault/internal/usecases/team"
	trashuc "promptvault/internal/usecases/trash"
	useruc "promptvault/internal/usecases/user"
)

// apiKeyValidatorAdapter адаптирует *apikeyuc.Service к узкому интерфейсу
// authmw.APIKeyValidator, чтобы middleware не зависел от usecases package.
type apiKeyValidatorAdapter struct {
	svc *apikeyuc.Service
}

func (a *apiKeyValidatorAdapter) ValidateKey(ctx context.Context, rawKey string) (userID uint, keyID uint, err error) {
	result, err := a.svc.ValidateKey(ctx, rawKey)
	if err != nil {
		return 0, 0, err
	}
	return result.UserID, result.KeyID, nil
}

// mcpPromptAdapter оборачивает *promptuc.Service и скрывает newly_unlocked_badges
// из сигнатур Create/Update — они не нужны MCP клиентам (Claude), которые работают
// через JSON-RPC и не имеют toast-UI. Адаптер позволяет не тянуть badge-логику в
// mcpserver package и сохранить узкий mcpserver.PromptService интерфейс.
type mcpPromptAdapter struct {
	*promptuc.Service
}

func (a *mcpPromptAdapter) Create(ctx context.Context, in promptuc.CreateInput) (*models.Prompt, error) {
	p, _, err := a.Service.Create(ctx, in)
	return p, err
}

func (a *mcpPromptAdapter) Update(ctx context.Context, id, userID uint, in promptuc.UpdateInput) (*models.Prompt, error) {
	p, _, err := a.Service.Update(ctx, id, userID, in)
	return p, err
}

// mcpCollectionAdapter — симметричный адаптер для *colluc.Service. Скрывает
// возвращаемый slice бейджей из Create, чтобы mcpserver.CollectionService
// оставался узким контрактом.
type mcpCollectionAdapter struct {
	*colluc.Service
}

// adminHealthAdapter — узкий интерфейс HealthCounter для adminhttp.Handler,
// делегирующий в AdminRepository.CountUsers. Лежит в app.go, чтобы не тянуть
// repo-slices в handler package.
type adminHealthAdapter struct {
	repo repo.AdminRepository
}

func (a *adminHealthAdapter) CountUsers(ctx context.Context) (total, admins, active, frozen int64, err error) {
	return a.repo.CountUsers(ctx)
}

func (a *mcpCollectionAdapter) Create(ctx context.Context, userID uint, name, description, color, icon string, teamID *uint) (*models.Collection, error) {
	c, _, err := a.Service.Create(ctx, userID, name, description, color, icon, teamID)
	return c, err
}

type App struct {
	cfg              *config.Config
	authSvc          *authuc.Service
	tokenValidator   authmw.TokenValidator
	apiKeyValidator  authmw.APIKeyValidator
	authHandler      *authhttp.Handler
	oauthHandler     *authhttp.OAuthHandler
	promptHandler     *prompthttp.Handler
	collectionHandler *collhttp.Handler
	tagHandler        *taghttp.Handler
	aiHandler         *aihttp.Handler
	searchHandler     *searchhttp.Handler
	teamHandler       *teamhttp.Handler
	userHandler       *userhttp.Handler
	starterHandler    *starterhttp.Handler
	trashHandler      *trashhttp.Handler
	apiKeyHandler     *apikeyhttp.Handler
	shareHandler      *sharehttp.Handler
	streakHandler     *streakhttp.Handler
	badgeHandler      *badgehttp.Handler
	adminauthHandler  *adminauthhttp.Handler
	adminHandler      *adminhttp.Handler
	feedbackHandler   *feedbackhttp.Handler
	changelogHandler  *changeloghttp.Handler
	// Следующие поля используются в MountRoutes для admin-middleware chain:
	userLookup adminmw.UserLookup
	auditSvc   *auditsvc.Service
	mcpServer         *mcpserver.MCPServer
	purgeLoop         *trashuc.PurgeLoop
	feedbackRL        *ratelimit.Limiter[uint]
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

	pinRepo := pgrepo.NewPinRepository(db)
	streakRepo := pgrepo.NewStreakRepository(db)
	streakSvc := streakuc.NewService(streakRepo)

	// Badges: загружает embedded catalog.json при старте, fail-fast на невалидном каталоге
	// (аналогично starter/changelog). Вызывается из prompt/collection usecases (шаги B4-B6).
	badgeRepo := pgrepo.NewBadgeRepository(db)
	badgeSvc, err := badgeuc.NewService(badgeRepo, streakSvc)
	if err != nil {
		slog.Error("badge.catalog.load_failed", "error", err)
		panic(fmt.Sprintf("badge catalog load failed: %v", err))
	}

	// Admin foundation (Этап 2):
	// - auditRepo / auditSvc — append-only журнал через audit_log таблицу.
	// - totpRepo / adminauthSvc — TOTP enrollment/verify + backup codes.
	// - adminRepo / adminSvc — admin actions (freeze, reset, grant/revoke badge).
	auditRepo := pgrepo.NewAuditRepository(db)
	auditSvc := auditsvc.NewService(auditRepo)
	totpRepo := pgrepo.NewTOTPRepository(db)
	adminauthSvc := adminauthuc.NewService(totpRepo, userRepo)
	adminRepo := pgrepo.NewAdminRepository(db)

	promptSvc := promptuc.NewService(promptRepo, tagRepo, collectionRepo, versionRepo, teamRepo, pinRepo, streakSvc, badgeSvc)
	collectionSvc := colluc.NewService(collectionRepo, teamRepo, badgeSvc)
	tagSvc := taguc.NewService(tagRepo, teamRepo)

	// Admin usecase (AU1) — зависит от auth/audit/badge сервисов, поэтому
	// собирается после promptSvc/collectionSvc, но до handlers.
	adminSvc := adminuc.NewService(adminRepo, userRepo, auditSvc, authSvc, badgeSvc)
	adminHealth := &adminHealthAdapter{repo: adminRepo}

	// AI
	orClient := openrouter.NewClient(cfg.AI.OpenRouterAPIKey, cfg.AI.OpenRouterBaseURL, time.Duration(cfg.AI.OpenRouterTimeoutSec)*time.Second)
	aiSvc := aiuc.NewService(orClient, &cfg.AI)

	// Search
	searchSvc := searchuc.NewService(promptRepo, collectionRepo, tagRepo)

	// Trash
	trashRepo := pgrepo.NewTrashRepository(db)
	trashSvc := trashuc.NewService(trashRepo, teamRepo)

	// API Keys
	apiKeyRepo := pgrepo.NewAPIKeyRepository(db)
	apiKeySvc := apikeyuc.NewService(apiKeyRepo, cfg.MCP.MaxKeysPerUser)

	// Share Links
	shareLinkRepo := pgrepo.NewShareLinkRepository(db)
	shareSvc := shareuc.NewService(shareLinkRepo, promptRepo, teamRepo, cfg.Server.FrontendURL)

	// MCP Server — promptSvc и collectionSvc оборачиваются в адаптеры, которые
	// скрывают возвращаемые badges-slices (MCP-клиенты не показывают toast-ов).
	var mcpSrv *mcpserver.MCPServer
	if cfg.MCP.Enabled {
		mcpSrv = mcpserver.NewMCPServer(
			apiKeySvc,
			&mcpPromptAdapter{Service: promptSvc},
			&mcpCollectionAdapter{Service: collectionSvc},
			tagSvc,
			searchSvc,
			60,
		)
	}

	purgeLoop := trashuc.NewPurgeLoop(trashRepo, 1*time.Hour, 30)

	// Feedback
	feedbackRepo := pgrepo.NewFeedbackRepository(db)
	feedbackSvc := feedbackuc.NewService(feedbackRepo)

	// Changelog
	changelogSvc, err := changeloguc.NewService(userRepo)
	if err != nil {
		slog.Error("changelog.load_failed", "error", err)
		panic(fmt.Sprintf("changelog load failed: %v", err))
	}

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
		apiKeyValidator:   &apiKeyValidatorAdapter{svc: apiKeySvc},
		authHandler:       authhttp.NewHandler(authSvc, adminauthSvc, changelogSvc, cfg.Server.SecureCookies),
		adminauthHandler:  adminauthhttp.NewHandler(adminauthSvc),
		adminHandler:      adminhttp.NewHandler(adminSvc, adminauthSvc, auditSvc, adminHealth),
		userLookup:        userRepo,
		auditSvc:          auditSvc,
		oauthHandler:      authhttp.NewOAuthHandler(oauthSvc, cfg.Server.FrontendURL, cfg.JWT.Secret, cfg.Server.SecureCookies),
		promptHandler:     prompthttp.NewHandler(promptSvc),
		collectionHandler: collhttp.NewHandler(collectionSvc),
		tagHandler:        taghttp.NewHandler(tagSvc),
		aiHandler:         aihttp.NewHandler(aiSvc),
		searchHandler:     searchhttp.NewHandler(searchSvc),
		teamHandler:       teamhttp.NewHandler(teamSvc),
		userHandler:       userhttp.NewHandler(useruc.NewService(userRepo)),
		starterHandler:    starterhttp.NewHandler(starterSvc),
		trashHandler:      trashhttp.NewHandler(trashSvc),
		apiKeyHandler:     apikeyhttp.NewHandler(apiKeySvc, cfg.MCP.MaxKeysPerUser),
		shareHandler:      sharehttp.NewHandler(shareSvc),
		streakHandler:     streakhttp.NewHandler(streakSvc),
		badgeHandler:      badgehttp.NewHandler(badgeSvc),
		feedbackHandler:   feedbackhttp.NewHandler(feedbackSvc),
		changelogHandler:  changeloghttp.NewHandler(changelogSvc),
		mcpServer:         mcpSrv,
		purgeLoop:         purgeLoop,
		feedbackRL:        ratelimit.NewLimiterWithWindow[uint](5, time.Hour),
	}
}

func (a *App) StartBackground() {
	a.purgeLoop.Start()
}

// Shutdown waits for background tasks to complete.
func (a *App) Shutdown(timeout time.Duration) {
	a.purgeLoop.Stop()
	a.feedbackRL.Close()
	a.authSvc.WaitBackground(timeout)
}

func (a *App) MountRoutes(r chi.Router) {
	// Protected routes принимают и JWT (SPA), и API-ключ `pvlt_*` (Chrome Extension, MCP-клиенты).
	// Префикс токена определяет путь валидации.
	authMiddleware := authmw.CombinedAuth(a.tokenValidator, a.apiKeyValidator)

	// MCP endpoint (outside /api, with pre-auth IP rate limit)
	if a.mcpServer != nil {
		mcpHandler := ratelimit.ByIP(120)(a.mcpServer.Handler())
		r.Method(http.MethodPost, "/mcp", mcpHandler)
		r.Method(http.MethodGet, "/mcp", mcpHandler)
		r.Method(http.MethodDelete, "/mcp", mcpHandler)
		slog.Info("mcp.server.mounted", "path", "/mcp")
	}

	r.Route("/api", func(r chi.Router) {
		// public — share links (rate limited: 60 req/min per IP)
		r.Route("/s", func(r chi.Router) {
			r.Use(ratelimit.ByIP(60))
			r.Get("/{token}", a.shareHandler.GetPublic)
		})

		// public — auth (rate limited: 20 req/min per IP)
		r.Route("/auth", func(r chi.Router) {
			r.Use(ratelimit.ByIP(20))
			r.Post("/register", a.authHandler.Register)
			r.Post("/login", a.authHandler.Login)
			r.Post("/verify-totp", a.authHandler.VerifyTOTP)
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
			r.Get("/search/suggest", a.searchHandler.Suggest)

			// Streaks
			r.Get("/streaks", a.streakHandler.Get)

			// Badges
			r.Get("/badges", a.badgeHandler.List)

			// Admin TOTP management (enrollment, verify, backup codes regeneration).
			// Сам enrollment доступен ТОЛЬКО admin-юзерам (RequireAdmin middleware),
			// но enroll endpoint не требует AdminAuditContext — до completion
			// enrollment это не destructive action над другими юзерами.
			r.Route("/admin/totp", func(r chi.Router) {
				r.Use(adminmw.RequireAdmin(a.userLookup))
				r.Post("/enroll", a.adminauthHandler.Enroll)
				r.Post("/verify-enrollment", a.adminauthHandler.ConfirmEnrollment)
				r.Post("/backup-codes/regenerate", a.adminauthHandler.RegenerateBackupCodes)
				r.Get("/status", a.adminauthHandler.Status)
			})

			// Admin user management + audit log + health dashboard.
			// Middleware chain: JWT (protected group выше) → RequireAdmin → AdminAuditContext.
			// Destructive actions требуют fresh TOTP в body (sudo mode) — проверка
			// в handler через adminauth.Service.Verify, не middleware.
			r.Route("/admin", func(r chi.Router) {
				r.Use(adminmw.RequireAdmin(a.userLookup))
				r.Use(adminmw.AdminAuditContext)

				r.Route("/users", func(r chi.Router) {
					r.Get("/", a.adminHandler.ListUsers)
					r.Get("/{id}", a.adminHandler.GetUserDetail)
					r.Post("/{id}/freeze", a.adminHandler.FreezeUser)
					r.Post("/{id}/unfreeze", a.adminHandler.UnfreezeUser)
					r.Post("/{id}/reset-password", a.adminHandler.ResetPassword)
					r.Post("/{id}/tier", a.adminHandler.ChangeTier)
					r.Post("/{id}/badges/{badge_id}/grant", a.adminHandler.GrantBadge)
					r.Delete("/{id}/badges/{badge_id}", a.adminHandler.RevokeBadge)
				})

				r.Get("/audit", a.adminHandler.ListAudit)
				r.Get("/health", a.adminHandler.Health)
			})

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

			// Trash
			r.Route("/trash", func(r chi.Router) {
				r.Get("/", a.trashHandler.List)
				r.Get("/count", a.trashHandler.Count)
				r.Delete("/", a.trashHandler.Empty)
				r.Post("/{type}/{id}/restore", a.trashHandler.Restore)
				r.Delete("/{type}/{id}", a.trashHandler.PermanentDelete)
			})

			// Prompts
			r.Route("/prompts", func(r chi.Router) {
				r.Get("/", a.promptHandler.List)
				r.Post("/", a.promptHandler.Create)
				r.Get("/pinned", a.promptHandler.ListPinned)
				r.Get("/recent", a.promptHandler.ListRecent)
				r.Get("/history", a.promptHandler.ListHistory)
				r.Get("/{id}", a.promptHandler.GetByID)
				r.Put("/{id}", a.promptHandler.Update)
				r.Delete("/{id}", a.promptHandler.Delete)
				r.Post("/{id}/favorite", a.promptHandler.ToggleFavorite)
				r.Post("/{id}/pin", a.promptHandler.TogglePin)
				r.Post("/{id}/use", a.promptHandler.IncrementUsage)
				r.Get("/{id}/versions", a.promptHandler.ListVersions)
				r.Post("/{id}/revert/{versionId}", a.promptHandler.RevertToVersion)
				r.Get("/{id}/share", a.shareHandler.Get)
				r.Post("/{id}/share", a.shareHandler.Create)
				r.Delete("/{id}/share", a.shareHandler.Delete)
			})

			// Feedback (5/hour per user)
			r.Route("/feedback", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						userID := authmw.GetUserID(r.Context())
						if userID > 0 && !a.feedbackRL.Allow(userID) {
							w.Header().Set("Content-Type", "application/json")
							w.Header().Set("Retry-After", "3600")
							w.WriteHeader(http.StatusTooManyRequests)
							_, _ = w.Write([]byte(`{"error":"Слишком много отзывов. Попробуйте через час"}`))
							return
						}
						next.ServeHTTP(w, r)
					})
				})
				r.Post("/", a.feedbackHandler.Submit)
			})

			// Changelog
			r.Get("/changelog", a.changelogHandler.List)
			r.Post("/changelog/seen", a.changelogHandler.MarkSeen)

			// Starter templates (onboarding wizard)
			r.Route("/starter", func(r chi.Router) {
				r.Get("/catalog", a.starterHandler.Catalog)
				r.Post("/complete", a.starterHandler.Complete)
			})

			// API Keys
			r.Route("/api-keys", func(r chi.Router) {
				r.Get("/", a.apiKeyHandler.List)
				r.Post("/", a.apiKeyHandler.Create)
				r.Delete("/{id}", a.apiKeyHandler.Revoke)
			})
		})
	})
}
