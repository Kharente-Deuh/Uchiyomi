// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGConfig struct {
	Host        string
	Username    string
	Password    string
	Database    string
	Schema      string
	SSLRequired bool
	Port        int
}

func (cfg *PGConfig) Validate() error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}

	if cfg.Username == "" {
		return errors.New("username is required")
	}

	if cfg.Password == "" {
		return errors.New("password is required")
	}

	if cfg.Port <= 0 {
		return errors.New("invalid port")
	}

	if cfg.Database == "" {
		return errors.New("database is required")
	}

	return nil
}

type PGDeps struct {
	Logger *slog.Logger
}

func (deps *PGDeps) Validate() error {
	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type PGDB struct {
	deps PGDeps
	DB   *gorm.DB
}

func NewPGDatabase(cfg PGConfig, deps PGDeps) (*PGDB, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "db.postgres")
	db, err := createPGDatabase(cfg, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("createPGDatabase: %w", err)
	}

	deps.Logger.Info("connected to database")

	return &PGDB{DB: db, deps: deps}, nil
}

func createPGDatabase(cfg PGConfig, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Username,
		cfg.Password,
		utils.Ternary(cfg.SSLRequired, "require", "disable"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         newGormLogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	return db, nil
}

func (pgdb *PGDB) Migrate() error {
	models := []any{
		&pgmodels.User{},
		&pgmodels.Comic{},
		&pgmodels.Chapter{},
		&pgmodels.FederatedIdentity{},
		&pgmodels.OIDCProvider{},
		&pgmodels.Session{},
		&pgmodels.PasswordCreds{},
		&pgmodels.LibraryEntry{},
	}

	for _, m := range models {
		if err := pgdb.DB.AutoMigrate(m); err != nil {
			return fmt.Errorf("pgdb.db.AutoMigrate: %w", err)
		}
	}

	pgdb.deps.Logger.Info("migrations applied", "models", len(models))

	return nil
}

func (pgdb *PGDB) Ping(ctx context.Context) error {
	db, err := pgdb.DB.DB()
	if err != nil {
		return fmt.Errorf("pgdb.DB.DB: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("db.PingContext: %w", err)
	}

	return nil
}

func (pgdb *PGDB) Close() {
	db, err := pgdb.DB.DB()
	if err != nil {
		pgdb.deps.Logger.Error("failed to get sql db", logging.Err(err))

		return
	}

	if err := db.Close(); err != nil {
		pgdb.deps.Logger.Error("failed to close db", logging.Err(err))

		return
	}

	pgdb.deps.Logger.Info("database connection closed")
}
