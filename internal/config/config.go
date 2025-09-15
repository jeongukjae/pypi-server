package config

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host                     string `mapstructure:"host"`
	Port                     int    `mapstructure:"port"`
	ReadHeaderTimeoutSeconds int    `mapstructure:"read_header_timeout_seconds"`
	GracefulShutdownSeconds  int    `mapstructure:"graceful_shutdown_seconds"`
}

type StorageConfig struct {
	Kind string `mapstructure:"kind"`
	Path string `mapstructure:"path"`
}

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`

	LogLevel string `mapstructure:"log_level"`
}

func MustInit() *Config {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("server.host", "")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.read_header_timeout_seconds", 5)
	viper.SetDefault("storage.kind", "local")
	viper.SetDefault("storage.path", "./data")

	viper.AutomaticEnv()
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("PYPI_SERVER")

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config")
	}

	return cfg
}
