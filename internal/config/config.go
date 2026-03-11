package config

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type Config struct {
	Primary  Primary        `mapstructure:"primary"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type Primary struct {
	Env string `mapstructure:"env" validate:"required,oneof=development staging production"`
}

type ServerConfig struct {
	Port               int      `mapstructure:"port" validate:"required"`
	ReadTimeout        int      `mapstructure:"read_timeout" validate:"required,min=1"`
	WriteTimeout       int      `mapstructure:"write_timeout" validate:"required,min=1"`
	IdleTimeout        int      `mapstructure:"idle_timeout" validate:"required,min=1"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host" validate:"required"`
	Port            int    `mapstructure:"port" validate:"required"`
	User            string `mapstructure:"user" validate:"required"`
	Password        string `mapstructure:"password" validate:"required"`
	DBName          string `mapstructure:"dbname" validate:"required"`
	SSLMode         string `mapstructure:"sslmode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" validate:"required,min=1"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time" validate:"required,min=1"`
}

type AuthConfig struct {
	SecretKey string `mapstructure:"secret_key" validate:"required"`
}

type RedisConfig struct {
	Address string `mapstructure:"address" validate:"required"`
}

var (
	instance *Config
	once     sync.Once
	loadErr  error
)

// Load reads config exactly once. Subsequent calls return the cached instance.
// Priority (highest → lowest):
//  1. Environment variables
//  2. Config file
//  3. Built-in defaults
func Load(cfgPath string) (*Config, error) {
	once.Do(func() {
		instance, loadErr = load(cfgPath)
	})
	return instance, loadErr
}

// MustLoad panics if config loading fails.
func MustLoad(cfgPath string) *Config {
	cfg, err := Load(cfgPath)
	if err != nil {
		panic(fmt.Sprintf("config: failed to load: %v", err))
	}
	return cfg
}

// load is an internal (non-singleton) loader so it can be tested in isolation.
func load(cfgPath string) (*Config, error) {
	v := viper.New()

	// 1. Set default values
	setDefaults(v)

	// 2. Read config file if provided
	if cfgPath != "" {
		v.SetConfigFile(cfgPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/app")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config: reading config file: %w", err)
		}
	}

	// 3. Read environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// 5. Validate config
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config: validate: %w", err)
	}

	return &cfg, nil
}

// setDefaults registers safe non-zero defaults so the app can start with a
// minimal config file (e.g. in local development).
func setDefaults(v *viper.Viper) {
	// primary
	v.SetDefault("primary.env", "development")

	// server
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 10)
	v.SetDefault("server.write_timeout", 10)
	v.SetDefault("server.idle_timeout", 10)
	v.SetDefault("server.cors_allowed_origins", []string{"*"})

	// database
	v.SetDefault("database.host", "")
	v.SetDefault("database.user", "")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 300)
	v.SetDefault("database.conn_max_idle_time", 60)

	// auth
	v.SetDefault("auth.secret_key", "secret")

	// redis
	v.SetDefault("redis.address", "localhost:6379")
}

// validate runs struct-level validation and returns a human-friendly error
// that lists every violated constraint at once (not just the first one).
func validate(cfg *Config) error {
	v := validator.New()

	if err := v.Struct(cfg); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) {
			msgs := make([]string, 0, len(errs))
			for _, e := range errs {
				msgs = append(msgs, fmt.Sprintf(
					"  • %s: failed '%s' constraint (got: %q)",
					e.Namespace(), e.Tag(), e.Value(),
				))
			}
			return fmt.Errorf("config: validation failed:\n%s", strings.Join(msgs, "\n"))
		}
		return fmt.Errorf("config: validation: %w", err)
	}
	return nil
}
