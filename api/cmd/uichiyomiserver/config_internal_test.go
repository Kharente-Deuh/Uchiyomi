// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/crypto"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "uchiyomi")
	t.Setenv("DB_PWD", "uchiyomi")
	t.Setenv("DB_NAME", "uchiyomi")
	t.Setenv("PUBLIC_URL", "https://manga.example.com")
	t.Setenv("OIDC_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, crypto.KeyLen)))
}

func TestNewConfigDecodesTheEncryptionKey(t *testing.T) {
	setRequiredEnv(t)

	c, err := newConfig()
	if err != nil {
		t.Fatalf("newConfig: %v", err)
	}

	if len(c.OIDC.EncryptionKey) != crypto.KeyLen {
		t.Errorf("len(EncryptionKey) = %d, want %d", len(c.OIDC.EncryptionKey), crypto.KeyLen)
	}

	if c.OIDC.PublicURL != "https://manga.example.com" {
		t.Errorf("PublicURL = %q", c.OIDC.PublicURL)
	}
}

func TestNewConfigRejectsAnUnusableEncryptionKey(t *testing.T) {
	tests := map[string]struct {
		value   string
		wantErr string
	}{
		"not base64": {
			value:   "not-base64!!",
			wantErr: "OIDC_ENCRYPTION_KEY",
		},
		"wrong length": {
			value:   base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr: "must be 32 bytes",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("OIDC_ENCRYPTION_KEY", tc.value)

			_, err := newConfig()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("newConfig() = %v, want an error mentioning %q", err, tc.wantErr)
			}
		})
	}
}

func TestNewConfigRequiresTheNewVariables(t *testing.T) {
	for _, name := range []string{"OIDC_ENCRYPTION_KEY", "PUBLIC_URL"} {
		t.Run(name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(name, "")

			if _, err := newConfig(); err == nil {
				t.Errorf("newConfig() without %s = nil, want an error", name)
			}
		})
	}
}
