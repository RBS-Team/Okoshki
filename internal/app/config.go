package app

import (
	"time"

	"github.com/spf13/viper"
)

type HTTPConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout     time.Duration `mapstructure:"idleTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
	Auth            AuthConfig    `mapstructure:"auth"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
	Mode  string `mapstructure:"mode"`
}

type AuthConfig struct {
	JWT  JWTConfig  `mapstructure:"jwt"`
}

type JWTConfig struct {
	SecretKey      string        `mapstructure:"secretKey"`
	AccessTokenTTL time.Duration `mapstructure:"accessTokenTTL"`
}

func BindViperEnv(v *viper.Viper) error {
	bindings := map[string]string{
		"db.password":                  "DB_PASSWORD",
		"db.user":                      "DB_USER",
		"db.host":                      "DB_HOST",
		"db.dbName":                    "DB_NAME",
		"auth.http.auth.jwt.secretKey": "JWT_SECRET",
	}

	for configKey, envKey := range bindings {
		if err := v.BindEnv(configKey, envKey); err != nil {
			return err
		}
	}

	return nil
}
