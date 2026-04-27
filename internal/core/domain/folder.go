package core_domain

import (
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	UserID         uuid.UUID  `json:"user_id"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

func CreateFolder(name string, userID uuid.UUID, ParentFolderID *uuid.UUID) Folder {
	var (
		id        = uuid.New()
		createdAt = time.Now()
	)

	return Folder{
		ID:             id,
		Name:           name,
		UserID:         userID,
		ParentFolderID: ParentFolderID,
		CreatedAt:      createdAt,
	}
}
