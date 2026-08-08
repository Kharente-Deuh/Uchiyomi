// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"errors"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type AuthService interface {
	LoginWithPwd(context.Context, LoginWithPwdOpts) (*sessions.IssuedSession, error)
	CreateUserWithPwd(context.Context, CreateUserWithPwdOpts) (*users.User, error)
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

var (
	ErrInvalidLoginPwd = errors.New("invalid login/password")
)
