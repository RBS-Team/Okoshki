package minio

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"strings"
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
	PublicURL string `mapstructure:"publicURL"`
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
	publicURL   string
}

func New(cfg Config) (*Client, error) {
	mClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	return &Client{minioClient: mClient, publicURL: strings.TrimRight(cfg.PublicURL, "/")}, nil
}

// BuildObjectURL returns a permanent public URL for the given object.
// Requires the bucket to have a public-read policy (set by EnsureBucket).
func (c *Client) BuildObjectURL(bucket, objectName string) string {
	return fmt.Sprintf("%s/%s/%s", c.publicURL, bucket, objectName)
}

// EnsureBucket creates the bucket if it does not exist and sets a public-read policy on it.
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

	policy := fmt.Sprintf(`{
		"Version":"2012-10-17",
		"Statement":[{
			"Effect":"Allow",
			"Principal":{"AWS":["*"]},
			"Action":["s3:GetObject"],
			"Resource":["arn:aws:s3:::%s/*"]
		}]
	}`, bucketName)

	if err := c.minioClient.SetBucketPolicy(ctx, bucketName, policy); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
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
