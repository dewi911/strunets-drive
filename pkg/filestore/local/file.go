package local

import (
	"os"
	"sync"
)

type File struct {
	file *os.File
	mu   sync.Mutex
	size int64
}

func NewFile(f *os.File) (*File, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &File{
		file: f,
		mu:   sync.Mutex{},
		size: info.Size(),
	}, nil
}

func (f *File) Read(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Read(p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Seek(offset, whence)
}

func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}

func (f *File) Size() int64 {
	return f.size
}

type Writer struct {
	file *os.File
	mu   sync.Mutex
}

func NewWriter(f *os.File) *Writer {
	return &Writer{f, sync.Mutex{}}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Write(p)
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
