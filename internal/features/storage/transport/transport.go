package transport

import (
	core_domain "GopherDrive/internal/core/domain"
	mwAuth "GopherDrive/internal/core/transport/http/middleware"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Service interface {
	CreateFile(ctx context.Context, reader io.Reader, file core_domain.File) (core_domain.File, error)
	DownloadFile(ctx context.Context, userID, fileID uuid.UUID) (io.ReadCloser, core_domain.File, error)
	DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (core_domain.File, error)
	CreateFolder(ctx context.Context, folder core_domain.Folder) (core_domain.Folder, error)
	GetFolderContent(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (core_domain.FolderContent, error)
	DeleteFolder(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]core_domain.File, error)
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func New(log *slog.Logger, service Service) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func getUserIDFromContext(log *slog.Logger, w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, err := mwAuth.GetUserID(r.Context())
	if err != nil {
		log.Info(err.Error())

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("user not found"))
		return uuid.Nil, false
	}

	return userID, true
}

func decodeJSON(closer io.ReadCloser, v interface{}, log *slog.Logger, w http.ResponseWriter, r *http.Request) bool {
	err := render.DecodeJSON(closer, v)
	if errors.Is(err, io.EOF) {
		log.Info("request body is empty")

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("request body empty"))
		return false
	}
	if err != nil {
		log.Info("failed to decode request body", logger.SlogErr(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("failed to decode request"))
		return false
	}

	return true
}

func parseUUID(uid string, log *slog.Logger, w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	newUID, err := uuid.Parse(uid)
	if err != nil {
		log.Info("invalid folder id", logger.SlogErr(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("invalid folder_id"))
		return uuid.Nil, false
	}

	return newUID, true
}
