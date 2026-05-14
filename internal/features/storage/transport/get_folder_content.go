package transport

import (
	"GopherDrive/internal/core/repository/postgres/pool"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (h *Handler) GetFolderContent(w http.ResponseWriter, r *http.Request) {
	const op = "storage.transport.GetFolderContent"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	userID, ok := getUserIDFromContext(log, w, r)
	if !ok {
		return
	}

	strFolderID := chi.URLParam(r, "id")

	var folderID *uuid.UUID
	if strFolderID != "" {
		id, ok := parseUUID(strFolderID, log, w, r)
		if !ok {
			return
		}
		folderID = &id
	}

	folderContent, err := h.service.GetFolderContent(r.Context(), userID, folderID)
	if errors.Is(err, pool.ErrFolderNotFound) {
		log.Info("folder not found", slog.Any("folder_id", folderID))

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, response.Error("folder not found"))
		return
	}
	if err != nil {
		log.Error("failed to get folder content", logger.SlogErr(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to get folder content"))
		return
	}

	log.Info("got folder content", slog.Any("folder_id", folderID))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, CreateFolderContentResponse(response.OK(), folderContent))
}
