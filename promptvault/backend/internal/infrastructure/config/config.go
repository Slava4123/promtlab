package config

type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	JWT      JWTConfig      `koanf:"jwt"`
	OAuth    OAuthConfig    `koanf:"oauth"`
	SMTP     SMTPConfig     `koanf:"smtp"`
	AI       AIConfig       `koanf:"ai"`
	Sentry   SentryConfig   `koanf:"sentry"`
	MCP      MCPConfig      `koanf:"mcp"`
}
