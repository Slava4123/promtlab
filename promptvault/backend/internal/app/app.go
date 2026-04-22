package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/email"
	pgrepo "promptvault/internal/infrastructure/postgres/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/middleware/ratelimit"

	analyticshttp "promptvault/internal/delivery/http/analytics"
	subscriptionhttp "promptvault/internal/delivery/http/subscription"
	webhookhttp "promptvault/internal/delivery/http/webhook"
	adminhttp "promptvault/internal/delivery/http/admin"
	adminauthhttp "promptvault/internal/delivery/http/adminauth"
	apikeyhttp "promptvault/internal/delivery/http/apikey"
	authhttp "promptvault/internal/delivery/http/auth"
	badgehttp "promptvault/internal/delivery/http/badge"
	changeloghttp "promptvault/internal/delivery/http/changelog"
	collhttp "promptvault/internal/delivery/http/collection"
	feedbackhttp "promptvault/internal/delivery/http/feedback"
	prompthttp "promptvault/internal/delivery/http/prompt"
	searchhttp "promptvault/internal/delivery/http/search"
	seohttp "promptvault/internal/delivery/http/seo"
	streakhttp "promptvault/internal/delivery/http/streak"
	sharehttp "promptvault/internal/delivery/http/share"
	starterhttp "promptvault/internal/delivery/http/starter"
	metadatahttp "promptvault/internal/delivery/http/metadata"
	oauthsrvhttp "promptvault/internal/delivery/http/oauth_server"
	taghttp "promptvault/internal/delivery/http/tag"
	teamhttp "promptvault/internal/delivery/http/team"
	trashhttp "promptvault/internal/delivery/http/trash"
	userhttp "promptvault/internal/delivery/http/user"
	"promptvault/internal/infrastructure/payment"
	"promptvault/internal/infrastructure/payment/tbank"
	"promptvault/internal/mcpserver"
	adminmw "promptvault/internal/middleware/admin"
	activityuc "promptvault/internal/usecases/activity"
	adminuc "promptvault/internal/usecases/admin"
	adminauthuc "promptvault/internal/usecases/adminauth"
	analyticsuc "promptvault/internal/usecases/analytics"
	apikeyuc "promptvault/internal/usecases/apikey"
	auditsvc "promptvault/internal/usecases/audit"
	authuc "promptvault/internal/usecases/auth"
	badgeuc "promptvault/internal/usecases/badge"
	changeloguc "promptvault/internal/usecases/changelog"
	colluc "promptvault/internal/usecases/collection"
	feedbackuc "promptvault/internal/usecases/feedback"
	promptuc "promptvault/internal/usecases/prompt"
	quotauc "promptvault/internal/usecases/quota"
	searchuc "promptvault/internal/usecases/search"
	streakuc "promptvault/internal/usecases/streak"
	shareuc "promptvault/internal/usecases/share"
	engagementuc "promptvault/internal/usecases/engagement"
	oauthsrvuc "promptvault/internal/usecases/oauth_server"
	subscriptionuc "promptvault/internal/usecases/subscription"
	starteruc "promptvault/internal/usecases/starter"
	taguc "promptvault/internal/usecases/tag"
	teamuc "promptvault/internal/usecases/team"
	trashuc "promptvault/internal/usecases/trash"
	useruc "promptvault/internal/usecases/user"
)

// Адаптеры вынесены в adapters.go для чистоты DI-графа.

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
	feedbackHandler       *feedbackhttp.Handler
	changelogHandler      *changeloghttp.Handler
	subscriptionHandler   *subscriptionhttp.Handler
	webhookHandler        *webhookhttp.Handler
	seoHandler            *seohttp.Handler
	// Phase 14 B.4
	analyticsHandler     *analyticshttp.Handler
	teamActivityHandler  *teamhttp.ActivityHandler
	// Phase 14 D
	teamBrandingHandler  *teamhttp.BrandingHandler
	oauthServerHandler    *oauthsrvhttp.Handler
	metadataHandler       *metadatahttp.Handler
	// Следующие поля используются в MountRoutes для admin-middleware chain:
	userLookup adminmw.UserLookup
	auditSvc   *auditsvc.Service
	mcpServer         *mcpserver.MCPServer
	purgeLoop         *trashuc.PurgeLoop
	expirationLoop    *subscriptionuc.ExpirationLoop
	renewalLoop       *subscriptionuc.RenewalLoop
	reminderLoop      *subscriptionuc.ReminderLoop
	reengagementLoop  *engagementuc.ReengagementLoop
	streakReminderLoop *engagementuc.StreakReminderLoop
	// Phase 14: audit + analytics фоновые воркеры.
	activityCleanupLoop *analyticsuc.CleanupLoop
	insightsLoop        *analyticsuc.InsightsComputeLoop
	feedbackRL        *ratelimit.Limiter[uint]
	// insightsRL — 1 refresh инсайтов в час на юзера (POST /api/analytics/insights/refresh).
	insightsRL *ratelimit.Limiter[uint]
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
	// Subscription / Quota repos
	planRepo := pgrepo.NewPlanRepository(db)
	quotaRepo := pgrepo.NewQuotaRepository(db)
	quotaSvc := quotauc.NewService(planRepo, quotaRepo, userRepo)
	quotaSvc.SetEmailNotifier(emailSvc, cfg.Server.FrontendURL)

	// Teams
	teamRepo := pgrepo.NewTeamRepository(db)
	teamSvc := teamuc.NewService(teamRepo, userRepo, quotaSvc)
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

	// Team activity feed (Phase 14) — продуктовые события внутри команды
	// (prompt/collection/share/member/role). Рядом с audit-инфраструктурой,
	// т.к. концептуально близко, но целевая аудитория — члены команды, не админы.
	activityRepo := pgrepo.NewTeamActivityRepository(db)
	activitySvc := activityuc.NewService(activityRepo, userRepo)

	// Analytics (Phase 14) — dashboard-агрегации, Smart Insights (Max-only),
	// retention cleanup.
	analyticsRepo := pgrepo.NewAnalyticsRepository(db)
	analyticsSvc := analyticsuc.NewService(analyticsRepo, promptRepo, teamRepo, userRepo, quotaSvc)
	// Q2: experimental insights flag — 4 неготовых типа скрыты до M8.
	analyticsSvc.SetExperimentalInsights(cfg.Analytics.ExperimentalInsights)

	// Phase 14 M-10: email-digest по Smart Insights. Rate-limit 1/неделя
	// через insight_notifications. Orchestration SMTP → repo → service.
	insightNotifRepo := pgrepo.NewInsightNotificationRepository(db)
	insightsNotifier := email.NewEmailInsightsNotifier(emailSvc, userRepo, insightNotifRepo, cfg.Server.FrontendURL)
	analyticsSvc.SetNotifier(insightsNotifier)

	// Phase B подключает analyticsSvc в MCP-server (ниже) и в HTTP handlers.

	// Cleanup loops — ежесуточно. Retention: Free=30д, Pro=90д, Max=365д (per-plan в SQL).
	activityCleanupLoop := analyticsuc.NewCleanupLoop(activityRepo, analyticsRepo, 24*time.Hour)
	insightsLoop := analyticsuc.NewInsightsComputeLoop(analyticsSvc, userRepo, teamRepo, 24*time.Hour)

	promptSvc := promptuc.NewService(promptRepo, tagRepo, collectionRepo, versionRepo, teamRepo, pinRepo, streakSvc, badgeSvc, quotaSvc)
	promptSvc.SetActivity(activitySvc)
	teamSvc.SetActivity(activitySvc)
	collectionSvc := colluc.NewService(collectionRepo, teamRepo, badgeSvc, quotaSvc)
	collectionSvc.SetActivity(activitySvc)
	tagSvc := taguc.NewService(tagRepo, teamRepo)

	// Subscription repos (нужны и для admin, и для subscription service)
	subscriptionRepo := pgrepo.NewSubscriptionRepository(db)
	paymentRepo := pgrepo.NewPaymentRepository(db)

	// Admin usecase (AU1) — зависит от auth/audit/badge сервисов, поэтому
	// собирается после promptSvc/collectionSvc, но до handlers.
	adminSvc := adminuc.NewService(adminRepo, userRepo, auditSvc, authSvc, badgeSvc, planRepo, subscriptionRepo)
	adminHealth := &adminHealthAdapter{repo: adminRepo}

	// Search
	searchSvc := searchuc.NewService(promptRepo, collectionRepo, tagRepo)

	// Trash
	trashRepo := pgrepo.NewTrashRepository(db)
	trashSvc := trashuc.NewService(trashRepo, teamRepo)

	// User service — общий для HTTP-хендлера и MCP (whoami).
	userSvc := useruc.NewService(userRepo)

	// API Keys
	apiKeyRepo := pgrepo.NewAPIKeyRepository(db)
	apiKeySvc := apikeyuc.NewService(apiKeyRepo, cfg.MCP.MaxKeysPerUser)

	// Share Links
	shareLinkRepo := pgrepo.NewShareLinkRepository(db)
	shareSvc := shareuc.NewService(shareLinkRepo, promptRepo, teamRepo, cfg.Server.FrontendURL, quotaSvc)
	shareSvc.SetActivity(activitySvc)
	// Phase 14 B.2: подключаем share_view_log write-path (Pro+ owner). Nil-safe.
	// План владельца читается из уже preload'ленного link.Prompt.User (M9).
	shareSvc.SetViewLogger(analyticsRepo)
	// Phase 14 D: branded share pages — BrandingProvider интерфейс,
	// teamSvc удовлетворяет ему методом GetBrandingForShare (H6).
	shareSvc.SetBrandingLookup(teamSvc)

	// OAuth 2.1 Authorization Server для внешних MCP-клиентов (Claude.ai и т.д.).
	// Canonical resource = public URL MCP-сервера + "/mcp" (RFC 8707 audience).
	oauthClientRepo := pgrepo.NewOAuthClientRepository(db)
	oauthCodeRepo := pgrepo.NewOAuthAuthorizationCodeRepository(db)
	oauthTokenRepo := pgrepo.NewOAuthTokenRepository(db)
	canonicalResource := strings.TrimRight(cfg.Server.FrontendURL, "/") + "/mcp"
	oauthSrvSvc := oauthsrvuc.NewService(oauthClientRepo, oauthCodeRepo, oauthTokenRepo, canonicalResource)

	// MCP Server — promptSvc и collectionSvc оборачиваются в адаптеры, которые
	// скрывают возвращаемые badges-slices (MCP-клиенты не показывают toast-ов).
	var mcpSrv *mcpserver.MCPServer
	if cfg.MCP.Enabled {
		resourceMetadataURL := strings.TrimRight(cfg.Server.FrontendURL, "/") + "/.well-known/oauth-protected-resource"
		mcpSrv = mcpserver.NewMCPServer(
			apiKeySvc,
			&mcpPromptAdapter{Service: promptSvc},
			&mcpCollectionAdapter{Service: collectionSvc},
			tagSvc,
			searchSvc,
			shareSvc,
			teamSvc,
			trashSvc,
			userSvc,
			activitySvc,  // Phase 14 B.3
			analyticsSvc, // Phase 14 B.3
			quotaSvc,
			mcpserver.Options{
				UserRPM:             60,
				OAuthValidator:      oauthSrvSvc,
				ResourceMetadataURL: resourceMetadataURL,
			},
		)
	}

	// Payment provider
	var paymentProvider payment.PaymentProvider
	if cfg.Payment.Enabled && cfg.Payment.TBankTerminalKey != "" {
		paymentProvider = tbank.NewProvider(tbank.Config{
			TerminalKey: cfg.Payment.TBankTerminalKey,
			Password:    cfg.Payment.TBankPassword,
			BaseURL:     cfg.Payment.TBankBaseURL,
		})
		slog.Info("payment.tbank.enabled")
	}

	subscriptionSvc := subscriptionuc.NewService(
		subscriptionRepo, planRepo, paymentRepo, userRepo,
		paymentProvider, &cfg.Payment,
	)

	purgeLoop := trashuc.NewPurgeLoop(trashRepo, 1*time.Hour, 30)

	// Email-уведомления для subscription loops. Если SMTP не сконфигурирован
	// (Configured()=false), notifier=nil — loops сами пропускают отправку.
	var subNotifier *subscriptionuc.EmailNotifier
	if emailSvc.Configured() {
		subNotifier = subscriptionuc.NewEmailNotifier(emailSvc, cfg.Server.FrontendURL)
	}

	expirationLoop := subscriptionuc.NewExpirationLoop(subscriptionRepo, userRepo, subNotifier, 1*time.Hour)
	// renewalLoop пытается продлить подписки за 48ч до конца периода;
	// если payment не настроен — Start() сам no-op'ит.
	renewalLoop := subscriptionuc.NewRenewalLoop(
		subscriptionRepo, planRepo, paymentRepo, userRepo,
		paymentProvider, subNotifier, &cfg.Payment,
		1*time.Hour, 48*time.Hour,
	)
	// reminderLoop — pre-expire напоминания для auto_renew=false подписок (M-5b).
	// Тикер 6ч: окна 3/1 день ловятся с запасом, спама нет (stage-флаг).
	reminderLoop := subscriptionuc.NewReminderLoop(subscriptionRepo, userRepo, subNotifier, 6*time.Hour)

	// reengagementLoop — письмо юзерам неактивным 14+ дней (M-5d). Раз в день.
	reengagementLoop := engagementuc.NewReengagementLoop(userRepo, emailSvc, cfg.Server.FrontendURL, 24*time.Hour)

	// streakReminderLoop — "не сломай серию" для юзеров со streak > 3 (M-16).
	// Тик раз в день; внутри check today и skip если уже отправляли.
	streakReminderLoop := engagementuc.NewStreakReminderLoop(streakRepo, userRepo, emailSvc, cfg.Server.FrontendURL, 24*time.Hour)

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
		promptHandler: func() *prompthttp.Handler {
			h := prompthttp.NewHandler(promptSvc, quotaSvc)
			// Phase 14 B.4: activity+users → склейка в GET /api/prompts/:id/history
			h.SetHistoryDeps(activitySvc, userRepo)
			return h
		}(),
		collectionHandler: collhttp.NewHandler(collectionSvc),
		tagHandler:        taghttp.NewHandler(tagSvc),
		searchHandler:     searchhttp.NewHandler(searchSvc),
		teamHandler:       teamhttp.NewHandler(teamSvc),
		userHandler:       userhttp.NewHandler(userSvc),
		starterHandler:    starterhttp.NewHandler(starterSvc),
		trashHandler:      trashhttp.NewHandler(trashSvc),
		apiKeyHandler:     apikeyhttp.NewHandler(apiKeySvc, teamSvc, cfg.MCP.MaxKeysPerUser),
		shareHandler:      sharehttp.NewHandler(shareSvc),
		streakHandler:     streakhttp.NewHandler(streakSvc),
		badgeHandler:      badgehttp.NewHandler(badgeSvc),
		feedbackHandler:   feedbackhttp.NewHandler(feedbackSvc),
		changelogHandler:      changeloghttp.NewHandler(changelogSvc),
		subscriptionHandler:  subscriptionhttp.NewHandler(subscriptionSvc, quotaSvc),
		webhookHandler:       webhookhttp.NewHandler(subscriptionSvc),
		seoHandler:           seohttp.NewHandler(promptSvc, cfg.Server.FrontendURL),
		// Phase 14 B.4: analytics + team activity HTTP handlers.
		// H5: plan-check вынесен в analytics.Service, handler больше не
		// нуждается в UserRepository.
		analyticsHandler:    analyticshttp.NewHandler(analyticsSvc),
		teamActivityHandler: teamhttp.NewActivityHandler(teamSvc, activitySvc),
		// Phase 14 D: team branding handler (GET/PUT /api/teams/:slug/branding)
		teamBrandingHandler: teamhttp.NewBrandingHandler(teamSvc),
		oauthServerHandler: oauthsrvhttp.NewHandler(
			oauthSrvSvc,
			func(ctx context.Context, refreshToken string) (uint, error) {
				// Переиспользуем auth.Service.Refresh: валидируем refresh cookie
				// и возвращаем userID. Side-effect (rotation пары токенов)
				// для OAuth-authorize flow не важен — Claude получит свой
				// OAuth access token, а не refresh пользовательской сессии.
				user, _, err := authSvc.Refresh(ctx, refreshToken)
				if err != nil {
					return 0, err
				}
				return user.ID, nil
			},
			cfg.Server.FrontendURL,
		),
		metadataHandler: metadatahttp.NewHandler(metadatahttp.Config{
			Issuer:         strings.TrimRight(cfg.Server.FrontendURL, "/"),
			ResourceServer: canonicalResource,
		}),
		mcpServer:            mcpSrv,
		expirationLoop:       expirationLoop,
		renewalLoop:          renewalLoop,
		reminderLoop:         reminderLoop,
		reengagementLoop:     reengagementLoop,
		streakReminderLoop:   streakReminderLoop,
		activityCleanupLoop:  activityCleanupLoop,
		insightsLoop:         insightsLoop,
		purgeLoop:         purgeLoop,
		feedbackRL:        ratelimit.NewLimiterWithWindow[uint](5, time.Hour, ratelimit.UintHash),
		insightsRL:        ratelimit.NewLimiterWithWindow[uint](1, time.Hour, ratelimit.UintHash),
	}
}

func (a *App) StartBackground() {
	a.purgeLoop.Start()
	a.expirationLoop.Start()
	a.renewalLoop.Start()
	a.reminderLoop.Start()
	a.reengagementLoop.Start()
	a.streakReminderLoop.Start()
	a.activityCleanupLoop.Start()
	a.insightsLoop.Start()
}

// Shutdown waits for background tasks to complete.
func (a *App) Shutdown(timeout time.Duration) {
	a.purgeLoop.Stop()
	a.expirationLoop.Stop()
	a.renewalLoop.Stop()
	a.reminderLoop.Stop()
	a.reengagementLoop.Stop()
	a.streakReminderLoop.Stop()
	a.activityCleanupLoop.Stop()
	a.insightsLoop.Stop()
	a.feedbackRL.Close()
	a.authSvc.WaitBackground(timeout)
}

