package cors

import (
	"net/http"

	gocors "github.com/go-chi/cors"

	"promptvault/internal/infrastructure/config"
)

func Middleware(cfg *config.Config) func(http.Handler) http.Handler {
	return gocors.Handler(gocors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
