package rest

import (
	"io"
	"strunetsdrive/internal/models"
)

type StorageService interface {
	UploadFile(username, filename string, content io.Reader) error
	DownloadFile(id string) (io.ReadCloser, string, error)
	ListFiles(username string) ([]*models.File, error)
}

type SessionService interface {
}
