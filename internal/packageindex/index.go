package packageindex

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"github.com/jeongukjae/pypi-server/internal/db"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

type PackageFile struct {
	FileName string
	FileType *string

	HashType  *string
	HashValue *string

	RequiresPython *string

	// We don't currently support gpg signature, so this field is always false.
	HasGpgSignature bool
}

type Index interface {
	ListPackages(ctx context.Context) ([]string, error)
	ListPackageFiles(ctx context.Context, packageName string) ([]PackageFile, error)
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

func (i *index) ListPackageFiles(ctx context.Context, packageName string) ([]PackageFile, error) {
	packageName = NormalizePackageName(packageName)

	if i.dbstore != nil {
		rows, err := i.dbstore.ListReleasesByPackageNameSimple(ctx, packageName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list package files from database")
		}

		// Add hash sum and py version info as PEP 503 suggests
		// https://peps.python.org/pep-0503/#specification
		files := make([]PackageFile, len(rows))
		for j, row := range rows {
			files[j] = PackageFile{
				FileName:        row.FileName,
				FileType:        row.FileType,
				HashType:        nil,
				HashValue:       nil,
				RequiresPython:  row.RequiresPython,
				HasGpgSignature: false,
			}

			switch {
			case row.Md5Digest != nil:
				files[j].HashType = pointer("md5")
				files[j].HashValue = row.Md5Digest
			case row.Sha256Digest != nil:
				files[j].HashType = pointer("sha256")
				files[j].HashValue = row.Sha256Digest
			case row.Blake2256Digest != nil:
				files[j].HashType = pointer("blake2_256")
				files[j].HashValue = row.Blake2256Digest
			}
		}
		return files, nil
	}

	file, err := i.strg.ListPackageFiles(ctx, packageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list package files from storage")
	}

	pkgFiles := make([]PackageFile, len(file))
	for i, f := range file {
		pkgFiles[i] = PackageFile{
			FileName: f,
		}
	}
	return pkgFiles, nil
}

func (i *index) DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error) {
	packageName = NormalizePackageName(packageName)

	return i.strg.ReadFile(ctx, packageName, fileName)
}

func pointer[T any](v T) *T {
	return &v
}
