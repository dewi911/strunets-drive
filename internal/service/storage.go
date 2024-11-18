package service

import (
	"fmt"
	"io"
	"strunetsdrive/internal/models"
	"strunetsdrive/pkg/encrypt"
	"strunetsdrive/pkg/filestore"
	"time"
)

type StoreService struct {
	repo      StoreRepository
	fileStore filestore.Store
}

func NewStoreService(repo StoreRepository, fileStore filestore.Store) *StoreService {
	return &StoreService{repo: repo, fileStore: fileStore}
}

func (s *StoreService) UploadFile(username, filename string, content io.Reader, size int64) (*models.File, error) {
	id := encrypt.GenerateUUID()
	path := fmt.Sprintf("%s/%s", username, id)
	//encryptedPath := encrypt.Encrypt(path)

	writer, err := s.fileStore.Create(path)
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
		Path:     path,
		Size:     written,
		Username: username,
	}
	if err := s.repo.SaveFile(fileInfo); err != nil {
		_ = s.fileStore.Delete(path)
		return nil, fmt.Errorf("failed to save file info: %w", err)
	}

	return fileInfo, nil
}

func (s *StoreService) DownloadFile(id string) (io.ReadSeekCloser, *models.File, error) {
	fileInfo, err := s.repo.GetFile(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	//decryptedPath := encrypt.Decrypt(fileInfo.Path)

	reader, err := s.fileStore.Open(fileInfo.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return reader, fileInfo, nil
}

func (s *StoreService) ListFiles(username string) ([]*models.File, error) {
	return s.repo.GetFileByUser(username)
}

func (s *StoreService) GetFileDownloadURL(fileID string) (string, error) {
	fileInfo, err := s.repo.GetFile(fileID)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	//decryptedPath := encrypt.Decrypt(fileInfo.Path)

	url, err := s.fileStore.(interface {
		GetPresignedURL(string, time.Duration) (string, error)
	}).
		GetPresignedURL(fileInfo.Path, time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}
