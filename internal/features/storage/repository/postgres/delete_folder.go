package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) DeleteFolder(ctx context.Context, userID, folderID uuid.UUID) ([]core_domain.File, error) {
	const op = "storage.repository.DeleteFolder"

	folder := make([]core_domain.File, 0)

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
		var file core_domain.File

		err = fillFile(&file, rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		folder = append(folder, file)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return folder, nil
}
