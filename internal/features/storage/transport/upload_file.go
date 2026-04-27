package transport

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"GopherDrive/internal/pkg/api/logger"
	"GopherDrive/internal/pkg/api/response"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	const op = "storage.transport.UploadFile"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var file core_domain.File

	userID, ok := getUserIDFromContext(log, w, r)
	if !ok {
		return
	}

	file.UserID = userID

	reader, err := r.MultipartReader()
	if err != nil {
		log.Info(fmt.Sprintf("Expected multipart/form-data, got %s", r.Header.Get("Content-Type")))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("Expected multipart/form-data"))
		return
	}

	newFile, err := h.readMultipartAndUploadFile(r, reader, file)
	if err != nil {
		switch {
		case errors.Is(err, ErrReadMultipartReader):
			log.Error("failed to read part of MultipartReader", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to read multipart/form-data"))
		case errors.Is(err, ErrReadFolderIDMultipart):
			log.Error("failed to read folder id", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to read folder_id"))
		case errors.Is(err, ErrInvalidFolderID):
			log.Info("invalid folder id", logger.SlogErr(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid folder_id"))
		case errors.Is(err, core_errors.ErrStorageCleanupFailed):
			log.Error("failed to cleanup storage after database error", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to upload file"))
		case errors.Is(err, core_errors.ErrFileAlreadyExists):
			log.Info("file already exists", slog.String("name", file.Name))

			render.Status(r, http.StatusConflict)
			render.JSON(w, r, response.Error("file already exists"))
		case errors.Is(err, core_errors.ErrUserNotFound):
			log.Info("user not found", slog.String("user_id", file.UserID.String()))

			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("user account no longer exists"))
		case errors.Is(err, core_errors.ErrFolderNotFound):
			log.Info("folder not found", slog.String("folder_id", file.FolderID.String()))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error("folder not found"))
		case errors.Is(err, ErrExtraDataAfterFile):
			log.Info(err.Error())

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("unexpected data after file"))
		case errors.Is(err, core_errors.ErrStorageCleanupFailedAfterDirtyFile):
			log.Error("failed to cleanup storage after dirty file", logger.SlogErr(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("unexpected data after file"))
		case errors.Is(err, ErrUnknownField):
			log.Info(err.Error())

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("unknown field in file"))
		case errors.Is(err, ErrMissingFilePart):
			log.Info("file part is missing")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("file part is missing"))

			log.Info("empty file name", logger.SlogErr(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("file name is empty"))
		default:
			log.Error("failed to upload file", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to upload file"))
		}
		return
	}

	log.Info("file read and created")

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, CreateFileResponse(response.OK(), newFile))
}

func (h *Handler) readMultipartAndUploadFile(
	r *http.Request,
	reader *multipart.Reader,
	file core_domain.File,
) (newFile core_domain.File, err error) {
	var fileProcessed bool

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return newFile, fmt.Errorf("%w: %w", ErrReadMultipartReader, err)
		}

		if fileProcessed {
			_, err = h.service.DeleteFile(context.WithoutCancel(r.Context()), newFile.UserID, newFile.ID)
			if err != nil {
				return newFile, fmt.Errorf("%w: %w", core_errors.ErrStorageCleanupFailed, err)
			}

			return newFile, fmt.Errorf("%w: %s", ErrExtraDataAfterFile, part.FormName())
		}

		switch part.FormName() {
		case "folder_id":
			rowID, err := io.ReadAll(io.LimitReader(part, 64))
			if err != nil {
				return newFile, fmt.Errorf("%w: %w", ErrReadFolderIDMultipart, err)
			}

			strID := strings.TrimSpace(string(rowID))
			if strID == "" {
				continue
			}

			folderID, err := uuid.Parse(strID)
			if err != nil {
				return newFile, fmt.Errorf("%w: %w", ErrInvalidFolderID, err)
			}

			file.FolderID = &folderID
			continue

		case "file":
			contentType := part.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			file.Name = part.FileName()
			file.ContentType = contentType

			createdFile, err := h.service.CreateFile(r.Context(), part, file)
			if err != nil {
				return newFile, err
			}

			newFile = createdFile
			fileProcessed = true
		default:
			return newFile, fmt.Errorf("%w: %s", ErrUnknownField, part.FormName())
		}
	}

	if !fileProcessed {
		return newFile, ErrMissingFilePart
	}

	return newFile, nil
}
