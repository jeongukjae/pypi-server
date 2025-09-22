package main

//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/config"
	"github.com/jeongukjae/pypi-server/internal/db"
	internalMw "github.com/jeongukjae/pypi-server/internal/middleware"
	"github.com/jeongukjae/pypi-server/internal/packageindex"
	"github.com/jeongukjae/pypi-server/internal/routes"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

func main() { //nolint:funlen // Function length is acceptable here for the sake of clarity.
	configFilePath := flag.String("config", "", "Path to config file")
	flag.Parse()

	ctx := context.Background()

	cfg := config.MustInit(configFilePath)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid log level")
	}
	zerolog.SetGlobalLevel(logLevel)

	strg, err := storage.New(ctx, &cfg.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage")
	}

	dbstore, err := initializeDBStore(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	index := packageindex.NewIndex(strg, dbstore)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(internalMw.Logger())

	if cfg.Server.EnableAccessLogger {
		e.Use(accessLogger())
	}

	routes.SetupSimpleRoutes(e, index)
	routes.SetupLegacyRoutes(e, index)

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Info().Msgf("Starting server at %s", addr)
		if err := e.Start(addr); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cfg.Server.GracefulShutdownSeconds))
	defer cancel()

	log.Info().Msg("Shutting down server")
	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}

	log.Info().Msg("Closing storage")
	if err := strg.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing storage")
	}

	log.Info().Msg("Closing database")
	if err := dbstore.Close(ctx); err != nil {
		log.Error().Err(err).Msg("Error closing database")
	}

	log.Info().Msg("Server stopped")
}

func initializeDBStore(ctx context.Context, cfg *config.Config) (db.Store, error) {
	dbstore, err := db.NewStore(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Migrating database")
	if err := dbstore.Migrate(ctx, cfg.Database.MigrationPath); err != nil {
		return nil, err
	}

	if _, err = dbstore.GetUserByUsername(ctx, cfg.Username); err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			return nil, errors.Wrap(err, "failed to get user by username")
		}

		log.Info().Msgf("Creating initial user: %s", cfg.Username)
		createdUser, err := dbstore.CreateUser(ctx, cfg.Username, cfg.Password, db.RoleAdmin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create initial user")
		}
		log.Info().Msgf("Initial user created (Only shown once): ID=%d, Username=%s, Password=%s", createdUser.ID, createdUser.Username, cfg.Password)
	}

	return dbstore, nil
}

func accessLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogMethod:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			l := log.Ctx(c.Request().Context())
			if v.Error == nil {
				l.Info().Str("method", v.Method).Str("URI", v.URI).Int("status", v.Status).Msg("request")
			} else {
				l.Error().Str("method", v.Method).Str("URI", v.URI).Int("status", v.Status).Err(v.Error).Msg("request")
			}
			return nil
		},
	})
}
