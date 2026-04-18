package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrFileAlreadyExists   = errors.New("user already exists")
	ErrFolderAlreadyExists = errors.New("folder already exists")

	ErrUserNotFound         = errors.New("user not found")
	ErrParentFolderNotFound = errors.New("parent folder not found")

	ErrFileNotFound   = errors.New("file not found")
	ErrFolderNotFound = errors.New("folder not found")
)

type File struct {
	ID          uuid.UUID
	Name        string
	UserID      uuid.UUID
	FolderID    *uuid.UUID
	S3Key       string
	Size        int64
	ContentType string
	CreatedAt   time.Time
}

type Folder struct {
	ID             uuid.UUID
	Name           string
	UserID         uuid.UUID
	ParentFolderID *uuid.UUID
	CreatedAt      time.Time
}

type FolderContent struct {
	Files   []File
	Folders []Folder
}
