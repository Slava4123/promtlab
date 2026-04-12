package config

type MCPConfig struct {
	Enabled        bool `koanf:"enabled"`
	MaxKeysPerUser int  `koanf:"max_keys_per_user"`
}
