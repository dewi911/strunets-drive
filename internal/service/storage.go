package service

import (
	"fmt"
	"io"
	"strunetsdrive/internal/models"
	"strunetsdrive/pkg/encrypt"
	"strunetsdrive/pkg/filestore"
	"time"
)

type Service struct {
	repo      StoreRepository
	fileStore filestore.Store
}

func NewService(repo StoreRepository, fileStore filestore.Store) *Service {
	return &Service{repo: repo, fileStore: fileStore}
}

func (s *Service) UploadFile(username, filename string, content io.Reader, size int64) (*models.File, error) {
	id := encrypt.GenerateUUID()
	path := fmt.Sprintf("%s/%s", username, id)
	encryptedPath := encrypt.Encrypt(path)

	writer, err := s.fileStore.Create(encryptedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer writer.Close()

	written, err := io.CopyN(writer, content, size)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	fileInfo := &models.File{
		ID:       id,
		Name:     filename,
		Path:     encrypt.Encrypt(path),
		Size:     written,
		Username: username,
	}
	if err := s.repo.SaveFile(fileInfo); err != nil {
		_ = s.fileStore.Delete(encryptedPath)
		return nil, fmt.Errorf("failed to save file info: %w", err)
	}

	return fileInfo, nil
}

func (s *Service) DownloadFile(id string) (io.ReadSeekCloser, *models.File, error) {
	fileInfo, err := s.repo.GetFile(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	decryptedPath := encrypt.Decrypt(fileInfo.Path)

	reader, err := s.fileStore.Open(decryptedPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return reader, fileInfo, nil
}

func (s *Service) ListFiles(username string) ([]*models.File, error) {
	return s.repo.GetFileByUser(username)
}

func (s *Service) GetFileDownloadURL(fileID string) (string, error) {
	fileInfo, err := s.repo.GetFile(fileID)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	decryptedPath := encrypt.Decrypt(fileInfo.Path)

	url, err := s.fileStore.(interface {
		GetPresignedURL(string, time.Duration) (string, error)
	}).
		GetPresignedURL(decryptedPath, time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}
