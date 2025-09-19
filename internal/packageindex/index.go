package packageindex

import (
	"context"
	"io"
	"path"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/db"
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
	ListPackages(ctx context.Context) ([]string, error)
	ListPackageFiles(ctx context.Context, packageName string) ([]PackageFile, error)
	DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error)
	UploadFile(ctx context.Context, req UploadFileRequest, content io.Reader) error
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
	return i.dbstore.ListPackages(ctx)
}

func (i *index) ListPackageFiles(ctx context.Context, packageName string) ([]PackageFile, error) {
	packageName = utils.NormalizePackageName(packageName)

	rows, err := i.dbstore.ListReleaseFiles(ctx, packageName)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list package files from database")
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

func (i *index) DownloadFile(ctx context.Context, packageName, fileName string) (io.ReadCloser, error) {
	packageName = utils.NormalizePackageName(packageName)

	row, err := i.dbstore.GetReleaseFile(ctx, packageName, fileName)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get release by file name from database")
		return nil, errors.Wrap(err, "failed to get release by file name from database")
	}
	if row == nil {
		return nil, errors.Errorf("release file %s/%s not found", packageName, fileName)
	}

	return i.strg.ReadFile(ctx, row.FilePath)
}

func (i *index) UploadFile(ctx context.Context, req UploadFileRequest, content io.Reader) error {
	filepath := path.Join(req.PackageName, uuid.NewString())
	if err := i.strg.WriteFile(ctx, filepath, content); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to write file to storage")
		return errors.Wrap(err, "failed to write file to storage")
	}

	if err := i.dbstore.CreateRelease(ctx, db.CreateReleaseRequest{
		PackageName:            req.PackageName,
		Version:                req.Version,
		MetadataVersion:        req.MetadataVersion,
		Summary:                req.Summary,
		Description:            req.Description,
		DescriptionContentType: req.DescriptionContentType,
		FileName:               req.FileName,
		FileType:               req.FileType,
		FilePath:               filepath,
		Pyversion:              req.Pyversion,
		RequiresPython:         req.RequiresPython,
		RequiresDist:           req.RequiresDist,
		Md5Digest:              req.Md5Digest,
		Sha256Digest:           req.Sha256Digest,
		Blake2256Digest:        req.Blake2256Digest,
	}); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create release in database")
		// unlink the file from storage back if db insert fails
		if err := i.strg.DeleteFile(ctx, filepath); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to delete file from storage after db insert failure")
		}

		return errors.Wrap(err, "failed to create release in database")
	}
	return nil
}

func pointer[T any](v T) *T {
	return &v
}
