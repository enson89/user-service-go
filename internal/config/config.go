//nolint:mnd
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string        `mapstructure:"secret"`
	ExpireHours time.Duration `mapstructure:"expireHours"`
}

type Config struct {
	App   AppConfig   `mapstructure:"app"`
	DB    DBConfig    `mapstructure:"db"`
	Redis RedisConfig `mapstructure:"redis"`
	JWT   JWTConfig   `mapstructure:"jwt"`
}

// nolint:nestif
// Load reads configuration from file and environment variables.
func Load() (*Config, error) {
	// Allow overriding any field via env vars
	viper.SetEnvPrefix("USER_SVC")
	viper.AutomaticEnv()

	// Defaults ensures your service has sane fallback values if neither a config file nor env var is present.
	viper.SetDefault("app.env", "dev")
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "postgres")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.name", "userdb")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "supersecretkey")
	viper.SetDefault("jwt.expireHours", 2)

	viper.SetConfigType("yaml")
	viper.AddConfigPath("./internal/config")

	env := viper.GetString("app.env")
	if env == "dev" {
		viper.SetConfigName("config.dev")
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read dev config: %w", err)
		}
	} else {
		viper.SetConfigName("config.prod")
		if err := viper.ReadInConfig(); err != nil {
			// ignore missing prod config, but fail on other errors
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to read prod config: %w", err)
			}
		}
	}

	var cfg Config
	// Maps viper’s internal registry into your Config struct.
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	// Convert expireHours from int to time.Duration into a Go time.Duration for easy use downstream
	cfg.JWT.ExpireHours = time.Duration(viper.GetInt("jwt.expireHours")) * time.Hour
	return &cfg, nil
}
