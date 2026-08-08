// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
)

var _ SessionService = (*Service)(nil)

type TTL struct {
	Idle     time.Duration
	Absolute time.Duration
}

func (t TTL) Validate() error {
	if t.Idle <= 0 {
		return errors.New("idle must be positive")
	}

	if t.Absolute < t.Idle {
		return errors.New("absolute must not be lower than idle")
	}

	return nil
}

type ServiceConfig struct {
	Password       TTL
	OIDC           TTL
	RenewThreshold time.Duration
}

func (cfg *ServiceConfig) Validate() error {
	if err := cfg.Password.Validate(); err != nil {
		return fmt.Errorf("password ttl: %w", err)
	}

	if err := cfg.OIDC.Validate(); err != nil {
		return fmt.Errorf("oidc ttl: %w", err)
	}

	if cfg.RenewThreshold <= 0 {
		return errors.New("renewThreshold must be positive")
	}

	shortest := min(cfg.OIDC.Idle, cfg.Password.Idle)
	if cfg.RenewThreshold >= shortest {
		return errors.New("renewThreshold must be lower than the shortest idle ttl")
	}

	return nil
}

type ServiceDeps struct {
	Repository SessionsRepository
	Now        func() time.Time
}

func (deps *ServiceDeps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
	}

	return nil
}

type Service struct {
	deps ServiceDeps
	cfg  ServiceConfig
}

func NewService(cfg ServiceConfig, deps ServiceDeps) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if deps.Now == nil {
		deps.Now = time.Now
	}

	return &Service{cfg: cfg, deps: deps}, nil
}

const tokenBytes = 32

func (s *Service) ttl(method AuthMethod) (TTL, error) {
	switch method {
	case AuthMethodPassword:
		return s.cfg.Password, nil
	case AuthMethodOIDC:
		return s.cfg.OIDC, nil
	default:
		return TTL{}, fmt.Errorf("unknown auth method %q", method)
	}
}

func (s *Service) Create(ctx context.Context, opts CreateSessionOpts) (*IssuedSession, error) {
	ttl, err := s.ttl(opts.AuthMethod)
	if err != nil {
		return nil, fmt.Errorf("s.ttl: %w", err)
	}

	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("rand.Read: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(token))

	session, err := s.deps.Repository.Insert(ctx, InsertSessionOpts{
		UserID:     opts.UserID,
		AuthMethod: opts.AuthMethod,
		TokenHash:  sum[:],
		ExpiresAt:  s.deps.Now().Add(ttl.Idle),
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.Insert: %w", err)
	}

	return &IssuedSession{Session: *session, Token: token}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*AuthenticatedSession, error) {
	if token == "" {
		return nil, ErrInvalidSession
	}

	sum := sha256.Sum256([]byte(token))

	session, user, err := s.deps.Repository.GetByTokenHash(ctx, sum[:])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidSession
		}

		return nil, fmt.Errorf("s.deps.Repository.GetByTokenHash: %w", err)
	}

	now := s.deps.Now()
	if !session.ExpiresAt.After(now) {
		return nil, ErrInvalidSession
	}

	ttl, err := s.ttl(session.AuthMethod)
	if err != nil {
		return nil, ErrInvalidSession
	}

	expiresAt, renewErr := s.renew(ctx, *session, ttl, now)
	session.ExpiresAt = expiresAt

	return &AuthenticatedSession{User: user, Session: *session, RenewErr: renewErr}, nil
}

func (s *Service) renew(ctx context.Context, session Session, ttl TTL, now time.Time) (time.Time, error) {
	if session.ExpiresAt.Sub(now) >= ttl.Idle-s.cfg.RenewThreshold {
		return session.ExpiresAt, nil
	}

	next := now.Add(ttl.Idle)
	if limit := session.CreatedAt.Add(ttl.Absolute); next.After(limit) {
		next = limit
	}

	if !next.After(session.ExpiresAt) {
		return session.ExpiresAt, nil
	}

	if err := s.deps.Repository.UpdateExpiry(ctx, session.ID, next); err != nil {
		return session.ExpiresAt, fmt.Errorf("s.deps.Repository.UpdateExpiry: %w", err)
	}

	return next, nil
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidSession
	}

	sum := sha256.Sum256([]byte(token))

	if err := s.deps.Repository.DeleteByTokenHash(ctx, sum[:]); err != nil {
		return fmt.Errorf("s.deps.Repository.DeleteByTokenHash: %w", err)
	}

	return nil
}

func (s *Service) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.deps.Repository.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("s.deps.Repository.DeleteByUserID: %w", err)
	}

	return nil
}
