package service

import (
	"io"
	"os"
	"path/filepath"
	"strunetsdrive/internal/models"
	"strunetsdrive/pkg/encrypt"
)

type Service struct {
	repo        StoreRepository
	storagePath string
}

func NewService(repo StoreRepository, storagePath string) *Service {
	return &Service{repo: repo, storagePath: storagePath}
}

func (s *Service) UploadFile(username, filename string, content io.Reader) error {
	id := encrypt.GenerateUUID()

	encryptedPath := encrypt.Encrypt(filepath.Join(s.storagePath, id))

	fullPath := filepath.Join(s.storagePath, id)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		return err
	}

	fileInfo := &models.File{
		ID:       id,
		Name:     filename,
		Path:     encryptedPath,
		Username: username,
	}

	return s.repo.SaveFile(fileInfo)
}

func (s *Service) DownloadFile(id string) (io.ReadCloser, string, error) {
	fileInfo, err := s.repo.GetFile(id)
	if err != nil {
		return nil, "", err
	}

	decryptedPath := encrypt.Decrypt(fileInfo.Path)

	file, err := os.Open(decryptedPath)
	if err != nil {
		return nil, "", err
	}

	return file, fileInfo.Name, nil
}

func (s *Service) ListFiles(username string) ([]*models.File, error) {
	return s.repo.GetFileByUser(username)
}
