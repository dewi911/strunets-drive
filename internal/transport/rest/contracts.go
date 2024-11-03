package rest

import (
	"io"
	"strunetsdrive/internal/models"
	"strunetsdrive/pkg/filestore"
)

type StorageService interface {
	UploadFile(username, filename string, content io.Reader) error
	DownloadFile(id string) (filestore.Reader, string, error)
	ListFiles(username string) ([]*models.File, error)
}

type SessionService interface {
}
