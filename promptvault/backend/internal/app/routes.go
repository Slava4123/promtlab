package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"promptvault/internal/infrastructure/metrics"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/middleware/ipallowlist"
	"promptvault/internal/middleware/ratelimit"
	adminmw "promptvault/internal/middleware/admin"
	sentrymw "promptvault/internal/middleware/sentry"
	"log/slog"
)

func (a *App) MountRoutes(r chi.Router) {
	// Protected routes принимают и JWT (SPA), и API-ключ `pvlt_*` (Chrome Extension, MCP-клиенты).
	// Префикс токена определяет путь валидации.
	authMiddleware := authmw.CombinedAuth(a.tokenValidator, a.apiKeyValidator)

	// byIP — shortcut с глобальным trust_proxy флагом. Без доверенного reverse-proxy
	// (nginx с real_ip_header) XFF/X-Real-IP могут быть подделаны клиентом → обход rate-limit.
	trustProxy := a.cfg.Server.TrustProxy
	byIP := func(rpm int) func(http.Handler) http.Handler {
		return ratelimit.ByIP(rpm, trustProxy)
	}

	// SEO endpoints (outside /api): /sitemap.xml для поисковиков. Rate-limit 60/IP —
	// типичный crawl-rate для bot-ов, защита от DoS на статичный список.
	// HEAD регистрируем явно: Google Search Console + Yandex Webmaster проверяют
	// sitemap через HEAD перед GET. Без HEAD получают 405 → «Недопустимый адрес».
	r.With(byIP(60)).Method(http.MethodGet, "/sitemap.xml", http.HandlerFunc(a.seoHandler.Sitemap))
	r.With(byIP(60)).Method(http.MethodHead, "/sitemap.xml", http.HandlerFunc(a.seoHandler.Sitemap))

	// Prometheus /metrics — text exposition. 404 если METRICS_ENABLED=false.
	// Защита через middleware ipallowlist (см. SERVER_METRICS_ALLOWLIST):
	// scrape Prometheus идёт изнутри Docker network минуя nginx, поэтому
	// trustForwarded=false — XFF на этом пути не имеет смысла.
	// Scrape rate-limit не нужен — endpoint кешируется в client_golang registry.
	r.With(ipallowlist.New(a.cfg.Server.MetricsAllowlist, false)).
		Method(http.MethodGet, "/metrics", metrics.Handler(a.cfg.Server.MetricsEnabled))

	// /p/{slug} — server-rendered HTML. nginx роутит сюда ТОЛЬКО bot-UA
	// (Yandexbot/Googlebot/Telegram/VK/...). Обычные юзеры получают SPA.
	// Без bot-routing этот endpoint всё равно будет отдаваться по прямому
	// curl/wget запросу — это OK (валидный HTML, индексируется).
	// HEAD аналогично — соц-парсеры превью часто HEAD-ят перед GET.
	r.With(byIP(60)).Method(http.MethodGet, "/p/{slug}", http.HandlerFunc(a.seoHandler.PromptHTML))
	r.With(byIP(60)).Method(http.MethodHead, "/p/{slug}", http.HandlerFunc(a.seoHandler.PromptHTML))

	// MCP endpoint (outside /api, with pre-auth IP rate limit)
	if a.mcpServer != nil {
		mcpHandler := byIP(120)(a.mcpServer.Handler())
		r.Method(http.MethodPost, "/mcp", mcpHandler)
		r.Method(http.MethodGet, "/mcp", mcpHandler)
		r.Method(http.MethodDelete, "/mcp", mcpHandler)
		slog.Info("mcp.server.mounted", "path", "/mcp")
	}

	// OAuth 2.1 metadata endpoints (RFC 9728 + RFC 8414). Публичные, без auth.
	// Rate-limit 60/IP — типичный discovery-rate для клиентов.
	r.With(byIP(60)).Get("/.well-known/oauth-protected-resource", a.metadataHandler.ProtectedResource)
	r.With(byIP(60)).Get("/.well-known/oauth-authorization-server", a.metadataHandler.AuthorizationServer)

	// OAuth 2.1 Authorization Server endpoints.
	// /register, /token, /revoke — публичные (RFC 7591/6749/7009), PKCE защищает code exchange.
	// /authorize — требует JWT-сессии залогиненного пользователя.
	r.Route("/oauth", func(r chi.Router) {
		r.Use(byIP(30)) // защита от brute-force /token
		r.Post("/register", a.oauthServerHandler.Register)
		r.Post("/token", a.oauthServerHandler.Token)
		r.Post("/revoke", a.oauthServerHandler.Revoke)
		// /oauth/authorize НЕ под authMiddleware: браузерный OAuth-flow
		// приносит сессию через refresh_token HttpOnly cookie, а не через
		// Authorization: Bearer. Handler читает cookie сам и редиректит
		// на /sign-in?return_url=... если сессии нет.
		r.Get("/authorize", a.oauthServerHandler.Authorize)
	})
	slog.Info("oauth.server.mounted", "endpoints", []string{
		"/oauth/register", "/oauth/token", "/oauth/authorize", "/oauth/revoke",
		"/.well-known/oauth-protected-resource", "/.well-known/oauth-authorization-server",
	})

	r.Route("/api", func(r chi.Router) {
		// public — plans (rate limited: 60 req/min per IP)
		r.Route("/plans", func(r chi.Router) {
			r.Use(byIP(60))
			r.Get("/", a.subscriptionHandler.ListPlans)
		})

		// public — share links (rate limited: 60 req/min per IP)
		r.Route("/s", func(r chi.Router) {
			r.Use(byIP(60))
			r.Get("/{token}", a.shareHandler.GetPublic)
		})

		// public — public prompts по slug (SEO). 60 req/min — типичный seo-crawl rate.
		r.Route("/public/prompts", func(r chi.Router) {
			r.Use(byIP(60))
			r.Get("/{slug}", a.promptHandler.GetPublic)
		})

		// OG-image generation для социальных превью (Telegram/VK/Twitter Cards).
		// ETag-cache 24h: 99% запросов после прогрева — 304 Not Modified без рендера.
		// HEAD: соц-парсеры (Twitter Card validator и т.д.) проверяют размер
		// картинки HEAD-запросом перед GET — без поддержки получают 405.
		r.Route("/og/prompts", func(r chi.Router) {
			r.Use(byIP(60))
			r.Method(http.MethodGet, "/{slug}", http.HandlerFunc(a.seoHandler.OGImage))
			r.Method(http.MethodHead, "/{slug}", http.HandlerFunc(a.seoHandler.OGImage))
		})

		// public — webhooks. T-Bank шлёт 1-5 уведомлений за цикл платежа;
		// 30 req/min per IP с запасом покрывает retry-поведение банка.
		// Защита от DoS на публичный endpoint без авторизации.
		// IP allowlist — defence-in-depth поверх SHA-256 подписи: даже если
		// атакующий получит Password терминала, он не сможет доставить webhook
		// с чужого IP. Пустой список — middleware no-op (dev/pre-prod).
		r.Route("/webhooks", func(r chi.Router) {
			r.Use(ipallowlist.New(a.cfg.Payment.WebhookAllowedIPs, a.cfg.Payment.WebhookTrustXFF))
			r.Use(byIP(30))
			r.Post("/tbank", a.webhookHandler.TBank)
		})

		// public — auth (rate limited: 20 req/min per IP)
		r.Route("/auth", func(r chi.Router) {
			r.Use(byIP(20))
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
			r.Use(byIP(60))
			r.Get("/auth/me", a.authHandler.Me)
			r.Post("/auth/set-password/initiate", a.authHandler.InitiateSetPassword)
			r.Post("/auth/set-password/confirm", a.authHandler.ConfirmSetPassword)
			r.Put("/auth/profile", a.authHandler.UpdateProfile)
			r.Put("/auth/password", a.authHandler.ChangePassword)
			// Phase 14 M-10: opt-in toggle Smart Insights email digest.
			r.Patch("/auth/notifications/insights", a.authHandler.SetInsightEmails)
			r.Get("/auth/linked-accounts", a.authHandler.LinkedAccounts)
			r.Delete("/auth/unlink/{provider}", a.authHandler.UnlinkProvider)
			r.Post("/auth/link/{provider}", a.oauthHandler.InitiateLink)
			r.Post("/auth/logout", a.authHandler.Logout)
			r.Get("/auth/referral", a.authHandler.Referral)

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

			// Teams
			r.Route("/teams", func(r chi.Router) {
				r.Get("/", a.teamHandler.List)
				r.Post("/", a.teamHandler.Create)
				r.Route("/{slug}", func(r chi.Router) {
					r.Get("/", a.teamHandler.GetBySlug)
					r.Put("/", a.teamHandler.Update)
					r.Delete("/", a.teamHandler.Delete)
					r.Get("/activity", a.teamActivityHandler.List) // Phase 14 B.4
				r.Get("/branding", a.teamBrandingHandler.Get)  // Phase 14 D
				r.Put("/branding", a.teamBrandingHandler.Set)  // Phase 14 D
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
				r.Get("/{id}/history", a.promptHandler.GetHistory) // Phase 14 B.4
				r.Post("/{id}/revert/{versionId}", a.promptHandler.RevertToVersion)
				r.Get("/{id}/share", a.shareHandler.Get)
				r.Post("/{id}/share", a.shareHandler.Create)
				r.Delete("/{id}/share", a.shareHandler.Delete)
			})

			// Phase 14 B.4: Analytics
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/personal", a.analyticsHandler.Personal)
				r.Get("/teams/{id}", a.analyticsHandler.Team)
				r.Get("/prompts/{id}", a.analyticsHandler.Prompt)
				r.Get("/insights", a.analyticsHandler.Insights)
				r.Get("/export", a.analyticsHandler.Export)
				// Phase 14.2: force-refresh insights (Max-only, 1 раз/час per user).
				r.With(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						uid := authmw.GetUserID(req.Context())
						if uid > 0 && !a.insightsRL.Allow(uid) {
							w.Header().Set("Content-Type", "application/json")
							w.Header().Set("Retry-After", "3600")
							w.WriteHeader(http.StatusTooManyRequests)
							_, _ = w.Write([]byte(`{"error":"Инсайты можно обновлять не чаще одного раза в час"}`))
							return
						}
						next.ServeHTTP(w, req)
					})
				}).Post("/insights/refresh", a.analyticsHandler.RefreshInsights)
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

			// Subscription
			r.Route("/subscription", func(r chi.Router) {
				r.Get("/", a.subscriptionHandler.GetSubscription)
				r.Get("/usage", a.subscriptionHandler.GetUsage)
				r.Get("/downgrade-preview", a.subscriptionHandler.GetDowngradePreview)
				r.Post("/checkout", a.subscriptionHandler.Checkout)
				r.Post("/cancel", a.subscriptionHandler.Cancel)
				r.Post("/pause", a.subscriptionHandler.Pause)
				r.Post("/resume", a.subscriptionHandler.Resume)
				r.Post("/downgrade", a.subscriptionHandler.Downgrade)
				r.Post("/auto-renew", a.subscriptionHandler.SetAutoRenew)
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
