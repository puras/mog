package oss

import (
	"context"
	"github.com/rs/xid"
	"io"
	"path/filepath"
	"time"
)

type Client interface {
	PubObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, options ...PutObjectOptions) (*PutObjectResult, error)
	GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)
	RemoveObject(ctx context.Context, bucketName, objectName string) error
	StatObject(ctx context.Context, bucketName, objectName string) (*ObjectStat, error)
}

type PutObjectOptions struct {
	ContentType  string
	UserMetadata map[string]string
}

type PutObjectResult struct {
	URL  string
	ETag string
}

type ObjectStat struct {
	Key          string
	ETag         string
	LastModified time.Time
	Size         int64
	ContentType  string
	UserMetadata map[string]string
}

func (o *ObjectStat) GetName() string {
	if name, ok := o.UserMetadata["name"]; ok {
		return name
	}
	return filepath.Base(o.Key)
}

func formatObjectName(objectName string) string {
	if objectName == "" {
		return xid.New().String()
	}
	if objectName[0] == '/' {
		objectName = objectName[1:]
	}
	return objectName
}
