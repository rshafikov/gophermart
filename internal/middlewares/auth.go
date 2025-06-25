package middlewares

import (
	"context"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type userService interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}

func Authenticater(jwtHandler security.JWTHandler, userService userService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			tokenHeader := r.Header.Get("Authorization")
			if tokenHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.Split(tokenHeader, "Bearer ")
			if len(token) != 2 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			payload, err := jwtHandler.ParseJWT(token[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			u, err := userService.GetByLogin(context.TODO(), payload.Subject)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			logger.L.Debug("user authenticated", zap.String("user", u.Login))
			ctx := context.WithValue(r.Context(), contextkeys.UserKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
