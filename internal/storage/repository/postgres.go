package repository

import (
	"GopherDrive/internal/storage"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{
		pool: pool,
	}
}

func (r *PGRepository) CreateFolder(ctx context.Context, folder *storage.Folder) (*storage.Folder, error) {
	const op = "storage.repository.CreateFolder"

	var newFolder storage.Folder

	query := "INSERT INTO folders (name, user_id, parent_folder_id) VALUES ($1, $2, $3) RETURNING *"
	row := r.pool.QueryRow(ctx, query, folder.Name, folder.UserID, folder.ParentFolderID)

	err := row.Scan(&newFolder.ID, &newFolder.Name, &newFolder.UserID, &newFolder.ParentFolderID, &newFolder.CreatedAt)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrFolderAlreadyExists)
		}
		if pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "folders_user_id_fkey":
				return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
			case "folders_parent_folder_id_fkey":
				return nil, fmt.Errorf("%s: %w", op, storage.ErrParentFolderNotFound)
			default:
				return nil, fmt.Errorf("%s: foreign key violation for %s", op, pgErr.ConstraintName)
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &newFolder, nil
}

func (r *PGRepository) CreateFile(ctx context.Context, file *storage.File) (*storage.File, error) {
	const op = "storage.repository.CreateFile"

	var newFile storage.File

	query := `INSERT INTO files (name, user_id, folder_id, s3_key, size, content_type)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`
	row := r.pool.QueryRow(ctx, query, file.Name, file.UserID, file.FolderID, file.S3Key, file.Size, file.ContentType)

	err := fillFile(&newFile, row)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrFileAlreadyExists)
		}
		if pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "files_user_id_fkey":
				return nil, fmt.Errorf("%s: %w", op, storage.ErrFileNotFound)
			case "files_folder_id_fkey":
				return nil, fmt.Errorf("%s: %w", op, storage.ErrFolderNotFound)
			default:
				return nil, fmt.Errorf("%s: foreign key violation for %s", op, pgErr.ConstraintName)
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &newFile, nil
}

func (r *PGRepository) GetFolderContent(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (*storage.FolderContent, error) {
	const op = "storage.repository.GetFolder"

	getFilesQuery := "SELECT * FROM files WHERE user_id = $1 AND folder_id IS NOT DISTINCT FROM $2"
	fileRows, err := r.pool.Query(ctx, getFilesQuery, userID, folderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer fileRows.Close()

	getFoldersQuery := "SELECT * FROM folders WHERE user_id = $1 AND parent_folder_id IS NOT DISTINCT FROM $2"
	folderRows, err := r.pool.Query(ctx, getFoldersQuery, userID, folderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer folderRows.Close()

	folderContent := storage.FolderContent{
		Files:   make([]storage.File, 0),
		Folders: make([]storage.Folder, 0),
	}

	for fileRows.Next() {
		var file storage.File

		err = fillFile(&file, fileRows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		folderContent.Files = append(folderContent.Files, file)
	}

	if err = fileRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for folderRows.Next() {
		var folder storage.Folder

		err = folderRows.Scan(&folder.ID, &folder.Name, &folder.UserID, &folder.ParentFolderID, &folder.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		folderContent.Folders = append(folderContent.Folders, folder)
	}

	if err = folderRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(folderContent.Files) == 0 && len(folderContent.Folders) == 0 && folderID != nil {
		if ok, _ := r.folderExist(ctx, userID, folderID); !ok {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrFolderNotFound)
		}
	}

	return &folderContent, nil
}

func (r *PGRepository) GetFileByID(ctx context.Context, userID, fileID uuid.UUID) (*storage.File, error) {
	const op = "storage.repository.GetFileByID"

	var file storage.File

	query := "SELECT * FROM files WHERE user_id = $1 AND id = $2"
	row := r.pool.QueryRow(ctx, query, userID, fileID)

	err := fillFile(&file, row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrFileNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &file, nil
}

func (r *PGRepository) DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (*storage.File, error) {
	const op = "storage.repository.DeleteFile"

	var file storage.File

	query := "DELETE FROM files WHERE user_id = $1 AND id = $2 RETURNING *"
	row := r.pool.QueryRow(ctx, query, userID, fileID)

	err := fillFile(&file, row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrFileNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &file, nil
}

func (r *PGRepository) DeleteFolder(ctx context.Context, userID, folderID uuid.UUID) ([]*storage.File, error) {
	const op = "storage.repository.DeleteFolder"

	folder := make([]*storage.File, 0)

	query := `WITH RECURSIVE
				target_folders AS (
				  SELECT id FROM folders
				  WHERE user_id = $1 AND id = $2 
				  UNION ALL
				  
				  SELECT f.id FROM folders f
				  JOIN target_folders ft ON f.parent_folder_id = ft.id),
				delete_action AS (
				  DELETE FROM folders
				  WHERE user_id = $1 AND id = $2)
			  
			  SELECT * from files f
			  JOIN target_folders fl ON f.folder_id = fl.id`

	rows, err := r.pool.Query(ctx, query, userID, folderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var file storage.File

		err = fillFile(&file, rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		folder = append(folder, &file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return folder, nil
}

func (r *PGRepository) folderExist(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (bool, error) {
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

func fillFile(file *storage.File, row pgx.Row) error {
	err := row.Scan(&file.ID, &file.Name, &file.UserID, &file.FolderID, &file.S3Key, &file.Size, &file.ContentType, &file.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}
