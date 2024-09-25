package repository

import (
	"github.com/jmoiron/sqlx"
	"strunetsdrive/internal/models"
	"time"
)

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) *PostgresRepo {
	return &PostgresRepo{db}
}

func (r *PostgresRepo) SaveFile(file *models.File) error {
	_, err := r.db.Exec(`
	INSERT INTO files (id, name, path, size, username, uploaded_at)
	Values ($1, $2, $3, $4, $5)
`, file.ID, file.Name, file.Path, file.Size, file.Username, time.Now())
	return err
}

func (r *PostgresRepo) FindFile(id string) (*models.File, error) {
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

func (r *PostgresRepo) GetFileByUser(username string) ([]*models.File, error) {
	rows, err := r.db.Query(`
	SELECT id, name, size, uploaded_at
	FROM files WHERE username = $1
	ORDER BY uploaded_at DESC
`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		var file models.File
		if err = rows.Scan(&file.ID, &file.Name, &file.Size, &file.UploadedAt); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	return files, nil
}
