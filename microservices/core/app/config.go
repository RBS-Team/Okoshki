package app

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/RBS-Team/Okoshki/internal/app"
	"github.com/RBS-Team/Okoshki/pkg/postgres"
)

type Config struct {
	Auth AuthConfig      `mapstructure:"auth"`
	DB   postgres.Config `mapstructure:"db"`
}

type AuthConfig struct {
	HTTP   app.HTTPConfig   `mapstructure:"http"`
	Logger app.LoggerConfig `mapstructure:"logger"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	configName := os.Getenv("CONFIG_FILE")
	if configName == "" {
		configName = "config.dev"
	}
	v.SetConfigName(configName)
	v.SetConfigType("yml")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := app.BindViperEnv(v); err != nil {
		return nil, fmt.Errorf("failed to bind env variables: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth config: %w", err)
	}

	return &cfg, nil
}
