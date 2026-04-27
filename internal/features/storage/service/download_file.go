package service

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

func (s *Service) DownloadFile(
	ctx context.Context,
	userID,
	fileID uuid.UUID,
) (reader io.ReadCloser, file core_domain.File, err error) {
	const op = "storage.service.DownloadFile"

	file, err = s.metadata.GetFileByID(ctx, userID, fileID)
	if err != nil {
		return reader, file, fmt.Errorf("%s: %w", op, err)
	}

	reader, err = s.storage.DownloadFile(ctx, file.S3Key)
	if err != nil {
		return reader, file, fmt.Errorf("%s: %w", op, err)
	}

	return reader, file, nil
}
