package handler

import (
	"GopherDrive/internal/lib/api/response"
	"GopherDrive/internal/storage"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrReadMultipartReader   = errors.New("failed to read part of MultipartReader")
	ErrReadFolderIDMultipart = errors.New("failed to read folder_id")
	ErrInvalidFolderID       = errors.New("failed to parse folder id")
	ErrExtraDataAfterFile    = errors.New("extra data after file part")
	ErrStorageCleanupFailed  = errors.New("failed to cleanup storage after uploading dirty file")
	ErrUnknownField          = errors.New("unknown field in file part")
	ErrMissingFilePart       = errors.New("file part is missing")
)

type CreateFolderRequest struct {
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id"`
}

type FileResponse struct {
	response.Response
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	UserID      uuid.UUID  `json:"user_id"`
	FolderID    *uuid.UUID `json:"folder_id"`
	S3Key       string     `json:"s3_key"`
	Size        int64      `json:"size"`
	ContentType string     `json:"content_type"`
	CreatedAt   time.Time  `json:"created_at"`
}

type FolderResponse struct {
	response.Response
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	UserID         uuid.UUID  `json:"user_id"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

type FolderContentResponse struct {
	response.Response
	storage.FolderContent
}

type DeleteFolderResponse struct {
	response.Response
	Files []*storage.File `json:"files"`
}
