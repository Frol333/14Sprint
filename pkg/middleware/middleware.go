package middleware

import (
	"net/http"

	"github.com/Frol333/14Sprint/pkg/service"
	"github.com/rs/zerolog/log"
)

var (
	CookieHeader = "token"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("uri", r.RequestURI).Str("method", r.Method).Send()
		next.ServeHTTP(w, r)
	})
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()
		for _, c := range cookies {
			if c.Name == "token" {
				ok, err := service.CheckAuth(r.Context(), c.Value)
				if err != nil {
					http.Error(w, "Failed to read token", http.StatusUnauthorized)
					return
				}
				if !ok {
					http.Error(w, "password is not exist", http.StatusUnauthorized)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
