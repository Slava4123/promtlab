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

	k.Load(confmap(defaults()), nil)

	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		// .env не найден — не критично
	}

	k.Load(env.Provider("", ".", func(s string) string {
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
			if !strings.HasPrefix(origin, "https://") {
				return nil, fmt.Errorf("CORS origin %q must use HTTPS in production", origin)
			}
		}
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
			"openrouter_api_key": "",
			"rate_limit_rpm":     10,
		},
	}
}

type confmap map[string]any

func (c confmap) ReadBytes() ([]byte, error) { return nil, nil }
func (c confmap) Read() (map[string]any, error) {
	return map[string]any(c), nil
}
