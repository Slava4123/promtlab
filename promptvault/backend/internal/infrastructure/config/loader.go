package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func Load() (*Config, error) {
	k := koanf.New(".")

	_ = k.Load(confmap(defaults()), nil)

	_ = k.Load(file.Provider(".env"), dotenv.Parser()) // .env не найден — не критично

	_ = k.Load(env.Provider("", ".", func(s string) string {
		s = strings.ToLower(s)

		// oauth.{github,google,yandex}.xxx — два уровня вложенности
		for _, p := range []string{"oauth_github_", "oauth_google_", "oauth_yandex_"} {
			if strings.HasPrefix(s, p) {
				parts := strings.SplitN(p[:len(p)-1], "_", 2)
				return parts[0] + "." + parts[1] + "." + s[len(p):]
			}
		}

		// section.xxx — первый _ разделяет секцию от ключа
		if i := strings.Index(s, "_"); i > 0 {
			return s[:i] + "." + s[i+1:]
		}
		return s
	}), nil)

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal failed: %w", err)
	}

	// Comma-separated origins: "https://a.ru,https://b.ru" → []string
	if len(cfg.Server.AllowedOrigins) == 1 && strings.Contains(cfg.Server.AllowedOrigins[0], ",") {
		cfg.Server.AllowedOrigins = strings.Split(cfg.Server.AllowedOrigins[0], ",")
		for i, o := range cfg.Server.AllowedOrigins {
			cfg.Server.AllowedOrigins[i] = strings.TrimSpace(o)
		}
	}

	// Comma-separated IP/CIDR: "212.233.80.7,91.194.226.0/23" → []string.
	// Поддерживаем тот же формат что и AllowedOrigins — коаnf даёт один элемент со всей строкой.
	if len(cfg.Payment.WebhookAllowedIPs) == 1 && strings.Contains(cfg.Payment.WebhookAllowedIPs[0], ",") {
		cfg.Payment.WebhookAllowedIPs = strings.Split(cfg.Payment.WebhookAllowedIPs[0], ",")
		for i, ip := range cfg.Payment.WebhookAllowedIPs {
			cfg.Payment.WebhookAllowedIPs[i] = strings.TrimSpace(ip)
		}
	}

	// Production safety checks
	if cfg.Server.IsProd() {
		if cfg.JWT.Secret == "dev-secret-change-me" {
			return nil, fmt.Errorf("JWT_SECRET must be changed in production")
		}
		if cfg.Server.FrontendURL == "http://localhost:5173" {
			return nil, fmt.Errorf("SERVER_FRONTEND_URL must be configured in production")
		}
		for _, origin := range cfg.Server.AllowedOrigins {
			if origin == "*" {
				return nil, fmt.Errorf("wildcard CORS origin (*) is not allowed in production")
			}
			// chrome-extension:// origins разрешены и в prod: расширение
			// аутентифицируется API-ключом, CORS здесь — формальный preflight.
			if strings.HasPrefix(origin, "chrome-extension://") {
				continue
			}
			if !strings.HasPrefix(origin, "https://") {
				return nil, fmt.Errorf("CORS origin %q must use HTTPS in production", origin)
			}
		}
	}

	// Sentry safety — если включён, DSN обязателен (иначе silent no-op в SDK,
	// что маскирует misconfiguration). Проверка применяется во всех env.
	if cfg.Sentry.Enabled && cfg.Sentry.Dsn == "" {
		return nil, fmt.Errorf("SENTRY_ENABLED=true but SENTRY_DSN is empty")
	}

	// Payment safety — если биллинг включён, все T-Bank ключи обязательны.
	// Fail-fast предотвращает запуск с броken-конфигом (юзер видел бы 501 на Checkout).
	if err := cfg.Payment.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// defaults возвращает вложенную map — формат должен совпадать с тем,
// что возвращает env.Provider (через Unflatten), иначе maps.Merge
// не сможет корректно объединить плоские ключи defaults с вложенными
// ключами из env, и значения без соответствующей env-переменной потеряются.
func defaults() map[string]any {
	return map[string]any{
		"server": map[string]any{
			"port":            "8080",
			"environment":     "development",
			"allowed_origins": []string{"http://localhost:5173"},
			"frontend_url":    "http://localhost:5173",
		},
		"database": map[string]any{
			"host":           "localhost",
			"port":           5432,
			"user":           "postgres",
			"password":       "postgres",
			"name":           "promptvault",
			"sslmode":        "disable",
			"sslrootcert":    "",
			"max_open_conns": 25,
			"max_idle_conns": 5,
		},
		"jwt": map[string]any{
			"secret":           "dev-secret-change-me",
			"access_duration":  "15m",
			"refresh_duration": "168h",
		},
		"oauth": map[string]any{
			"callback_base": "http://localhost:8080",
		},
		"ai": map[string]any{
			"openrouter_api_key":       "",
			"openrouter_base_url":      "https://openrouter.ai/api/v1",
			"openrouter_timeout_seconds": 300,
			"rate_limit_rpm":           10,
		},
		"sentry": map[string]any{
			"enabled":            false,
			"dsn":                "",
			"environment":        "development",
			"release":            "",
			"traces_sample_rate": 0.0,
			"debug":              false,
		},
		"mcp": map[string]any{
			"enabled":          false,
			"max_keys_per_user": 5,
		},
		"payment": map[string]any{
			"enabled":              false,
			"tbank_terminal_key":   "",
			"tbank_password":       "",
			"tbank_base_url":       "https://securepay.tinkoff.ru/v2",
			"webhook_base_url":     "",
			"success_url":          "/settings?payment=success",
			"fail_url":             "/settings?payment=failure",
			"receipt_enabled":      false,
			"taxation":             "usn_income",
			"recurrent_enabled":    true,
			"webhook_allowed_ips":  []string{},
			"webhook_trust_xff":    false,
		},
	}
}

type confmap map[string]any

func (c confmap) ReadBytes() ([]byte, error) { return nil, nil }
func (c confmap) Read() (map[string]any, error) {
	return map[string]any(c), nil
}
