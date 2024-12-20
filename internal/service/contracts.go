package service

import (
	"context"
	"strunetsdrive/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByID(ctx context.Context, id int) (models.User, error)
	GetByCredentials(ctx context.Context, username, password string) (models.User, error)
	Exist(ctx context.Context, username string) (bool, error)
}

type StoreRepository interface {
	SaveFile(file *models.File) error
	DeleteFile(fileID string) error
	SaveFolder(folder *models.Folder) error
	GetRootFolder(username string) (*models.Folder, error)
	GetFolderContent(folderID string) (*models.Folder, error)
	GetFile(id string) (*models.File, error)
	GetFileByUser(username string) ([]*models.File, error)
	GetFileById(fileID, username string) (*models.File, error)
	GetUserByUsername(username string) (*models.User, error)
	GetCompleteHierarchy(username string) ([]*models.Folder, error)
	GetFolderHierarchy(username string) ([]*models.Folder, error)
}
type SessionRepository interface {
	Create(ctx context.Context, token models.RefreshSession) error
	GetToken(ctx context.Context, token string) (*models.RefreshSession, error)
}
