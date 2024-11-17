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
	GetFile(id string) (*models.File, error)
	GetFileByUser(username string) ([]*models.File, error)
	GetUserByUsername(username string) (*models.User, error)
}
type SessionRepository interface {
	Create(ctx context.Context, token models.RefreshSession) error
	GetToken(ctx context.Context, token string) (*models.RefreshSession, error)
}
