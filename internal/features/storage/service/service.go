package service

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"io"

	"github.com/google/uuid"
)

type MetadataRepository interface {
	SaveFile(ctx context.Context, file core_domain.File) error
	SaveFolder(ctx context.Context, folder core_domain.Folder) error
	GetFolderContent(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (core_domain.FolderContent, error)
	GetFileByID(ctx context.Context, userID, fileID uuid.UUID) (core_domain.File, error)
	DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (core_domain.File, error)
	DeleteFolder(ctx context.Context, userID, folderID uuid.UUID) ([]core_domain.File, error)
}

type ObjectStorageRepository interface {
	UploadFile(ctx context.Context, key string, reader io.Reader, contentType string) error
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
	DeleteFiles(ctx context.Context, files []string) error
}

type Service struct {
	metadata MetadataRepository
	storage  ObjectStorageRepository
}

func New(repo MetadataRepository, storage ObjectStorageRepository) *Service {
	return &Service{
		metadata: repo,
		storage:  storage,
	}
}
