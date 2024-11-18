package rest

import (
	"context"
	"io"
	"strunetsdrive/internal/models"
)

type StorageService interface {
	UploadFile(username, filename string, content io.Reader, size int64) (*models.File, error)
	DownloadFile(id string) (io.ReadSeekCloser, *models.File, error)
	ListFiles(username string) ([]*models.File, error)
	GetFileDownloadURL(fileID string) (string, error)
}

type UserService interface {
	SingUp(ctx context.Context, input models.SignUpInput) error
	Login(ctx context.Context, input models.LoginInput) (string, string, error)
	ParseToken(ctx context.Context, token string) (int, string, error)
	RefreshSession(ctx context.Context, refreshToken string) (string, string, error)
}
