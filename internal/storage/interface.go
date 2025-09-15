package storage

import (
	"context"
	"errors"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/config"
)

type Storage interface {
	ListPackages(context.Context) ([]string, error)
	ListPackageFiles(context.Context, string) ([]string, error)
	ReadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error)
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
