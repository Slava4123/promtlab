package config

import "fmt"

type DatabaseConfig struct {
	Host         string `koanf:"host"`
	Port         int    `koanf:"port"`
	User         string `koanf:"user"`
	Password     string `koanf:"password"`
	Name         string `koanf:"name"`
	SSLMode      string `koanf:"sslmode"`
	SSLRootCert  string `koanf:"sslrootcert"`
	MaxOpenConns int    `koanf:"max_open_conns"`
	MaxIdleConns int    `koanf:"max_idle_conns"`
}

func (c DatabaseConfig) DSN() string {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
	if c.SSLRootCert != "" {
		dsn += "&sslrootcert=" + c.SSLRootCert
	}
	return dsn
}
