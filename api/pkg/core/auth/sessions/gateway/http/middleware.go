// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type sessionService interface {
	Authenticate(ctx context.Context, token string) (*sessions.AuthenticatedSession, error)
}

type AuthenticatorDeps struct {
	SessionService sessionService
	Cookies        *CookieManager
	Logger         *slog.Logger
	Now            func() time.Time
}

func (deps *AuthenticatorDeps) Validate() error {
	if deps.SessionService == nil {
		return errors.New("sessionService is required")
	}

	if deps.Cookies == nil {
		return errors.New("cookies is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type Authenticator struct {
	deps AuthenticatorDeps
}

func NewAuthenticator(deps AuthenticatorDeps) (*Authenticator, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if deps.Now == nil {
		deps.Now = time.Now
	}

	deps.Logger = deps.Logger.With("component", "sessions.gateway.http")

	return &Authenticator{deps: deps}, nil
}

type userCtxKey struct{}

func UserFrom(ctx context.Context) (*users.User, bool) {
	user, ok := ctx.Value(userCtxKey{}).(*users.User)

	return user, ok
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token := a.deps.Cookies.Read(r)
		if token == "" {
			httputils.WriteError(w, a.deps.Logger, http.StatusUnauthorized, "")

			return
		}

		authenticated, err := a.deps.SessionService.Authenticate(ctx, token)
		if err != nil {
			if errors.Is(err, sessions.ErrInvalidSession) {
				a.deps.Cookies.Clear(w)
				httputils.WriteError(w, a.deps.Logger, http.StatusUnauthorized, "")

				return
			}

			a.deps.Logger.ErrorContext(ctx, "failed to authenticate session", logging.Err(err))
			httputils.WriteError(w, a.deps.Logger, http.StatusInternalServerError, "")

			return
		}

		if authenticated.RenewErr != nil {
			a.deps.Logger.WarnContext(ctx, "failed to renew session", logging.Err(authenticated.RenewErr))
		}

		a.deps.Cookies.Set(w, token, authenticated.Session.ExpiresAt, a.deps.Now())

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, userCtxKey{}, authenticated.User)))
	})
}
