package models

import "time"

//type File struct {
//	ID         string    `json:"id" db:"id"`
//	Name       string    `json:"name" db:"name"`
//	Path       string    `json:"path" db:"path"`
//	Size       int64     `json:"size" db:"size"`
//	Username   string    `json:"username" db:"username"`
//	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
//}

type File struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Path       string    `db:"path"`
	Size       int64     `db:"size"`
	Username   string    `db:"username"`
	UploadedAt time.Time `db:"uploaded_at"`
	IsDir      bool      `db:"is_dir"`
	FolderID   string    `db:"folder_id"`
}

type Folder struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	ParentID  string    `db:"parent_id"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"created_at"`
	Files     []*File   `db:"-"`
	Folders   []*Folder `db:"-"`
}
