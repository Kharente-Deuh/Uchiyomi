// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type gormLogger struct {
	gormlogger.Interface
}

func newGormLogger(log *slog.Logger) gormlogger.Interface {
	level := gormlogger.Warn
	if log.Enabled(context.Background(), slog.LevelDebug) {
		level = gormlogger.Info
	}

	return gormLogger{
		Interface: gormlogger.NewSlogLogger(log, gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
		}),
	}
}

func (l gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return gormLogger{Interface: l.Interface.LogMode(level)}
}

func (l gormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		err = nil
	}

	l.Interface.Trace(ctx, begin, fc, err)
}
