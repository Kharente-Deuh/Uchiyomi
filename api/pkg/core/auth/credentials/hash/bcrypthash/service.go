// SPDX-License-Identifier: AGPL-3.0-or-later

package bcrypthash

import (
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash"
	"golang.org/x/crypto/bcrypt"
)

var _ hash.HashService = (*Service)(nil)

const maxBcryptInputLen = 72

type Config struct {
	Cost int
}

func (cfg *Config) Validate() error {
	if cfg.Cost < bcrypt.MinCost {
		return fmt.Errorf("cost must be at least %d", bcrypt.MinCost)
	}

	if cfg.Cost > bcrypt.MaxCost {
		return fmt.Errorf("cost must be lower than %d", bcrypt.MaxCost+1)
	}

	return nil
}

type Service struct {
	cfg Config
}

func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	return &Service{cfg: cfg}, nil
}

func (s *Service) Hash(toHash []byte) ([]byte, error) {
	if !s.isByteLenOkToHash(toHash) {
		return nil, hash.ErrStringTooLong
	}

	h, err := bcrypt.GenerateFromPassword(toHash, s.cfg.Cost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrHashTooShort) {
			err = hash.ErrHashTooShort
		} else if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			err = hash.ErrStringTooLong
		} else {
			err = fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
		}

		return nil, err
	}

	return h, nil
}

func (s *Service) Match(hashed []byte, toCompare []byte) (bool, error) {
	if !s.isByteLenOkToHash(toCompare) {
		return false, nil
	}

	err := bcrypt.CompareHashAndPassword(hashed, toCompare)
	if err != nil {
		if errors.Is(err, bcrypt.ErrHashTooShort) ||
			errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) ||
			errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return false, nil
		}

		return false, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", err)
	}

	return true, nil
}

func (s *Service) isByteLenOkToHash(b []byte) bool {
	return len(b) <= maxBcryptInputLen
}
