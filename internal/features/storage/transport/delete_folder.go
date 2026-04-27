package transport

import (
	core_errors "GopherDrive/internal/core/errors"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	const op = "storage.transport.DeleteFolder"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	userID, ok := getUserIDFromContext(log, w, r)
	if !ok {
		return
	}

	strFolderID := chi.URLParam(r, "id")
	folderID, ok := parseUUID(strFolderID, log, w, r)
	if !ok {
		return
	}

	files, err := h.service.DeleteFolder(r.Context(), userID, folderID)
	if errors.Is(err, core_errors.ErrDeleteCleanupFailed) {
		log.Error("failed to cleanup storage", logger.SlogErr(err))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, DeleteFolderResponse{
			Response: response.OK(),
			Files:    files,
		})
		return
	}
	if err != nil {
		log.Error("failed to delete folder", slog.String("folder_id", folderID.String()))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to delete folder"))
		return
	}

	log.Info("folder deleted", slog.String("folder_id", folderID.String()))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, CreateDeleteFolderResponse(response.OK(), files))
}
