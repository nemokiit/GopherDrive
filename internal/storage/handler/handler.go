package handler

import (
	mwAuth "GopherDrive/internal/infra/middleware"
	"GopherDrive/internal/lib/api/logger"
	"GopherDrive/internal/lib/api/response"
	"GopherDrive/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Service interface {
	CreateFile(ctx context.Context, reader io.Reader, file *storage.File) (*storage.File, error)
	DownloadFile(ctx context.Context, userID, fileID uuid.UUID) (io.ReadCloser, *storage.File, error)
	DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (*storage.File, error)
	CreateFolder(ctx context.Context, folder *storage.Folder) (*storage.Folder, error)
	GetFolderContent(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (*storage.FolderContent, error)
	DeleteFolder(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]*storage.File, error)
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

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	const op = "storage.handler.UploadFile"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var file storage.File

	userID, err := mwAuth.GetUserID(r.Context())
	if err != nil {
		log.Info(err.Error())

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("user not found"))
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

	newFile, err := h.readMultipartAndUploadFile(r, reader, &file)
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
		case errors.Is(err, storage.ErrStorageCleanupFailed):
			log.Error("failed to cleanup storage after database error", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to upload file"))
		case errors.Is(err, storage.ErrFileAlreadyExists):
			log.Info("file already exists", slog.String("name", file.Name))

			render.Status(r, http.StatusConflict)
			render.JSON(w, r, response.Error("file already exists"))
		case errors.Is(err, storage.ErrUserNotFound):
			log.Info("user not found", slog.String("user_id", file.UserID.String()))

			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("user account no longer exists"))
		case errors.Is(err, storage.ErrFolderNotFound):
			log.Info("folder not found", slog.String("folder_id", file.FolderID.String()))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error("folder not found"))
		case errors.Is(err, ErrExtraDataAfterFile):
			log.Info(err.Error())

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("unexpected data after file"))
		case errors.Is(err, ErrStorageCleanupFailed):
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
		default:
			log.Error("failed to upload file", logger.SlogErr(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to upload file"))
		}
		return
	}

	log.Info("file read and created")

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, fillFileResponse(newFile))
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	const op = "storage.handler.DownloadFile"

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
	if errors.Is(err, storage.ErrFileNotFound) {
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

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	const op = "storage.handler.DeleteFile"

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
		case errors.Is(err, storage.ErrDeleteCleanupFailed):
			log.Error("failed to cleanup storage", logger.SlogErr(err))

			render.Status(r, http.StatusOK)
			render.JSON(w, r, fillFileResponse(file))
		case errors.Is(err, storage.ErrFileNotFound):
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
	render.JSON(w, r, fillFileResponse(file))
}

func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	const op = "storage.handler.CreateFolder"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var folder storage.Folder

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

	log.Info("request body decoded", slog.String("folder_id", req.ParentFolderID))

	if req.ParentFolderID != "" {
		folderID, ok := parseUUID(req.ParentFolderID, log, w, r)
		if !ok {
			return
		}
		folder.ParentFolderID = &folderID
	}

	createdFolder, err := h.service.CreateFolder(r.Context(), &folder)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrFolderAlreadyExists):
			log.Info("folder already exists", slog.String("name", folder.Name))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("folder already exists"))
		case errors.Is(err, storage.ErrUserNotFound):
			log.Info("user not found", slog.String("user_id", folder.UserID.String()))

			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("user account no longer exists"))
		case errors.Is(err, storage.ErrParentFolderNotFound):
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
	render.JSON(w, r, fillFolderResponse(createdFolder))
}

func (h *Handler) GetFolderContent(w http.ResponseWriter, r *http.Request) {
	const op = "storage.handler.GetFolderContent"

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
	if errors.Is(err, storage.ErrFolderNotFound) {
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
	render.JSON(w, r, FolderContentResponse{
		Response:      response.OK(),
		FolderContent: *folderContent,
	})
}

func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	const op = "storage.handler.DeleteFolder"

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
	if errors.Is(err, storage.ErrDeleteCleanupFailed) {
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
	render.JSON(w, r, DeleteFolderResponse{
		Response: response.OK(),
		Files:    files,
	})
}

func (h *Handler) readMultipartAndUploadFile(r *http.Request, reader *multipart.Reader, file *storage.File) (*storage.File, error) {
	var newFile storage.File

	var fileProcessed bool

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrReadMultipartReader, err)
		}

		if fileProcessed {
			_, err = h.service.DeleteFile(context.WithoutCancel(r.Context()), newFile.UserID, newFile.ID)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrStorageCleanupFailed, err)
			}

			return nil, fmt.Errorf("%w: %s", ErrExtraDataAfterFile, part.FormName())
		}

		switch part.FormName() {
		case "folder_id":
			rowID, err := io.ReadAll(io.LimitReader(part, 64))
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrReadFolderIDMultipart, err)
			}

			strID := strings.TrimSpace(string(rowID))
			if strID == "" {
				continue
			}

			folderID, err := uuid.Parse(strID)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidFolderID, err)
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
				return nil, err
			}

			newFile = *createdFile
			fileProcessed = true
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnknownField, part.FormName())
		}
	}

	if !fileProcessed {
		return nil, ErrMissingFilePart
	}

	return &newFile, nil
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

func fillFileResponse(file *storage.File) *FileResponse {
	return &FileResponse{
		Response:    response.OK(),
		ID:          file.ID,
		Name:        file.Name,
		UserID:      file.UserID,
		FolderID:    file.FolderID,
		S3Key:       file.S3Key,
		Size:        file.Size,
		ContentType: file.ContentType,
		CreatedAt:   file.CreatedAt,
	}
}

func fillFolderResponse(folder *storage.Folder) *FolderResponse {
	return &FolderResponse{
		Response:       response.OK(),
		ID:             folder.ID,
		Name:           folder.Name,
		UserID:         folder.UserID,
		ParentFolderID: folder.ParentFolderID,
		CreatedAt:      folder.CreatedAt,
	}
}
