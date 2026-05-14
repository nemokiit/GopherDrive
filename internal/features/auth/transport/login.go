package transport

import (
	core_errors "GopherDrive/internal/core/errors"
	"GopherDrive/internal/core/repository/postgres/pool"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "auth.transport.Login"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req Request

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Info("request body is empty")

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("request is empty"))
		return
	}
	if err != nil {
		log.Info("failed to decode request body", logger.SlogErr(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.String("email", req.Email))

	user, accessToken, refreshToken, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, pool.ErrUserNotFound) {
			log.Info("user not found", slog.String("email", req.Email))
		} else if errors.Is(err, core_errors.ErrInvalidCredentials) {
			log.Info("invalid password")
		} else {
			log.Error("failed to login user", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to login user"))
			return
		}

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("invalid credentials"))
		return
	}

	log.Info("user logged", slog.Any("id", user.ID))

	cookie := createCookie("refresh_token", refreshToken, 24*60*60)
	http.SetCookie(w, cookie)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, CreateUserResponse(response.OK(), accessToken, user))
}
