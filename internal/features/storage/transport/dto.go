package transport

import (
	core_domain "GopherDrive/internal/core/domain"
	"GopherDrive/internal/pkg/api/response"
	"errors"
)

var (
	ErrReadMultipartReader   = errors.New("failed to read part of MultipartReader")
	ErrReadFolderIDMultipart = errors.New("failed to read folder_id")
	ErrInvalidFolderID       = errors.New("failed to parse folder id")
	ErrExtraDataAfterFile    = errors.New("extra data after file part")
	ErrUnknownField          = errors.New("unknown field in file part")
	ErrMissingFilePart       = errors.New("file part is missing")
)

type CreateFolderRequest struct {
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id"`
}

type FileResponse struct {
	response.Response
	File core_domain.File
}

type FolderResponse struct {
	response.Response
	Folder core_domain.Folder
}

type FolderContentResponse struct {
	response.Response
	FolderContent core_domain.FolderContent
}

type DeleteFolderResponse struct {
	response.Response
	Files []core_domain.File `json:"files"`
}

func CreateFileResponse(response response.Response, file core_domain.File) *FileResponse {
	return &FileResponse{
		Response: response,
		File:     file,
	}
}

func CreateFolderResponse(response response.Response, folder core_domain.Folder) *FolderResponse {
	return &FolderResponse{
		Response: response,
		Folder:   folder,
	}
}

func CreateFolderContentResponse(response response.Response, folderContent core_domain.FolderContent) *FolderContentResponse {
	return &FolderContentResponse{
		Response:      response,
		FolderContent: folderContent,
	}
}

func CreateDeleteFolderResponse(response response.Response, files []core_domain.File) *DeleteFolderResponse {
	return &DeleteFolderResponse{
		Response: response,
		Files:    files,
	}
}
