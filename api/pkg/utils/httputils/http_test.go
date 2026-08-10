// SPDX-License-Identifier: AGPL-3.0-or-later

package httputils_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

const contentTypeJSON = "application/json"

type errorBody struct {
	Message string `json:"message"`
}

func testLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

type failingWriter struct {
	http.ResponseWriter
	err error
}

func (f *failingWriter) Write([]byte) (int, error) { return 0, f.err }

func TestWriteJSONSuccess(t *testing.T) {
	t.Parallel()

	logger, logs := testLogger()
	rec := httptest.NewRecorder()

	httputils.WriteJSON(rec, logger, http.StatusCreated, map[string]any{"id": 42, "name": "one piece"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contentTypeJSON)
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got["name"] != "one piece" {
		t.Errorf("body[name] = %v, want %q", got["name"], "one piece")
	}

	if logs.Len() != 0 {
		t.Errorf("no logs expected on happy path, got: %s", logs.String())
	}
}

func TestWriteJSONPreservesTypeThroughGenerics(t *testing.T) {
	t.Parallel()

	type payload struct {
		Required bool `json:"required"`
	}

	logger, _ := testLogger()
	rec := httptest.NewRecorder()

	httputils.WriteJSON(rec, logger, http.StatusOK, payload{Required: true})

	if body := strings.TrimSpace(rec.Body.String()); body != `{"required":true}` {
		t.Errorf("body = %s, want %s", body, `{"required":true}`)
	}
}

func TestWriteJSONEncodeFailureFallsBackTo500(t *testing.T) {
	t.Parallel()

	logger, logs := testLogger()
	rec := httptest.NewRecorder()

	httputils.WriteJSON(rec, logger, http.StatusOK, make(chan int))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contentTypeJSON)
	}

	var got errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("fallback body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Message != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("message = %q, want %q", got.Message, http.StatusText(http.StatusInternalServerError))
	}

	if !strings.Contains(logs.String(), "failed to encode response body") {
		t.Errorf("encoding failure not logged, logs: %s", logs.String())
	}
}

func TestWriteJSONNoPartialBodyOnEncodeFailure(t *testing.T) {
	t.Parallel()

	type partial struct {
		Bad  chan int `json:"bad"`
		Good string   `json:"good"`
	}

	logger, _ := testLogger()
	rec := httptest.NewRecorder()

	httputils.WriteJSON(rec, logger, http.StatusOK, partial{Good: "visible", Bad: make(chan int)})

	if strings.Contains(rec.Body.String(), "visible") {
		t.Errorf("fragment of failed response written to client: %q", rec.Body.String())
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestWriteJSONNilLoggerDoesNotPanic(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	httputils.WriteJSON(rec, nil, http.StatusOK, map[string]string{"ok": "true"})
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec2 := httptest.NewRecorder()
	httputils.WriteJSON(rec2, nil, http.StatusOK, make(chan int))
	if rec2.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec2.Code, http.StatusInternalServerError)
	}
}

func TestWriteJSONLogsWriteFailure(t *testing.T) {
	t.Parallel()

	logger, logs := testLogger()
	w := &failingWriter{ResponseWriter: httptest.NewRecorder(), err: errors.New("connection reset")}

	httputils.WriteJSON(w, logger, http.StatusOK, map[string]string{"a": "b"})

	if !strings.Contains(logs.String(), "failed to write response body") {
		t.Errorf("write failure not logged, logs: %s", logs.String())
	}

	if !strings.Contains(logs.String(), "connection reset") {
		t.Errorf("failure cause missing from logs: %s", logs.String())
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		msg         string
		wantMessage string
		status      int
	}{
		"empty message falls back to status text": {
			status:      http.StatusInternalServerError,
			msg:         "",
			wantMessage: "Internal Server Error",
		},
		"404 without message": {
			status:      http.StatusNotFound,
			msg:         "",
			wantMessage: "Not Found",
		},
		"explicit message preserved": {
			status:      http.StatusBadRequest,
			msg:         "slug is required",
			wantMessage: "slug is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger, _ := testLogger()
			rec := httptest.NewRecorder()

			httputils.WriteError(rec, logger, tc.status, tc.msg)

			if rec.Code != tc.status {
				t.Errorf("status = %d, want %d", rec.Code, tc.status)
			}

			if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
				t.Errorf("Content-Type = %q, want %q", ct, contentTypeJSON)
			}

			var got errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
			}

			if got.Message != tc.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tc.wantMessage)
			}
		})
	}
}

func TestWriteErrorUnknownStatusHasEmptyStatusText(t *testing.T) {
	t.Parallel()

	logger, _ := testLogger()
	rec := httptest.NewRecorder()

	httputils.WriteError(rec, logger, 799, "")

	var got errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Message != "" {
		t.Errorf("message = %q, want %q for unknown status", got.Message, "")
	}
}
