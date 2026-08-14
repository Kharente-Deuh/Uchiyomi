// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/crypto"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type cfg struct {
	Logger struct {
		Level logging.LogLevel `env:"LOG_LEVEL" envDefault:"info"`
	}
	Covers struct {
		Dir string `env:"COVERS_DIR" envDefault:"data/cache/covers"`
	}
	Downloads struct {
		Dir          string        `env:"DOWNLOADS_DIR" envDefault:"data/downloads"`
		RateLimit    time.Duration `env:"CHAPTER_DOWNLOAD_RATE_LIMIT" envDefault:"500ms"`
		ScanInterval time.Duration `env:"CHAPTER_DOWNLOAD_SCAN_INTERVAL" envDefault:"15m"`
	}
	PG struct {
		Host        string `env:"DB_HOST,required"`
		Username    string `env:"DB_USER,required"`
		Password    string `env:"DB_PWD,required"`
		Database    string `env:"DB_NAME,required"`
		Schema      string `env:"DB_SCHEMA"`
		SSLRequired bool   `env:"DB_SSL" envDefault:"false"`
		Port        int    `env:"DB_PORT" envDefault:"5432"`
	}
	OIDC struct {
		PublicURL            string `env:"PUBLIC_URL,required,notEmpty"`
		EncryptionKeyB64     string `env:"OIDC_ENCRYPTION_KEY,required,notEmpty"`
		EncryptionKey        []byte
		RevalidationInterval time.Duration `env:"OIDC_REVALIDATION_INTERVAL" envDefault:"15m"`
	}
	Http struct {
		AllowedOrigins []string `env:"CORS_ORIGINS" envSeparator:","`
		Port           int      `env:"PORT" envDefault:"3000"`
	}
	Runtime struct {
		UID int `env:"PUID" envDefault:"65532"`
		GID int `env:"PGID" envDefault:"65532"`
	}
}

func newConfig() (*cfg, error) {
	c, err := env.ParseAs[cfg]()
	if err != nil {
		return nil, fmt.Errorf("env.ParseAs: %w", err)
	}

	if err := c.decodeEncryptionKey(); err != nil {
		return nil, err
	}

	if err := c.validatePublicURL(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *cfg) decodeEncryptionKey() error {
	key, err := base64.StdEncoding.DecodeString(c.OIDC.EncryptionKeyB64)
	if err != nil {
		return fmt.Errorf("OIDC_ENCRYPTION_KEY is not valid base64: %w", err)
	}

	if len(key) != crypto.KeyLen {
		return fmt.Errorf("OIDC_ENCRYPTION_KEY must be %d bytes once decoded, got %d", crypto.KeyLen, len(key))
	}

	c.OIDC.EncryptionKey = key

	return nil
}

func (c *cfg) validatePublicURL() error {
	u, err := url.Parse(c.OIDC.PublicURL)
	if err != nil {
		return fmt.Errorf("PUBLIC_URL is not a valid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("PUBLIC_URL must have an http or https scheme, got %q", c.OIDC.PublicURL)
	}

	if u.Host == "" {
		return fmt.Errorf("PUBLIC_URL must have a host, got %q", c.OIDC.PublicURL)
	}

	return nil
}
