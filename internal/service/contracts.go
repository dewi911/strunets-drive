package service

import (
	"context"
	"strunetsdrive/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) error
}

type StoreRepository interface {
	SaveFile(file *models.File) error
	GetFile(id string) (*models.File, error)
	GetFileByUser(username string) ([]*models.File, error)
	GetUserByUsername(username string) (*models.User, error)
}
type SessionRepository interface {
}
