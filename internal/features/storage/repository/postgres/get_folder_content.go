package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	"GopherDrive/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetFolderContent(
	ctx context.Context,
	userID uuid.UUID,
	folderID *uuid.UUID,
) (folderContent core_domain.FolderContent, err error) {
	const op = "storage.repository.GetFolder"

	getFilesQuery := "SELECT * FROM files WHERE user_id = $1 AND folder_id IS NOT DISTINCT FROM $2"
	fileRows, err := r.pool.Query(ctx, getFilesQuery, userID, folderID)
	if err != nil {
		return folderContent, fmt.Errorf("%s: %w", op, err)
	}
	defer fileRows.Close()

	getFoldersQuery := "SELECT * FROM folders WHERE user_id = $1 AND parent_folder_id IS NOT DISTINCT FROM $2"
	folderRows, err := r.pool.Query(ctx, getFoldersQuery, userID, folderID)
	if err != nil {
		return folderContent, fmt.Errorf("%s: %w", op, err)
	}
	defer folderRows.Close()

	folderContent = core_domain.CreateFolderContent()

	for fileRows.Next() {
		var file core_domain.File

		err = fillFile(&file, fileRows)
		if err != nil {
			return folderContent, fmt.Errorf("%s: %w", op, err)
		}

		folderContent.Files = append(folderContent.Files, &file)
	}
	if err = fileRows.Err(); err != nil {
		return folderContent, fmt.Errorf("%s: %w", op, err)
	}

	for folderRows.Next() {
		var folder core_domain.Folder

		err = folderRows.Scan(&folder.ID, &folder.Name, &folder.UserID, &folder.ParentFolderID, &folder.CreatedAt)
		if err != nil {
			return folderContent, fmt.Errorf("%s: %w", op, err)
		}

		folderContent.Folders = append(folderContent.Folders, &folder)
	}
	if err = folderRows.Err(); err != nil {
		return folderContent, fmt.Errorf("%s: %w", op, err)
	}

	if len(folderContent.Files) == 0 && len(folderContent.Folders) == 0 && folderID != nil {
		if ok, _ := r.folderExist(ctx, userID, folderID); !ok {
			return folderContent, fmt.Errorf("%s: %w", op, pool.ErrFolderNotFound)
		}
	}

	return folderContent, nil
}

func (r *Repository) folderExist(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (bool, error) {
	var ok int

	query := "SELECT 1 FROM folders WHERE user_id = $1 AND id = $2"
	err := r.pool.QueryRow(ctx, query, userID, folderID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
