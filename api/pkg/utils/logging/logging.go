// SPDX-License-Identifier: AGPL-3.0-or-later

package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
	LogLevelWarn  LogLevel = "warn"
	LogLevelDebug LogLevel = "debug"
)

func (l *LogLevel) UnmarshalText(b []byte) error {
	if !IsLogLevel(string(b)) {
		return fmt.Errorf("invalid log level %q", b)
	}

	*l = LogLevel(b)

	return nil
}

func IsLogLevel(s string) bool {
	switch LogLevel(s) {
	case LogLevelInfo, LogLevelError, LogLevelWarn, LogLevelDebug:
		return true
	default:
		return false
	}
}

type Config struct {
	Writer io.Writer
	Level  LogLevel
}

func New(cfg Config) *slog.Logger {
	level := slog.LevelInfo
	_ = level.UnmarshalText([]byte(cfg.Level))

	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	return slog.New(tint.NewTextHandler(w, &tint.Options{
		Level:      level,
		AddSource:  cfg.Level == LogLevelDebug,
		TimeFormat: time.DateTime,
	}))
}

func Err(err error) slog.Attr {
	return tint.Err(err)
}
