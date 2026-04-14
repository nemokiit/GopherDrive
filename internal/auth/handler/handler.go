package handler

import (
	"GopherDrive/internal/auth"
	"GopherDrive/internal/lib/api/logger"
	"GopherDrive/internal/lib/api/response"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Service interface {
	Register(ctx context.Context, email string, password string) (uuid.UUID, string, string, error)
	Login(ctx context.Context, email string, password string) (uuid.UUID, string, string, error)
	Refresh(ctx context.Context, cookie string) (uuid.UUID, string, string, error)
}

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	response.Response
	AccessToken string    `json:"access_token"`
	ID          uuid.UUID `json:"id"`
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func New(service Service, log *slog.Logger) *Handler {
	return &Handler{
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
		log.Info("failed to decode request body", logger.SlogErr(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.Any("email", req.Email))

	id, accessToken, refreshToken, err := h.service.Register(r.Context(), req.Email, req.Password)
	if errors.Is(err, auth.ErrUserAlreadyExists) {
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

	log.Info("user registered", slog.Any("id", id))

	cookie := getCookie("refresh_token", refreshToken, 24*60*60)
	http.SetCookie(w, cookie)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, Response{
		Response:    response.OK(),
		AccessToken: accessToken,
		ID:          id,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handler.Login"

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

	log.Info("request body decoded", slog.Any("email", req.Email))

	id, accessToken, refreshToken, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			log.Info("user not found", slog.String("email", req.Email))
		} else if errors.Is(err, auth.ErrInvalidCredentials) {
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

	log.Info("user logged", slog.Any("id", id))

	cookie := getCookie("refresh_token", refreshToken, 24*60*60)
	http.SetCookie(w, cookie)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, Response{
		Response:    response.OK(),
		AccessToken: accessToken,
		ID:          id,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handler.Refresh"

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
	if errors.Is(err, auth.ErrTokenNotFound) {
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

	cookie := getCookie("refresh_token", refreshToken, 24*60*60)
	http.SetCookie(w, cookie)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, Response{
		Response:    response.OK(),
		AccessToken: accessToken,
		ID:          id,
	})
}

func getCookie(name string, value string, ttl int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   ttl,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
