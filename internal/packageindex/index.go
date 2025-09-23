package packageindex

import (
	"context"
	"io"
	"path"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/storage"
	"github.com/jeongukjae/pypi-server/internal/utils"
)

type PackageFile struct {
	FileName string
	FileType string

	HashType  *string
	HashValue *string

	RequiresPython *string

	// We don't currently support gpg signature, so this field is always false.
	HasGpgSignature bool
}

type Authorization struct {
	Username string
	Password string
}

type UploadFileRequest struct {
	PackageName string
	Version     string
	FileName    string
	FileType    string

	MetadataVersion        string
	Summary                *string
	Description            *string
	DescriptionContentType *string
	Pyversion              *string
	RequiresPython         *string
	RequiresDist           []string
	Md5Digest              *string
	Sha256Digest           *string
	Blake2256Digest        *string
}

//go:generate go tool go.uber.org/mock/mockgen -source=index.go -destination=./index_mock.go -package=packageindex Index

type Index interface {
	// TODO: Add authorization to these methods.
	ListPackages(ctx context.Context) ([]string, error)
	ListPackageFiles(ctx context.Context, packageName string) ([]string, error)
	DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error)

	UploadFile(ctx context.Context, auth Authorization, req UploadFileRequest, content io.Reader) error
}

func NewIndex(strg storage.Storage) Index {
	return &index{
		strg: strg,
	}
}

type index struct {
	strg storage.Storage
}

func (i *index) ListPackages(ctx context.Context) ([]string, error) {
	return i.strg.ListPackages(ctx)
}

func (i *index) ListPackageFiles(ctx context.Context, packageName string) ([]string, error) {
	packageName = utils.NormalizePackageName(packageName)

	files, err := i.strg.ListPackageFiles(ctx, packageName)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list package files from storage")
		return nil, errors.Wrap(err, "failed to list package files from storage")
	}

	return files, nil
}

func (i *index) DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error) {
	packageName = utils.NormalizePackageName(packageName)
	return i.strg.ReadFile(ctx, path.Join(packageName, fileName))
}

func (i *index) UploadFile(ctx context.Context, auth Authorization, req UploadFileRequest, content io.Reader) error {
	filepath := path.Join(req.PackageName, req.FileName)
	if err := i.strg.WriteFile(ctx, filepath, content); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to write file to storage")
		return errors.Wrap(err, "failed to write file to storage")
	}
	return nil
}
