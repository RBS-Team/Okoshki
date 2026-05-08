package minio

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`
	UseSSL    bool   `mapstructure:"useSSL"`
}

type ObjectInfo struct {
	BucketName   string
	ObjectName   string
	Reader       io.Reader
	Size         int64
	ContentType  string
	OriginalName string
}

type Client struct {
	minioClient *minio.Client
}

func New(cfg Config) (*Client, error) {
	mClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	return &Client{minioClient: mClient}, nil
}

// Проверяет существует ли бакет с bucketName, если нет, то создает его
func (c *Client) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := c.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := c.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		log.Printf("Bucket %s создан", bucketName)
	}
	return nil
}

func (c *Client) Upload(ctx context.Context, obj ObjectInfo) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := c.EnsureBucket(ctx, obj.BucketName); err != nil {
		return "", err
	}

	exts, _ := mime.ExtensionsByType(obj.ContentType)
	ext := ".bin"
	if len(exts) > 0 {
		ext = exts[0]
	}
	objectName := uuid.New().String() + ext

	_, err := c.minioClient.PutObject(ctx, obj.BucketName, objectName, obj.Reader, obj.Size, minio.PutObjectOptions{
		ContentType: obj.ContentType,
		UserMetadata: map[string]string{
			"name": obj.OriginalName,
		},
	})
	if err != nil {
		return "", fmt.Errorf("upload object: %w", err)
	}
	return objectName, nil
}

func (c *Client) Remove(ctx context.Context, bucket, objectName string) error {
	if err := c.minioClient.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}

func (c *Client) GetFileURL(ctx context.Context, bucket, objectName string) (string, error) {
	params := make(url.Values)
	u, err := c.minioClient.PresignedGetObject(ctx, bucket, objectName, time.Hour, params)
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}
	return u.String(), nil
}
