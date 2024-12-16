package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"strings"
	"strunetsdrive/internal/models"
	"time"
)

type StoreRepo struct {
	db *sqlx.DB
}

func NewStoreRepo(db *sqlx.DB) *StoreRepo {
	return &StoreRepo{db}
}

func (r *StoreRepo) GetRootFolder(username string) (*models.Folder, error) {
	query := `
        SELECT id, name, username 
        FROM folders 
        WHERE username = $1 AND name = 'Root' AND parent_id IS NULL
    `
	var folder models.Folder
	err := r.db.QueryRow(query, username).Scan(&folder.ID, &folder.Name, &folder.Username)
	if err != nil {
		return nil, fmt.Errorf("could not find root folder for user %s: %w", username, err)
	}
	return &folder, nil
}

func (r *StoreRepo) GetFolderContent(folderID string) (*models.Folder, error) {
	folder := &models.Folder{}

	var parentID *string

	err := r.db.QueryRow(`
    SELECT id, name, parent_id, username, created_at
    FROM folders 
    WHERE id = $1
    `, folderID).Scan(
		&folder.ID,
		&folder.Name,
		&parentID,
		&folder.Username,
		&folder.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if parentID != nil {
		folder.ParentID = *parentID
	}

	rows, err := r.db.Query(`
    SELECT id, name, parent_id, username, created_at
    FROM folders 
    WHERE parent_id = $1
    `, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		subfolder := &models.Folder{}
		err := rows.Scan(
			&subfolder.ID,
			&subfolder.Name,
			&subfolder.ParentID,
			&subfolder.Username,
			&subfolder.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		folder.Folders = append(folder.Folders, subfolder)
	}

	fileRows, err := r.db.Query(`
    SELECT id, name, path, size, username, uploaded_at, is_dir
    FROM files 
    WHERE folder_id = $1 AND is_dir = false
    `, folderID)
	if err != nil {
		return nil, err
	}
	defer fileRows.Close()

	for fileRows.Next() {
		file := &models.File{}
		err := fileRows.Scan(
			&file.ID,
			&file.Name,
			&file.Path,
			&file.Size,
			&file.Username,
			&file.UploadedAt,
			&file.IsDir,
		)
		if err != nil {
			return nil, err
		}
		folder.Files = append(folder.Files, file)
	}

	return folder, nil
}

func (r *StoreRepo) GetFile(id string) (*models.File, error) {
	var file models.File
	err := r.db.QueryRow(`
	SELECT id, name, path, size, username,uploaded_at
	FROM files WHERE id = $1
`, id).Scan(&file.ID, &file.Name, &file.Path, &file.Size, &file.Username, &file.UploadedAt)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (r *StoreRepo) GetFileByUser(username string) ([]*models.File, error) {
	rows, err := r.db.Query(`
    SELECT id, name, path, size, uploaded_at, is_dir, folder_id
    FROM files 
    WHERE username = $1 AND is_dir = false
    ORDER BY uploaded_at DESC
    `, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		var file models.File
		if err = rows.Scan(
			&file.ID,
			&file.Name,
			&file.Path,
			&file.Size,
			&file.UploadedAt,
			&file.IsDir,
			&file.FolderID,
		); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, nil
}

func (r *StoreRepo) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
	SELECT id, username, password, created_at
	FROM users WHERE username = $1
`, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *StoreRepo) SaveFile(file *models.File) error {
	query := `
    INSERT INTO files (
        id, name, path, size, username, uploaded_at, 
        is_dir, folder_id
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.Exec(query,
		file.ID,
		file.Name,
		file.Path,
		file.Size,
		file.Username,
		time.Now(),
		file.IsDir,
		file.FolderID,
	)
	return err
}

func (r *StoreRepo) SaveFolder(folder *models.Folder) error {
	var parentPathArray []string
	err := r.db.QueryRow(`
        SELECT COALESCE(path_array, ARRAY[]::VARCHAR[]) 
        FROM folders 
        WHERE id = $1
    `, folder.ParentID).Scan(pq.Array(&parentPathArray))

	if err != nil {
		return fmt.Errorf("get parent path_array: %v", err)
	}

	newPathArray := append(parentPathArray, folder.ParentID)

	query := `
    INSERT INTO folders (id, name, parent_id, username, path_array)
    VALUES ($1, $2, $3, $4, $5)
    `
	_, err = r.db.Exec(query,
		folder.ID,
		folder.Name,
		folder.ParentID,
		folder.Username,
		pq.Array(newPathArray),
	)
	return err
}

func (r *StoreRepo) DeleteFile(fileID string) error {
	query := `DELETE FROM files WHERE id = $1`
	_, err := r.db.Exec(query, fileID)
	return err
}

func (r *StoreRepo) GetFolderHierarchy(username string) ([]*models.Folder, error) {
	query := `
        WITH RECURSIVE folder_hierarchy AS (
			SELECT
				f.id,
				f.name,
				f.parent_id,
				f.username,
				f.created_at,
				0 as level,
				f.path_array
			FROM folders f
			WHERE f.username = $1
			  AND f.parent_id IS NULL
		
			UNION ALL
		
			SELECT
				f.id,
				f.name,
				f.parent_id,
				f.username,
				f.created_at,
				fh.level + 1,
				f.path_array
			FROM folders f
					 JOIN folder_hierarchy fh ON f.parent_id = fh.id
		)
		SELECT
			id,
			name,
			parent_id,
			username,
			created_at,
			level,
			path_array
		FROM folder_hierarchy
		ORDER BY path_array, level;
    `

	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folderMap := make(map[string]*models.Folder)
	var rootFolders []*models.Folder

	for rows.Next() {
		folder := &models.Folder{}
		var level int
		var parentID *string
		var pathArrayBytes []byte

		err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&parentID,
			&folder.Username,
			&folder.CreatedAt,
			&level,
			&pathArrayBytes,
		)
		if err != nil {
			return nil, err
		}

		if pathArrayBytes != nil {
			pathStr := string(pathArrayBytes)
			folder.PathArray = parsePgArray(pathStr)
		}

		if parentID != nil {
			folder.ParentID = *parentID
		}

		folderMap[folder.ID] = folder

		if parentID == nil {
			rootFolders = append(rootFolders, folder)
		} else {
			parent := folderMap[*parentID]
			if parent != nil {
				parent.Folders = append(parent.Folders, folder)
			}
		}
	}

	return rootFolders, nil
}

func (r *StoreRepo) GetCompleteHierarchy(username string) ([]*models.Folder, error) {
	query := `
        WITH RECURSIVE folder_hierarchy AS (
            SELECT
                f.id,
                f.name,
                f.parent_id,
                f.username,
                f.created_at,
                0 as level,
                f.path_array,
                f.id::text AS path
            FROM folders f
            WHERE f.username = $1
              AND f.parent_id IS NULL 

            UNION ALL

            SELECT
                f.id,
                f.name,
                f.parent_id,
                f.username,
                f.created_at,
                fh.level + 1,
                f.path_array,
                fh.path || '/' || f.id::text
            FROM folders f
            JOIN folder_hierarchy fh ON f.parent_id = fh.id
        )
        SELECT 
            fh.id,
            fh.name,
            fh.parent_id,
            fh.username,
            fh.created_at,
            fh.level,
            fh.path_array,
            f.id AS file_id,
            f.name AS file_name,
            f.path AS file_path,
            f.size AS file_size,
            f.uploaded_at AS file_uploaded_at,
            f.is_dir AS file_is_dir
        FROM folder_hierarchy fh
        LEFT JOIN files f ON f.folder_id = fh.id
        WHERE fh.username = $1
        ORDER BY fh.path, fh.level;
    `

	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folderMap := make(map[string]*models.Folder)
	var rootFolders []*models.Folder

	for rows.Next() {
		folder := &models.Folder{}
		var level int
		var parentID *string
		var pathArrayBytes []byte

		var fileID, fileName, filePath *string
		var fileSize *int64
		var fileUploadedAt *time.Time
		var fileIsDir *bool

		err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&parentID,
			&folder.Username,
			&folder.CreatedAt,
			&level,
			&pathArrayBytes,
			&fileID,
			&fileName,
			&filePath,
			&fileSize,
			&fileUploadedAt,
			&fileIsDir,
		)
		if err != nil {
			return nil, err
		}

		if pathArrayBytes != nil {
			pathStr := string(pathArrayBytes)
			folder.PathArray = parsePgArray(pathStr)
		}

		if parentID != nil {
			folder.ParentID = *parentID
		}

		existingFolder, exists := folderMap[folder.ID]
		if !exists {
			folderMap[folder.ID] = folder

			if parentID == nil {
				rootFolders = append(rootFolders, folder)
			} else {
				parent := folderMap[*parentID]
				if parent != nil {
					parent.Folders = append(parent.Folders, folder)
				}
			}
		} else {
			folder = existingFolder
		}

		if fileID != nil {
			file := &models.File{
				ID:         *fileID,
				Name:       *fileName,
				Path:       *filePath,
				Size:       *fileSize,
				Username:   folder.Username,
				UploadedAt: *fileUploadedAt,
				IsDir:      *fileIsDir,
				FolderID:   folder.ID,
			}
			folder.Files = append(folder.Files, file)
		}
	}

	return rootFolders, nil
}

func parsePgArray(arrayStr string) []string {
	if arrayStr == "" || arrayStr == "{}" {
		return []string{}
	}
	arrayStr = strings.Trim(arrayStr, "{}")
	if arrayStr == "" {
		return []string{}
	}
	return strings.Split(arrayStr, ",")
}
