package transport

import (
	core_errors "GopherDrive/internal/core/errors"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	const op = "auth.transport.Refresh"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	cookieReq, err := r.Cookie("refresh_token")
	if err != nil {
		log.Info("cookie missing", logger.SlogErr(err))

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("missing refresh token"))
		return
	}

	id, accessToken, refreshToken, err := h.service.Refresh(r.Context(), cookieReq.Value)
	if errors.Is(err, core_errors.ErrTokenNotFound) {
		log.Info("unauthorized refresh token")

		deleteCookie := &http.Cookie{
			Name:   "refresh_token",
			MaxAge: -1,
			Path:   "/",
		}

		http.SetCookie(w, deleteCookie)

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("unauthorized cookie"))
		return
	}
	if err != nil {
		log.Error("failed to refresh user", logger.SlogErr(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to refresh user"))
		return
	}

	log.Info("user refreshed", slog.Any("id", id))

	cookie := createCookie("refresh_token", refreshToken, 24*60*60)
	http.SetCookie(w, cookie)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, CreateRefreshTokenResponse(response.OK(), accessToken))
}
