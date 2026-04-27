package service

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) GetFolderContent(
	ctx context.Context,
	userID uuid.UUID,
	folderID *uuid.UUID,
) (folderContent core_domain.FolderContent, err error) {
	const op = "storage.service.GetFolderContent"

	folderContent, err = s.metadata.GetFolderContent(ctx, userID, folderID)
	if err != nil {
		return folderContent, fmt.Errorf("%s: %w", op, err)
	}

	return folderContent, nil
}
