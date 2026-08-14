// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newTestGormLogger(buf *bytes.Buffer, level slog.Level) gormlogger.Interface {
	log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level}))

	return newGormLogger(log)
}

func TestGormLoggerOmitsHandledErrors(t *testing.T) {
	var buf bytes.Buffer
	l := newTestGormLogger(&buf, slog.LevelInfo)
	fc := func() (string, int64) { return "SELECT 1", int64(0) }

	l.Trace(context.Background(), time.Now(), fc, gorm.ErrRecordNotFound)
	l.Trace(context.Background(), time.Now(), fc, gorm.ErrDuplicatedKey)

	if buf.Len() != 0 {
		t.Errorf("handled errors were logged: %s", buf.String())
	}
}

func TestGormLoggerLogsUnexpectedErrors(t *testing.T) {
	var buf bytes.Buffer
	l := newTestGormLogger(&buf, slog.LevelInfo)
	fc := func() (string, int64) { return "SELECT 1", int64(0) }

	l.Trace(context.Background(), time.Now(), fc, errors.New("connection reset"))

	if !strings.Contains(buf.String(), "connection reset") {
		t.Errorf("unexpected error was not logged: %s", buf.String())
	}
}
