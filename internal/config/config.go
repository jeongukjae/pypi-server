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
	EnableAccessLogger       bool   `mapstructure:"enable_access_logger"`
}

type LocalConfig struct {
	Path string `mapstructure:"path"`
}

type S3Config struct {
	Bucket       string `mapstructure:"bucket"`
	Prefix       string `mapstructure:"prefix"`
	Region       string `mapstructure:"region"`
	Endpoint     string `mapstructure:"endpoint"`
	UsePathStyle bool   `mapstructure:"use_path_style"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
}

type StorageConfig struct {
	Kind string `mapstructure:"kind"`

	Local LocalConfig `mapstructure:"local"`
	S3    S3Config    `mapstructure:"s3"`
}

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`

	LogLevel string `mapstructure:"log_level"`
	HTPasswd string `mapstructure:"htpasswd"`
}

func MustInit(configFilePath *string) *Config {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("server.host", "")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.read_header_timeout_seconds", 5)
	viper.SetDefault("server.graceful_shutdown_seconds", 10)
	viper.SetDefault("server.enable_access_logger", true)
	viper.SetDefault("storage.kind", "local")
	viper.SetDefault("storage.local.path", "./data")
	viper.SetDefault("htpasswd", "./htpasswd")

	viper.AutomaticEnv()
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("PYPI_SERVER")

	if configFilePath != nil && *configFilePath != "" {
		viper.SetConfigFile(*configFilePath)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Msg("failed to read config file")
		}
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config")
	}

	return cfg
}
