package errors

import "errors"

var (
	ErrStorageCleanupFailed               = errors.New("failed to cleanup storage after database error")
	ErrDeleteCleanupFailed                = errors.New("record deleted from database, but failed to remove object from storage")
	ErrStorageCleanupFailedAfterDirtyFile = errors.New("failed to cleanup storage after uploading dirty file")

	ErrNameIsEmpty = errors.New("name must not be empty")

	ErrInvalidCredentials = errors.New("invalid credentials")
)
