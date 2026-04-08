package auth

import (
	"GopherDrive/internal/lib/api/response"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	response.Response
	ID uuid.UUID
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) Handler {
	return Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handler.Register"

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
		log.Info("failed to decode request body", response.SlogErr(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.Any("req", req))

	id, err := h.service.Register(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrUserAlreadyExists) {
		log.Info("user already exists", slog.String("email", req.Email))

		render.Status(r, http.StatusConflict)
		render.JSON(w, r, response.Error("user already exists"))
		return
	}
	if err != nil {
		log.Error("failed to register user", response.SlogErr(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to register user"))
		return
	}

	log.Info("user registered", slog.Any("id", id))

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, Response{
		Response: response.OK(),
		ID:       id,
	})
}
