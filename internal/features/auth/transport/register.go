package transport

import (
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "auth.transport.Register"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req Request

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Info("request body is empty")

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("request body empty"))
		return
	}
	if err != nil {
		log.Info("failed to decode request body", logger.SlogErr(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.Any("email", req.Email))

	user, accessToken, refreshToken, err := h.service.Register(r.Context(), req.Email, req.Password)
	if errors.Is(err, pool.ErrUserAlreadyExists) {
		log.Info("user already exists", slog.String("email", req.Email))

		render.Status(r, http.StatusConflict)
		render.JSON(w, r, response.Error("user already exists"))
		return
	}
	if err != nil {
		log.Error("failed to register user", logger.SlogErr(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to register user"))
		return
	}

	log.Info("user registered", slog.Any("id", user.ID))

	cookie := createCookie("refresh_token", refreshToken, 24*60*60)
	http.SetCookie(w, cookie)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, CreateUserResponse(response.OK(), accessToken, user))
}
