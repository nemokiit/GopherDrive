package core_domain

type FolderContent struct {
	Files   []*File   `json:"files"`
	Folders []*Folder `json:"folders"`
}

func CreateFolderContent() FolderContent {
	var (
		files   = make([]*File, 0)
		folders = make([]*Folder, 0)
	)

	return FolderContent{
		Files:   files,
		Folders: folders,
	}
}
