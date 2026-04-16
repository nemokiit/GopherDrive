CREATE UNIQUE INDEX idx_folders_unique_name_root
ON folders (user_id, name)
WHERE parent_folder_id IS NULL;

CREATE UNIQUE INDEX idx_folders_unique_name_nested
ON folders (user_id, parent_folder_id, name)
WHERE parent_folder_id IS NOT NULL;

CREATE UNIQUE INDEX idx_files_unique_name_root
ON files (user_id, name)
WHERE folder_id IS NULL;

CREATE UNIQUE INDEX isd_files_unique_name_nested
ON files (user_id, folder_id, name)
WHERE folder_id IS NOT NULL;