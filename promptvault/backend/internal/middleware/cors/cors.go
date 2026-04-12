package cors

import (
	"net/http"
	"strings"

	gocors "github.com/go-chi/cors"

	"promptvault/internal/infrastructure/config"
)

// Middleware настраивает CORS для SPA + Chrome Extension.
//
// Chrome Extension имеет origin вида `chrome-extension://<id>`, где <id> — случайный
// hash генерируемый при установке. Пропускаем любой `chrome-extension://*` через
// AllowOriginFunc (расширение авторизуется API-ключом, не cookie — запрос
// незащищён CSRF).
func Middleware(cfg *config.Config) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(cfg.Server.AllowedOrigins))
	for _, o := range cfg.Server.AllowedOrigins {
		allowed[o] = struct{}{}
	}

	return gocors.Handler(gocors.Options{
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			// Chrome/Edge extension origins — trusted через Bearer API-key.
			if strings.HasPrefix(origin, "chrome-extension://") {
				return true
			}
			_, ok := allowed[origin]
			return ok
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Timezone", "X-Client"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
