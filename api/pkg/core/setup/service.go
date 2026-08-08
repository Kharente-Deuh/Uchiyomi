// SPDX-License-Identifier: AGPL-3.0-or-later

package setup

import (
	"context"
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

var _ SetupService = (*Service)(nil)

type Deps struct {
	AuthService     auth.AuthService
	UsersRepository users.UsersRepository
	Transactor      transaction.Transactor
	SessionService  sessions.SessionService
}

func (deps *Deps) Validate() error {
	if deps.AuthService == nil {
		return errors.New("authService is required")
	}

	if deps.UsersRepository == nil {
		return errors.New("usersRepository is required")
	}

	if deps.Transactor == nil {
		return errors.New("transactor is required")
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

func (s *Service) IsSetupRequired(ctx context.Context) (bool, error) {
	count, err := s.deps.UsersRepository.CountAdmins(ctx)
	if err != nil {
		return false, fmt.Errorf("s.deps.UsersRepository.CountAdmins: %w", err)
	}

	return count == 0, nil
}

func (s *Service) DoSetup(ctx context.Context, opts DoSetupOpts) (*sessions.IssuedSession, error) {
	var user *users.User

	err := s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{Isolation: transaction.IsolationSerializable},
		func(ctx context.Context) error {
			required, err := s.IsSetupRequired(ctx)
			if err != nil {
				return fmt.Errorf("s.IsSetupRequired: %w", err)
			}

			if !required {
				return ErrSetupNotNeeded
			}

			user, err = s.deps.AuthService.CreateUserWithPwd(ctx, auth.CreateUserWithPwdOpts{
				Name:     opts.Username,
				Password: opts.Password,
				IsAdmin:  true,
			})
			if err != nil {
				return fmt.Errorf("s.deps.AuthService.CreateUserWithPwd: %w", err)
			}

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
	}

	session, err := s.deps.SessionService.Create(ctx, sessions.CreateSessionOpts{
		UserID:     user.ID,
		AuthMethod: sessions.AuthMethodPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSessionNotIssued, err)
	}

	return session, nil
}
