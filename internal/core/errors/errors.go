package core_errors

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrFileNotFound         = errors.New("file not found")
	ErrFolderNotFound       = errors.New("folder not found")
	ErrParentFolderNotFound = errors.New("parent folder not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrFileAlreadyExists    = errors.New("file already exists")
	ErrFolderAlreadyExists  = errors.New("folder already exists")

	ErrStorageCleanupFailed = errors.New("failed to cleanup storage after database error")
	ErrDeleteCleanupFailed  = errors.New("record deleted from database, but failed to remove object from storage")

	ErrStorageCleanupFailedAfterDirtyFile = errors.New("failed to cleanup storage after uploading dirty file")

	ErrNameIsEmpty = errors.New("name must not be empty")

	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTokenNotFound = errors.New("token not found")
)
