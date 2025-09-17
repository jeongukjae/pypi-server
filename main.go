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
	"github.com/jeongukjae/pypi-server/internal/packageindex"
	"github.com/jeongukjae/pypi-server/internal/routes"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

func main() { //nolint:funlen // Function length is acceptable here for the sake of clarity.
	configFilePath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cfg := config.MustInit(configFilePath)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid log level")
	}
	zerolog.SetGlobalLevel(logLevel)

	strg, err := storage.New(&cfg.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage")
	}

	var dbstore db.Store

	ctx := context.Background()

	log.Info().Msg("Initializing database")
	dbstore, err = db.NewStore(ctx, &cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	log.Info().Msg("Migrating database")
	if err := dbstore.Migrate(ctx, cfg.Database.MigrationPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	index := packageindex.NewIndex(strg, dbstore)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	if cfg.Server.EnableAccessLogger {
		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:      true,
			LogStatus:   true,
			LogError:    true,
			HandleError: true,
			LogMethod:   true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				reqID := c.Response().Header().Get(echo.HeaderXRequestID)
				if v.Error == nil {
					log.Info().Str("method", v.Method).Str("URI", v.URI).Int("status", v.Status).Str("requestId", reqID).Msg("request")
				} else {
					log.Error().Str("method", v.Method).Str("URI", v.URI).Int("status", v.Status).Str("requestId", reqID).Err(v.Error).Msg("request")
				}
				return nil
			},
		}))
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
