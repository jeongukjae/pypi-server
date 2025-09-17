package storage

import (
	"context"
	"errors"
	"io"

	"github.com/rs/zerolog/log"

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

func New(cfg *config.StorageConfig) (Storage, error) {
	switch cfg.Kind {
	case "local":
		log.Info().Msgf("Using local storage at path: %s", cfg.Path)
		return NewLocalStorage(cfg.Path)
	default:
		return nil, errors.New("unknown storage kind: " + cfg.Kind)
	}
}
