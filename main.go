package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/config"
	"github.com/jeongukjae/pypi-server/internal/routes"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

func main() {
	cfg := config.MustInit()
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

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	routes.SetupSimpleRoutes(e, strg)

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

	ctx := context.Background()
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

	log.Info().Msg("Server stopped")
}
