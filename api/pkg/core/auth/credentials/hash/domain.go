// SPDX-License-Identifier: AGPL-3.0-or-later

package hash

import "errors"

type HashService interface {
	Hash(toHash []byte) ([]byte, error)
	Match(hash []byte, toCompare []byte) (bool, error)
}

var (
	ErrHashTooShort  = errors.New("hash must be at least characters")
	ErrStringTooLong = errors.New("string must not exceed 72 bytes")
)
