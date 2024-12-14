package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"
)

type MinioStore struct {
	client     *minio.Client
	bucketName string
}
type ObjectInfo struct {
	Path         string
	Size         int64
	ContentType  string
	LastModified time.Time
	IsDirectory  bool
}

func NewStore(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL bool) (*MinioStore, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	exist, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exist {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinioStore{
		client:     client,
		bucketName: bucketName,
	}, nil

}

type MinioWriter struct {
	ctx        context.Context
	client     *minio.Client
	bucketName string
	objectName string
	pipeline   *io.PipeWriter
}

func (m *MinioStore) Create(path string) (io.WriteCloser, error) {
	reader, writer := io.Pipe()
	ctx := context.Background()

	go func() {
		_, err := m.client.PutObject(ctx, m.bucketName, path, reader, -1, minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		})
		if err != nil {
			reader.CloseWithError(err)
		}
	}()

	return &MinioWriter{
		ctx:        ctx,
		client:     m.client,
		bucketName: m.bucketName,
		objectName: path,
		pipeline:   writer,
	}, nil
}

func (m *MinioWriter) Write(p []byte) (n int, err error) {
	return m.pipeline.Write(p)
}

func (m *MinioWriter) Close() error {
	return m.pipeline.Close()
}

type MinioReader struct {
	ctx    context.Context
	object *minio.Object
}

func (m *MinioStore) Open(path string) (io.ReadSeekCloser, error) {
	ctx := context.Background()
	object, err := m.client.GetObject(ctx, m.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to open object: %w", err)
	}

	return &MinioReader{
		ctx:    ctx,
		object: object,
	}, nil
}

func (m *MinioReader) Read(p []byte) (n int, err error) {
	return m.object.Read(p)
}

func (m *MinioReader) Seek(offset int64, whence int) (int64, error) {
	return m.object.Seek(offset, whence)
}

func (m *MinioReader) Close() error {
	return m.object.Close()
}

func (m *MinioStore) GetPresignedURL(path string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(context.Background(), m.bucketName, path, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *MinioStore) GetPresignedURLWithParams(path string, expires time.Duration, params url.Values) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(context.Background(),
		s.bucketName,
		path,
		expires,
		params)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL with params: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *MinioStore) Delete(path string) error {
	return s.client.RemoveObject(context.Background(), s.bucketName, path, minio.RemoveObjectOptions{})
}

func (m *MinioStore) CreateDirectory(path string) error {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	_, err := m.client.PutObject(
		context.Background(),
		m.bucketName,
		path,
		bytes.NewReader([]byte{}),
		0,
		minio.PutObjectOptions{ContentType: "application/x-directory"},
	)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func (m *MinioStore) MoveObject(sourcePath, destPath string) error {
	src := minio.CopySrcOptions{
		Bucket: m.bucketName,
		Object: sourcePath,
	}
	dst := minio.CopyDestOptions{
		Bucket: m.bucketName,
		Object: destPath,
	}

	_, err := m.client.CopyObject(context.Background(), dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	err = m.Delete(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to delete source object after move: %w", err)
	}

	return nil
}

func (m *MinioStore) ListObjects(prefix string) ([]ObjectInfo, error) {
	ctx := context.Background()
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	}

	var objects []ObjectInfo
	for object := range m.client.ListObjects(ctx, m.bucketName, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		info := ObjectInfo{
			Path:         object.Key,
			Size:         object.Size,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
			IsDirectory:  strings.HasSuffix(object.Key, "/"),
		}
		objects = append(objects, info)
	}

	return objects, nil
}

func (m *MinioStore) GetObjectInfo(path string) (*ObjectInfo, error) {
	ctx := context.Background()
	info, err := m.client.StatObject(ctx, m.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &ObjectInfo{
		Path:         info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		IsDirectory:  strings.HasSuffix(info.Key, "/"),
	}, nil
}

func (m *MinioStore) ObjectExists(path string) (bool, error) {
	_, err := m.GetObjectInfo(path)
	if err != nil {
		if errResponse, ok := err.(minio.ErrorResponse); ok {
			if errResponse.Code == "NoSuchKey" || errResponse.Code == "NoSuchObject" {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (m *MinioStore) DeleteDirectory(path string) error {
	ctx := context.Background()

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	objectsCh := make(chan minio.ObjectInfo)
	opts := minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}

	go func() {
		defer close(objectsCh)
		for object := range m.client.ListObjects(ctx, m.bucketName, opts) {
			if object.Err != nil {
				return
			}
			objectsCh <- object
		}
	}()

	objectsToRemove := make(chan minio.ObjectInfo, 1000)
	go func() {
		defer close(objectsToRemove)
		for object := range objectsCh {
			objectsToRemove <- object
		}
	}()

	errorCh := m.client.RemoveObjects(ctx, m.bucketName, objectsToRemove, minio.RemoveObjectsOptions{})

	var deleteErrors []error
	for err := range errorCh {
		if err.Err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("failed to delete %s: %w", err.ObjectName, err.Err))
		}
	}

	if len(deleteErrors) > 0 {
		var errMsg strings.Builder
		for _, err := range deleteErrors {
			errMsg.WriteString(err.Error() + "; ")
		}
		return fmt.Errorf("failed to delete some objects: %s", errMsg.String())
	}

	return nil
}

func (m *MinioStore) SafeDeleteDirectory(path string) error {
	objects, err := m.ListObjects(path)
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	if len(objects) > 0 {
		return fmt.Errorf("directory %s is not empty", path)
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return m.Delete(path)
}

func (m *MinioStore) GetDirectorySize(path string) (int64, error) {
	ctx := context.Background()
	var totalSize int64

	opts := minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}

	for object := range m.client.ListObjects(ctx, m.bucketName, opts) {
		if object.Err != nil {
			return 0, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		totalSize += object.Size
	}

	return totalSize, nil
}

func (m *MinioStore) DeleteDirectoryParallel(path string) error {
	ctx := context.Background()
	workers := 10

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	objectsCh := make(chan string, 1000)
	errorsCh := make(chan error, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for objectName := range objectsCh {
				err := m.Delete(objectName)
				if err != nil {
					errorsCh <- fmt.Errorf("failed to delete %s: %w", objectName, err)
					return
				}
			}
		}()
	}

	opts := minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}

	for object := range m.client.ListObjects(ctx, m.bucketName, opts) {
		if object.Err != nil {
			close(objectsCh)
			return fmt.Errorf("failed to list objects: %w", object.Err)
		}
		objectsCh <- object.Key
	}

	close(objectsCh)
	wg.Wait()
	close(errorsCh)

	var errors []error
	for err := range errorsCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("deletion errors occurred: %v", errors)
	}

	return nil
}
