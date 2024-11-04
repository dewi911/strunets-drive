package rest

import (
	"io"
	"strunetsdrive/internal/models"
)

type StorageService interface {
	UploadFile(username, filename string, content io.Reader, size int64) (*models.File, error)
	DownloadFile(id string) (io.ReadSeekCloser, *models.File, error)
	ListFiles(username string) ([]*models.File, error)
}

type SessionService interface {
}
