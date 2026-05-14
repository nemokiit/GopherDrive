package transport

import (
	"GopherDrive/internal/core/repository/postgres/pool"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	const op = "storage.transport.DownloadFile"

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

	reader, file, err := h.service.DownloadFile(r.Context(), userID, fileID)
	if errors.Is(err, pool.ErrFileNotFound) {
		log.Info("file not found", slog.String("file_id", fileID.String()))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("file not found"))
		return
	}
	if err != nil {
		log.Error("failed to download file", slog.String("file_id", fileID.String()))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to download file"))
		return
	}
	defer func() { _ = reader.Close() }()

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.PathEscape(file.Name)))
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))

	_, err = io.Copy(w, reader)
	if err != nil {
		log.Error("failed to stream file to response", logger.SlogErr(err))
		return
	}

	log.Info("file downloaded", slog.String("file_id", file.ID.String()))
}
