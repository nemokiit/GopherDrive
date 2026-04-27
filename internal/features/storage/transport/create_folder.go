package transport

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	const op = "storage.transport.CreateFolder"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var folder core_domain.Folder

	userID, ok := getUserIDFromContext(log, w, r)
	if !ok {
		return
	}
	folder.UserID = userID

	var req CreateFolderRequest
	ok = decodeJSON(r.Body, &req, log, w, r)
	if !ok {
		return
	}
	folder.Name = req.Name

	log.Info("request body decoded", slog.String("folder_id", req.ParentFolderID))

	if req.ParentFolderID != "" {
		folderID, ok := parseUUID(req.ParentFolderID, log, w, r)
		if !ok {
			return
		}
		folder.ParentFolderID = &folderID
	}

	createdFolder, err := h.service.CreateFolder(r.Context(), folder)
	if err != nil {
		switch {
		case errors.Is(err, core_errors.ErrFolderAlreadyExists):
			log.Info("folder already exists", slog.String("name", folder.Name))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("folder already exists"))
		case errors.Is(err, core_errors.ErrUserNotFound):
			log.Info("user not found", slog.String("user_id", folder.UserID.String()))

			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("user account no longer exists"))
		case errors.Is(err, core_errors.ErrParentFolderNotFound):
			log.Info("parent folder not found", slog.String("folder_id", folder.ParentFolderID.String()))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("parent folder not found"))
		default:
			log.Error("failed to create folder", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to create folder"))
		}
		return
	}

	log.Info("folder created", slog.String("folder_id", createdFolder.ID.String()))

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, CreateFolderResponse(response.OK(), folder))
}
