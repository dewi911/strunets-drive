package rest

import (
	"context"
	"io"
	"strunetsdrive/internal/models"
)

type StorageService interface {
	CreateFolder(username, folderName, parentID string) (*models.Folder, error)
	UploadFile(username, filename string, content io.Reader, size int64, folderID string) (*models.File, error)
	DownloadFilesAsZip(username string) (io.ReadSeekCloser, error)
	DownloadFolderAsZip(folderID string) (io.ReadSeekCloser, error)
	DownloadSelectedFilesAsZip(username string, fileIDs []string) (io.ReadSeekCloser, error)
	DownloadFile(id string) (io.ReadSeekCloser, *models.File, error)
	DeleteFile(username, fileID string) error
	ListFiles(username string) ([]*models.File, error)
	GetFileDownloadURL(fileID string) (string, error)
	GetFolderContent(id, username string) (*models.Folder, error)
	GetRootFolder(username string) (*models.Folder, error)
	GetCompleteHierarchy(username string) ([]*models.Folder, error)
	GetFolderHierarchy(username string) ([]*models.Folder, error)
}

type UserService interface {
	SingUp(ctx context.Context, input models.SignUpInput) error
	Login(ctx context.Context, input models.LoginInput) (string, string, error)
	ParseToken(ctx context.Context, token string) (int, string, error)
	RefreshSession(ctx context.Context, refreshToken string) (string, string, error)
}
