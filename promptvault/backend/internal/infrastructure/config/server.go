package config

type ServerConfig struct {
	Port           string   `koanf:"port"`
	Environment    string   `koanf:"environment"`
	AllowedOrigins []string `koanf:"allowed_origins"`
	FrontendURL    string   `koanf:"frontend_url"`
	SecureCookies  bool     `koanf:"secure_cookies"`
}

func (c ServerConfig) IsDev() bool {
	return c.Environment == "development"
}

func (c ServerConfig) IsProd() bool {
	return c.Environment == "production"
}
