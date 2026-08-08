// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type cfg struct {
	Logger struct {
		Level logging.LogLevel `env:"LOG_LEVEL" envDefault:"info"`
	}
	Http struct {
		AllowedOrigins []string `env:"CORS_ORIGINS" envSeparator:","`
		Port           int      `env:"PORT" envDefault:"3000"`
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
}

func newConfig() (*cfg, error) {
	c, err := env.ParseAs[cfg]()
	if err != nil {
		return nil, fmt.Errorf("env.ParseAs: %w", err)
	}

	return &c, nil
}
