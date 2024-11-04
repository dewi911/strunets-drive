package local

//import (
//	"os"
//	"path/filepath"
//	"strunetsdrive/pkg/filestore"
//)
//
//type Store struct {
//	rootDir string
//}
//
//func NewStore(rootDir string) (*Store, error) {
//	if err := os.MkdirAll(rootDir, 0700); err != nil {
//		return nil, err
//	}
//	return &Store{rootDir: rootDir}, nil
//}
//
//func (s *Store) Create(path string) (filestore.Writer, error) {
//	fullPath := filepath.Join(s.rootDir, path)
//
//	dir := filepath.Dir(fullPath)
//	if err := os.MkdirAll(dir, 0700); err != nil {
//		return nil, err
//	}
//
//	file, err := os.Create(fullPath)
//	if err != nil {
//		return nil, err
//	}
//
//	return NewWriter(file), nil
//}
//
//func (s *Store) Open(path string) (filestore.Reader, error) {
//	fullPath := filepath.Join(s.rootDir, path)
//
//	file, err := os.Open(fullPath)
//	if err != nil {
//		return nil, err
//	}
//
//	return NewFile(file)
//}
//
//func (s *Store) Remove(path string) error {
//	return os.Remove(filepath.Join(s.rootDir, path))
//}
//
//func (s *Store) Exists(path string) (bool, error) {
//	_, err := os.Stat(filepath.Join(s.rootDir, path))
//	if os.IsNotExist(err) {
//		return false, nil
//	}
//	if err != nil {
//		return false, err
//	}
//	return true, nil
//}
