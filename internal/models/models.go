package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type File struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Path       string    `json:"path" db:"path"`
	Size       int64     `json:"size" db:"size"`
	Username   string    `json:"username" db:"username"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

type FileResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	DownloadURL string    `json:"download_url,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
