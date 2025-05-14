package database

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"strings"
)

type MigrateLogger struct {
	logger *zap.Logger
}

func NewMigrateLogger(logger *zap.Logger) *MigrateLogger {
	return &MigrateLogger{logger: logger}
}

func (z *MigrateLogger) Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	message = strings.TrimSpace(message)
	z.logger.Info("Migration event", zap.String("migration", message))
}

func (z *MigrateLogger) Verbose() bool {
	return z.logger.Level() == zap.DebugLevel
}

func MigrationUp(sourceURL, databaseURL string, logger *zap.Logger) error {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		if !errors.Is(err, migrate.ErrNilVersion) {
			return err
		}
	}

	if version == 0 {
		logger.Info("No existing database version detected")
	} else {
		logger.Info("Current migration version", zap.Uint("version", version), zap.Bool("dirty", dirty))
	}

	m.Log = &MigrateLogger{logger: logger}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("Database schema is up to date, no migration required")
			return nil
		}
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}

func MigrationDown(sourceURL, databaseURL string, logger *zap.Logger) error {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return err
	}

	m.Log = &MigrateLogger{logger: logger}

	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	logger.Info("Database migration down completed successfully")
	return nil
}
