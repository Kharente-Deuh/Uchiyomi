// SPDX-License-Identifier: AGPL-3.0-or-later

package bcrypthash_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash/bcrypthash"
	"golang.org/x/crypto/bcrypt"
)

func newService(t *testing.T) *bcrypthash.Service {
	t.Helper()

	s, err := bcrypthash.New(bcrypthash.Config{Cost: bcrypt.MinCost})
	if err != nil {
		t.Fatalf("bcrypthash.New: %v", err)
	}

	return s
}

func TestHashAcceptsPasswordWithinBcryptLimit(t *testing.T) {
	t.Parallel()

	s := newService(t)

	h, err := s.Hash([]byte("Dragonicadu33!"))
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if len(h) == 0 {
		t.Fatal("Hash returned an empty hash")
	}
}

func TestHashAcceptsExactly72Bytes(t *testing.T) {
	t.Parallel()

	s := newService(t)

	if _, err := s.Hash(bytes.Repeat([]byte("a"), 72)); err != nil {
		t.Fatalf("Hash: %v", err)
	}
}

func TestHashRejectsMoreThan72Bytes(t *testing.T) {
	t.Parallel()

	s := newService(t)

	_, err := s.Hash(bytes.Repeat([]byte("a"), 73))
	if !errors.Is(err, hash.ErrStringTooLong) {
		t.Fatalf("Hash err = %v, want %v", err, hash.ErrStringTooLong)
	}
}

func TestMatchAcceptsCorrectPassword(t *testing.T) {
	t.Parallel()

	s := newService(t)
	pwd := []byte("Dragonicadu33!")

	h, err := s.Hash(pwd)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := s.Match(h, pwd)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}

	if !ok {
		t.Fatal("Match = false, want true")
	}
}

func TestMatchRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	s := newService(t)

	h, err := s.Hash([]byte("Dragonicadu33!"))
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := s.Match(h, []byte("wrongpassword"))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}

	if ok {
		t.Fatal("Match = true, want false")
	}
}
