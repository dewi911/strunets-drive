package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"time"
)

type MinioStore struct {
	client     *minio.Client
	bucketName string
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
