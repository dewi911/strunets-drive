package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strunetsdrive/internal/models"
)

type Users struct {
	db *sqlx.DB
}

func NewUsers(db *sqlx.DB) *Users {
	return &Users{db}
}

func (r *Users) Create(ctx context.Context, user models.User) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "create",
		"user":       user,
	}

	query := "INSERT INTO users(name,password, created_at) values ($1,$2, now())"

	_, err := r.db.ExecContext(ctx, query, user.Username, user.Password)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("failed to create user query")

		return errors.Wrap(err, "failed to create user query")
	}

	return nil
}

func (r *Users) GetByID(ctx context.Context, id int) (models.User, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "GetByID",
		"id":         id,
	}

	var user models.User

	query := "SELECT id, username, password, created_at FROM users WHERE id = $1"

	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("failed to get user by id")

		return models.User{}, errors.Wrap(err, "failed to get user by id")

	}

	return user, nil
}
