package filestore

type Reader interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
	Size() int64
}

type Writer interface {
	Write(p []byte) (n int, err error)
	Close() error
}

type Store interface {
	Create(path string) (Writer, error)
	Open(path string) (Reader, error)
	Remove(path string) error
	Exists(path string) (bool, error)
}
