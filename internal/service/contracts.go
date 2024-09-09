package service

import (
	"context"
	"strunetsdrive/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) error
}

type SessionRepository interface {
}
