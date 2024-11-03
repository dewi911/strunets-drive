package service

import (
	"io"
	"strunetsdrive/internal/models"
	"strunetsdrive/pkg/encrypt"
	"strunetsdrive/pkg/filestore"
)

type Service struct {
	repo        StoreRepository
	storagePath string
	fileStore   filestore.Store
}

func NewService(repo StoreRepository, storagePath string, fileStore filestore.Store) *Service {
	return &Service{repo: repo, storagePath: storagePath, fileStore: fileStore}
}

func (s *Service) UploadFile(username, filename string, content io.Reader) error {
	id := encrypt.GenerateUUID()
	path := encrypt.Encrypt(username + "/" + id)

	//encryptedPath := encrypt.Encrypt(filepath.Join(s.storagePath, id))

	//fullPath := filepath.Join(s.storagePath, id)
	//file, err := os.Create(fullPath)
	//if err != nil {
	//	return err
	//}
	//defer file.Close()

	writer, err := s.fileStore.Create(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, content)
	if err != nil {
		return err
	}

	fileInfo := &models.File{
		ID:       id,
		Name:     filename,
		Path:     encrypt.Encrypt(path),
		Username: username,
	}

	return s.repo.SaveFile(fileInfo)
}

func (s *Service) DownloadFile(id string) (filestore.Reader, string, error) {
	fileInfo, err := s.repo.GetFile(id)
	if err != nil {
		return nil, "", err
	}

	decryptedPath := encrypt.Decrypt(fileInfo.Path)

	reader, err := s.fileStore.Open(decryptedPath)
	if err != nil {
		return nil, "", err
	}

	return reader, fileInfo.Name, nil
}

func (s *Service) ListFiles(username string) ([]*models.File, error) {
	return s.repo.GetFileByUser(username)
}
