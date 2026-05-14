package pool

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrFileNotFound         = errors.New("file not found")
	ErrFolderNotFound       = errors.New("folder not found")
	ErrParentFolderNotFound = errors.New("parent folder not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrFileAlreadyExists    = errors.New("file already exists")
	ErrFolderAlreadyExists  = errors.New("folder already exists")
)
