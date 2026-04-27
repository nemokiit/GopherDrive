package service

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

type sizeCounter struct {
	source io.Reader
	size   int64
}

func (s *Service) CreateFile(
	ctx context.Context,
	reader io.Reader,
	file core_domain.File,
) (newFile core_domain.File, err error) {
	const op = "storage.service.UploadFile"

	if err = isValidName(file.Name); err != nil {
		return newFile, fmt.Errorf("%s: %w", op, err)
	}

	file.S3Key = uuid.New().String()

	counter := &sizeCounter{source: reader}

	err = s.storage.UploadFile(ctx, file.S3Key, counter, file.ContentType)
	if err != nil {
		return newFile, fmt.Errorf("%s: %w", op, err)
	}

	newFile = core_domain.CreateFile(file.Name, file.UserID, file.FolderID, file.S3Key, counter.size, file.ContentType)

	err = s.metadata.SaveFile(ctx, newFile)
	if err != nil {
		s3Err := s.storage.DeleteFile(context.WithoutCancel(ctx), file.S3Key)
		if s3Err != nil {
			return core_domain.File{}, errors.Join(
				fmt.Errorf("%s: %w", op, err),
				fmt.Errorf("%w: %w", core_errors.ErrStorageCleanupFailed, s3Err),
			)
		}
		return core_domain.File{}, fmt.Errorf("%s: %w", op, err)
	}

	return newFile, nil
}

func (sc *sizeCounter) Read(p []byte) (int, error) {
	n, err := sc.source.Read(p)
	sc.size += int64(n)
	return n, err
}

func isValidName(name string) error {
	if len(name) > 255 {
		return errors.New("name is too long")
	}
	if strings.TrimSpace(name) == "" {
		return core_errors.ErrNameIsEmpty
	}

	return nil
}
