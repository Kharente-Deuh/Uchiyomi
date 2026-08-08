// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

var _ AuthService = (*Service)(nil)

type Deps struct {
	HashService     hash.HashService
	UsersRepository users.UsersRepository
	PwdRepository   password.PasswordCredsRepository
	Transactor      transaction.Transactor
	SessionService  sessions.SessionService
}

func (deps *Deps) Validate() error {
	if deps.HashService == nil {
		return fmt.Errorf("hashService is required")
	}

	if deps.UsersRepository == nil {
		return fmt.Errorf("usersRepository is required")
	}

	if deps.PwdRepository == nil {
		return fmt.Errorf("pwdRepository is required")
	}

	if deps.Transactor == nil {
		return fmt.Errorf("transactor is required")
	}

	if deps.SessionService == nil {
		return errors.New("sessionService is required")
	}

	return nil
}

type Service struct {
	deps Deps
}

func New(deps Deps) (*Service, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &Service{deps: deps}, nil
}

func (s *Service) CreateUserWithPwd(ctx context.Context, opts CreateUserWithPwdOpts) (*users.User, error) {
	pwdHash, err := s.deps.HashService.Hash([]byte(opts.Password))
	if err != nil {
		return nil, fmt.Errorf("s.deps.HashService.Hash: %w", err)
	}

	var user *users.User

	err = s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		user, err = s.deps.UsersRepository.Create(ctx, users.CreateUserOpts{IsAdmin: opts.IsAdmin, Name: opts.Name})
		if err != nil {
			return fmt.Errorf("s.deps.UsersRepository.Create: %w", err)
		}

		_, err = s.deps.PwdRepository.Create(ctx, password.UpsertPasswordCredsOpts{
			UserID: user.ID,
			Hash:   string(pwdHash),
		})
		if err != nil {
			return fmt.Errorf("s.deps.PwdRepository.Create: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
	}

	return user, nil
}

func (s *Service) LoginWithPwd(ctx context.Context, opts LoginWithPwdOpts) (*sessions.IssuedSession, error) {
	user, err := s.deps.UsersRepository.GetByUsername(ctx, opts.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.equalizeLoginTiming(opts.Password)

			return nil, ErrInvalidLoginPwd
		}

		return nil, fmt.Errorf("s.deps.UsersRepository.GetByUsername: %w", err)
	}

	pwd, err := s.deps.PwdRepository.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.equalizeLoginTiming(opts.Password)

			return nil, ErrInvalidLoginPwd
		}

		return nil, fmt.Errorf("s.deps.PwdRepository.GetByUserID: %w", err)
	}

	match, err := s.deps.HashService.Match([]byte(pwd.Hash), []byte(opts.Password))
	if err != nil {
		return nil, fmt.Errorf("s.deps.HashService.Match: %w", err)
	}

	if !match {
		return nil, ErrInvalidLoginPwd
	}

	session, err := s.deps.SessionService.Create(ctx, sessions.CreateSessionOpts{
		UserID:     user.ID,
		AuthMethod: sessions.AuthMethodPassword,
	})

	if err != nil {
		return nil, fmt.Errorf("s.deps.SessionService.Create: %w", err)
	}

	return session, nil
}

func (s *Service) equalizeLoginTiming(password string) {
	_, _ = s.deps.HashService.Hash([]byte(password))
}
