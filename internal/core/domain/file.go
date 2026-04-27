package core_domain

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	UserID      uuid.UUID  `json:"user_id"`
	FolderID    *uuid.UUID `json:"folder_id"`
	S3Key       string     `json:"s3_key"`
	Size        int64      `json:"size"`
	ContentType string     `json:"content_type"`
	CreatedAt   time.Time  `json:"created_at"`
}

func CreateFile(name string, userID uuid.UUID, folderID *uuid.UUID, s3Key string, size int64, contentType string) File {
	var (
		id        = uuid.New()
		createdAt = time.Now()
	)

	return File{
		ID:          id,
		Name:        name,
		UserID:      userID,
		FolderID:    folderID,
		S3Key:       s3Key,
		Size:        size,
		ContentType: contentType,
		CreatedAt:   createdAt,
	}
}
