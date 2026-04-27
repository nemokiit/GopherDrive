package service

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) DeleteFolder(
	ctx context.Context,
	userID uuid.UUID,
	folderID uuid.UUID,
) (files []core_domain.File, err error) {
	const op = "storage.service.DeleteFolder"

	files, err = s.metadata.DeleteFolder(ctx, userID, folderID)
	if err != nil {
		return files, fmt.Errorf("%s: %w", op, err)
	}

	keys := make([]string, 0, len(files))
	for _, file := range files {
		keys = append(keys, file.S3Key)
	}

	err = s.storage.DeleteFiles(context.WithoutCancel(ctx), keys)
	if err != nil {
		return files, fmt.Errorf("%s: %w: %w", op, core_errors.ErrDeleteCleanupFailed, err)
	}

	return files, nil
}
