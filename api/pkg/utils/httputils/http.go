// SPDX-License-Identifier: AGPL-3.0-or-later

package httputils

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type errorResponse struct {
	Message string `json:"message"`
}

const fallbackErrorBody = `{"message":"Internal Server Error"}`

func WriteJSON[T any](w http.ResponseWriter, logger *slog.Logger, status int, data T) {
	if logger == nil {
		logger = slog.Default()
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		logger.Error("failed to encode response body",
			logging.Err(err),
			slog.Int("status", status),
		)
		write(w, logger, http.StatusInternalServerError, []byte(fallbackErrorBody))

		return
	}

	write(w, logger, status, buf.Bytes())
}

func WriteError(w http.ResponseWriter, logger *slog.Logger, status int, msg string) {
	if msg == "" {
		msg = http.StatusText(status)
	}

	WriteJSON(w, logger, status, errorResponse{Message: msg})
}

func write(w http.ResponseWriter, logger *slog.Logger, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		logger.Error("failed to write response body",
			logging.Err(err),
			slog.Int("status", status),
		)
	}
}
