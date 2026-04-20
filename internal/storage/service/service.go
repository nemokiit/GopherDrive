package service

import (
	"GopherDrive/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

type MetadataRepository interface {
	CreateFolder(ctx context.Context, folder *storage.Folder) (*storage.Folder, error)
	CreateFile(ctx context.Context, file *storage.File) (*storage.File, error)
	GetFolderContent(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (*storage.FolderContent, error)
	GetFileByID(ctx context.Context, userID, fileID uuid.UUID) (*storage.File, error)
	DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (*storage.File, error)
	DeleteFolder(ctx context.Context, userID, folderID uuid.UUID) ([]*storage.File, error)
}

type ObjectStorage interface {
	UploadFile(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
	DeleteFiles(ctx context.Context, files []string) error
}

type Service struct {
	repo    MetadataRepository
	storage ObjectStorage
}

func New(repo MetadataRepository, storage ObjectStorage) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
	}
}

func (s *Service) CreateFile(ctx context.Context, reader io.Reader, file *storage.File) (*storage.File, error) {
	const op = "storage.service.UploadFile"

	if err := isValidName(file.Name); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	file.S3Key = uuid.New().String()

	err := s.storage.UploadFile(ctx, file.S3Key, reader, file.Size, file.ContentType)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newFile, err := s.repo.CreateFile(ctx, file)
	if err != nil {
		s3Err := s.storage.DeleteFile(context.WithoutCancel(ctx), file.S3Key)
		if s3Err != nil {
			return nil, errors.Join(
				fmt.Errorf("%s: %w", op, err),
				fmt.Errorf("%w: %w", storage.ErrStorageCleanupFailed, s3Err),
			)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newFile, nil
}

func (s *Service) DownloadFile(ctx context.Context, userID, fileID uuid.UUID) (io.ReadCloser, *storage.File, error) {
	const op = "storage.service.DownloadFile"

	file, err := s.repo.GetFileByID(ctx, userID, fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	reader, err := s.storage.DownloadFile(ctx, file.S3Key)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return reader, file, nil
}

func (s *Service) DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (*storage.File, error) {
	const op = "storage.service.DeleteFile"

	file, err := s.repo.DeleteFile(ctx, userID, fileID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.storage.DeleteFile(context.WithoutCancel(ctx), file.S3Key)
	if err != nil {
		return file, fmt.Errorf("%s: %w: %w", op, storage.ErrDeleteCleanupFailed, err)
	}

	return file, nil
}

func (s *Service) CreateFolder(ctx context.Context, folder *storage.Folder) (*storage.Folder, error) {
	const op = "storage.service.CreateFolder"

	if err := isValidName(folder.Name); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newFolder, err := s.repo.CreateFolder(ctx, folder)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newFolder, nil
}

func (s *Service) GetFolderContent(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (*storage.FolderContent, error) {
	const op = "storage.service.GetFolderContent"

	folderContent, err := s.repo.GetFolderContent(ctx, userID, folderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return folderContent, nil
}

func (s *Service) DeleteFolder(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]*storage.File, error) {
	const op = "storage.service.DeleteFolder"

	files, err := s.repo.DeleteFolder(ctx, userID, folderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	keys := make([]string, 0, len(files))
	for _, file := range files {
		keys = append(keys, file.S3Key)
	}

	err = s.storage.DeleteFiles(context.WithoutCancel(ctx), keys)
	if err != nil {
		return files, fmt.Errorf("%s: %w: %w", op, storage.ErrDeleteCleanupFailed, err)
	}

	return files, nil
}

func isValidName(name string) error {
	if len(name) > 255 {
		return errors.New("name is too long")
	}
	if strings.TrimSpace(name) == "" {
		return storage.ErrNameIsEmpty
	}

	return nil
}
