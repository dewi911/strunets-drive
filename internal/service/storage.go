package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strunetsdrive/internal/models"
	"strunetsdrive/pkg/encrypt"
	"strunetsdrive/pkg/filestore"
	"time"
)

type bufferReadSeekCloser struct {
	*bytes.Reader
}

func (b *bufferReadSeekCloser) Close() error {
	return nil
}

type StoreService struct {
	repo      StoreRepository
	fileStore filestore.Store
}

func NewStoreService(repo StoreRepository, fileStore filestore.Store) *StoreService {
	return &StoreService{repo: repo, fileStore: fileStore}
}

func (s *StoreService) CreateFolder(username, folderName, parentID string) (*models.Folder, error) {
	if parentID == "" {
		rootFolder, err := s.repo.GetRootFolder(username)
		if err != nil {
			return nil, err
		}
		parentID = rootFolder.ID
	}

	folder := &models.Folder{
		ID:       encrypt.GenerateUUID(),
		Name:     folderName,
		ParentID: parentID,
		Username: username,
	}

	if err := s.repo.SaveFolder(folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *StoreService) UploadFile(username, filename string, content io.Reader, size int64, folderID string) (*models.File, error) {
	if folderID == "" {
		rootFolder, err := s.repo.GetRootFolder(username)
		if err != nil {
			return nil, err
		}
		folderID = rootFolder.ID
	}

	id := encrypt.GenerateUUID()
	path := fmt.Sprintf("%s/%s/%s", username, folderID, id)

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
		ID:         id,
		Name:       filename,
		Path:       path,
		Size:       written,
		Username:   username,
		FolderID:   folderID,
		IsDir:      false,
		UploadedAt: time.Now(),
	}

	if err := s.repo.SaveFile(fileInfo); err != nil {
		_ = s.fileStore.Delete(path)
		return nil, fmt.Errorf("failed to save file info: %w", err)
	}

	return fileInfo, nil
}

func (s *StoreService) DownloadFilesAsZip(username string) (io.ReadSeekCloser, error) {
	files, err := s.repo.GetFileByUser(username)
	if err != nil {
		return nil, err
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	for _, file := range files {
		reader, _, err := s.DownloadFile(file.ID)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(zipFile, reader)
		if err != nil {
			return nil, err
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return &bufferReadSeekCloser{bytes.NewReader(zipBuffer.Bytes())}, nil
}

func (s *StoreService) DownloadFolderAsZip(folderID string) (io.ReadSeekCloser, error) {
	folderContent, err := s.repo.GetFolderContent(folderID)
	if err != nil {
		return nil, err
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	var addToZip func(folder *models.Folder, basePath string) error
	addToZip = func(folder *models.Folder, basePath string) error {
		for _, file := range folder.Files {
			reader, _, err := s.DownloadFile(file.ID)
			if err != nil {
				return err
			}
			defer reader.Close()

			zipPath := filepath.Join(basePath, file.Name)
			zipFile, err := zipWriter.Create(zipPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(zipFile, reader)
			if err != nil {
				return err
			}
		}

		for _, subfolder := range folder.Folders {
			subfolderPath := filepath.Join(basePath, subfolder.Name)

			subfolderContent, err := s.repo.GetFolderContent(subfolder.ID)
			if err != nil {
				return err
			}

			err = addToZip(subfolderContent, subfolderPath)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = addToZip(folderContent, "")
	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return &bufferReadSeekCloser{bytes.NewReader(zipBuffer.Bytes())}, nil
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

func (s *StoreService) DeleteFile(username, fileID string) error {
	fileInfo, err := s.repo.GetFile(fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if fileInfo.Username != username {
		return fmt.Errorf("unauthorized to delete this file")
	}

	if err := s.fileStore.Delete(fileInfo.Path); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	if err := s.repo.DeleteFile(fileID); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
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

func (s *StoreService) GetFolderContent(id string) (*models.Folder, error) {
	return s.repo.GetFolderContent(id)
}

func (s *StoreService) GetRootFolder(username string) (*models.Folder, error) {
	return s.repo.GetRootFolder(username)
}
