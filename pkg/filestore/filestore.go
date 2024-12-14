package filestore

import (
	"io"
	"strunetsdrive/pkg/filestore/minio"
	"time"
)

type Store interface {
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadSeekCloser, error)
	GetPresignedURL(path string, expires time.Duration) (string, error)
	Delete(path string) error
	CreateDirectory(path string) error
	MoveObject(sourcePath, destPath string) error
	ListObjects(prefix string) ([]minio.ObjectInfo, error)
	GetObjectInfo(path string) (*minio.ObjectInfo, error)
	ObjectExists(path string) (bool, error)
	DeleteDirectory(path string) error
	SafeDeleteDirectory(path string) error
	GetDirectorySize(path string) (int64, error)
	DeleteDirectoryParallel(path string) error
}
