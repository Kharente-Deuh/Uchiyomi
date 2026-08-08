// SPDX-License-Identifier: AGPL-3.0-or-later

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const KeyLen = 32

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

type Config struct {
	Key []byte
}

func (cfg *Config) Validate() error {
	if len(cfg.Key) != KeyLen {
		return fmt.Errorf("key must be %d bytes, got %d", KeyLen, len(cfg.Key))
	}

	return nil
}

type Cipher struct {
	aead cipher.AEAD
}

func New(cfg Config) (*Cipher, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	block, err := aes.NewCipher(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}

	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Seal(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("io.ReadFull: %w", err)
	}

	return c.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (c *Cipher) Open(ciphertext []byte) ([]byte, error) {
	nonceSize := c.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	plaintext, err := c.aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}
