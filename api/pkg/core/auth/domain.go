// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type AuthService interface {
	LoginWithPwd(context.Context, LoginWithPwdOpts) (*LoginResult, error)
	CreateUserWithPwd(context.Context, CreateUserWithPwdOpts) (*users.User, error)
	Logout(context.Context, LogoutOpts) (*LogoutResult, error)
	StartOIDCLogin(context.Context, StartOIDCLoginOpts) (*OIDCStart, error)
	FinishOIDCLogin(context.Context, FinishOIDCLoginOpts) (*OIDCLoginResult, error)
	BackchannelLogout(context.Context, string) error
}

type LoginResult struct {
	Session *sessions.IssuedSession
	User    *users.User
}

type CreateUserWithPwdOpts struct {
	Name     string
	Password string
	IsAdmin  bool
}

type LoginWithPwdOpts struct {
	Username string
	Password string
}

type LogoutOpts struct {
	Token   string
	Session sessions.Session
}

type LogoutResult struct {
	EndSessionURL string
}

type StartOIDCLoginOpts struct {
	Redirect     string
	ProviderSlug string
}

type OIDCStart struct {
	ExpiresAt        time.Time
	AuthCodeURL      string
	StateCookieValue string
}

type FinishOIDCLoginOpts struct {
	Code             string
	State            string
	ErrorParam       string
	StateCookieValue string
}

type OIDCLoginResult struct {
	Session  *sessions.IssuedSession
	Redirect string
}

var (
	ErrInvalidLoginPwd    = errors.New("invalid login/password")
	ErrOIDCState          = errors.New("oidc state is missing, expired or invalid")
	ErrOIDCDenied         = errors.New("oidc provider denied the request")
	ErrOIDCNotAllowed     = errors.New("oidc claims do not satisfy the allowed values")
	ErrOIDCNoAccount      = errors.New("oidc login did not resolve to an account")
	ErrOIDCUnavailable    = errors.New("oidc provider is unavailable")
	ErrLogoutTokenInvalid = errors.New("logout token is invalid")
)
