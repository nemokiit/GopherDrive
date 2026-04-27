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

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	const op = "storage.transport.DeleteFile"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	userID, ok := getUserIDFromContext(log, w, r)
	if !ok {
		return
	}

	strFileID := chi.URLParam(r, "id")
	fileID, ok := parseUUID(strFileID, log, w, r)
	if !ok {
		return
	}

	file, err := h.service.DeleteFile(r.Context(), userID, fileID)
	if err != nil {
		switch {
		case errors.Is(err, core_errors.ErrDeleteCleanupFailed):
			log.Error("failed to cleanup storage", logger.SlogErr(err))

			render.Status(r, http.StatusOK)
			render.JSON(w, r, CreateFileResponse(response.OK(), file))
		case errors.Is(err, core_errors.ErrFileNotFound):
			log.Info("file not found", slog.String("file_id", fileID.String()))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("file not found"))
		default:
			log.Error("failed to delete file", slog.String("file_id", fileID.String()))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to delete file"))
		}
		return
	}

	log.Info("file deleted", slog.String("filed_id", file.ID.String()))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, CreateFileResponse(response.OK(), file))
}
