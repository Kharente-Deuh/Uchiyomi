// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"net/http"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

const (
	componentMigrations       = "migrations"
	componentAsura            = "asura"
	componentSessions         = "sessions"
	componentOIDCRevalidation = "oidc-revalidation"
	componentDB               = "db"
)

const notReadyMessage = "service is starting"

func NewHealthRegistry(db Database) *health.Registry {
	reg := health.NewRegistry(componentMigrations, componentAsura, componentSessions, componentOIDCRevalidation)
	reg.AddProbe(componentDB, db.Ping)

	return reg
}

func (a *App) requireMigrations(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.deps.Health.LatchStatus(componentMigrations) != health.StatusOK {
			httputils.WriteError(w, a.deps.Logger, http.StatusServiceUnavailable, notReadyMessage)

			return
		}

		next.ServeHTTP(w, r)
	})
}
