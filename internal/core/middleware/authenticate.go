package middleware

import (
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"GopherDrive/internal/pkg/tokens"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

func Authenticate(log *slog.Logger, secretToken string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		wrappedFunc := func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.Authenticate"

			log = log.With(
				slog.String("op", op),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Info("missing authorization header")

				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error("missing authorization header"))
				return
			}

			authHeaderSplit := strings.Split(authHeader, " ")
			if len(authHeaderSplit) != 2 || authHeaderSplit[0] != "Bearer" {
				log.Info("invalid authorization format")

				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error("invalid authorization format"))
				return
			}

			id, err := tokens.ParseAccessToken(authHeaderSplit[1], secretToken)
			if errors.Is(err, jwt.ErrTokenExpired) {
				log.Info("token has expired")

				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error("token expired"))
				return
			}
			if err != nil {
				log.Info("invalid token", logger.SlogErr(err))

				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error("invalid token"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, id)
			newRequest := r.WithContext(ctx)

			next.ServeHTTP(w, newRequest)
		}

		return http.HandlerFunc(wrappedFunc)
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}
	return id, nil
}
