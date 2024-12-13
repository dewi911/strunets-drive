package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type FileResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Path        string    `json:"path,omitempty"`
	DownloadURL string    `json:"download_url,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
	FolderID    string    `json:"folder_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
