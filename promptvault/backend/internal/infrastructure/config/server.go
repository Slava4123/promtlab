package config

type ServerConfig struct {
	Port           string   `koanf:"port"`
	Environment    string   `koanf:"environment"`
	AllowedOrigins []string `koanf:"allowed_origins"`
	FrontendURL    string   `koanf:"frontend_url"`
	SecureCookies  bool     `koanf:"secure_cookies"`
	// TrustProxy — доверять X-Forwarded-For/X-Real-IP для определения клиентского IP.
	// Ставить true ТОЛЬКО если backend за доверенным reverse-proxy (nginx/cloudflare),
	// который затирает incoming XFF. Иначе атакующий подменяет заголовок и обходит rate-limit.
	TrustProxy bool `koanf:"trust_proxy"`
}

func (c ServerConfig) IsDev() bool {
	return c.Environment == "development"
}

func (c ServerConfig) IsProd() bool {
	return c.Environment == "production"
}
