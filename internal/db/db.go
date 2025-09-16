package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/config"
	"github.com/jeongukjae/pypi-server/internal/db/dbgen"
)

//go:generate go tool go.uber.org/mock/mockgen -source=db.go -destination=./db_mock.go -package=db Store

type ListReleasesByPackageNameSimpleRow struct {
	Version         string
	FileName        string
	FileType        *string
	Md5Digest       *string
	Sha256Digest    *string
	Blake2256Digest *string
	RequiresPython  *string
}

type Store interface {
	Migrate(ctx context.Context, migrationQueryPath string) error
	ListPackagesSimple(ctx context.Context) ([]string, error)
	ListReleasesByPackageNameSimple(ctx context.Context, packageName string) ([]ListReleasesByPackageNameSimpleRow, error)
	Close(ctx context.Context) error
}

func New(ctx context.Context, cfg *config.DatabaseConfig) (Store, error) {
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

func (d *db) ListPackagesSimple(ctx context.Context) ([]string, error) {
	querier := dbgen.New(d.pool)
	packages, err := querier.ListPackagesSimple(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list packages")
	}
	return packages, nil
}

func (d *db) ListReleasesByPackageNameSimple(ctx context.Context, packageName string) ([]ListReleasesByPackageNameSimpleRow, error) {
	querier := dbgen.New(d.pool)
	rows, err := querier.ListReleasesByPackageNameSimple(ctx, pgtype.Text{String: packageName, Valid: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list releases by package name")
	}

	result := make([]ListReleasesByPackageNameSimpleRow, len(rows))
	for i, row := range rows {
		result[i] = ListReleasesByPackageNameSimpleRow{
			Version:         row.Version,
			FileName:        row.FileName.String,
			FileType:        getStringFromText(row.FileType),
			Md5Digest:       getStringFromText(row.Md5Digest),
			Sha256Digest:    getStringFromText(row.Sha256Digest),
			Blake2256Digest: getStringFromText(row.Blake2256Digest),
			RequiresPython:  getStringFromText(row.RequiresPython),
		}
	}
	return result, nil
}

func (d *db) Close(ctx context.Context) error {
	d.pool.Close()
	return nil
}

func getStringFromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
