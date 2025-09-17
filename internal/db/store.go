package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/config"
	"github.com/jeongukjae/pypi-server/internal/utils"
)

//go:generate go tool go.uber.org/mock/mockgen -source=db.go -destination=./db_mock.go -package=db Store

type ListReleasesResponse struct {
	Version         string
	FileName        string
	FileType        string
	Md5Digest       *string
	Sha256Digest    *string
	Blake2256Digest *string
	RequiresPython  *string
}

type CreateReleaseRequest struct {
	PackageName            string
	Version                string
	MetadataVersion        string
	Summary                *string
	Description            *string
	DescriptionContentType *string
	FileName               string
	FileType               string
	FilePath               string
	Pyversion              *string
	RequiresPython         *string
	RequiresDist           []string
	Md5Digest              *string
	Sha256Digest           *string
	Blake2256Digest        *string
}

type Store interface {
	Migrate(ctx context.Context, migrationQueryPath string) error
	Close(ctx context.Context) error

	GetPackageByName(ctx context.Context, packageName string) (*Package, error)
	ListPackages(ctx context.Context) ([]string, error)
	CreateRelease(ctx context.Context, arg CreateReleaseRequest) error
	ListReleaseFiles(ctx context.Context, packageName string) ([]ListReleasesResponse, error)
	GetRelease(ctx context.Context, packageName, version string) (*GetReleaseRow, error)
	GetReleaseFile(ctx context.Context, packageName, fileName string) (*GetReleaseFileByNameRow, error)
}

func NewStore(ctx context.Context, cfg *config.DatabaseConfig) (Store, error) {
	pool, err := pgxpool.New(
		ctx,
		fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
			cfg.SSLMode,
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	instance := &db{
		pool: pool,
		cfg:  cfg,
	}
	return instance, nil
}

type db struct {
	pool *pgxpool.Pool
	cfg  *config.DatabaseConfig
}

func (d *db) Migrate(ctx context.Context, migrationQueryPath string) error {
	poolConn, err := d.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire connection from pool")
	}
	defer poolConn.Release()

	conn := poolConn.Conn()
	migrator, err := migrate.NewMigrator(ctx, conn, d.cfg.MigrationTableName)
	if err != nil {
		return errors.Wrap(err, "failed to create migrator")
	}

	if err := migrator.LoadMigrations(os.DirFS(migrationQueryPath)); err != nil {
		return errors.Wrap(err, "failed to load migrations")
	}

	log.Debug().Int("count", len(migrator.Migrations)).Msg("loaded migrations")

	if err := migrator.Migrate(ctx); err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	return nil
}

func (d *db) Close(ctx context.Context) error {
	d.pool.Close()
	return nil
}

func (d *db) GetPackageByName(ctx context.Context, packageName string) (*Package, error) {
	querier := New(d.pool)
	pkg, err := querier.GetPackageByName(ctx, packageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get package by name")
	}
	return &pkg, err
}

func (d *db) ListPackages(ctx context.Context) ([]string, error) {
	querier := New(d.pool)
	packages, err := querier.ListPackagesSimple(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list packages")
	}
	return packages, nil
}

func (d *db) CreateRelease(ctx context.Context, arg CreateReleaseRequest) error {
	version, err := utils.ParseVersion(arg.Version)
	if err != nil {
		return errors.Wrap(err, "failed to parse version")
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	commit := false
	defer func() {
		if !commit {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg("failed to rollback transaction")
			}
		}
	}()

	querier := New(tx)

	// 1. Check if the package exists, if not create it.
	log.Debug().Str("package", arg.PackageName).Msg("Checking if package exists")
	needsUpdatePackage := true
	currentPackage, err := querier.GetPackageByName(ctx, arg.PackageName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.Wrap(err, "failed to upsert package")
	} else if errors.Is(err, pgx.ErrNoRows) {
		needsUpdatePackage = false
		currentPackage, err = querier.CreatePackage(ctx, CreatePackageParams{
			Name:          arg.PackageName,
			Summary:       getTextFromString(arg.Summary),
			LatestVersion: pgtype.Text{String: arg.Version, Valid: true},
		})

		if err != nil {
			return errors.Wrap(err, "failed to create package")
		}
	}

	// 2. Upsert the release
	log.Debug().Str("package", arg.PackageName).Str("version", arg.Version).Msg("Upserting release")
	release, err := querier.UpsertRelease(ctx, UpsertReleaseParams{
		Version:                version.String(),
		PackageName:            arg.PackageName,
		MetadataVersion:        arg.MetadataVersion,
		Summary:                getTextFromString(arg.Summary),
		Description:            getTextFromString(arg.Description),
		DescriptionContentType: getTextFromString(arg.DescriptionContentType),
	})
	if err != nil {
		return errors.Wrap(err, "failed to upsert release")
	}

	// 3. Create the release file
	log.Debug().Str("package", arg.PackageName).Str("version", arg.Version).Str("file", arg.FileName).Msg("Creating release file")
	_, err = querier.CreateReleaseFile(ctx, CreateReleaseFileParams{
		PackageName:     arg.PackageName,
		Version:         release.Version,
		FileName:        arg.FileName,
		FileType:        arg.FileType,
		FilePath:        arg.FilePath,
		Pyversion:       getTextFromString(arg.Pyversion),
		RequiresPython:  getTextFromString(arg.RequiresPython),
		RequiresDist:    arg.RequiresDist,
		Md5Digest:       getTextFromString(arg.Md5Digest),
		Sha256Digest:    getTextFromString(arg.Sha256Digest),
		Blake2256Digest: getTextFromString(arg.Blake2256Digest),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create release file")
	}

	// 4. Update the package's latest version if needed
	log.Debug().Str("package", arg.PackageName).Str("version", arg.Version).Msg("Checking if package latest version needs update")
	if needsUpdatePackage {
		log.Debug().Str("current", currentPackage.LatestVersion.String).Str("new", version.String()).Msg("Comparing versions to update latest version")
		latestVersion := currentPackage.LatestVersion
		summary := currentPackage.Summary
		if !latestVersion.Valid {
			latestVersion.String = version.String()
			latestVersion.Valid = true
			summary = getTextFromString(arg.Summary)
		}

		if parsedLatestVersion, err := utils.ParseVersion(latestVersion.String); err != nil || version.Compare(parsedLatestVersion) > 0 {
			latestVersion.String = version.String()
			latestVersion.Valid = true
			summary = getTextFromString(arg.Summary)
		}

		if err := querier.UpdatePackageLatestVersion(ctx, UpdatePackageLatestVersionParams{
			Name:          arg.PackageName,
			LatestVersion: latestVersion,
			Summary:       summary,
		}); err != nil {
			return errors.Wrap(err, "failed to update package latest version")
		}
	}

	log.Debug().Str("package", arg.PackageName).Str("version", arg.Version).Msg("Committing transaction")
	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	commit = true
	return nil
}

func (d *db) ListReleaseFiles(ctx context.Context, packageName string) ([]ListReleasesResponse, error) {
	querier := New(d.pool)
	rows, err := querier.ListReleaseFilesByPackageNameSimple(ctx, packageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list releases by package name")
	}

	result := make([]ListReleasesResponse, len(rows))
	for i, row := range rows {
		result[i] = ListReleasesResponse{
			Version:         row.Version,
			FileName:        row.FileName,
			FileType:        row.FileType,
			Md5Digest:       getStringFromText(row.Md5Digest),
			Sha256Digest:    getStringFromText(row.Sha256Digest),
			Blake2256Digest: getStringFromText(row.Blake2256Digest),
			RequiresPython:  getStringFromText(row.RequiresPython),
		}
	}
	return result, nil
}

func (d *db) GetRelease(ctx context.Context, packageName, version string) (*GetReleaseRow, error) {
	querier := New(d.pool)
	row, err := querier.GetRelease(ctx, GetReleaseParams{
		PackageName: packageName,
		Version:     version,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get release")
	}

	return &row, nil
}

func (d *db) GetReleaseFile(ctx context.Context, packageName, fileName string) (*GetReleaseFileByNameRow, error) {
	querier := New(d.pool)
	row, err := querier.GetReleaseFileByName(ctx, GetReleaseFileByNameParams{
		PackageName: packageName,
		FileName:    fileName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get release file by name")
	}

	return &row, nil
}

func getTextFromString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

func getStringFromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
