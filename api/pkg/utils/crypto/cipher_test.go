// SPDX-License-Identifier: AGPL-3.0-or-later

package crypto_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/crypto"
)

func testKey() []byte {
	key := make([]byte, crypto.KeyLen)
	for i := range key {
		key[i] = byte(i)
	}

	return key
}

func newCipher(t *testing.T) *crypto.Cipher {
	t.Helper()

	c, err := crypto.New(crypto.Config{Key: testKey()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

func TestNewRejectsKeysOfTheWrongLength(t *testing.T) {
	t.Parallel()

	tests := map[string]int{"empty": 0, "too short": 16, "too long": 64}

	for name, size := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, err := crypto.New(crypto.Config{Key: make([]byte, size)})
			if err == nil {
				t.Fatalf("New(%d bytes) = nil, want an error", size)
			}

			if c != nil {
				t.Error("New returned a cipher alongside the error")
			}
		})
	}
}

func TestSealOpenRoundTrip(t *testing.T) {
	t.Parallel()

	c := newCipher(t)
	plaintext := []byte("s3cr3t")

	sealed, err := c.Seal(plaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if bytes.Contains(sealed, plaintext) {
		t.Error("the plaintext appears verbatim in the ciphertext")
	}

	got, err := c.Open(sealed)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Errorf("Open() = %q, want %q", got, plaintext)
	}
}

func TestSealIsNotDeterministic(t *testing.T) {
	t.Parallel()

	c := newCipher(t)

	first, err := c.Seal([]byte("same"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	second, err := c.Seal([]byte("same"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if bytes.Equal(first, second) {
		t.Error("two seals of the same plaintext produced the same ciphertext")
	}
}

func TestOpenRejectsTamperedCiphertext(t *testing.T) {
	t.Parallel()

	c := newCipher(t)

	sealed, err := c.Seal([]byte("s3cr3t"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	tampered := bytes.Clone(sealed)
	tampered[len(tampered)-1] ^= 0xff

	if _, err := c.Open(tampered); !errors.Is(err, crypto.ErrInvalidCiphertext) {
		t.Errorf("Open(tampered) = %v, want ErrInvalidCiphertext", err)
	}
}

func TestOpenRejectsTruncatedCiphertext(t *testing.T) {
	t.Parallel()

	c := newCipher(t)

	if _, err := c.Open([]byte("short")); !errors.Is(err, crypto.ErrInvalidCiphertext) {
		t.Errorf("Open(short) = %v, want ErrInvalidCiphertext", err)
	}
}

func TestOpenRejectsAnotherKeysCiphertext(t *testing.T) {
	t.Parallel()

	sealed, err := newCipher(t).Seal([]byte("s3cr3t"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	other, err := crypto.New(crypto.Config{Key: make([]byte, crypto.KeyLen)})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := other.Open(sealed); !errors.Is(err, crypto.ErrInvalidCiphertext) {
		t.Errorf("Open(other key) = %v, want ErrInvalidCiphertext", err)
	}
}
