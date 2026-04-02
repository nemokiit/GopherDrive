package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func NewLogger(log *slog.Logger) func(next http.HandlerFunc) http.Handler {
	return func(next http.HandlerFunc) http.Handler {
		log = log.With(slog.String("component", "middleware/logger"))
		log.Info("logger middleware enabled")

		wrappedFn := func(w http.ResponseWriter, r *http.Request) {
			wrappedW := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			timeNow := time.Now()

			defer func() {
				log.Info("request completed",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.String("request_id", middleware.GetReqID(r.Context())),
					slog.Int("status", wrappedW.Status()),
					slog.Int("bytes", wrappedW.BytesWritten()),
					slog.Duration("duration", time.Since(timeNow)),
				)
			}()

			next.ServeHTTP(wrappedW, r)
		}

		return http.HandlerFunc(wrappedFn)
	}
}
