// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			ctx := r.Context()

			clientIP := middleware.GetClientIP(ctx)
			if clientIP == "" {
				clientIP = r.RemoteAddr
			}

			a.deps.Logger.InfoContext(ctx, fmt.Sprintf("[%s] %s", r.Method, r.RequestURI),
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"clientIP", clientIP,
				"requestID", middleware.GetReqID(ctx),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
