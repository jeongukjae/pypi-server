package packageindex

import (
	"context"
	"io"

	"github.com/jeongukjae/pypi-server/internal/db"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

type Index interface {
	ListPackages(ctx context.Context) ([]string, error)
	ListPackageFiles(ctx context.Context, packageName string) ([]string, error)
	DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error)
}

func NewIndex(
	strg storage.Storage,
	dbstore db.Store,
) Index {
	return &index{
		strg:    strg,
		dbstore: dbstore,
	}
}

type index struct {
	strg    storage.Storage
	dbstore db.Store
}

func (i *index) ListPackages(ctx context.Context) ([]string, error) {
	if i.dbstore != nil {
		return i.dbstore.ListPackagesSimple(ctx)
	}
	return i.strg.ListPackages(ctx)
}

func (i *index) ListPackageFiles(ctx context.Context, packageName string) ([]string, error) {
	if i.dbstore != nil {
		rows, err := i.dbstore.ListReleasesByPackageNameSimple(ctx, packageName)
		if err != nil {
			return nil, err
		}

		// Add hash sum and py version info as PEP 503 suggests
		// https://peps.python.org/pep-0503/#specification
		files := make([]string, len(rows))
		for j, row := range rows {
			files[j] = row.FileName
		}
		return files, nil
	}

	return i.strg.ListPackageFiles(ctx, packageName)
}

func (i *index) DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error) {
	return i.strg.ReadFile(ctx, packageName, fileName)
}
