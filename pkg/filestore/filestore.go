package filestore

import (
	"io"
	"time"
)

type Store interface {
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadSeekCloser, error)
	GetPresignedURL(path string, expires time.Duration) (string, error)
	Delete(path string) error
}
