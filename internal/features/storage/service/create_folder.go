package service

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"fmt"
)

func (s *Service) CreateFolder(
	ctx context.Context,
	folder core_domain.Folder,
) (newFolder core_domain.Folder, err error) {
	const op = "storage.service.CreateFolder"

	if err = isValidName(folder.Name); err != nil {
		return newFolder, fmt.Errorf("%s: %w", op, err)
	}

	newFolder = core_domain.CreateFolder(folder.Name, folder.UserID, folder.ParentFolderID)

	err = s.metadata.SaveFolder(ctx, folder)
	if err != nil {
		return newFolder, fmt.Errorf("%s: %w", op, err)
	}

	return newFolder, nil
}
