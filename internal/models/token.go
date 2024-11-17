package models

import "time"

type RefreshSession struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
}
