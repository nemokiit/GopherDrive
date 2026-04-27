package service

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (file core_domain.File, err error) {
	const op = "storage.service.DeleteFile"

	file, err = s.metadata.DeleteFile(ctx, userID, fileID)
	if err != nil {
		return file, fmt.Errorf("%s: %w", op, err)
	}

	err = s.storage.DeleteFile(context.WithoutCancel(ctx), file.S3Key)
	if err != nil {
		return file, fmt.Errorf("%s: %w: %w", op, core_errors.ErrDeleteCleanupFailed, err)
	}

	return file, nil
}
