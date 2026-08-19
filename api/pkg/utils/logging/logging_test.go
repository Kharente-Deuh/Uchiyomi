// SPDX-License-Identifier: AGPL-3.0-or-later

package logging_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

func TestIsLogLevel(t *testing.T) {
	t.Parallel()

	for _, level := range []string{"info", "error", "warn", "debug"} {
		if !logging.IsLogLevel(level) {
			t.Errorf("IsLogLevel(%q) = false, want true", level)
		}
	}

	if logging.IsLogLevel("INFO") || logging.IsLogLevel("trace") || logging.IsLogLevel("") {
		t.Fatal("unknown log levels must be rejected")
	}
}

func TestLogLevelUnmarshalText(t *testing.T) {
	t.Parallel()

	var level logging.LogLevel
	if err := level.UnmarshalText([]byte("debug")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}

	if level != logging.LogLevelDebug {
		t.Errorf("level = %q, want %q", level, logging.LogLevelDebug)
	}

	if err := level.UnmarshalText([]byte("nope")); err == nil {
		t.Fatal("UnmarshalText(nope) = nil, want error")
	}
}

func TestNewWritesAtConfiguredLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := logging.New(logging.Config{Writer: &buf, Level: logging.LogLevelError})

	logger.Info("quiet")
	logger.Error("loud")

	out := buf.String()
	if strings.Contains(out, "quiet") {
		t.Errorf("info log leaked at error level: %q", out)
	}

	if !strings.Contains(out, "loud") {
		t.Errorf("error log missing: %q", out)
	}
}

func TestNewDefaultsWriterToStdout(t *testing.T) {
	t.Parallel()

	logger := logging.New(logging.Config{Level: logging.LogLevelInfo})
	if logger == nil {
		t.Fatal("New returned nil")
	}
}

func TestErr(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := logging.New(logging.Config{Writer: &buf, Level: logging.LogLevelError})
	logger.Error("failed", logging.Err(errors.New("boom")))

	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("logged output = %q, want to contain boom", buf.String())
	}
}
