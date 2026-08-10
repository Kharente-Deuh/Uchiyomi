// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodOIDC     AuthMethod = "oidc"
)

var ErrInvalidSession = errors.New("invalid session")

type Session struct {
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ProviderSID *string
	ProviderID  *uuid.UUID
	AuthMethod  AuthMethod
	ID          uuid.UUID
	UserID      uuid.UUID
}

type IssuedSession struct {
	Token string
	Session
}

type AuthenticatedSession struct {
	User     *users.User
	RenewErr error
	Session  Session
}

type SessionService interface {
	Create(context.Context, CreateSessionOpts) (*IssuedSession, error)
	Authenticate(context.Context, string) (*AuthenticatedSession, error)
	Revoke(context.Context, string) error
	RevokeAllForUser(context.Context, uuid.UUID) error
	RevokeForProvider(context.Context, uuid.UUID, uuid.UUID) error
	RevokeByProviderAndSID(context.Context, uuid.UUID, string) error
}

type CreateSessionOpts struct {
	ProviderSID *string
	ProviderID  *uuid.UUID
	AuthMethod  AuthMethod
	UserID      uuid.UUID
}

type SessionsRepository interface {
	Insert(context.Context, InsertSessionOpts) (*Session, error)
	GetByTokenHash(context.Context, []byte) (*Session, *users.User, error)
	UpdateExpiry(context.Context, uuid.UUID, time.Time) error
	DeleteByTokenHash(context.Context, []byte) error
	DeleteByUserID(context.Context, uuid.UUID) error
	DeleteByUserAndProvider(context.Context, uuid.UUID, uuid.UUID) error
	DeleteByProviderAndSID(context.Context, uuid.UUID, string) error
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type InsertSessionOpts struct {
	ExpiresAt   time.Time
	ProviderSID *string
	ProviderID  *uuid.UUID
	AuthMethod  AuthMethod
	TokenHash   []byte
	UserID      uuid.UUID
}
