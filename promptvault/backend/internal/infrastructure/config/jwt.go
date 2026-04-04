package config

type JWTConfig struct {
	Secret          string `koanf:"secret"`
	AccessDuration  string `koanf:"access_duration"`
	RefreshDuration string `koanf:"refresh_duration"`
}
