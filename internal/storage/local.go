package storage

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/config"
)

type LocalStorage struct {
	cfg *config.LocalConfig
}

func NewLocalStorage(cfg *config.LocalConfig) *LocalStorage {
	log.Info().Msgf("Using local storage at path: %s", cfg.Path)
	return &LocalStorage{cfg: cfg}
}

func (s *LocalStorage) ListPackages(context.Context) ([]string, error) {
	osFiles, err := os.ReadDir(s.cfg.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, err
	}

	packages := make([]string, 0, len(osFiles))
	for _, f := range osFiles {
		if f.IsDir() {
			packages = append(packages, f.Name())
		}
	}

	return packages, nil
}

func (s *LocalStorage) ListPackageFiles(_ context.Context, packageName string) ([]string, error) {
	osFiles, err := os.ReadDir(path.Join(s.cfg.Path, packageName))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, err
	}

	files := make([]string, 0, len(osFiles))
	for _, f := range osFiles {
		if !f.IsDir() {
			files = append(files, f.Name())
		}
	}

	return files, nil
}

func (s *LocalStorage) ReadFile(_ context.Context, filepath string) (io.ReadCloser, error) {
	return os.Open(path.Join(s.cfg.Path, filepath))
}

func (s *LocalStorage) WriteFile(_ context.Context, filepath string, content io.Reader) error {
	fullPath := path.Join(s.cfg.Path, filepath)
	parentPath := path.Dir(fullPath)
	if err := os.MkdirAll(parentPath, 0750); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, content)
	return err
}

func (s *LocalStorage) DeleteFile(_ context.Context, filepath string) error {
	return os.Remove(path.Join(s.cfg.Path, filepath))
}

func (s *LocalStorage) Close() error {
	return nil
}
