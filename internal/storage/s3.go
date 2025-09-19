package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/jeongukjae/pypi-server/internal/config"
)

type S3Storage struct {
	bucket string
	prefix string
	client *s3.Client
}

func NewS3Storage(ctx context.Context, cfg *config.S3Config) (*S3Storage, error) {
	cred := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(cred),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.UsePathStyle
	})
	return &S3Storage{bucket: cfg.Bucket, prefix: cfg.Prefix, client: client}, nil
}

func (s *S3Storage) ListPackages(ctx context.Context) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(s.prefix),
		Delimiter: aws.String("/"),
	}
	resp, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}
	var packages []string
	for _, cp := range resp.CommonPrefixes {
		name := strings.TrimPrefix(*cp.Prefix, s.prefix)
		name = strings.TrimSuffix(name, "/")
		if name != "" {
			packages = append(packages, name)
		}
	}
	return packages, nil
}

func (s *S3Storage) ListPackageFiles(ctx context.Context, packageName string) ([]string, error) {
	prefix := path.Join(s.prefix, packageName) + "/"
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}
	resp, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, obj := range resp.Contents {
		name := strings.TrimPrefix(*obj.Key, prefix)
		if name != "" && !strings.HasSuffix(name, "/") {
			files = append(files, name)
		}
	}
	return files, nil
}

func (s *S3Storage) ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	key := path.Join(s.prefix, filePath)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	resp, err := s.client.GetObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return resp.Body, nil
}

func (s *S3Storage) WriteFile(ctx context.Context, filePath string, content io.Reader) error {
	key := path.Join(s.prefix, filePath)
	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   content,
	})
	return err
}

func (s *S3Storage) DeleteFile(ctx context.Context, filePath string) error {
	key := path.Join(s.prefix, filePath)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) Close() error {
	return nil
}
