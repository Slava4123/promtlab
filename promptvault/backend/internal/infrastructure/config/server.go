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
	// MetricsEnabled — включает /metrics endpoint (Prometheus text exposition).
	// По умолчанию false — endpoint возвращает 404. Endpoint без auth, защищён
	// IP-allowlist через MetricsAllowlist + middleware ipallowlist.
	MetricsEnabled bool `koanf:"metrics_enabled"`
	// MetricsAllowlist — IP/CIDR, которым разрешён доступ к /metrics.
	// Применяется ipallowlist middleware. Пустой список = no-op (всё пропускается).
	// Default покрывает Docker bridge networks (RFC 1918 диапазон 172.16.0.0/12)
	// + loopback — Prometheus scrape'ит api изнутри Docker network минуя nginx,
	// поэтому XFF-парсинг не нужен (trustForwarded=false в месте использования).
	MetricsAllowlist []string `koanf:"metrics_allowlist"`
}

func (c ServerConfig) IsDev() bool {
	return c.Environment == "development"
}

func (c ServerConfig) IsProd() bool {
	return c.Environment == "production"
}
