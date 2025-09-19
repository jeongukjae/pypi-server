package storage

import (
	"context"
	"errors"
	"io"

	"github.com/jeongukjae/pypi-server/internal/config"
)

//go:generate go tool go.uber.org/mock/mockgen -source=interface.go -destination=./interface_mock.go -package=storage Storage

type Storage interface {
	ListPackages(context.Context) ([]string, error)
	ListPackageFiles(context.Context, string) ([]string, error)
	ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
	WriteFile(ctx context.Context, path string, content io.Reader) error
	DeleteFile(ctx context.Context, path string) error
	Close() error
}

func New(ctx context.Context, cfg *config.StorageConfig) (Storage, error) {
	switch cfg.Kind {
	case "local":
		return NewLocalStorage(&cfg.Local), nil
	case "s3":
		return NewS3Storage(ctx, &cfg.S3)
	default:
		return nil, errors.New("unknown storage kind: " + cfg.Kind)
	}
}
