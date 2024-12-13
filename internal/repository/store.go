package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strunetsdrive/internal/models"
	"time"
)

type StoreRepo struct {
	db *sqlx.DB
}

func NewStoreRepo(db *sqlx.DB) *StoreRepo {
	return &StoreRepo{db}
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
	query := `
    INSERT INTO folders (id, name, parent_id, username)
    VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.Exec(query,
		folder.ID,
		folder.Name,
		folder.ParentID,
		folder.Username,
	)
	return err
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
		&parentID, // Use pointer here
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

func (r *StoreRepo) DeleteFile(fileID string) error {
	query := `DELETE FROM files WHERE id = $1`
	_, err := r.db.Exec(query, fileID)
	return err
}
