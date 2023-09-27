package oss

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClientConfig struct {
	Domain          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

var _ Client = (*MinioClient)(nil)

func NewMinioClient(config MinioClientConfig) (*MinioClient, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if exists, err := client.BucketExists(ctx, config.BucketName); err != nil {
		return nil, err
	} else if !exists {
		if err := client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	return &MinioClient{
		config: config,
		client: client,
	}, nil
}

type MinioClient struct {
	config MinioClientConfig
	client *minio.Client
}

func (o *MinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, options ...PutObjectOptions) (*PutObjectResult, error) {
	if bucketName == "" {
		bucketName = o.config.BucketName
	}

	var opt PutObjectOptions
	if len(options) > 0 {
		opt = options[0]
	}
	objectName = formatObjectName(objectName)
	output, err := o.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType:  opt.ContentType,
		UserMetadata: opt.UserMetadata,
	})
	if err != nil {
		return nil, err
	}

	return &PutObjectResult{
		URL:  o.config.Domain + "/" + objectName,
		ETag: output.ETag,
	}, nil
}

func (o *MinioClient) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	if bucketName == "" {
		bucketName = o.config.BucketName
	}
	objectName = formatObjectName(objectName)
	return o.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}

func (o *MinioClient) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	if bucketName == "" {
		bucketName = o.config.BucketName
	}
	objectName = formatObjectName(objectName)
	return o.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

func (o *MinioClient) StatObject(ctx context.Context, bucketName, objectName string) (*ObjectStat, error) {
	if bucketName == "" {
		bucketName = o.config.BucketName
	}

	objectName = formatObjectName(objectName)
	info, err := o.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &ObjectStat{
		Key:          info.Key,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		UserMetadata: info.UserMetadata,
	}, nil
}
