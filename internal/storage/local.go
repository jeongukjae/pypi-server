package storage

import (
	"context"
	"io"
	"os"
	"path"
)

type LocalStorage struct {
	path string
}

func NewLocalStorage(path string) (*LocalStorage, error) {
	return &LocalStorage{path: path}, nil
}

func (s *LocalStorage) ListPackages(context.Context) ([]string, error) {
	osFiles, err := os.ReadDir(s.path)
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
	osFiles, err := os.ReadDir(path.Join(s.path, packageName))
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

func (s *LocalStorage) ReadFile(_ context.Context, packageName, fileName string) (io.ReadCloser, error) {
	return os.Open(path.Join(s.path, packageName, fileName))
}

func (s *LocalStorage) Close() error {
	return nil
}
