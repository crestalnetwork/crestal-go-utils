package xdb

import (
	"fmt"
	"log/slog"

	"github.com/avast/retry-go/v4"
	"github.com/samber/oops"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config is the configuration for the postgres database
// If SecretManagerPath is set, it will try to load the config from aws ssm parameter store, ignore other fields.
// But the Name field will override the DB Name field in SecretManager value, if both are set.
type Config struct {
	SecretManagerPath string
	Host              string
	Port              string
	Username          string
	Password          string
	Name              string
	TranslateError    bool `default:"true"`
}

// New creates a new postgres database connection
func New(config Config) (*gorm.DB, error) {
	if config.SecretManagerPath != "" {
		err := loadFromAwsSecretsManager(&config)
		if err != nil {
			return nil, err
		}
	}
	if config.Host == "" {
		return nil, oops.Errorf("host is required")
	}
	if config.Port == "" {
		return nil, oops.Errorf("port is required")
	}
	if config.Username == "" {
		return nil, oops.Errorf("username is required")
	}
	if config.Password == "" {
		return nil, oops.Errorf("password is required")
	}
	if config.Name == "" {
		return nil, oops.Errorf("db name is required")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", config.Host, config.Port, config.Username, config.Password, config.Name)
	var db *gorm.DB
	var err error
	err = retry.Do(func() error {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			TranslateError: config.TranslateError,
		})
		if err != nil {
			return oops.With("host", config.Host, "port", config.Port, "user", config.Username, "db-name", config.Name).
				Wrapf(err, "connect to db failed")
		}
		return nil
	}, retry.OnRetry(func(n uint, err error) {
		slog.Info("retrying to connect to db", "n", n, "err", err)
	}))
	return db, nil
}
